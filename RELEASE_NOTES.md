# 🎉 GoSlide v1.0.0

**Markdown-driven interactive presentations — single binary, offline-first.**

## ✨ Highlights

- **14 themes** — default, dark, corporate, minimal, hacker, dracula, midnight, gruvbox, solarized, catppuccin-mocha, ink-wash, instagram, western, pixel
- **8 accent colors** — 112 visual combinations out of the box
- **12 slide layouts** — from simple title slides to multi-column and grid-cards
- **Interactive components** — charts, diagrams, sortable tables, tabs, sliders, toggles, expandable cards
- **API dashboards** — live data from backend APIs with auto-refresh and proxy support
- **Static export** — `goslide build` produces a single self-contained HTML (~7MB), works offline
- **Host mode** — serve a directory as a presentation library with index page
- **Speaker view** — press S for timer, notes, and next slide preview
- **Live reload** — edit your .md, browser auto-refreshes and keeps your slide position

## 📦 Downloads

| Platform | File |
|----------|------|
| Windows (x64) | `goslide-windows-amd64.exe` |
| macOS (Intel) | `goslide-darwin-amd64` |
| macOS (Apple Silicon) | `goslide-darwin-arm64` |
| Linux (x64) | `goslide-linux-amd64` |
| Linux (ARM64) | `goslide-linux-arm64` |

## 🚀 Quick Start

```bash
# Create a presentation
goslide init

# Serve with live reload
goslide serve talk.md

# Export as standalone HTML
goslide build talk.md

# Host multiple presentations
goslide host ./slides
```

## 📋 Full Feature List

### Rendering
- Markdown → Reveal.js slides with `---` separators
- YAML frontmatter for presentation config
- 14 themes × 8 accent colors
- 12 layout templates (two-column, code-preview, grid-cards, etc.)
- Fragment animations (fade-in, fade-up, highlight-current)
- Custom slide transitions (slide, fade, perspective, flip, convex, concave, zoom)
- CJK support with bundled Noto Sans TC font

### Components
- **Charts**: bar, line, pie, radar, sparkline (Chart.js)
- **Diagrams**: Mermaid.js (flowcharts, sequence, ERD)
- **Tables**: sortable with click-to-sort headers
- **Tabs + Panels**: tabbed content switching
- **Slider**: range input with live value display
- **Toggle**: switch control with panel visibility binding
- **Expandable Cards**: grid layout with click-to-expand detail overlay
- **API Component**: fetch from proxied APIs with 7 render types (metric, chart, table, json, log, image, markdown)
- **Embed HTML**: raw HTML/CSS/JS execution
- **Embed Iframe**: embedded external pages

### Infrastructure
- Single binary with all assets embedded (~8MB)
- Live reload via WebSocket with slide position preservation
- API reverse proxy with header injection and env var expansion
- Custom theme overrides via goslide.yaml
- Static HTML export (single file, works offline)
- Host mode with index page and directory watching
- Speaker view with notes, timer, next slide preview
- Lightweight presenter sync (viewers see presenter's slide + jump button)
- Progress bar with click-to-jump

### CLI
- `goslide serve` — single file with live reload
- `goslide host` — directory with index page
- `goslide build` — static HTML export
- `goslide init` — scaffold from templates (basic, demo, corporate)
- `goslide list` — list presentations with metadata
