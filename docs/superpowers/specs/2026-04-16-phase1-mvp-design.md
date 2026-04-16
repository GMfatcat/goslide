# GoSlide Phase 1 (MVP) Design Spec

> **Status: COMPLETED (2026-04-16)**
>
> All 15 implementation tasks done. Manual testing passed. Bugfixes applied:
> - Enter/Backspace keyboard navigation added (Reveal.js 5.x doesn't bind these by default)
> - Base font size set to 36px (browser default 16px too small for presentations)
> - Navigation controls colored with accent for dark theme visibility
> - WS broadcaster tests stabilized with registration wait
>
> Remaining discussion: page number indicator mechanism (deferred to future session)

## Decision Log

| # | Question | Decision |
|---|----------|----------|
| 1 | MVP scope | PRD Phase 1 all 10 items, no carry-over |
| 2 | Parser strategy | Pre-split (line scan) then per-slide goldmark |
| 3 | Test strategy | Go unit + HTML golden (3-5 examples) + MANUAL_CHECKLIST.md |
| 4 | Asset vendoring | `scripts/vendor.sh` + commit artifacts + SHA256 checksum + VERSIONS.md |
| 5 | CLI framework | spf13/cobra |
| 6 | Live reload | Full page reload + sessionStorage slide index, fsnotify watcher |
| 7 | Theme injection | Static CSS files via go:embed + `data-accent` attribute on body |
| 8 | Unknown syntax | Warn + fallback; Levenshtein <= 2 against whitelist -> error + "did you mean?" |
| 9 | WebSocket library | coder/websocket |
| 10 | Case sensitivity | Enum fields (theme/accent/layout/transition/fragment-style) normalize to lowercase silently |

---

## 1. Architecture Overview

```
                           +-----------------------------------+
  user .md file            |             goslide               |
  goslide.yaml     ---->   |                                   |  ----> localhost:3000
                           |  +-----------------------------+  |        (HTML + WS)
                           |  | internal/parser             |  |
                           |  |   Split -> Frontmatter -> IR|  |
                           |  +-------------+---------------+  |
                           |                |                  |
                           |                v                  |
                           |  +-----------------------------+  |
                           |  | internal/ir                 |  |
                           |  |   Presentation{Slides, ...} |  |
                           |  |   .Validate() -> []Error    |  |<-- Phase 2+
                           |  +-------------+---------------+  |   extension point
                           |                |                  |
                           |                v                  |
                           |  +-----------------------------+  |
                           |  | internal/renderer           |  |
                           |  |   IR + theme -> HTML        |  |
                           |  |   (html/template + embed)   |  |
                           |  +-------------+---------------+  |
                           |                |                  |
                           |                v                  |
                           |  +-----------------------------+  |
  fsnotify ------------>   |  | internal/server             |  |
                           |  |   HTTP + WebSocket + Watch  |  |
                           |  |   /themes/*, /fonts/*, /ws  |  |
                           |  +-----------------------------+  |
                           +-----------------------------------+
```

### Package Tree

```
goslide/
├── cmd/goslide/main.go              # cobra root, version injection
├── internal/
│   ├── ir/
│   │   ├── presentation.go          # Presentation, Slide, Meta types
│   │   ├── validate.go              # Validate() []Error (whitelist + fuzzy)
│   │   └── validate_test.go
│   ├── parser/
│   │   ├── split.go                 # frontmatter + slide splitting + code-fence state machine
│   │   ├── split_test.go
│   │   ├── frontmatter.go           # YAML -> ir.Frontmatter
│   │   ├── slide.go                 # per-slide: HTML comments + goldmark body
│   │   └── parser.go                # Parse([]byte) (*ir.Presentation, error)
│   ├── renderer/
│   │   ├── renderer.go              # Render(ir.Presentation, theme) ([]byte, error)
│   │   ├── renderer_test.go         # includes golden tests
│   │   └── testdata/golden/         # *.md -> *.html pairs
│   ├── theme/
│   │   ├── theme.go                 # accent whitelist, data-accent assembly
│   │   └── assets.go                # embed.FS for themes/ + fonts/
│   ├── server/
│   │   ├── server.go                # Run(Options) error
│   │   ├── handlers.go              # /, /themes/*, /fonts/*, /ws
│   │   ├── watch.go                 # fsnotify + debounce
│   │   └── reload.go                # WS broadcaster (coder/websocket)
│   └── cli/
│       ├── root.go                  # cobra root cmd + --verbose
│       └── serve.go                 # cobra serve cmd
├── web/
│   ├── static/                      # vendor.sh output: reveal/, fonts/
│   │   ├── reveal/
│   │   │   ├── reveal.js
│   │   │   └── reveal.css
│   │   ├── fonts/
│   │   │   ├── NotoSansTC.woff2
│   │   │   └── JetBrainsMono.woff2
│   │   ├── runtime.js               # our live reload + toast runtime
│   │   ├── CHECKSUMS.sha256
│   │   └── VERSIONS.md
│   ├── themes/
│   │   ├── tokens.css               # accent vars, font-face, toast
│   │   ├── default.css
│   │   ├── dark.css
│   │   └── layouts.css              # 5 layout CSS rules
│   └── templates/
│       └── slide.html               # html/template with Reveal.js
├── scripts/
│   └── vendor.sh                    # download assets + SHA256 verify
├── examples/
│   └── demo.md                      # manual checklist target
├── MANUAL_CHECKLIST.md
├── go.mod
└── README.md
```

---

## 2. IR Types

```go
package ir

import "html/template"

type Presentation struct {
    Source   string        // file path, for error messages
    Meta     Frontmatter
    Slides   []Slide
    Warnings []Error       // accumulated by Validate()
}

type Slide struct {
    Index    int           // 1-based, for error messages
    Meta     SlideMeta     // merged final metadata (HTML comments > frontmatter)
    BodyHTML template.HTML // goldmark output (Phase 2 adds Components field)
}

type SlideMeta struct {
    Title         string
    Layout        string                  // default|title|section|two-column|code-preview
    Transition    string
    Fragments     bool
    FragmentStyle string
    Regions       map[string]template.HTML // "left"/"right"/"code"/"preview" etc.
}

type Frontmatter struct {
    Title         string   `yaml:"title"`
    Author        string   `yaml:"author"`
    Date          string   `yaml:"date"`
    Tags          []string `yaml:"tags"`
    Theme         string   `yaml:"theme"`
    Accent        string   `yaml:"accent"`
    Transition    string   `yaml:"transition"`
    Fragments     bool     `yaml:"fragments"`
    FragmentStyle string   `yaml:"fragment-style"`
}

type Error struct {
    Slide    int    // 0 = presentation-level
    Severity string // "error" | "warning"
    Code     string // "unknown-layout", "typo-suggestion", etc.
    Message  string
    Hint     string // e.g. `did you mean "two-column"?`
}
```

---

## 3. Data Flow (`goslide serve hello.md`)

1. `cli/serve.go` -> `server.Run(opts)`
2. Read `hello.md` bytes -> `parser.Parse(bytes)` -> `*ir.Presentation`
3. Apply CLI flag overrides (theme/accent) onto `presentation.Meta`
4. `presentation.Validate()` -> `[]Error`; any `severity == "error"` -> fatal exit with all errors printed
5. `renderer.Render(presentation, theme)` -> `[]byte` HTML (in-memory cache)
6. `GET /` -> serve cached HTML
7. `GET /themes/dark.css`, `/fonts/NotoSansTC.woff2` -> serve from `embed.FS`
8. `GET /ws` -> WebSocket upgrade, add to broadcaster
9. `fsnotify` detects `hello.md` change (100ms debounce) -> re-run steps 2-5
   - On success: broadcaster sends `{"type":"reload"}` + `{"type":"ok"}`
   - On failure: log to stderr, keep `lastGood`, broadcaster sends `{"type":"error","message":"..."}`
10. Browser JS receives `reload` -> save `Reveal.getIndices()` to sessionStorage -> `location.reload()`
11. On page load: JS reads sessionStorage -> `Reveal.slide(h, v, f)` to restore position

---

## 4. Error Handling

### 4.1 Error Categories

| Category | Source | Severity | Handling |
|----------|--------|----------|----------|
| Setup error | CLI / I/O / env | fatal | cobra `RunE` returns error, exit 1 |
| Validation error | `ir.Validate()` | error or warning | error -> abort serve; warning -> stderr, continue |
| Runtime error | watcher re-parse | recoverable | keep lastGood, WS toast, log to stderr |

### 4.2 Validation Rules

| Rule | Severity | Code |
|------|----------|------|
| frontmatter `theme` not in {default, dark} | error | `unknown-theme` |
| frontmatter `accent` not in 8-color whitelist | error | `unknown-accent` |
| frontmatter `transition` not in 6 valid values | warning | `unknown-transition` |
| slide `layout` typo of Phase 1 name (Levenshtein <= 2) | error | `typo-suggestion` |
| slide `layout` is Phase 2+ name (in future whitelist) | warning | `future-layout` |
| slide `layout` completely unknown | warning | `unknown-layout` |
| Phase 2 component fence (`~~~chart:bar` etc.) | warning | `future-component` |
| `fragments: true` but slide has no list | warning | `fragments-noop` |
| `two-column` layout but missing `<!-- left -->` or `<!-- right -->` | error | `missing-region` |

### 4.3 Whitelist Layers

- `phase1Layouts = {default, title, section, two-column, code-preview}` -> typo target
- `futureLayouts = {three-column, image-left, image-right, quote, split-heading, top-bottom, grid-cards, blank}` -> hit = warning
- Neither -> Levenshtein against union; <= 2 -> error; > 2 -> warning

### 4.4 Case Handling

All enum fields (theme, accent, transition, layout, fragment-style) are normalized to lowercase via `strings.ToLower(strings.TrimSpace(...))` at parser entry. Levenshtein comparison runs on normalized values.

### 4.5 Validation Output Format

```
talk.md:
  warning [unknown-transition] frontmatter: transition "swirl" not recognized (using default)
  warning [future-layout] slide 4: layout "grid-cards" not implemented in Phase 1 (using default)
  error   [typo-suggestion] slide 7: layout "two-colum" -- did you mean "two-column"?
  error   [missing-region] slide 9: layout "two-column" but no <!-- left --> region found

2 errors, 2 warnings -- refusing to serve
```

### 4.6 Runtime Error Recovery

```
[reload] talk.md changed but parse failed: yaml: line 3: mapping values not allowed
[reload] keeping previous version
```

WS pushes `{"type":"error","message":"..."}` -> frontend shows red toast in bottom-right. On next successful parse, push `{"type":"ok"}` -> toast disappears.

---

## 5. Testing Strategy

### 5.1 Test Matrix

| Layer | Tool | What | Run |
|-------|------|------|-----|
| Unit | `go test` + `testify/require` | pure functions (split/frontmatter/validate/theme/fuzzy) | `go test ./internal/...` |
| Golden | `go test` + custom helper | full pipeline output (parser -> renderer) | same, with `-update` flag |
| Manual | human + browser | Reveal.js behavior, CSS visuals, live reload feel | Phase 1 sign-off |

### 5.2 Minimum Required Unit Tests

| Package | Required Cases |
|---------|---------------|
| `parser/split` | plain text / with frontmatter / multi-slide / `---` inside fenced code / `---` as setext heading / Windows CRLF |
| `parser/frontmatter` | complete / missing fields / unknown fields ignored / YAML syntax error returns SetupError |
| `parser/slide` | HTML comment metadata / region splitting (two-column) / goldmark body / case normalize |
| `ir/validate` | one case per rule in section 4.2 + Levenshtein boundary (distance=2 / =3) + case-insensitive hit |
| `theme` | all 8 accents queryable / unknown accent returns error / data-accent attribute correct |
| `renderer` | default theme / dark theme / fragment flag maps to correct Reveal.js attribute |
| `server` | lastGood mechanism (bad md -> keeps previous) / debounce / WS broadcast |

### 5.3 Golden Test Setup

Location: `internal/renderer/testdata/golden/`

Structure per example: `{name}.md` (input), `{name}.html` (expected output), `{name}.stderr` (expected validation messages, may be empty).

3-5 examples, each with a single intent:
- `basic-default-theme` — minimal 2-slide with default theme
- `two-column-dark` — two-column layout with dark theme
- `fragments-on` — slide with fragments: true and highlight-current
- `all-typos` — validation errors collection (no .html output expected)
- `crlf-windows` — Windows line endings

`-update` flag: `go test -run TestGolden -update ./internal/renderer` regenerates golden files. Must `git diff testdata/` and review before commit.

### 5.4 Subagent Delivery Discipline

Before each task completion, subagent must run:
1. `go test ./... -race`
2. `go vet ./...`
3. `gofmt -l .` (must be empty output)
4. List new/modified test case names in delivery message
5. If golden files changed, list affected examples and diff summary

---

## 6. Frontend Runtime

### 6.1 Template: `web/templates/slide.html`

Single html/template file. Go injects `Presentation` data at render time.

Key attributes:
- `<body data-accent="{{.Meta.Accent}}">` for accent CSS switching
- `<section data-layout="{{.Meta.Layout}}" data-transition="{{.Meta.Transition}}">` per slide
- Reveal.js loaded as UMD build (vendor artifact)
- `runtime.js` loaded last (our code)

### 6.2 `web/static/runtime.js` (~50 lines)

Responsibilities:
1. On page load: restore slide index from sessionStorage, call `Reveal.slide(h, v, f)`
2. Open WebSocket to `/ws`
3. On `{"type":"reload"}`: save `Reveal.getIndices()` to sessionStorage, `location.reload()`
4. On `{"type":"error"}`: show red toast with message
5. On `{"type":"ok"}`: hide toast
6. On WS close: show "server disconnected" toast

### 6.3 Theme CSS Structure

```
web/themes/
├── tokens.css      # 8 accent vars, @font-face, toast style, accent attribute selectors
├── default.css     # --slide-* variables for white theme
├── dark.css        # --slide-* variables for dark theme
└── layouts.css     # [data-layout="..."] rules for 5 layouts
```

Accent switching via `body[data-accent="teal"] { --slide-accent: var(--accent-teal); }` in `tokens.css`.

### 6.4 Reveal.js Configuration

- Version: 5.x latest stable (UMD build)
- Plugins: none (we parse markdown ourselves)
- `hash: false` (avoid conflict with sessionStorage reload)
- `controls: true`, `progress: true`, `overview: true`, `center: true`, `keyboard: true`
- `slideNumber: true` (enables G + number + Enter jump)

### 6.5 Route Table

| Path | Source | Behavior |
|------|--------|----------|
| `/` | renderer cache | HTML, re-rendered on file change |
| `/themes/{name}.css` | embed.FS | static |
| `/static/reveal/*` | embed.FS | static (vendor) |
| `/static/runtime.js` | embed.FS | static (our code) |
| `/fonts/*.woff2` | embed.FS | static, `Cache-Control: max-age=31536000` |
| `/ws` | coder/websocket | upgrade |

MIME: `font/woff2` explicitly set for `.woff2` files.

---

## 7. CLI

### 7.1 Command Structure

```
goslide (root)            --verbose, --version
├── serve <file.md>       -p/--port, -t/--theme, -a/--accent, --no-open, --no-watch
├── host  <dir>           (stub: "not implemented")
├── init                  (stub)
├── list  [dir]           (stub)
└── build <file.md>       (stub)
```

Phase 1 implements `serve` only. Other commands registered as stubs returning `"not implemented: available in a future release"`.

### 7.2 Flag Precedence

```
CLI flag  >  frontmatter  >  default
```

CLI override applied to `presentation.Meta` before `Validate()`.

### 7.3 Version Injection

```go
// cmd/goslide/main.go
var version = "dev"
// build: go build -ldflags "-X main.version=0.1.0" -o goslide ./cmd/goslide
```

### 7.4 Auto-Open Browser

`exec.Command` with OS-specific launcher (`rundll32`/`open`/`xdg-open`). Skipped with `--no-open`. Not unit tested (OS-dependent, covered by manual checklist).

### 7.5 Graceful Shutdown

`signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)` -> close watcher -> close WS -> `srv.Shutdown(5s timeout)`.

---

## 8. Asset Vendoring

### 8.1 `scripts/vendor.sh`

Downloads pinned versions of Reveal.js + fonts to `web/static/`. Generates or verifies `CHECKSUMS.sha256`.

Workflow:
1. First run: `bash scripts/vendor.sh --update-checksums` (download + generate checksums)
2. `git commit web/static/` including checksums
3. Subsequent runs: `bash scripts/vendor.sh` (verify only)
4. Upgrade: change version -> delete old files -> `--update-checksums` -> review diff -> commit

### 8.2 `web/static/VERSIONS.md`

| Asset | Version | License | Source |
|-------|---------|---------|--------|
| Reveal.js | 5.1.0 | MIT | https://github.com/hakimel/reveal.js |
| Noto Sans TC | v2.008 | OFL-1.1 | https://github.com/notofonts/noto-cjk |
| JetBrains Mono | 2.304 | OFL-1.1 | https://github.com/JetBrains/JetBrainsMono |

### 8.3 go:embed

```go
// paths relative to .go file location; all: prefix includes dotfiles
//go:embed all:../../web/themes
var ThemeFS embed.FS

//go:embed all:../../web/static
var StaticFS embed.FS

//go:embed all:../../web/templates
var TemplateFS embed.FS
```

---

## 9. Dependencies

```
github.com/spf13/cobra          v1.8.x
github.com/yuin/goldmark        v1.7.x
github.com/fsnotify/fsnotify    v1.8.x
coder.com/websocket             v1.8.x
gopkg.in/yaml.v3                latest
github.com/agnivade/levenshtein v1.2.x
github.com/stretchr/testify     v1.9.x   (test only)
```

7 direct dependencies. No framework, no ORM, no config library beyond cobra/pflag.

---

## 10. Phases 2-5 Skeleton

### Phase 2 — Components

- Component code block parser (`~~~chart:bar` -> IR `Slide.Components []Component`)
- L1: chart (bar/line/pie/radar/sparkline via Chart.js), mermaid, sortable table
- L2: tabs/panel, slider, toggle, reactive runtime (`$variable` binding)
- Remaining 7 layouts: three-column, image-left, image-right, quote, split-heading, top-bottom, grid-cards, blank
- Expandable card pattern (detail-on-demand)
- 3 themes: corporate, minimal, hacker
- Completion: `examples/components-demo.md` showcasing all L1+L2 + all layouts + all 5 themes x 8 accents
- Risk: Chart.js/Mermaid binary bloat (~3-5MB); reactive runtime vs Reveal.js fragment state conflict; grid-cards expand pattern Esc key conflict with overview mode

### Phase 3 — API Integration

- `goslide.yaml` full config parsing
- Go server API proxy (prefix match -> reverse proxy + header injection + env var expansion)
- `api` component with render types: metric, chart:*, table, json, log, image, markdown
- Dashboard layout (multi-render-item grid)
- `chat` render type (streaming LLM via SSE)
- `embed:html`, `embed:iframe`
- Polling/refresh, server-side cache
- Completion: `examples/api-demo.md` with mock HTTP server showing metric + refresh chart + chat
- Risk: SSRF prevention in proxy; SSE flush behavior; consider introducing headless browser e2e tests

### Phase 4 — Host Mode & Collaboration

- `goslide host <dir>`: directory scan -> multi-presentation routing (`/talks/{name}`)
- Index page (list all .md with title/date/tags from frontmatter)
- Directory file watcher (add/delete/modify .md)
- Presenter sync via WebSocket (session management, follow/unfollow)
- `goslide init --template basic|demo|corporate`
- `goslide list [dir]`
- Progress bar with click-to-jump
- Completion: two machines, A runs `goslide host ./slides`, B joins with `?sync=xxx`, navigation syncs
- Risk: concurrent map in multi-session WS; init template content must showcase all features up to current phase

### Phase 5 — Polish & Export

- `goslide build <file.md>` -> self-contained static HTML (`--single-file` inlines everything)
- Custom theme overrides via `goslide.yaml` (`theme.overrides` map)
- `goslide generate` (LLM-powered slide generation, provider abstraction)
- Speaker view (timer, notes `<!-- notes: ... -->`, next slide preview)
- PDF export (headless browser, optional feature)
- CI/CD pipeline (GitHub Actions: test + build + release)
- README / docs / examples polish
- Completion: `goslide build examples/demo.md --single-file` produces one .html that works offline; GitHub release with Windows + Linux binaries
- Risk: base64 fonts in single-file (~7MB HTML); headless Chrome cross-platform; LLM provider abstraction scope
