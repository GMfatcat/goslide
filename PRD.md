# GoSlide — Product Requirements Document

## 1. Overview

GoSlide is a Markdown-driven interactive presentation system built with Go and Reveal.js. It targets internal teams that need high-interactivity slides — live charts, API-driven dashboards, expandable cards, embedded LLM demos — without leaving the comfort of Markdown authoring.

GoSlide is **not** a PowerPoint replacement. It fills the gap between static slide decks and full web applications, giving presenters a way to showcase backend services, live data, and interactive demos inside a single `.md` file served as a browser-based presentation.

### Design principles

- **Progressive complexity** — write plain Markdown for basic slides; opt into interactivity only when needed.
- **Single binary** — Go compiles to one executable with all assets embedded via `go:embed`. No Node.js, no npm, no runtime dependencies.
- **Offline-first** — designed for air-gapped factory and intranet environments. All fonts, JS libraries, and CSS are bundled.
- **AI-friendly** — the Markdown spec is simple enough that any LLM can generate valid slides from a topic prompt.

### Target users

- Internal development teams doing technical presentations and demo showcases.
- Engineers who want to present backend service capabilities without building a frontend.
- Anyone who prefers writing Markdown over dragging boxes in PowerPoint.

---

## 2. Architecture

### Tech stack

| Layer | Technology | Role |
|-------|-----------|------|
| Server | Go (net/http) | Markdown parsing, file serving, API proxy, WebSocket sync |
| Markdown parser | goldmark + custom extensions | Parse MD into Reveal.js section structure |
| Presentation engine | Reveal.js | Slide rendering, navigation, transitions, speaker view |
| Charts | Chart.js | Bar, line, pie, sparkline, radar |
| Diagrams | Mermaid.js | Flowcharts, sequence diagrams, ERD |
| Fonts | Noto Sans TC, JetBrains Mono | Bundled via go:embed for CJK + monospace |
| Reactive runtime | Custom (~15 lines JS) | Lightweight event bus for `$variable` bindings |

### Binary packaging

All static assets (Reveal.js, Chart.js, Mermaid.js, CSS themes, fonts) are embedded into the Go binary via `go:embed`. The resulting `goslide` (or `goslide.exe`) is a self-contained executable with zero external dependencies.

Estimated binary size: ~15-20MB (fonts account for most of it).

---

## 3. Running modes

### 3.1 Local serve (single file)

```bash
goslide serve talk.md
```

Serves a single `.md` file as a presentation. Automatically opens the default browser to `localhost:3000`. Watches the file for changes and triggers live reload via WebSocket.

Intended for: individual authoring, rehearsal, presenting from your own machine.

### 3.2 Host server (directory)

```bash
goslide host ./slides
```

Serves an entire directory of `.md` files as a presentation library. Each file becomes a presentation at a URL path derived from its filename:

```
./slides/aoi-architecture.md  →  http://host:8080/talks/aoi-architecture
./slides/git-onboarding.md    →  http://host:8080/talks/git-onboarding
```

The root path (`/`) serves an index page listing all available presentations with titles, dates, and tags extracted from frontmatter.

Intended for: team knowledge sharing, persistent presentation hosting on an internal server.

---

## 4. Markdown spec

### 4.1 Basic structure

A GoSlide `.md` file uses `---` as slide separators and an optional YAML frontmatter block at the top:

```markdown
---
title: Presentation title
theme: dark
transition: slide
---

# First slide

Content here.

---

# Second slide

More content.
```

Rules:
- `---` at line start = slide separator (after frontmatter is closed).
- `#` = slide title.
- Standard Markdown (paragraphs, lists, bold, italic, code blocks, images, links, tables) is supported.
- HTML comments `<!-- -->` are used for per-slide metadata and content region markers.

### 4.2 Configuration hierarchy

Settings are resolved in order of precedence (highest wins):

```
Slide-level HTML comments  >  Frontmatter  >  Project config file  >  Defaults
```

**Frontmatter** — lives at the top of the `.md` file. Covers presentation-wide settings:

```yaml
---
title: My talk
theme: dark
accent: teal
transition: slide
fragments: true
fragment-style: highlight-current
---
```

**Project config file** — optional `goslide.yaml` in the same directory. Covers infrastructure settings (API proxies, base theme, shared defaults) and settings shared across multiple presentations:

