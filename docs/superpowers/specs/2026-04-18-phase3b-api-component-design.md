# GoSlide Phase 3b — API Component + Render Types Design Spec

> **Status: COMPLETED (2026-04-18)**
>
> Key learnings:
> - Phase 2b subagent's TrimLeft on YAML content broke nested indentation — removed
> - Full-width render types (log/json/table) need explicit `width: 100%` to override flex centering

## Decision Log

| # | Question | Decision |
|---|----------|----------|
| 1 | Fetch location | Client-side fetch via proxy (JS `fetch('/api/...')`) |
| 2 | Render type scope | All 7 types at once; markdown displays as plain text `<pre>` |
| 3 | Path extraction | Simple dot notation + array index (`data.lines[0].yield`) |
| 4 | Polling + cache | Client-side `setInterval` polling; server cache deferred |

---

## 1. Data Flow

### Markdown syntax

```yaml
~~~api
url: /api/mock/metrics
method: GET
refresh: 5s
render:
  - type: metric
    path: lines[0].yield
    label: Yield
    unit: "%"
    color: green
  - type: chart:bar
    path: lines
    title: Yield by Line
~~~
```

### Go side

No changes needed. `api` is an already-recognized component type. `extractComponents` parses the YAML into `Component.Params` containing `url`, `method`, `refresh`, `render[]`. The `data-params` attribute carries everything to the frontend.

### Frontend flow

```
Reveal.on('ready') → initAllComponents() →
  [data-type="api"] → initApiComponent(el) →
    parse data-params → { url, method, refresh, render[] }
    fetchAndRender():
      fetch(url, {method}) → JSON response →
      for each render item:
        extractPath(response, item.path) → data
        renderItem(container, item, data)
    if refresh: setInterval(fetchAndRender, refreshMs)
```

### Polling lifecycle

- On `ready`: init all api components, start polling for those on current slide
- On `slidechanged`: pause polling on non-visible slides, resume on current slide
- Polling function stored as `el._fetchFn`, interval ID as `el._pollInterval`

---

## 2. Path Extraction

Simple dot notation parser (~20 lines JS):

```javascript
function extractPath(obj, path) {
  if (!path) return obj;
  var parts = path.replace(/\[(\d+)\]/g, '.$1').split('.');
  var current = obj;
  for (var i = 0; i < parts.length; i++) {
    if (current == null) return undefined;
    current = current[parts[i]];
  }
  return current;
}
```

Examples:
- `yield` → `obj.yield`
- `lines[0].name` → `obj.lines[0].name`
- `data.items` → `obj.data.items`
- (empty) → entire response object

---

## 3. Render Types

All render items are placed inside a `<div class="goslide-api-items">` container with flex-wrap layout.

### `metric`

Large number card. Extracts single value via `path`.

```html
<div class="goslide-metric">
  <div class="goslide-metric-value" style="color: {color}">96.2%</div>
  <div class="goslide-metric-label">Yield</div>
</div>
```

### `chart:*`

Reuses Phase 2b `buildChartConfig`. Auto-detects labels from `object[]`:
- If data is `object[]`, first string key → labels, first number key → data values
- If data is `number[]`, uses directly with render item's labels

### `table`

Reuses Phase 2b table builder. Auto-derives columns from `object[]` keys. Optional `columns` field to select/order specific fields.

### `json`

```html
<pre class="goslide-json">JSON.stringify(data, null, 2)</pre>
```

Dark background, mono font, scrollable, max-height 300px.

### `log`

Terminal-style scrolling text. Black background, green mono text.

```html
<pre class="goslide-log">[2026-04-18 10:00:01] System startup\n...</pre>
```

Data can be `string` or `string[]` (joined with `\n`).

### `image`

Renders `<img>`. Data is URL or base64 string. Auto-detects: if starts with `data:` or `http` → use as `src`; otherwise → prepend `data:image/png;base64,`.

### `markdown`

Displays raw markdown as plain text in `<pre>`. No frontend markdown parser (would require new vendor dependency).

---

## 4. Polling Control

### Start

```javascript
function initApiComponent(el) {
  var params = JSON.parse(decodeAttr(el.getAttribute('data-params')));
  var refreshMs = parseRefresh(params.refresh);

  el._fetchFn = function() { fetchAndRender(el, params); };
  el._fetchFn();

  if (refreshMs > 0) {
    el._refreshMs = refreshMs;
    if (el.closest('section') === Reveal.getCurrentSlide()) {
      el._pollInterval = setInterval(el._fetchFn, refreshMs);
    }
  }
}
```

