# GoSlide Phase 4a — Host Mode Design Spec

> **Status: DRAFT**

## Decision Log

| # | Question | Decision |
|---|----------|----------|
| 1 | Architecture | Independent `HostRun` + shared helpers (not extending serve's `app`) |
| 2 | URL routing | `/talks/{name}` prefix, index at `/` |
| 3 | Index page | Go template, server-side render |
| 4 | Directory watcher | Watch directory level, classify events (create/write/delete) |
| 5 | Cache strategy | Render all at startup, re-render single file on change |

---

## 1. Host App Architecture

### `hostApp` struct

```go
type hostApp struct {
    dir       string
    port      int
    mux       *http.ServeMux
    broadcast *broadcaster
    mu        sync.RWMutex
    pages     map[string]*pageEntry
    indexHTML  string
    cfg       *config.Config
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
```

### HostRun flow

```
1. config.Load(dir)
2. Scan dir/*.md → Parse → Validate → Render → store in pages map
3. renderIndex() → generate index HTML from pages
4. setupHostRoutes() → register all routes
5. watchDir(dir) → fsnotify on directory
6. ListenAndServe on :8080
```

---

## 2. Route Table

| Path | Source | Behavior |
|------|--------|----------|
| `/` | index template | List all presentations |
| `/talks/{name}` | pages map | Individual slide HTML |
| `/themes/*` | embed.FS | Static (shared with serve) |
| `/static/*` | embed.FS | Static (shared with serve) |
| `/fonts/*` | embed.FS | Static + cache header (shared) |
| `/ws` | coder/websocket | Live reload broadcast |
| `/api/*` | proxy | If goslide.yaml has proxy config |

### Slide routing

`/talks/aoi-architecture` → lookup `pages["aoi-architecture"]` → serve cached HTML.

File-to-name mapping: `filepath.Base(path)` with `.md` stripped, e.g.:
- `./slides/aoi-architecture.md` → `aoi-architecture`
- `./slides/git-onboarding.md` → `git-onboarding`

404 if name not in pages map.

---

## 3. Index Page

### Template: `web/templates/index.html`

```html
<!DOCTYPE html>
<html lang="zh-TW">
<head>
  <meta charset="UTF-8">
  <title>GoSlide</title>
  <link rel="stylesheet" href="/themes/tokens.css">
  <link rel="stylesheet" href="/themes/default.css">
  <style>
    body { font-family: var(--font-sans); max-width: 800px; margin: 2rem auto; padding: 0 1rem; color: var(--slide-text); background: var(--slide-bg); }
    h1 { color: var(--slide-heading); }
    .talk-list { list-style: none; padding: 0; }
    .talk-item { padding: 1rem; margin: 0.5rem 0; background: var(--slide-card-bg); border-radius: 0.5rem; border: 1px solid var(--slide-border); }
    .talk-item a { color: var(--slide-accent); text-decoration: none; font-size: 1.2em; font-weight: 700; }
    .talk-item a:hover { filter: brightness(1.2); }
    .talk-meta { color: var(--slide-muted); font-size: 0.85em; margin-top: 0.3rem; }
    .talk-tags span { background: var(--slide-accent); color: white; padding: 0.1em 0.5em; border-radius: 0.3em; font-size: 0.75em; margin-right: 0.3em; }
  </style>
</head>
<body>
  <h1>Presentations</h1>
  <ul class="talk-list">
    {{range .Pages}}
    <li class="talk-item">
      <a href="/talks/{{.Name}}">{{.Title}}</a>
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
</body>
</html>
```

### Data source

`renderIndex()` iterates `pages` map, sorts by title or date, passes to template.

### Live update

When .md files are added/removed, `renderIndex()` is called again and `indexHTML` is updated. WS broadcast triggers browser reload on index page too.

---

## 4. Directory Watcher

### Watch setup

```go
func watchDir(dir string, onChange func(event string, path string)) (func(), error)
```

Single `fsnotify.Add(dir)` on the directory. Events classified:

```
fsnotify.Create *.md →
  Parse → Validate → Render → add to pages → rebuild index → WS reload

fsnotify.Write *.md →
  Re-render file → update pages entry → WS reload

fsnotify.Remove *.md →
  Remove from pages → rebuild index → WS reload

Non-.md events → ignored
```

Debounce 100ms (same as serve mode, reuse `debouncer` struct).

### Error recovery

Parse/render failure on a specific file → log error, skip that file, keep other files serving. The failed file is removed from pages (or kept at last-good if it was a Write event).

---

## 5. Shared Helpers

Extract from `handlers.go` into reusable functions:

```go
func setupStaticRoutes(mux *http.ServeMux) {
    // themes, static, fonts — same code as current setupRoutes minus / and /ws
}
```

Both `app.setupRoutes()` (serve) and `hostApp.setupHostRoutes()` (host) call `setupStaticRoutes(mux)`.

`openBrowser(url)` already standalone — no change needed.

---

## 6. CLI

### `internal/cli/host.go`

```go
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
```

Default port 8080 (different from serve's 3000).

Remove the host stub from `root.go`.

---

## 7. Testing

| Test | Method |
|------|--------|
| Directory scan + pages | TempDir with 3 .md files → newHostApp → assert len(pages)==3 |
| Index page content | httptest GET `/` → assert contains all titles |
| Slide routing | httptest GET `/talks/{name}` → assert contains slide HTML |
| 404 on unknown | httptest GET `/talks/nonexist` → 404 |
| File create event | Write new .md → debounce → assert pages grows |
| File delete event | Delete .md → assert pages shrinks |
| File modify event | Modify .md title → assert HTML updates |

---

## 8. Files Changed Summary

| Action | File |
|--------|------|
| Create | `internal/server/host.go` — hostApp, HostRun, host routes, dir watcher |
| Create | `internal/server/host_test.go` |
| Create | `web/templates/index.html` |
| Modify | `internal/server/handlers.go` — extract setupStaticRoutes |
| Modify | `internal/cli/root.go` — remove host stub |
| Create | `internal/cli/host.go` — host command with flags |
| Modify | `MANUAL_CHECKLIST.md` — Phase 4a checklist |
