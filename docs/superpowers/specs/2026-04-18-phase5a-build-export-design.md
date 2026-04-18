# GoSlide Phase 5a — Static HTML Export Design Spec

> **Status: DRAFT**

## Decision Log

| # | Question | Decision |
|---|----------|----------|
| 1 | Export mode | Single-file only (all assets inline, ~6-7MB) |
| 2 | API/WS handling | API components left empty, WS removed via static mode flag |
| 3 | JS strategy | Keep all JS, add `data-mode="static"` to skip WS/API at runtime |

---

## 1. CLI

```
goslide build <file.md> [flags]

Flags:
  -o, --output string    Output file (default: {name}.html)
```

- `goslide build talk.md` → produces `talk.html`
- `goslide build talk.md -o slides.html` → produces `slides.html`
- Remove build stub from `root.go`

---

## 2. Build Flow

```
1. Read .md file
2. parser.Parse → ir.Presentation
3. Apply CLI overrides (theme/accent if flags added later)
4. pres.Validate() — warnings printed, errors abort
5. renderer.Render(pres) → HTML string
6. inlineAssets(html) → replace <link>/<script src> with inline content
7. inlineFonts(css) → replace url('/fonts/...') with base64 data URIs
8. Add data-mode="static" to <body> tag
9. Write output file
```

---

## 3. Asset Inlining

### `inlineAssets(html string) string`

Reads assets from `web.StaticFS` and `web.ThemeFS` (go:embed). Replacements:

| Original | Replacement |
|----------|-------------|
| `<link rel="stylesheet" href="/themes/tokens.css">` | `<style>{tokens.css content}</style>` |
| `<link rel="stylesheet" href="/themes/{theme}.css">` | `<style>{theme CSS content}</style>` |
| `<link rel="stylesheet" href="/themes/layouts.css">` | `<style>{layouts CSS content}</style>` |
| `<script src="/static/chartjs/chart.min.js"></script>` | `<script>{chart.min.js content}</script>` |
| `<script src="/static/mermaid/mermaid.min.js"></script>` | `<script>{mermaid.min.js content}</script>` |
| `<script src="/static/reveal/reveal.js"></script>` | `<script>{reveal.js content}</script>` |
| `<script src="/static/runtime.js"></script>` | `<script>{runtime.js content}</script>` |
| `<script src="/static/reactive.js"></script>` | `<script>{reactive.js content}</script>` |
| `<script src="/static/components.js"></script>` | `<script>{components.js content}</script>` |

Implementation: `strings.Replace` for each known path. Read file content from embed.FS.

### `inlineFonts(css string) string`

In the CSS content (after inlining tokens.css), replace font URLs:

```
url('/fonts/NotoSansTC-Regular.woff2') → url('data:font/woff2;base64,...')
url('/fonts/NotoSansTC-Bold.woff2') → url('data:font/woff2;base64,...')
url('/fonts/JetBrainsMono-Regular.woff2') → url('data:font/woff2;base64,...')
```

Read font files from `web.StaticFS`, base64 encode, replace.

---

## 4. Static Mode

### `<body>` tag modification

```go
html = strings.Replace(html, "<body ", "<body data-mode=\"static\" ", 1)
```

### runtime.js changes

Add at the top of the IIFE, after `'use strict'`:

```javascript
var isStatic = document.body.dataset.mode === 'static';
```

Wrap WS connection block in `if (!isStatic) { ... }`.
Wrap presenter tracking block in `if (!isStatic) { ... }`.

Fragments, page number, and session storage restore remain active.

### components.js changes

In `initAllComponents`, before the api branch:

```javascript
if (type === 'api') {
    if (isStatic) {
        el.innerHTML = '<div class="goslide-api-static" style="color:var(--slide-muted);font-size:0.75em;text-align:center;padding:1rem;">API data requires goslide serve</div>';
    } else {
        initApiComponent(el);
    }
    initialized[id] = true;
    return;
}
```

Add `var isStatic = document.body.dataset.mode === 'static';` at the top of the IIFE.

Chart, table, mermaid, iframe init unchanged — they work with static data-params.

API polling slidechanged handler: wrap in `if (!isStatic)`.

---

## 5. Testing

### Go unit tests (`internal/builder/builder_test.go`)

| Test | Assertion |
|------|-----------|
| Build basic .md | Output file exists, contains inline CSS/JS |
| No external references | Output does not contain `<link href="/` or `<script src="/` |
| Fonts inlined | Output contains `data:font/woff2;base64` |
| Static mode flag | Output contains `data-mode="static"` |
| Reveal.js present | Output contains `Reveal.initialize` |
| Slides rendered | Output contains slide content from .md |

### Manual checklist

```markdown
## Static Export (Phase 5a)
- [ ] `goslide build examples/demo.md` produces demo.html
- [ ] demo.html opens in browser offline (no server needed)
- [ ] All slides render correctly
- [ ] Charts display with correct data
- [ ] Mermaid diagrams render as SVG
- [ ] Tables are sortable
- [ ] Tabs/slider/toggle work
- [ ] Card overlay works
- [ ] Fragments work
- [ ] Page numbers display
- [ ] No WS connection attempt (no console errors)
- [ ] API slides show "requires goslide serve" message
- [ ] Keyboard navigation works
- [ ] Dark theme: `goslide build examples/demo.md --theme dark` (if theme flag added)
```

---

## 6. Files Changed Summary

| Action | File |
|--------|------|
| Create | `internal/builder/builder.go` — Build + inlineAssets + inlineFonts |
| Create | `internal/builder/builder_test.go` |
| Create | `internal/cli/build.go` — build command |
| Modify | `internal/cli/root.go` — remove build stub |
| Modify | `web/static/runtime.js` — add isStatic check around WS/presenter |
| Modify | `web/static/components.js` — add isStatic check around api init/polling |
| Modify | `MANUAL_CHECKLIST.md` — Phase 5a checklist |