### Pause/Resume on slide change

```javascript
Reveal.on('slidechanged', function() {
  document.querySelectorAll('.goslide-component[data-type="api"]').forEach(function(el) {
    var isVisible = el.closest('section') === Reveal.getCurrentSlide();
    if (!isVisible && el._pollInterval) {
      clearInterval(el._pollInterval);
      el._pollInterval = null;
    } else if (isVisible && el._refreshMs && !el._pollInterval) {
      el._fetchFn();
      el._pollInterval = setInterval(el._fetchFn, el._refreshMs);
    }
  });
});
```

### Refresh parsing

```javascript
function parseRefresh(s) {
  if (!s) return 0;
  var m = String(s).match(/^(\d+)(s|ms)?$/);
  if (!m) return 0;
  var val = parseInt(m[1]);
  if (m[2] === 'ms') return val;
  return val * 1000;
}
```

---

## 5. Error Handling

- `fetch` fails → display `<pre class="goslide-api-error">API error: {message}</pre>` inside the component
- On next successful fetch (during polling), error is replaced with rendered content
- `extractPath` returns `undefined` → render item is skipped silently

---

## 6. CSS

```css
.goslide-api-error {
  color: var(--accent-red);
  font-family: var(--font-mono);
  font-size: 0.75em;
}
.goslide-api-items {
  display: flex;
  flex-wrap: wrap;
  gap: 1rem;
  justify-content: center;
}
.goslide-metric {
  text-align: center;
  padding: 0.8rem 1.2rem;
  background: var(--slide-card-bg);
  border-radius: 0.75rem;
  border: 1px solid var(--slide-border, rgba(0,0,0,0.1));
  min-width: 8rem;
}
.goslide-metric-value {
  font-size: 2em;
  font-weight: 700;
  font-family: var(--font-mono);
}
.goslide-metric-label {
  font-size: 0.7em;
  color: var(--slide-muted);
  margin-top: 0.3em;
}
.goslide-json {
  background: var(--slide-code-bg);
  color: var(--slide-code-text);
  font-family: var(--font-mono);
  font-size: 0.65em;
  padding: 1rem;
  border-radius: 0.5rem;
  overflow: auto;
  max-height: 300px;
}
.goslide-log {
  background: #0a0a0a;
  color: #00ff00;
  font-family: var(--font-mono);
  font-size: 0.65em;
  padding: 1rem;
  border-radius: 0.5rem;
  overflow-y: auto;
  max-height: 300px;
  white-space: pre-wrap;
}
.goslide-markdown-raw {
  background: var(--slide-card-bg);
  font-family: var(--font-mono);
  font-size: 0.7em;
  padding: 1rem;
  border-radius: 0.5rem;
  white-space: pre-wrap;
}
```

---

## 7. Mock API Expansion

Add `/log` endpoint to `examples/mock-api/main.go`:

```go
http.HandleFunc("/log", func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]any{
        "entries": []string{
            "[2026-04-18 10:00:01] System startup",
            "[2026-04-18 10:00:02] Camera initialized",
            "[2026-04-18 10:00:03] Model loaded: SegFormer v2",
            "[2026-04-18 10:00:05] First inspection completed",
        },
    })
})
```

---

## 8. JS File Strategy

All api component code goes in `components.js` (same file as chart/table). Reasons:
- Reuses `buildChartConfig`, table builder, `decodeAttr`, `resolveColor`
- `components.js` grows from ~210 lines to ~360 lines — acceptable for a single-responsibility file (L1 + API component rendering)

---

## 9. Testing

### Go side

No new Go tests needed. `api` component uses the same pipeline as all other components.

### Manual checklist

```markdown
## API Component (Phase 3b)
- [ ] api fetch — /api/mock/metrics returns data and renders
- [ ] render metric — large number card with value + unit
- [ ] render chart:bar — API data rendered as bar chart
- [ ] render chart:pie — API data rendered as pie chart
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
```

---

## 10. Files Changed Summary

| Action | File |
|--------|------|
| Modify | `web/static/components.js` — api init, 7 render types, path extraction, polling |
| Modify | `web/themes/tokens.css` — metric/json/log/markdown/error CSS |
| Modify | `examples/mock-api/main.go` — add /log endpoint |
| Modify | `examples/demo.md` — api component demo slides |
| Modify | `MANUAL_CHECKLIST.md` — Phase 3b checklist |
