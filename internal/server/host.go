package server

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/user/goslide/internal/config"
	"github.com/user/goslide/internal/ir"
	"github.com/user/goslide/internal/parser"
	"github.com/user/goslide/internal/renderer"
	"github.com/user/goslide/web"
)

type HostOptions struct {
	Dir    string
	Port   int
	NoOpen bool
}

type pageEntry struct {
	Name     string
	Title    string
	Author   string
	Date     string
	Tags     []string
	HTML     string
	FilePath string
}

type hostApp struct {
	dir       string
	cfg       *config.Config
	mux       *http.ServeMux
	broadcast *broadcaster
	mu        sync.RWMutex
	pages     map[string]*pageEntry
	indexHTML string
}

func newHostApp(opts HostOptions) (*hostApp, error) {
	h := &hostApp{
		dir:       opts.Dir,
		mux:       http.NewServeMux(),
		broadcast: newBroadcaster(),
		pages:     make(map[string]*pageEntry),
	}

	if err := h.scanDirectory(); err != nil {
		return nil, err
	}
	h.rebuildIndex()
	h.setupHostRoutes()

	cfg, err := config.Load(opts.Dir)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}
	h.cfg = cfg
	if len(cfg.API.Proxy) > 0 {
		setupProxy(h.mux, cfg.API.Proxy)
	}

	return h, nil
}

func (h *hostApp) scanDirectory() error {
	entries, err := os.ReadDir(h.dir)
	if err != nil {
		return fmt.Errorf("read directory %s: %w", h.dir, err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		path := filepath.Join(h.dir, entry.Name())
		if err := h.loadFile(path); err != nil {
			log.Printf("[host] skip %s: %v", entry.Name(), err)
		}
	}
	return nil
}

func (h *hostApp) loadFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	pres, err := parser.Parse(data, path)
	if err != nil {
		return err
	}

	valErrs := pres.Validate()
	if ir.HasErrors(valErrs) {
		return fmt.Errorf("validation failed")
	}

	html, err := renderer.Render(pres)
	if err != nil {
		return err
	}

	if h.cfg != nil && len(h.cfg.Theme.Overrides) > 0 {
		html = injectThemeOverrides(html, h.cfg.Theme.Overrides)
	}

	name := strings.TrimSuffix(filepath.Base(path), ".md")

	h.mu.Lock()
	h.pages[name] = &pageEntry{
		Name:     name,
		Title:    pres.Meta.Title,
		Author:   pres.Meta.Author,
		Date:     pres.Meta.Date,
		Tags:     pres.Meta.Tags,
		HTML:     html,
		FilePath: path,
	}
	h.mu.Unlock()

	return nil
}

func (h *hostApp) rebuildIndex() {
	tmplFS, err := fs.Sub(web.TemplateFS, "templates")
	if err != nil {
		log.Printf("[host] index template error: %v", err)
		return
	}
	tmpl, err := template.New("index.html").ParseFS(tmplFS, "index.html")
	if err != nil {
		log.Printf("[host] index template parse error: %v", err)
		return
	}

	h.mu.RLock()
	pages := make([]*pageEntry, 0, len(h.pages))
	for _, p := range h.pages {
		pages = append(pages, p)
	}
	h.mu.RUnlock()

	sort.Slice(pages, func(i, j int) bool {
		return pages[i].Name < pages[j].Name
	})

	var buf bytes.Buffer
	tmpl.Execute(&buf, struct{ Pages []*pageEntry }{Pages: pages})

	h.mu.Lock()
	h.indexHTML = buf.String()
	h.mu.Unlock()
}

func (h *hostApp) setupHostRoutes() {
	h.mux.HandleFunc("/", h.handleHostIndex)
	h.mux.HandleFunc("/talks/", h.handleTalk)
	h.mux.HandleFunc("/ws", h.broadcast.handleWS)
	setupStaticRoutes(h.mux)
}

func (h *hostApp) handleHostIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	h.mu.RLock()
	html := h.indexHTML
	h.mu.RUnlock()
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

func (h *hostApp) handleTalk(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/talks/")
	name = strings.TrimSuffix(name, "/")
	if name == "" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	h.mu.RLock()
	page, ok := h.pages[name]
	h.mu.RUnlock()

	if !ok {
		http.NotFound(w, r)
		return
	}

	indexLink := `<a href="/" id="goslide-home" style="position:fixed;top:1.2rem;left:1.5rem;z-index:100;font-family:var(--font-sans);font-size:0.85rem;color:var(--slide-muted);text-decoration:none;opacity:0.7;transition:opacity 0.2s;" onmouseover="this.style.opacity='1'" onmouseout="this.style.opacity='0.7'">← Index</a>`
	html := strings.Replace(page.HTML, "</body>", indexLink+"</body>", 1)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

func (h *hostApp) handleFileEvent(event string, path string) {
	if !strings.HasSuffix(path, ".md") {
		return
	}
	name := strings.TrimSuffix(filepath.Base(path), ".md")

	switch event {
	case "create", "write":
		if err := h.loadFile(path); err != nil {
			log.Printf("[host] %s %s failed: %v", event, filepath.Base(path), err)
			if event == "write" {
				log.Printf("[host] keeping previous version of %s", name)
				return
			}
		} else {
			log.Printf("[host] %s %s loaded", event, filepath.Base(path))
		}
		h.rebuildIndex()
	case "remove":
		h.mu.Lock()
		delete(h.pages, name)
		h.mu.Unlock()
		h.rebuildIndex()
		log.Printf("[host] removed %s", filepath.Base(path))
	}

	h.broadcast.send(reloadMsg{Type: "reload"})
}

func HostRun(opts HostOptions) error {
	h, err := newHostApp(opts)
	if err != nil {
		return err
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("watch %s: %w", opts.Dir, err)
	}
	defer watcher.Close()

	if err := watcher.Add(opts.Dir); err != nil {
		return fmt.Errorf("watch %s: %w", opts.Dir, err)
	}

	go func() {
		for {
			select {
			case ev, ok := <-watcher.Events:
				if !ok {
					return
				}
				var eventType string
				if ev.Has(fsnotify.Create) {
					eventType = "create"
				} else if ev.Has(fsnotify.Write) {
					eventType = "write"
				} else if ev.Has(fsnotify.Remove) || ev.Has(fsnotify.Rename) {
					eventType = "remove"
				} else {
					continue
				}
				evPath := ev.Name
				evType := eventType
				db := newDebouncer(100*time.Millisecond, func() {
					h.handleFileEvent(evType, evPath)
				})
				db.trigger()
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Printf("[host] watch error: %v", err)
			}
		}
	}()

	addr := fmt.Sprintf(":%d", opts.Port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("bind port %d: %w\nhint: try --port %d", opts.Port, err, opts.Port+1)
	}

	srv := &http.Server{Handler: h.mux}
	url := fmt.Sprintf("http://localhost:%d", opts.Port)
	fmt.Printf("GoSlide hosting %s at %s\n", opts.Dir, url)
	fmt.Printf("  %d presentations loaded\n", len(h.pages))

	if !opts.NoOpen {
		openBrowser(url)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
			log.Printf("[host] serve error: %v", err)
		}
	}()

	<-ctx.Done()
	fmt.Println("\nShutting down...")
	shutCtx, shutCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutCancel()
	return srv.Shutdown(shutCtx)
}
