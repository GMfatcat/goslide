# Phase 4a: Host Mode Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add `goslide host <dir>` command that serves a directory of `.md` files as a presentation library with an index page, individual talk routes at `/talks/{name}`, and live reload on file changes.

**Architecture:** New `hostApp` struct in `internal/server/host.go` manages N presentations. Scans directory at startup, renders all `.md` files, serves index at `/` and talks at `/talks/{name}`. Directory-level fsnotify watcher handles create/write/delete events. Static asset routes extracted into shared helper used by both serve and host modes.

**Tech Stack:** Go 1.21.6, existing parser/renderer/config packages. No new dependencies.

**Shell rules (Windows):** Never chain commands with `&&`. Use SEPARATE Bash calls for `git add`, `git commit`, `go test`, `go build`. Use `GOTOOLCHAIN=local` prefix for ALL go commands. Use `-C` flag for go/git to specify directory.

---

## Task 1: Extract Shared Static Routes

**Files:**
- Modify: `internal/server/handlers.go`
- Modify: `internal/server/server_test.go`

- [ ] **Step 1: Extract setupStaticRoutes from handlers.go**

Read `internal/server/handlers.go`. Refactor `setupRoutes` to extract the static asset routes into a standalone function. Replace the file with:

```go
package server

import (
	"io/fs"
	"net/http"
	"strings"

	"github.com/user/goslide/web"
)

func setupStaticRoutes(mux *http.ServeMux) {
	themeSub, _ := fs.Sub(web.ThemeFS, "themes")
	mux.Handle("/themes/", http.StripPrefix("/themes/", http.FileServer(http.FS(themeSub))))

	staticSub, _ := fs.Sub(web.StaticFS, "static")
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticSub))))

	fontsSub, _ := fs.Sub(web.StaticFS, "static/fonts")
	mux.Handle("/fonts/", http.StripPrefix("/fonts/", addCacheHeader(http.FileServer(http.FS(fontsSub)))))
}

func (a *app) setupRoutes() {
	a.mux.HandleFunc("/", a.handleIndex)
	a.mux.HandleFunc("/ws", a.broadcast.handleWS)
	setupStaticRoutes(a.mux)
}

func (a *app) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	a.mu.RLock()
	html := a.cachedHTML
	a.mu.RUnlock()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

func addCacheHeader(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, ".woff2") {
			w.Header().Set("Content-Type", "font/woff2")
		}
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		h.ServeHTTP(w, r)
	})
}
```

- [ ] **Step 2: Run all tests to verify refactor didn't break anything**

Run: `GOTOOLCHAIN=local go test -C D:/CLAUDE-CODE-GOSLIDE ./internal/... -count=1`

Expected: all PASS.

- [ ] **Step 3: Commit**

```
git -C D:/CLAUDE-CODE-GOSLIDE add internal/server/handlers.go
```
```
git -C D:/CLAUDE-CODE-GOSLIDE commit -m "refactor(server): extract setupStaticRoutes for shared use by serve and host modes"
```

---

## Task 2: Index Page Template

**Files:**
- Create: `web/templates/index.html`

- [ ] **Step 1: Create index.html template**

Create `web/templates/index.html`:

```html
<!DOCTYPE html>
<html lang="zh-TW">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>GoSlide</title>
  <link rel="stylesheet" href="/themes/tokens.css">
  <link rel="stylesheet" href="/themes/default.css">
  <style>
    body {
      font-family: var(--font-sans);
      max-width: 800px;
      margin: 2rem auto;
      padding: 0 1rem;
      color: var(--slide-text);
      background: var(--slide-bg);
    }
    h1 { color: var(--slide-heading); margin-bottom: 1.5rem; }
    .talk-list { list-style: none; padding: 0; }
    .talk-item {
      padding: 1rem;
      margin: 0.5rem 0;
      background: var(--slide-card-bg);
      border-radius: 0.5rem;
      border: 1px solid var(--slide-border, rgba(0,0,0,0.1));
    }
    .talk-item a {
      color: var(--slide-accent);
      text-decoration: none;
      font-size: 1.2em;
      font-weight: 700;
    }
    .talk-item a:hover { filter: brightness(1.2); }
    .talk-meta {
      color: var(--slide-muted);
      font-size: 0.85em;
      margin-top: 0.3rem;
    }
    .talk-tags { margin-top: 0.3rem; }
    .talk-tags span {
      background: var(--slide-accent);
      color: white;
      padding: 0.1em 0.5em;
      border-radius: 0.3em;
      font-size: 0.75em;
      margin-right: 0.3em;
    }
    .empty { color: var(--slide-muted); font-style: italic; }
  </style>
</head>
<body>
  <h1>Presentations</h1>
  {{if .Pages}}
  <ul class="talk-list">
    {{range .Pages}}
    <li class="talk-item">
      <a href="/talks/{{.Name}}">{{if .Title}}{{.Title}}{{else}}{{.Name}}{{end}}</a>
      <div class="talk-meta">
        {{if .Author}}{{.Author}}{{end}}
        {{if .Date}} · {{.Date}}{{end}}
      </div>
      {{if .Tags}}
      <div class="talk-tags">
        {{range .Tags}}<span>{{.}}</span>{{end}}
      </div>
      {{end}}
    </li>
    {{end}}
  </ul>
  {{else}}
  <p class="empty">No presentations found. Add .md files to this directory.</p>
  {{end}}
</body>
</html>
```