```yaml
theme: corporate
accent: blue
api:
  proxy:
    /api/aoi:
      target: http://192.168.1.100:8000
      headers:
        Authorization: "Bearer ${AOI_TOKEN}"
    /api/ollama:
      target: http://spark-c037:11434
```

**Per-slide overrides** — HTML comments immediately after a slide separator:

```markdown
---
<!-- transition: fade -->
<!-- layout: two-column -->

# This slide only
```

When only a `.md` file exists (no config file), everything works with sensible defaults. Config files are opt-in for advanced use cases.

---

## 5. Layout system

GoSlide provides 12 built-in layout templates. Each layout is selected via `layout:` in frontmatter or per-slide HTML comments. Content regions within a layout are delimited by HTML comments (`<!-- left -->`, `<!-- right -->`, etc.).

### 5.1 Layout catalog

| Layout | Description | Regions |
|--------|-------------|---------|
| `default` | Title + centered content | (none — single flow) |
| `title` | Large centered title + subtitle | (none — single flow) |
| `section` | Chapter divider with accent line | (none — single flow) |
| `two-column` | Two equal columns | `<!-- left -->`, `<!-- right -->` |
| `three-column` | Three equal columns | `<!-- col1 -->`, `<!-- col2 -->`, `<!-- col3 -->` |
| `image-left` | Image on left, text on right | `<!-- image -->`, `<!-- text -->` |
| `image-right` | Text on left, image on right | `<!-- text -->`, `<!-- image -->` |
| `code-preview` | Code on left, output/explanation on right | `<!-- code -->`, `<!-- preview -->` |
| `quote` | Centered blockquote with attribution | (none — single flow, uses `>` and `—`) |
| `split-heading` | Large heading on left, body content on right | `<!-- heading -->`, `<!-- body -->` |
| `top-bottom` | Visual/embed on top, text on bottom | `<!-- top -->`, `<!-- bottom -->` |
| `grid-cards` | 2×2 (or 2×N) card grid | Uses `card` component blocks |
| `blank` | Empty — full custom HTML/component control | (none) |

### 5.2 Layout usage example

```markdown
---
<!-- layout: two-column -->

# Go vs Python

<!-- left -->
## Go
- Static typing
- Single binary output
- Fast compilation

<!-- right -->
## Python
- Dynamic typing
- Rich ML ecosystem
- Rapid prototyping
```

### 5.3 Default layout

If no `layout` is specified, `default` is used. It centers the title and flows content below it — suitable for 80% of slides.

---

## 6. Interactive component system

Components are the core differentiator of GoSlide. They are embedded in Markdown via fenced code blocks with custom language tags. The component system has three layers of increasing complexity.

### 6.1 Layer 1 — Declarative (data in YAML, zero JS)

Users provide static data directly in the code block. GoSlide renders it as an interactive widget.

#### chart

Renders Chart.js-based charts with hover tooltips.

```
~~~chart:bar
title: Yield by production line
labels: ["Line A", "Line B", "Line C", "Line D"]
data: [96.2, 93.8, 97.1, 91.5]
unit: "%"
color: teal
~~~
```

Supported chart types: `bar`, `line`, `pie`, `radar`, `sparkline`.

Parameters:
- `title` (string) — chart title.
- `labels` (string[]) — axis labels or legend entries.
- `data` (number[] | number[][]) — single dataset or multiple datasets.
- `datasets` (object[]) — for multi-series: `[{label, data, color}]`.
- `unit` (string) — suffix for tooltip values.
- `color` (string) — color ramp name. For multi-series, use `datasets[].color`.
- `stacked` (bool) — stack bars/areas. Default `false`.

#### table

Renders a sortable, interactive table.

```
~~~table
columns: [Name, Role, Department]
rows:
  - ["Alice", "Engineer", "AOI"]
  - ["Bob", "PM", "R&D"]
  - ["Carol", "Lead", "AI Tools"]
sortable: true
~~~
```

#### mermaid

Renders Mermaid.js diagrams.

```
~~~mermaid
graph TD
    A[Image Capture] --> B[Preprocessing]
    B --> C[SegFormer Inference]
    C --> D[Defect Classification]
    D --> E{Pass?}
    E -->|Yes| F[OK]
    E -->|No| G[NG]
~~~
```

### 6.2 Layer 2 — Reactive controls (declarative + `$variable` binding)

Interactive controls that communicate with other components on the same slide via a lightweight reactive store.

#### Runtime

GoSlide injects a minimal reactive event bus into every slide page:

```javascript
// Auto-injected by GoSlide — not user-written
const store = {};
const listeners = {};
window.GoSlide = {
  set(key, value) {
    store[key] = value;
    (listeners[key] || []).forEach(fn => fn(value));
  },
  get(key) { return store[key]; },
  on(key, fn) {
    (listeners[key] = listeners[key] || []).push(fn);
  }
};
```

#### tabs + panel

```
~~~tabs
id: compare
labels: ["Plan A", "Plan B", "Plan C"]
~~~

~~~panel:compare-0
Plan A details here...
~~~

~~~panel:compare-1
Plan B details here...
~~~

~~~panel:compare-2
Plan C details here...
~~~
```

Tabs toggle visibility of corresponding panels. Panel IDs follow the pattern `{tabs.id}-{index}`.

#### slider

```
~~~slider
id: threshold
label: Yield threshold
min: 80
max: 100
value: 95
step: 0.5
unit: "%"
~~~
```

Publishes its value to `$threshold` in the reactive store.

#### toggle

```
~~~toggle
id: show_details
label: Show details
default: false
~~~
```

Publishes `true`/`false` to `$show_details`.

#### Variable binding

Any component parameter can reference a reactive variable using `$` prefix:

```
~~~chart:bar
title: Line status
labels: ["A", "B", "C"]
data: [96.2, 93.8, 97.1]
threshold: "$threshold"
~~~
```

When `$threshold` changes (e.g., from a slider), the chart re-renders with the new threshold line. Components below the threshold turn red; above turn green.

### 6.3 Layer 3 — API-driven components

Connect Markdown slides to live backend services. All API requests are proxied through the Go server (see section 4.2 config for proxy setup).

#### api component

```
~~~api
url: /api/aoi/status?line=A
refresh: 5s
render:
  - type: metric
    path: yield
    label: Yield
    unit: "%"
    color: green
  - type: chart:pie
    path: defects
    title: Defect distribution
  - type: chart:sparkline
    path: trend
    title: Trend
~~~
```

Parameters:
- `url` (string) — API endpoint (proxied through Go server).
- `method` (string) — HTTP method. Default `GET`.
- `refresh` (duration) — polling interval. Omit for one-shot fetch.
- `body` (object) — request body for POST.
- `layout` (string) — `auto` (default), `dashboard`, `horizontal`, `vertical`.
- `render` (object[]) — list of render items.

#### Render types

| Type | Input | Output |
|------|-------|--------|
| `metric` | Single number (via JSONPath `path`) | Large number card with label, unit, optional color |
| `chart:bar/line/pie/sparkline/radar` | Array or object | Chart.js chart |
| `table` | Array of objects | Sortable table with auto-detected columns |
| `json` | Any JSON | Collapsible tree viewer |
| `log` | String or string[] | Terminal-style scrolling text (monospace, dark bg) |
| `image` | Base64 string or URL | Rendered image |
| `markdown` | Markdown string | Rendered Markdown content |
| `chat` | (special) | Chat interface for LLM interaction |

#### Render item parameters

- `type` (string) — render type from table above.
- `path` (string) — JSONPath expression to extract data from API response. Omit to use full response.
- `label` / `title` (string) — display label.
- `unit` (string) — suffix for values.
- `color` (string) — color ramp name.
- `span` (int) — grid span for dashboard layout. Default `1`.
- `columns` (string[]) — for `table` type, which fields to show and in what order.

#### Dashboard layout

When `layout: dashboard` is specified, render items are arranged in a responsive grid. Items flow left-to-right, top-to-bottom. `span: 2` makes an item take two columns.

#### Chat render type (LLM integration)

```
~~~api
url: /api/ollama/v1/chat/completions
method: POST
render: chat
model: nemotron-nano:30b
system: You are an AOI optical inspection expert.
placeholder: Ask me anything about AOI...
~~~
```

Renders a chat interface on the slide. User (or presenter) types a question, the response streams in via the API proxy. Messages are maintained in-session.

### 6.4 Layer 4 — Full custom (escape hatch)

For anything the built-in components don't cover.

#### embed:html

```
~~~embed:html
<div id="demo"></div>
<script>
  fetch('/api/aoi/latest')
    .then(r => r.json())
    .then(data => {
      document.getElementById('demo').innerHTML =
        `Yield: ${data.yield}%`;
    });
</script>
~~~
```

Full HTML/CSS/JS execution within the slide. Has access to `window.GoSlide` for reactive variable binding.

