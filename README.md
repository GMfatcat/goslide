# 🎯 GoSlide

**Markdown-driven interactive presentations — single binary, offline-first.**

GoSlide turns `.md` files into Reveal.js presentations with live charts, diagrams, API dashboards, and interactive controls. No Node.js. No npm. Just one Go binary.

> 📖 [中文版 README](README_zh-TW.md)

---

## ⚡ Quick Start

### Option 1: Download Binary

Download the latest release for your platform from [GitHub Releases](https://github.com/GMfatcat/goslide/releases):

| Platform | File |
|----------|------|
| Windows (x64) | `goslide-windows-amd64.exe` |
| macOS (Intel) | `goslide-darwin-amd64` |
| macOS (Apple Silicon) | `goslide-darwin-arm64` |
| Linux (x64) | `goslide-linux-amd64` |
| Linux (ARM64) | `goslide-linux-arm64` |

Rename to `goslide` (or `goslide.exe` on Windows), place in your PATH, and you're ready.

### Option 2: Install via Go

```bash
go install github.com/GMfatcat/goslide/cmd/goslide@latest
```

### Get Started

```bash
# Create a presentation
goslide init

# Serve with live reload
goslide serve talk.md

# Export as standalone HTML
goslide build talk.md
```

## ✨ Features

| Feature | Description |
|---------|-------------|
| 📝 **Markdown authoring** | Write slides in plain `.md` — frontmatter for config, `---` for slide breaks |
| 🎨 **22 themes** | default, dark, corporate, minimal, hacker, dracula, midnight, gruvbox, solarized, catppuccin-mocha, ink-wash, instagram, western, pixel, nord-light, paper, catppuccin-latte, chalk, synthwave, forest, rose, amoled |
| 📊 **Charts** | Bar, line, pie, radar, sparkline via Chart.js |
| 🔀 **Diagrams** | Mermaid.js flowcharts, sequence diagrams, ERD |
| 📋 **Tables** | Sortable tables with click-to-sort headers |
| 🎛️ **Interactive controls** | Tabs, sliders, toggles with reactive `$variable` binding |
| 🃏 **Expandable cards** | Grid layout with click-to-expand detail overlays |
| 🌐 **API dashboards** | Live data from backend APIs with auto-refresh |
| 🔌 **API proxy** | Built-in reverse proxy with auth header injection |
| 📦 **Single binary** | All assets embedded via `go:embed` (~8MB) |
| 🔄 **Live reload** | Edit `.md` → browser auto-refreshes, keeps slide position |
| 🖥️ **Speaker view** | Press `S` for timer, notes, next slide preview |
| 📤 **Static export** | `goslide build` → one `.html` file, works offline |
| 🏠 **Host mode** | Serve a directory as a presentation library |
| 📡 **Presenter sync** | Viewers see presenter's current slide + jump button |

## 🎨 Themes

22 built-in themes × 8 accent colors = **176 visual combinations**.

```yaml
---
theme: dracula
accent: pink
---
```

👉 [Full Theme Catalog](docs/THEMES.md)

## 📐 Layouts

12 slide layouts via HTML comments:

```markdown
---
<!-- layout: two-column -->

# Title

<!-- left -->
Left content

<!-- right -->
Right content
```

Available: `default`, `title`, `section`, `two-column`, `code-preview`, `three-column`, `image-left`, `image-right`, `quote`, `split-heading`, `top-bottom`, `grid-cards`, `blank`

## 📦 Components

Interactive components via fenced code blocks:

```markdown
~~~chart:bar
title: Revenue
labels: ["Q1", "Q2", "Q3"]
data: [100, 150, 200]
color: teal
~~~
```

👉 [Full Component Reference](docs/COMPONENTS.md)

## ⌨️ CLI

```bash
goslide serve <file.md>     # Serve with live reload
goslide host <directory>    # Host multiple presentations
goslide build <file.md>     # Export as standalone HTML
goslide init                # Scaffold new presentation
goslide list [directory]    # List presentations
```

👉 [Full CLI Reference](docs/CLI.md)

### AI slide generation

Generate a full presentation from a topic using any OpenAI-compatible LLM
endpoint (OpenAI, OpenRouter, Ollama, vllm, sglang, etc.).

Add a `generate:` section to `goslide.yaml`:

```yaml
generate:
  base_url: https://api.openai.com/v1
  model: gpt-4o
  api_key_env: OPENAI_API_KEY
  timeout: 120s
```

Export the API key, then run:

```bash
export OPENAI_API_KEY=sk-...
goslide generate "Introduction to Kubernetes"            # simple mode
goslide generate my-prompt.md -o talk.md                 # advanced mode
goslide generate --dump-prompt > system.txt              # inspect prompt
```

Advanced mode reads a `prompt.md` file:

```markdown
---
topic: Kubernetes Architecture
audience: Backend engineers
slides: 15
theme: dark
language: en
---
Emphasize Pod/Service/Ingress. End with a Q&A slide.
```

The command refuses to overwrite an existing output file unless `--force` is
passed. Generated Markdown is sanity-checked against the parser; common
issues (unclosed code fences, missing frontmatter terminator) are auto-fixed
with a transparent report.

## ⚙️ Configuration

Optional `goslide.yaml` in the same directory as your `.md` file:

```yaml
# API proxy — routes browser requests through Go server to upstream APIs
api:
  proxy:
    /api/backend:
      target: http://localhost:8000
      headers:
        Authorization: "Bearer ${API_TOKEN}"

# Custom theme overrides
theme:
  overrides:
    slide-bg: "#1e1e2e"
    slide-accent: "#f38ba8"
```

> **Note:** When `goslide.yaml` has proxy config, GoSlide will attempt to connect to the upstream targets on every proxied request. If an upstream is not running, you'll see `proxy error` in the console and 502 responses in the browser — this only affects API component slides, not the rest of your presentation.

### 🧪 Testing API Components with Mock Server

GoSlide includes a mock API server for testing API-driven slides:

```bash
# Terminal 1: Start mock API server
go run examples/mock-api/main.go
# → Mock API running on http://localhost:9999

# Terminal 2: Copy the example config and serve
cp examples/goslide.yaml.example examples/goslide.yaml
go run ./cmd/goslide serve examples/demo.md --no-open
# → Open http://localhost:3000, navigate to API Dashboard slides
```

The example config (`goslide.yaml.example`) proxies `/api/mock` to `localhost:9999`. Rename it to `goslide.yaml` to activate. When done testing, you can remove or rename the config to avoid proxy errors when the mock server isn't running.

## 📡 Presenter Sync

GoSlide has a lightweight presenter sync feature. The presenter's current slide is broadcast to all viewers in real-time.

**Presenter** opens with `?role=presenter`:
```
http://localhost:3000?role=presenter
```

**Viewers** open the normal URL:
```
http://localhost:3000
```

When the presenter navigates slides, viewers see a small indicator at the bottom-left:

```
┌──────────────────────────┐
│  Presenter: 5/12  [Jump] │
└──────────────────────────┘
```

- Viewers can click **Jump** to go to the presenter's current slide
- Viewers can freely browse on their own — they are never forced to follow
- Works in both `serve` and `host` mode

### Speaker View

Press **S** on any slide to open the speaker view in a new window. It shows:
- Current slide + next slide preview
- Speaker notes (from `<!-- notes -->` in your markdown)
- Elapsed time

## 🏗️ Build from Source

```bash
git clone https://github.com/GMfatcat/goslide.git
cd goslide
bash scripts/vendor.sh --update-checksums
go build -o goslide ./cmd/goslide
```

**Requirements:** Go 1.21+

## 🎬 Transitions

```yaml
---
transition: perspective  # 3D Y-axis rotation
---
```

Available: `slide` (default), `fade`, `convex`, `concave`, `zoom`, `none`, `perspective`, `flip`

## 📝 Speaker Notes

```markdown
# My Slide

Content here.

<!-- notes -->

Speaker notes — visible in speaker view (press S).
```

## 🙏 Acknowledgments

GoSlide is built on the shoulders of these excellent open-source projects:

| Project | Role | License |
|---------|------|---------|
| [Reveal.js](https://revealjs.com/) | Slide rendering engine | MIT |
| [Chart.js](https://www.chartjs.org/) | Charts (bar, line, pie, radar, sparkline) | MIT |
| [Mermaid](https://mermaid.js.org/) | Diagrams (flowcharts, sequence, ERD) | MIT |
| [goldmark](https://github.com/yuin/goldmark) | Markdown parser | MIT |
| [cobra](https://github.com/spf13/cobra) | CLI framework | Apache-2.0 |
| [fsnotify](https://github.com/fsnotify/fsnotify) | File system watcher | BSD-3 |
| [coder/websocket](https://github.com/coder/websocket) | WebSocket library | MIT |
| [Noto Sans TC](https://fonts.google.com/noto/specimen/Noto+Sans+TC) | CJK font | OFL-1.1 |
| [JetBrains Mono](https://www.jetbrains.com/lp/mono/) | Monospace font | OFL-1.1 |
| [Press Start 2P](https://fonts.google.com/specimen/Press+Start+2P) | Pixel font (pixel theme) | OFL-1.1 |
| [Rye](https://fonts.google.com/specimen/Rye) | Western font (western theme) | OFL-1.1 |

Theme color palettes inspired by: [Dracula](https://draculatheme.com/), [Catppuccin](https://github.com/catppuccin/catppuccin), [Gruvbox](https://github.com/morhetz/gruvbox), [Solarized](https://ethanschoonover.com/solarized/), [Nord](https://www.nordtheme.com/).

## 📄 License

MIT
