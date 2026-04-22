# 🎉 GoSlide v1.5.0

## What's New

### 📄 PDF export

New `goslide export-pdf <file.md>` command renders your deck to a PDF
via headless Chrome. Whatever you see in `goslide serve` — fonts,
themes, Chart.js, Mermaid diagrams, LLM-baked API results — renders
in the PDF, because the renderer is literally a real browser.

```bash
goslide export-pdf talk.md
goslide export-pdf talk.md -o handout.pdf
goslide export-pdf talk.md --notes                 # include speaker notes
goslide export-pdf talk.md --paper-size a4-landscape
```

**Paper sizes** (default `slide-16x9`):

| Preset | Dimensions | Intended use |
|--------|-----------|--------------|
| `slide-16x9` | 1920 × 1080 px | On-screen presentation look (default) |
| `slide-4x3`  | 1600 × 1200 px | Legacy projector aspect |
| `a4-landscape` | 297 × 210 mm | Print-friendly handout |
| `letter-landscape` | 11 × 8.5 in | US letter handout |

**Fragment animations** are collapsed to their final state — one slide,
one PDF page.

### 🧭 Chrome discovery: install-once, no bundled chromium

GoSlide stays a single ~8MB binary. `export-pdf` locates Chrome in
this order:

1. `GOSLIDE_CHROME_PATH` env var (explicit override)
2. PATH — searches `chrome`, `chromium`, `chromium-browser`,
   `google-chrome`, `microsoft-edge`
3. Platform-specific known install locations
   (`C:\Program Files\Google\Chrome\Application\chrome.exe`,
   `/Applications/Google Chrome.app/...`, etc.)

If nothing is found, the command exits non-zero with an actionable
message listing every location it checked. No auto-download.

### 🧩 How it works

Under the hood `export-pdf` is a thin wrapper:

1. Runs the existing `goslide build` to produce static HTML with
   everything (LLM bakes included) already inlined.
2. Launches Chrome headless against that HTML with reveal.js's
   `?print-pdf` mode.
3. Waits for `window.__goslideReady` (new front-end marker that fires
   once Mermaid promises settle — async-safe rendering).
4. Calls Chrome DevTools `Page.printToPDF` with the resolved paper
   dimensions and writes the bytes to the output path.

The `Launcher` interface keeps unit tests deterministic (fake launcher
for orchestrator tests) and the real `ChromedpLauncher` only runs when
Chrome is actually present. The integration test auto-skips when
`FindChrome()` errors, so CI without Chrome stays green.

## Compatibility

No breaking changes. v1.4.0 decks export unchanged; the new command is
additive. Existing `goslide build` / `serve` / `generate` / `host` are
untouched.

**Go version:** still 1.21.6. Chromedp is pinned to v0.10.0 — the last
release compatible with Go 1.21.

## Requirements

- **Chrome / Edge / Chromium installed locally.** Developers typically
  have one of these; if not, any distribution works. Set
  `GOSLIDE_CHROME_PATH` if you want to pick a specific binary.

## Out of scope (future work)

- Fragment-per-page mode.
- Custom header/footer, page numbers, bookmarks.
- Password-protected PDFs.
- Auto-downloading or bundling Chromium.

## Full Changelog

See [v1.4.0...v1.5.0](https://github.com/GMfatcat/goslide/compare/v1.4.0...v1.5.0) for all changes.