- [ ] **Step 2: Verify embed compiles**

Run: `GOTOOLCHAIN=local go build -C D:/CLAUDE-CODE-GOSLIDE ./web/`

- [ ] **Step 3: Commit**

```
git -C D:/CLAUDE-CODE-GOSLIDE add web/templates/index.html
```
```
git -C D:/CLAUDE-CODE-GOSLIDE commit -m "feat: add index.html template for host mode presentation listing"
```

---

## Task 3: Host App Core

**Files:**
- Create: `internal/server/host.go`
- Create: `internal/server/host_test.go`

- [ ] **Step 1: Write failing tests**

Create `internal/server/host_test.go`:

```go
package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func writeMD(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	require.NoError(t, os.WriteFile(path, []byte(content), 0644))
	return path
}

func TestHostApp_ScanDirectory(t *testing.T) {
	dir := t.TempDir()
	writeMD(t, dir, "talk-a.md", "---\ntitle: Talk A\nauthor: Alice\n---\n\n# Slide 1\n")
	writeMD(t, dir, "talk-b.md", "---\ntitle: Talk B\ndate: 2026-01-01\ntags: [go, slides]\n---\n\n# Slide 1\n")
	writeMD(t, dir, "not-md.txt", "ignore me")

	h, err := newHostApp(HostOptions{Dir: dir, Port: 0})
	require.NoError(t, err)
	require.Len(t, h.pages, 2)
	require.Contains(t, h.pages, "talk-a")
	require.Contains(t, h.pages, "talk-b")
	require.Equal(t, "Talk A", h.pages["talk-a"].Title)
	require.Equal(t, "Alice", h.pages["talk-a"].Author)
}

func TestHostApp_IndexPage(t *testing.T) {
	dir := t.TempDir()
	writeMD(t, dir, "demo.md", "---\ntitle: Demo Talk\n---\n\n# Hello\n")

	h, err := newHostApp(HostOptions{Dir: dir, Port: 0})
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	h.mux.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), "Demo Talk")
	require.Contains(t, rec.Body.String(), "/talks/demo")
}

func TestHostApp_ServeTalk(t *testing.T) {
	dir := t.TempDir()
	writeMD(t, dir, "my-talk.md", "---\ntitle: My Talk\n---\n\n# Slide 1\n\nContent.\n")

	h, err := newHostApp(HostOptions{Dir: dir, Port: 0})
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/talks/my-talk", nil)
	rec := httptest.NewRecorder()
	h.mux.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), "Slide 1")
	require.Contains(t, rec.Body.String(), "reveal.js")
}

func TestHostApp_TalkNotFound(t *testing.T) {
	dir := t.TempDir()
	writeMD(t, dir, "exists.md", "# Slide\n")

	h, err := newHostApp(HostOptions{Dir: dir, Port: 0})
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/talks/nonexist", nil)
	rec := httptest.NewRecorder()
	h.mux.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
}

func TestHostApp_EmptyDirectory(t *testing.T) {
	dir := t.TempDir()

	h, err := newHostApp(HostOptions{Dir: dir, Port: 0})
	require.NoError(t, err)
	require.Len(t, h.pages, 0)

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	h.mux.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), "No presentations found")
}

func TestHostApp_StaticAssets(t *testing.T) {
	dir := t.TempDir()

	h, err := newHostApp(HostOptions{Dir: dir, Port: 0})
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/themes/tokens.css", nil)
	rec := httptest.NewRecorder()
	h.mux.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), "--accent-blue")
}

func TestHostApp_FileCreate(t *testing.T) {
	dir := t.TempDir()
	writeMD(t, dir, "initial.md", "---\ntitle: Initial\n---\n\n# Slide\n")

	h, err := newHostApp(HostOptions{Dir: dir, Port: 0})
	require.NoError(t, err)
	require.Len(t, h.pages, 1)

	writeMD(t, dir, "new-talk.md", "---\ntitle: New Talk\n---\n\n# New\n")
	h.handleFileEvent("create", filepath.Join(dir, "new-talk.md"))

	h.mu.RLock()
	count := len(h.pages)
	h.mu.RUnlock()
	require.Equal(t, 2, count)
}

func TestHostApp_FileDelete(t *testing.T) {
	dir := t.TempDir()
	writeMD(t, dir, "to-delete.md", "---\ntitle: Delete Me\n---\n\n# Slide\n")

	h, err := newHostApp(HostOptions{Dir: dir, Port: 0})
	require.NoError(t, err)
	require.Len(t, h.pages, 1)

	os.Remove(filepath.Join(dir, "to-delete.md"))
	h.handleFileEvent("remove", filepath.Join(dir, "to-delete.md"))

	h.mu.RLock()
	count := len(h.pages)
	h.mu.RUnlock()
	require.Equal(t, 0, count)
}

func TestHostApp_FileModify(t *testing.T) {
	dir := t.TempDir()
	writeMD(t, dir, "modify.md", "---\ntitle: Original\n---\n\n# Old\n")

	h, err := newHostApp(HostOptions{Dir: dir, Port: 0})
	require.NoError(t, err)
	require.Equal(t, "Original", h.pages["modify"].Title)

	writeMD(t, dir, "modify.md", "---\ntitle: Updated\n---\n\n# New\n")
	h.handleFileEvent("write", filepath.Join(dir, "modify.md"))

	h.mu.RLock()
	title := h.pages["modify"].Title
	h.mu.RUnlock()
	require.Equal(t, "Updated", title)
}
```