#### embed:iframe

```
~~~embed:iframe
url: http://localhost:3000/dashboard
height: 500
~~~
```

Embeds an external page in an iframe.

---

## 7. Expandable detail-on-demand pattern

A cross-cutting interaction pattern available on `grid-cards` layout and other container components. Cards display a summary view; clicking a card expands it into a modal-like overlay showing full detail content. Clicking outside or pressing `Escape` collapses it back.

### 7.1 Markdown syntax

The `card` code block uses `---` as an internal separator between summary (above) and detail (below):

```markdown
---
layout: grid-cards
columns: 2
expand: true
---

# AOI System Overview

~~~card
icon: clock
color: blue
title: Image capture
desc: High-speed camera acquisition
---
## Image capture

| Capture speed | Resolution | Lighting modes |
|:---:|:---:|:---:|
| 120 fps | 5MP | 4 |

The image capture module uses MIL to interface with industrial cameras.
Each station captures under multiple lighting conditions.
~~~

~~~card
icon: grid
color: teal
title: Defect detection
desc: SegFormer + YOLO hybrid pipeline
---
## Defect detection

Accuracy: 98.7% | Latency: 12ms | 8 defect classes

The pipeline combines SegFormer for semantic segmentation
with YOLO for discrete defect detection.
~~~
```

### 7.2 Behavior

- Cards render in a grid (2×N by default, configurable via `columns`).
- Each card shows `icon`, `title`, and `desc` in summary view.
- Click → card content fades into a centered overlay (85% of slide area) showing the detail content below `---`.
- Overlay background dims the rest of the slide.
- Dismiss: click outside, press `Escape`, or click close button.
- Detail content supports full Markdown rendering including nested components (charts, tables, code blocks).
- `expand: true` is opt-in. Without it, `grid-cards` renders as static cards.

### 7.3 Applicability

The expand pattern is not limited to `grid-cards`. It can be applied to:
- Timeline events (click an event to see details).
- Flowchart nodes (click a step to see implementation details).
- Any container where overview → detail drilling is useful.

---

## 8. Navigation and controls

### 8.1 Slide navigation model

GoSlide uses **horizontal single-axis navigation only** (no vertical sub-slides). The detail-on-demand pattern replaces the need for sub-slides.

### 8.2 Keyboard shortcuts

| Key | Action |
|-----|--------|
| `→` / `Space` / `Enter` | Next slide |
| `←` / `Backspace` | Previous slide |
| `Escape` | Toggle overview mode / close expanded card |
| `F` | Toggle fullscreen |
| `S` | Open speaker view |
| `G` + number + `Enter` | Jump to slide N |
| `Home` | First slide |
| `End` | Last slide |

### 8.3 Overview mode

Pressing `Escape` (when no overlay is open) enters overview mode: all slides shrink to thumbnails in a grid. Click any thumbnail to jump to that slide. Press `Escape` again to return to normal view.

### 8.4 Progress bar

A thin progress indicator at the bottom of the viewport. Clickable — clicking a position jumps to the proportional slide. Non-intrusive by default; becomes fully visible on hover.

### 8.5 Host mode: sync control

Host mode supports two audience modes, selected via URL parameter:

```
/talks/my-talk              → Free browsing (default)
/talks/my-talk?sync=abc123  → Presenter-controlled sync
```

**Free browsing**: each viewer navigates independently. Suitable for async review ("here's the link, read it when you want").

**Presenter sync**: the presenter's navigation is broadcast to all connected viewers via WebSocket. The Go server manages sync sessions.

- Presenter opens the talk and a session ID is generated.
- Viewers join with `?sync={session_id}`.
- Viewer navigation is locked to presenter's current slide.
- Viewers can temporarily "unfollow" (button in corner) to browse freely, then re-follow with one click.

---

## 9. Transitions and animations

### 9.1 Slide transitions

Transition effects between slides. Set globally in frontmatter, override per-slide via HTML comment.

Available transitions (provided by Reveal.js):
- `slide` (default) — horizontal push.
- `fade` — cross-fade.
- `convex` — 3D convex rotation.
- `concave` — 3D concave rotation.
- `zoom` — zoom in/out.
- `none` — instant switch.

```markdown
---
transition: slide
---

# Slide 1

---
<!-- transition: fade -->

# Slide 2 (fades in)
```

### 9.2 Fragment animations (progressive reveal)

