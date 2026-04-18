package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/user/goslide/internal/config"
	"github.com/user/goslide/internal/ir"
	"github.com/user/goslide/internal/parser"
	"github.com/user/goslide/internal/renderer"
)

type Options struct {
	File    string
	Port    int
	Theme   string
	Accent  string
	NoOpen  bool
	NoWatch bool
	Verbose bool
}

type app struct {
	opts       Options
	cfg        *config.Config
	mux        *http.ServeMux
	broadcast  *broadcaster
	mu         sync.RWMutex
	cachedHTML string
}

func newApp(opts Options) (*app, error) {
	a := &app{
		opts:      opts,
		mux:       http.NewServeMux(),
		broadcast: newBroadcaster(),
	}
	a.setupRoutes()

	cfg, err := config.Load(filepath.Dir(opts.File))
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}
	a.cfg = cfg
	if len(cfg.API.Proxy) > 0 {
		setupProxy(a.mux, cfg.API.Proxy)
	}

	if err := a.loadAndRender(); err != nil {
		return nil, err
	}
	return a, nil
}

func (a *app) loadAndRender() error {
	data, err := os.ReadFile(a.opts.File)
	if err != nil {
		return fmt.Errorf("open file %s: %w", a.opts.File, err)
	}

	pres, err := parser.Parse(data, a.opts.File)
	if err != nil {
		return fmt.Errorf("parse %s: %w", a.opts.File, err)
	}

	if a.opts.Theme != "" {
		pres.Meta.Theme = a.opts.Theme
	}
	if a.opts.Accent != "" {
		pres.Meta.Accent = a.opts.Accent
	}

	valErrs := pres.Validate()
	if len(valErrs) > 0 {
		fmt.Fprint(os.Stderr, ir.FormatErrors(a.opts.File, valErrs))
	}
	if ir.HasErrors(valErrs) {
		return fmt.Errorf("validation failed for %s", a.opts.File)
	}

	html, err := renderer.Render(pres)
	if err != nil {
		return fmt.Errorf("render %s: %w", a.opts.File, err)
	}

	if a.cfg != nil && len(a.cfg.Theme.Overrides) > 0 {
		html = injectThemeOverrides(html, a.cfg.Theme.Overrides)
	}

	a.mu.Lock()
	a.cachedHTML = html
	a.mu.Unlock()

	return nil
}

func (a *app) reload() {
	if err := a.loadAndRender(); err != nil {
		log.Printf("[reload] %s changed but failed: %v", a.opts.File, err)
		log.Printf("[reload] keeping previous version")
		a.broadcast.send(reloadMsg{Type: "error", Message: err.Error()})
		return
	}
	log.Printf("[reload] %s reloaded successfully", a.opts.File)
	a.broadcast.send(reloadMsg{Type: "ok"})
	a.broadcast.send(reloadMsg{Type: "reload"})
}

func Run(opts Options) error {
	a, err := newApp(opts)
	if err != nil {
		return err
	}

	var stopWatch func()
	if !opts.NoWatch {
		stopWatch, err = watchFile(opts.File, a.reload)
		if err != nil {
			return fmt.Errorf("watch %s: %w", opts.File, err)
		}
		defer stopWatch()
	}

	addr := fmt.Sprintf(":%d", opts.Port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("bind port %d: %w\nhint: try --port %d", opts.Port, err, opts.Port+1)
	}

	srv := &http.Server{Handler: a.mux}

	url := fmt.Sprintf("http://localhost:%d", opts.Port)
	fmt.Printf("GoSlide serving %s at %s\n", opts.File, url)

	if !opts.NoOpen {
		openBrowser(url)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
			log.Printf("[server] serve error: %v", err)
		}
	}()

	<-ctx.Done()
	fmt.Println("\nShutting down...")

	shutCtx, shutCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutCancel()
	return srv.Shutdown(shutCtx)
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	cmd.Start()
}

func injectThemeOverrides(html string, overrides map[string]string) string {
	var sb strings.Builder
	sb.WriteString("<style>:root {")
	for k, v := range overrides {
		sb.WriteString(" --")
		sb.WriteString(k)
		sb.WriteString(": ")
		sb.WriteString(v)
		sb.WriteString(";")
	}
	sb.WriteString(" }</style>")
	return strings.Replace(html, "</head>", sb.String()+"</head>", 1)
}
