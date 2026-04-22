# PDF Export Design

**Date:** 2026-04-22
**Status:** Design approved, ready for implementation plan
**Scope:** Phase 7b — PDF export via headless Chrome

---

## 1. Overview

Add a new `goslide export-pdf` command that produces a PDF of the
presentation by:

1. Running the existing `build` pipeline to produce a static HTML.
2. Launching Chrome / Edge / Chromium in headless mode against the
   built HTML with the reveal.js `?print-pdf` query.
3. Calling Chrome DevTools Protocol `Page.printToPDF` to get PDF bytes.
4. Writing the PDF to the requested output path.

Fidelity matches what the viewer sees in a browser — fonts, themes,
Chart.js, Mermaid, LLM-baked API results, everything — because the
renderer is literally a real browser. No native PDF synthesis.

**Design principles:**

- **Reuse `build`, don't reimplement.** Runtime components (LLM
  transformer, reactive, Chart.js) are already rendered to static
  inert HTML by `goslide build`. PDF export consumes that artefact.
- **Require Chrome; fail fast.** Look in PATH and standard install
  locations. If not found, print an actionable message and exit with
  non-zero status. No auto-download.
- **No bundled Chromium.** Keeping the single-binary ~8MB story intact.

---

## 2. CLI Surface

### 2.1 New command

```
goslide export-pdf <file.md> [flags]
```

### 2.2 Flags

| Flag | Default | Purpose |
|------|---------|---------|
| `-o, --output <path>` | `<name>.pdf` (derived from input `<name>.md`) | Output PDF path |
| `--paper-size <size>` | `slide-16x9` | Paper size preset (see §2.3) |
| `--notes` | off | Include speaker notes below each slide |
| `-t, --theme <name>` | — | Same override semantics as `goslide build` |
| `-a, --accent <name>` | — | Same override semantics as `goslide build` |

Fragments are always collapsed (one slide → one page). Animation
intent is preserved by reveal.js's own fragment-flattening logic —
no flag for per-fragment pages in the MVP.

### 2.3 Paper sizes

| Preset | Dimensions | Ratio | Notes |
|--------|-----------|-------|-------|
| `slide-16x9` (default) | 1920 × 1080 px | 16:9 | Matches on-screen presentation |
| `slide-4x3` | 1600 × 1200 px | 4:3 | Legacy projector aspect |
| `a4-landscape` | 297 × 210 mm | 1.414 | Printed "notes handout" look |
| `letter-landscape` | 11 × 8.5 in | 1.294 | US letter |

Unknown preset → error listing valid options.

---

## 3. Architecture

### 3.1 Package layout

```
internal/
├── pdfexport/                 # NEW package
│   ├── chrome.go              # FindChrome() (PATH + known install paths)
│   ├── chrome_test.go
│   ├── export.go              # Export(opts Options) error — orchestrator
│   ├── export_test.go
│   ├── papersize.go           # paper-size preset table + resolution
│   └── papersize_test.go
└── cli/
    └── export_pdf.go          # Cobra command, flag parsing
```

New Go dependency: `github.com/chromedp/chromedp` (chromium DevTools
Protocol client). Needs Go 1.21 compatibility — verified at plan time.

### 3.2 Dependency direction

```
cli/export_pdf.go
  ↓
pdfexport.Export
  ├── pdfexport.FindChrome  (returns path or error)
  ├── builder.Build         (writes static HTML to a temp file)
  └── chromedp driving headless browser → PDF bytes → os.WriteFile
```

No cycles. `internal/pdfexport` depends only on `internal/builder` and
the chromedp library.

### 3.3 Chrome discovery

`FindChrome()` tries, in order:

1. `GOSLIDE_CHROME_PATH` env var (user override)
2. PATH search for: `chrome`, `chromium`, `chromium-browser`, `google-chrome`, `microsoft-edge`
3. Platform-specific known paths:
   - **Windows:** `C:\Program Files\Google\Chrome\Application\chrome.exe`, `C:\Program Files (x86)\Google\Chrome\Application\chrome.exe`, `C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe`, `%LOCALAPPDATA%\Google\Chrome\Application\chrome.exe`
   - **macOS:** `/Applications/Google Chrome.app/Contents/MacOS/Google Chrome`, `/Applications/Chromium.app/Contents/MacOS/Chromium`, `/Applications/Microsoft Edge.app/Contents/MacOS/Microsoft Edge`
   - **Linux:** already covered by PATH search

Returns `(path string, err error)`. Error on not-found contains the
checked locations and installation guidance.

### 3.4 Export pipeline

1. Resolve and validate `Options` (paper-size preset → pixel dimensions).
2. Call `FindChrome()`; return that error if nothing found.
3. Call `builder.Build` with a temp directory output — produces a
   static HTML alongside its assets. (Static export is already a
   single-file HTML with everything inlined — simplifies URL loading.)
4. Start chromedp with the located binary. Use headless mode.
5. Navigate to `file://<temp-html>?print-pdf` plus `&showNotes=true`
   when `--notes` is set.
