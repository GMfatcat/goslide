# 🎯 GoSlide

**Markdown-driven interactive presentations — single binary, offline-first.**

GoSlide turns `.md` files into Reveal.js presentations with live charts, diagrams, API dashboards, and interactive controls. No Node.js. No npm. Just one Go binary.

> 📖 [中文版 README](README_zh-TW.md)

---

## ⚡ Quick Start

```bash
# Install
go install github.com/GMfatcat/goslide/cmd/goslide@latest

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
| 🎨 **14 themes** | default, dark, corporate, minimal, hacker, dracula, midnight, gruvbox, solarized, catppuccin-mocha, ink-wash, instagram, western, pixel |
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

14 built-in themes × 8 accent colors = **112 visual combinations**.

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

## ⚙️ Configuration

Optional `goslide.yaml` in the same directory:

```yaml
# API proxy
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

## 📄 License

MIT
