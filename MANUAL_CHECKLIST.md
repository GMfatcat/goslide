# GoSlide Phase 1 Manual Checklist

Run: `go run ./cmd/goslide serve examples/demo.md`

## Rendering
- [ ] Default theme: white background, dark text, Noto Sans TC loaded (DevTools Network: 200)
- [ ] `--theme dark`: dark background, light text
- [ ] `--accent teal`: links and active elements use teal color
- [ ] Layout: default — centered content
- [ ] Layout: title — large centered title
- [ ] Layout: section — centered with accent underline
- [ ] Layout: two-column — two equal columns
- [ ] Layout: code-preview — code left, preview right
- [ ] Code blocks use JetBrains Mono
- [ ] CJK characters render correctly

## Navigation
- [ ] Right arrow / Space / Enter → next slide
- [ ] Left arrow / Backspace → previous slide
- [ ] Esc → overview mode (thumbnail grid), Esc again exits
- [ ] F → fullscreen toggle
- [ ] Home → first slide
- [ ] End → last slide
- [ ] G + number + Enter → jump to slide N

## Fragments
- [ ] Slide with `fragments: true`: list items appear one at a time
- [ ] `highlight-current`: previous items dim to 40%

## Live Reload
- [ ] Edit demo.md text → browser auto-reloads, stays on same slide
- [ ] Change frontmatter theme → reload switches theme
- [ ] Introduce YAML error → red toast appears, previous version still usable
- [ ] Fix error → toast disappears, slide reloads

## CLI
- [ ] `goslide serve` (no args) → shows usage, exit code 1
- [ ] `goslide serve nonexistent.md` → error message, exit code 1
- [ ] `goslide --version` → prints version
- [ ] `goslide host .` → "not implemented" message

## Cross-Platform
- [ ] Windows: CRLF line endings in .md parse correctly
- [ ] Build produces working binary: `go build -o goslide.exe ./cmd/goslide`

## Components (Phase 2b)
- [ ] chart:bar — bar chart displays, hover shows tooltip with unit
- [ ] chart:line — line chart with data points
- [ ] chart:pie — pie chart with legend
- [ ] chart:radar — radar chart (test by changing demo type to radar)
- [ ] chart:sparkline — mini line chart, no axes, no legend (test by changing demo type to sparkline)
- [ ] chart accent color — chart colors match specified or default accent
- [ ] mermaid — flowchart renders as SVG, not raw text
- [ ] mermaid dark theme — switch to `--theme dark`, mermaid auto-uses dark palette
- [ ] table sortable — click header to sort, arrow indicator appears
- [ ] table number sort — numeric column sorts by value not string
- [ ] table non-sortable — remove `sortable: true`, headers not clickable
- [ ] component lazy init — navigate to chart slide, verify canvas appears on arrival
- [ ] live reload — edit chart YAML, save, chart re-renders after reload