- [ ] **Step 2: Implement host.go**

Create `internal/server/host.go`:

```go
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
	iofs "io/fs"
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
	mux       *http.ServeMux
	broadcast *broadcaster
	mu        sync.RWMutex
	pages     map[string]*pageEntry
	indexHTML  string
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
	tmplFS, err := iofs.Sub(web.TemplateFS, "templates")
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

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(page.HTML))
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

	db := newDebouncer(100*time.Millisecond, func() {})
	go func() {
		var pendingEvent string
		var pendingPath string
		for {
			select {
			case ev, ok := <-watcher.Events:
				if !ok {
					return
				}
				if ev.Has(fsnotify.Create) {
					pendingEvent = "create"
				} else if ev.Has(fsnotify.Write) {
					pendingEvent = "write"
				} else if ev.Has(fsnotify.Remove) || ev.Has(fsnotify.Rename) {
					pendingEvent = "remove"
				} else {
					continue
				}
				pendingPath = ev.Name
				pe := pendingEvent
				pp := pendingPath
				db = newDebouncer(100*time.Millisecond, func() {
					h.handleFileEvent(pe, pp)
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
```

- [ ] **Step 3: Run tests**

Run: `GOTOOLCHAIN=local go test -C D:/CLAUDE-CODE-GOSLIDE/internal/server -run TestHostApp -v`

Run: `GOTOOLCHAIN=local go test -C D:/CLAUDE-CODE-GOSLIDE ./internal/... -count=1 -race`

Expected: all PASS.

- [ ] **Step 4: Commit**

```
git -C D:/CLAUDE-CODE-GOSLIDE add internal/server/host.go internal/server/host_test.go
```
```
git -C D:/CLAUDE-CODE-GOSLIDE commit -m "feat(server): add host mode with multi-presentation routing, index page, and directory watcher"
```

---

## Task 4: CLI — Host Command

**Files:**
- Create: `internal/cli/host.go`
- Modify: `internal/cli/root.go`

- [ ] **Step 1: Create host.go**

Create `internal/cli/host.go`:

