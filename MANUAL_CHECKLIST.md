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

## Reactive Components (Phase 2c)
- [ ] tabs — button bar displays, click switches active style
- [ ] tabs → panel — clicking tab shows/hides corresponding panel content
- [ ] slider — range input displays with label + live value
- [ ] slider accent — slider color matches accent
- [ ] slider value — dragging updates value display in real-time
- [ ] toggle — switch style, click toggles on/off
- [ ] toggle accent — on-state color matches accent
- [ ] toggle → panel — toggle on shows panel, off hides it
- [ ] dark theme — all L2 controls visible on dark theme
- [ ] console debug — browser console `GoSlide.get('threshold')` returns current value
- [ ] fragment coexistence — slide with both fragments + slider, no interference

## Expandable Cards (Phase 2d)
- [ ] grid-cards layout — cards in 2-column grid
- [ ] card summary — icon, title, desc displayed
- [ ] card hover — slight lift effect
- [ ] card click → overlay with detail content
- [ ] overlay close — click ✕ button
- [ ] overlay close — click backdrop
- [ ] overlay close — Esc key
- [ ] Esc precedence — closes overlay, not Reveal.js overview
- [ ] keyboard lock — arrows don't navigate while overlay open
- [ ] overlay dark theme — panel bg matches theme
- [ ] detail content — markdown rendered (headings, tables, lists)

## API Proxy (Phase 3a)
- [ ] goslide.yaml absent → starts normally without proxy
- [ ] goslide.yaml present → proxy routes work
- [ ] /api/mock/status → returns mock JSON (requires mock-api running)
- [ ] Header injection → upstream receives configured headers
- [ ] env var expansion → ${TOKEN} replaced with actual env var value
- [ ] embed:html → custom HTML/JS executes in slide (click button works)
- [ ] embed:iframe → iframe displays external page
- [ ] upstream unreachable → 502 response

## API Component (Phase 3b)
- [ ] api fetch — /api/mock/metrics returns data and renders
- [ ] render metric — large number card with value + unit
- [ ] render chart:bar — API data rendered as bar chart
- [ ] render table — object array auto-derives columns and rows
- [ ] render json — formatted JSON display
- [ ] render log — terminal style black/green text
- [ ] render image — base64 or URL image displays
- [ ] render markdown — plain text display
- [ ] polling — refresh: 5s re-fetches every 5 seconds
- [ ] polling pause — leave slide stops polling
- [ ] polling resume — return to slide resumes polling
- [ ] error — upstream unreachable shows red error message
- [ ] dark theme — all render types visible on dark theme

## Host Mode (Phase 4a)
- [ ] `goslide host examples/slides --no-open` starts on port 8080
- [ ] Index page at `/` lists both presentations with titles
- [ ] Click presentation link → renders slides at `/talks/{name}`
- [ ] `/talks/nonexist` → 404
- [ ] Static assets (themes, fonts) work in host mode
- [ ] Add new .md to directory → appears in index after reload
- [ ] Delete .md → removed from index after reload
- [ ] Modify .md → slides update after live reload
- [ ] goslide.yaml proxy works in host mode