6. Wait for a `ready` signal. Reveal.js's print-pdf mode renders all
   slides synchronously to DOM, but Chart.js / Mermaid need a moment.
   Use chromedp's `WaitReady` on a known-present selector (e.g.
   `.reveal .slides > section:last-child`) plus a short settled delay
   (~500ms after idle network).
7. Call `page.PrintToPDF` with:
   - `PaperWidth`, `PaperHeight` (in inches — chromedp takes inches)
   - `PrintBackground: true`
   - `PreferCSSPageSize: true` (let reveal.js's own CSS `@page` rules win)
   - `MarginTop/Bottom/Left/Right: 0`
8. `os.WriteFile` the returned bytes to `opts.Output`.
9. Cleanup temp HTML (unless `--verbose` — then print its path).

### 3.5 Timing and correctness

Reveal.js `?print-pdf` mode injects its own page-break CSS and
synchronously lays out every slide. Chart.js renders from the
component.js init loop, which runs on `DOMContentLoaded`. Mermaid
renders async (returns promises). To catch everything:

- `chromedp.WaitReady(".reveal .slides section:last-of-type")` — DOM
  ready.
- `chromedp.Sleep(500 * time.Millisecond)` — paint settle + async
  mermaid render.
- Add a JS-evaluated marker: on-page script sets
  `window.__goslideReady = true` when it's done initialising
  components; chromedp polls `window.__goslideReady === true`.

The last approach is most reliable and deterministic; it's part of
this task's implementation (small addition to `web/static/runtime.js`
or similar).

---

## 4. Error handling

| Condition | Behaviour | Exit |
|-----------|-----------|------|
| Chrome not found | Print list of searched locations + "install Chrome/Edge/Chromium and retry, or set GOSLIDE_CHROME_PATH to an explicit binary" | 1 |
| Unknown `--paper-size` value | List valid presets, exit | 1 |
| `builder.Build` fails (validation, render) | Forward builder's error message | 1 |
| Chrome launch fails (e.g. profile locked, crashed) | chromedp error + "try closing other Chrome instances" | 1 |
| `PrintToPDF` returns empty bytes | Error "Chrome produced empty PDF; re-run with --verbose for diagnostics" | 1 |
| Output file write fails | OS-level error + path | 1 |

Temp HTML is cleaned up on both success and failure paths (`defer`).

---

## 5. Testing

### 5.1 Unit

- `papersize_test.go` — preset resolution: known → dimensions,
  unknown → error with valid list.
- `chrome_test.go` — FindChrome() with mocked filesystem / PATH:
  env var wins, PATH next, then platform paths, then error.

### 5.2 Integration (skipped when Chrome absent)

- `export_test.go` — full `Export()` pipeline with a small fixture
  deck (`testdata/fixture.md`). Test uses `t.Skip` when `FindChrome()`
  returns an error — CI without Chrome just skips. Locally + in the
  GitHub Actions linux runner (which has Chrome pre-installed on
  `ubuntu-latest`) it runs end-to-end:
  - Export the fixture
  - Open the produced PDF
  - Assert it's valid PDF (starts with `%PDF-`)
  - Assert page count matches slide count
  - Assert file size is reasonable (> 10KB, < 50MB)

PDF content-level assertions (text extraction, font matching) are
out of scope — the integration test treats PDF as opaque beyond
format validity + page count.

### 5.3 CLI

- `cli/export_pdf_test.go` — flag parsing, error on missing input
  file, `--help` text sanity.

### 5.4 No new LLM / network calls in any test.

---

## 6. Out of scope

- Fragment-per-page mode.
- Custom header/footer, page numbers.
- Embedded fonts override (browser's choice stands).
- PDF bookmarks / outline.
- Password protection / encryption.
- Auto-download Chromium.
- Bundled Chromium binary.
- Mobile / iOS Safari rendering.

Any of these can be future features once there's real demand. The MVP
covers "make a reasonable PDF of my deck for distribution".

---

## 7. Success criteria

- `goslide export-pdf talk.md` produces `talk.pdf` that opens in any
  PDF viewer, renders the same theme colours / fonts / charts that
  `goslide serve` shows.
- One page per slide; fragments collapsed to their final state.
- `--notes` appends speaker notes beneath each slide.
- On a machine without Chrome, command exits non-zero with an
  actionable message.
- Existing `goslide build` / `serve` are unchanged; no regressions in
  any existing test package.
- No new Go dependencies beyond `github.com/chromedp/chromedp` and
  its transitive deps.

---

## 8. Release note snippet (draft for v1.5.0)

```
### 📄 PDF export

New `goslide export-pdf talk.md` command renders your deck to a PDF
via headless Chrome. Charts, Mermaid, themes, LLM-baked API results
— whatever renders in the browser renders in the PDF.

Requires Chrome / Edge / Chromium installed locally (set
GOSLIDE_CHROME_PATH to override auto-discovery). No bundled Chromium;
GoSlide stays a single ~8MB binary.

Flags: --paper-size (slide-16x9 default, also slide-4x3, a4-landscape,
letter-landscape), --notes to include speaker notes, -o/--output for
the destination path.
```