```go
package cli

import (
	"github.com/spf13/cobra"
	"github.com/user/goslide/internal/server"
)

var (
	hostPort   int
	hostNoOpen bool
)

var hostCmd = &cobra.Command{
	Use:   "host <directory>",
	Short: "Serve a directory of presentations",
	Args:  cobra.ExactArgs(1),
	RunE:  runHost,
}

func init() {
	hostCmd.Flags().IntVarP(&hostPort, "port", "p", 8080, "port number")
	hostCmd.Flags().BoolVar(&hostNoOpen, "no-open", false, "don't auto-open browser")
	rootCmd.AddCommand(hostCmd)
}

func runHost(cmd *cobra.Command, args []string) error {
	return server.HostRun(server.HostOptions{
		Dir:    args[0],
		Port:   hostPort,
		NoOpen: hostNoOpen,
	})
}
```

- [ ] **Step 2: Remove host stub from root.go**

Read `internal/cli/root.go`. In the `stubs` slice, remove the `host` entry. Change:

```go
	stubs := []struct {
		use, short string
	}{
		{"host <directory>", "Serve a directory of presentations"},
		{"init", "Scaffold a new presentation"},
		{"list [directory]", "List presentations in a directory"},
		{"build <file.md>", "Export presentation as static HTML"},
	}
```

To:

```go
	stubs := []struct {
		use, short string
	}{
		{"init", "Scaffold a new presentation"},
		{"list [directory]", "List presentations in a directory"},
		{"build <file.md>", "Export presentation as static HTML"},
	}
```

- [ ] **Step 3: Verify build**

Run: `GOTOOLCHAIN=local go build -C D:/CLAUDE-CODE-GOSLIDE -o D:/CLAUDE-CODE-GOSLIDE/goslide.exe ./cmd/goslide`

- [ ] **Step 4: Test CLI**

Run: `D:/CLAUDE-CODE-GOSLIDE/goslide.exe host --help`

Expected: shows host usage with `--port` and `--no-open` flags.

- [ ] **Step 5: Commit**

```
git -C D:/CLAUDE-CODE-GOSLIDE add internal/cli/host.go internal/cli/root.go
```
```
git -C D:/CLAUDE-CODE-GOSLIDE commit -m "feat(cli): add host command replacing stub"
```

---

## Task 5: Test Examples + Demo + Checklist

**Files:**
- Create: `examples/slides/intro.md`
- Create: `examples/slides/features.md`
- Modify: `MANUAL_CHECKLIST.md`

- [ ] **Step 1: Create test slides for host mode**

Create `examples/slides/intro.md`:

```markdown
---
title: GoSlide Introduction
author: GoSlide Team
date: 2026-04-18
tags: [intro, overview]
---

# GoSlide

Markdown-driven interactive presentations.

---

# Features

- Single binary
- Offline-first
- Live reload
```

Create `examples/slides/features.md`:

```markdown
---
title: GoSlide Features Deep Dive
author: GoSlide Team
date: 2026-04-18
tags: [features, technical]
---

# Charts

~~~chart:bar
title: Demo Chart
labels: ["A", "B", "C"]
data: [10, 20, 30]
~~~

---

# Mermaid

~~~mermaid
graph LR
    A --> B --> C
~~~
```

- [ ] **Step 2: Update MANUAL_CHECKLIST.md**

Read `MANUAL_CHECKLIST.md`. Append:

```markdown

## Host Mode (Phase 4a)
- [ ] `goslide host examples/slides --no-open` → starts on port 8080
- [ ] Index page at `/` lists both presentations with titles
- [ ] Click presentation link → renders slides at `/talks/{name}`
- [ ] `/talks/nonexist` → 404
- [ ] Static assets (themes, fonts) work in host mode
- [ ] Add new .md to directory → appears in index after reload
- [ ] Delete .md → removed from index after reload
- [ ] Modify .md → slides update after live reload
- [ ] goslide.yaml proxy works in host mode
```

- [ ] **Step 3: Run all tests + build**

Run: `GOTOOLCHAIN=local go test -C D:/CLAUDE-CODE-GOSLIDE ./... -count=1 -race`

Run: `GOTOOLCHAIN=local go build -C D:/CLAUDE-CODE-GOSLIDE -ldflags "-X main.version=0.7.0" -o D:/CLAUDE-CODE-GOSLIDE/goslide.exe ./cmd/goslide`

- [ ] **Step 4: Commit**

```
git -C D:/CLAUDE-CODE-GOSLIDE add examples/slides/ MANUAL_CHECKLIST.md
```
```
git -C D:/CLAUDE-CODE-GOSLIDE commit -m "feat: add host mode test slides and Phase 4a manual checklist"
```