Content within a slide can appear incrementally on each click/advance.

**Slide-level flag** (recommended for most cases):

```markdown
---
<!-- fragments: true -->
<!-- fragment-style: highlight-current -->

# Key points

- First point (visible immediately)
- Second point (appears on click)
- Third point (appears on next click)
```

When `fragments: true`, all list items after the first become fragments. This is the simplest approach and covers 80% of use cases.

**Manual fragment markers** (for fine-grained control):

```markdown
This text is always visible.

<!-- fragment -->

This appears on first click.

<!-- fragment -->

This appears on second click.
```

### 9.3 Fragment styles

| Style | Behavior |
|-------|----------|
| `fade-in` (default) | Fades in from transparent |
| `fade-up` | Fades in while sliding up slightly |
| `highlight-current` | Current fragment is fully opaque; previous fragments dim to 40% opacity |

`highlight-current` is recommended for technical presentations — it keeps the audience focused on the point currently being discussed.

---

## 10. Theme system

### 10.1 Built-in themes

| Theme | Description | Best for |
|-------|-------------|----------|
| `default` | White background, dark text, clean and neutral | Daily internal presentations |
| `dark` | Dark background, light text, code-friendly | Technical demos, code-heavy talks |
| `corporate` | Structured with accent bars, subtle footer, muted palette | Formal reports to management |
| `minimal` | Maximum whitespace, large type, very sparse | Keynote-style talks, one idea per slide |
| `hacker` | Terminal aesthetic, monospace font, green/amber on black | Live coding demos, hacker culture |

### 10.2 Theme selection and accent override

```markdown
---
theme: dark
accent: teal
---
```

`theme` sets the overall visual style. `accent` overrides the primary highlight color used for links, active elements, chart defaults, and interactive controls.

Available accent colors: `blue`, `teal`, `purple`, `coral`, `amber`, `green`, `red`, `pink`.

5 themes × 8 accents = 40 visual combinations out of the box.

### 10.3 Technical implementation

Each theme is a CSS file defining CSS custom properties:

```css
:root {
  --slide-bg: #1a1a2e;
  --slide-text: #e0e0e0;
  --slide-heading: #ffffff;
  --slide-accent: var(--accent-teal);
  --slide-code-bg: #0f0f1a;
  --slide-code-text: #e0e0e0;
  --slide-border: rgba(255, 255, 255, 0.1);
  --slide-muted: #888888;
  --slide-card-bg: #252540;
  --slide-font-sans: 'Noto Sans TC', sans-serif;
  --slide-font-mono: 'JetBrains Mono', monospace;
}
```

All components (charts, cards, tables, API widgets) inherit from these variables, ensuring visual consistency when switching themes.

### 10.4 Bundled fonts

| Font | Purpose | Approx. size |
|------|---------|-------------|
| Noto Sans TC | Body text, headings (full CJK + Latin) | ~4MB |
| JetBrains Mono | Code blocks, `hacker` theme, `log` render type | ~1MB |

Fonts are embedded via `go:embed` and served by the Go server. No external font loading.

### 10.5 Future: custom themes

Not in initial version. The architecture supports future customization via `goslide.yaml`:

```yaml
theme:
  base: dark
  overrides:
    slide-bg: "#1e1e2e"
    slide-accent: "#f38ba8"
```

---

## 11. API proxy

### 11.1 Purpose

Slide-embedded API requests are proxied through the Go server rather than being fetched directly from the browser. This solves:
- **CORS** — no cross-origin issues.
- **Security** — internal API addresses are not exposed to the client.
- **Auth injection** — Go server can inject auth tokens from config/environment.
- **Caching** — frequently polled endpoints can be cached server-side.

### 11.2 Configuration

In `goslide.yaml`:

```yaml
api:
  proxy:
    /api/aoi:
      target: http://192.168.1.100:8000
      headers:
        Authorization: "Bearer ${AOI_TOKEN}"
    /api/ollama:
      target: http://spark-c037:11434
    /api/translate:
      target: http://spark-c037:8080
  cache:
    default_ttl: 5s
```

Environment variables in header values (e.g., `${AOI_TOKEN}`) are expanded at runtime.

### 11.3 Routing

Any request from the slide frontend to a path matching a configured proxy prefix is forwarded:

```
Browser: GET /api/aoi/status?line=A
  → Go server: GET http://192.168.1.100:8000/status?line=A
    (with injected Authorization header)
  → Response proxied back to browser
```

---

## 12. CLI design

### 12.1 Command structure

```
goslide <command> [flags] [arguments]
```

### 12.2 Commands

#### serve

Serve a single Markdown file as a presentation.

```
goslide serve <file.md> [flags]

Flags:
  -p, --port int          Port number (default: 3000)
  -t, --theme string      Override theme
  -a, --accent string     Override accent color
      --no-open           Don't auto-open browser
      --no-watch          Disable file watching / live reload
```

#### host

Serve a directory of presentations.

```
goslide host <directory> [flags]

Flags:
  -p, --port int          Port number (default: 8080)
      --index             Enable index page listing all presentations (default: true)
      --sync              Enable presenter sync feature (default: false)
```

#### build (future)

Export a presentation as self-contained static HTML.

```
goslide build <file.md> [flags]

Flags:
  -o, --output string     Output directory (default: ./build)
      --single-file       Inline all assets into one HTML file
```

#### init

Scaffold a new presentation with example content.

```
goslide init [flags]

Flags:
  -t, --template string   Template to use: basic, demo, corporate (default: basic)
```

Creates a `talk.md` in the current directory with frontmatter, sample slides demonstrating various layouts and components, and a `goslide.yaml` with commented-out configuration examples.

#### list

List presentations in a directory.

```
goslide list [directory]
```

Outputs a table of all `.md` files with their title, theme, and slide count.

### 12.3 Global flags

```
  -v, --version    Print version
  -h, --help       Print help
      --verbose    Verbose logging
```

---

## 13. AI slide generation

### 13.1 Design goal

A user should be able to generate a complete, valid GoSlide presentation by giving a topic to any LLM. The Markdown spec is intentionally simple to maximize generation reliability.

### 13.2 Generation prompt template

GoSlide can ship with a built-in system prompt (stored in the binary) that describes the Markdown spec, available layouts, and component syntax. This prompt is used by:

1. `goslide generate` command (future) — calls a configured LLM endpoint to generate slides from a topic.
2. External use — users can feed this prompt to any LLM manually.

### 13.3 Spec simplicity for AI

The AI-facing spec subset should be minimal:
- `---` separates slides.
- YAML frontmatter for `title`, `theme`, `layout`.
- HTML comments for per-slide layout and region markers.
- Standard Markdown for content.
- Component code blocks for charts and cards.

Advanced features (API bindings, reactive variables, embed:html) are outside the AI-generation scope and are added manually by the author.

---

## 14. Project structure

```
goslide/
├── cmd/
│   └── goslide/
│       └── main.go              # CLI entry point
├── internal/
│   ├── server/
│   │   ├── server.go            # HTTP server (serve + host modes)
│   │   ├── proxy.go             # API proxy handler
│   │   ├── websocket.go         # Live reload + presenter sync
│   │   └── handlers.go          # Route handlers
│   ├── parser/
│   │   ├── parser.go            # Markdown → slide structure
│   │   ├── frontmatter.go       # YAML frontmatter extraction
│   │   ├── components.go        # Component code block parsing
│   │   └── layout.go            # Layout region splitting
│   ├── renderer/
│   │   ├── renderer.go          # Slide structure → Reveal.js HTML
│   │   ├── components.go        # Component → HTML/JS generation
│   │   └── theme.go             # Theme CSS injection
│   └── config/
│       └── config.go            # goslide.yaml parsing and merging
├── web/
│   ├── static/
│   │   ├── reveal/              # Reveal.js (embedded)
│   │   ├── chartjs/             # Chart.js (embedded)
│   │   ├── mermaid/             # Mermaid.js (embedded)
│   │   ├── fonts/               # Noto Sans TC, JetBrains Mono
│   │   └── runtime.js           # GoSlide reactive runtime
│   ├── themes/
│   │   ├── default.css
│   │   ├── dark.css
│   │   ├── corporate.css
│   │   ├── minimal.css
│   │   └── hacker.css
│   ├── layouts/
│   │   └── *.css                # Per-layout CSS
│   └── templates/
│       ├── slide.html           # Main slide HTML template
│       ├── index.html           # Host mode index page
│       └── speaker.html         # Speaker view template
├── templates/
│   ├── basic.md                 # Init template: basic
│   ├── demo.md                  # Init template: demo (showcases components)
│   └── corporate.md             # Init template: corporate
├── go.mod
├── go.sum
└── README.md
```

---

## 15. Development phases

### Phase 1 — Core (MVP)

Deliver a working `goslide serve` with basic Markdown → Reveal.js rendering.

- [x] Go HTTP server with `go:embed` static assets
- [x] Goldmark-based Markdown parser with `---` slide splitting
- [x] YAML frontmatter parsing
- [x] Layout system: `default`, `title`, `section`, `two-column`, `code-preview`
- [x] Theme system: `default`, `dark` themes with accent color support
- [x] Font embedding: Noto Sans TC, JetBrains Mono
- [x] File watcher + WebSocket live reload
- [x] CLI: `goslide serve` with basic flags
- [x] Fragment animation: slide-level `fragments: true`
- [x] Keyboard navigation (arrow keys, fullscreen, overview)

### Phase 2 — Components

Add the interactive component layer.

- [x] Component code block parser (fenced blocks with custom language tags)
- [x] L1 components: `chart:bar`, `chart:line`, `chart:pie`, `mermaid`, `table`
- [x] L2 components: `tabs` + `panel`, `slider`, `toggle`
- [x] Reactive runtime (`$variable` binding)
- [x] Remaining layouts: `three-column`, `image-left`, `image-right`, `quote`, `split-heading`, `top-bottom`, `grid-cards`, `blank`
- [x] Expandable detail-on-demand pattern
- [x] Remaining themes: `corporate`, `minimal`, `hacker`

### Phase 3 — API integration

Connect slides to live backend services.

- [x] API proxy in Go server
- [x] `api` component with render types: `metric`, `chart:*`, `table`, `json`, `log`, `image`, `markdown`
- [x] Dashboard layout for multi-render-item arrangement
- [ ] `chat` render type for LLM integration (deferred to Phase 5)
- [x] `embed:html` and `embed:iframe` components
- [x] `goslide.yaml` config file parsing with proxy settings
- [x] Polling / refresh support

### Phase 4 — Host mode and collaboration

Multi-presentation hosting and presenter sync.

- [x] CLI: `goslide host` with directory watching
- [x] Index page generation (list all presentations)
- [x] Presenter sync via WebSocket (lightweight: broadcast + jump button)
- [x] CLI: `goslide init` with templates
- [x] CLI: `goslide list`
- [x] Progress bar with click-to-jump
- [ ] Speaker view (timer, notes, next slide preview) (moved to Phase 5)

### Phase 5 — Polish and export

- [ ] CLI: `goslide build` (static HTML export)
- [ ] Single-file export (all assets inlined)
- [ ] Custom theme overrides via config
- [ ] `goslide generate` command (LLM-powered slide generation)
- [ ] Per-slide speaker notes (`<!-- notes: ... -->`)
- [ ] PDF export (via headless browser)

---

## Appendix A: Complete frontmatter reference

```yaml
---
# Presentation metadata
title: string              # Presentation title (shown in index, browser tab)
author: string             # Author name
date: string               # Date string (free format)
tags: string[]             # Tags for index page filtering

# Theme
theme: string              # default | dark | corporate | minimal | hacker
accent: string             # blue | teal | purple | coral | amber | green | red | pink

# Transitions
transition: string         # slide | fade | convex | concave | zoom | none

# Fragments
fragments: bool            # Enable progressive reveal on lists (default: false)
fragment-style: string     # fade-in | fade-up | highlight-current
---
```

## Appendix B: Per-slide HTML comment options

```markdown
---
<!-- layout: two-column -->
<!-- transition: fade -->
<!-- fragments: true -->
<!-- fragment-style: highlight-current -->
```

## Appendix C: Component quick reference

```
~~~chart:bar|line|pie|radar|sparkline
title, labels, data, datasets, unit, color, stacked
~~~

~~~table
columns, rows, sortable
~~~

~~~mermaid
(mermaid diagram source)
~~~

~~~tabs
id, labels
~~~

~~~panel:{tabs-id}-{index}
(markdown content)
~~~

~~~slider
id, label, min, max, value, step, unit
~~~

~~~toggle
id, label, default
~~~

~~~api
url, method, refresh, body, layout, render[]
  render[].type: metric|chart:*|table|json|log|image|markdown|chat
  render[].path, label, title, unit, color, span, columns
~~~

~~~embed:html
(raw HTML/CSS/JS)
~~~

~~~embed:iframe
url, height
~~~

~~~card
icon, color, title, desc
---
(expanded detail content — full markdown)
~~~
```
