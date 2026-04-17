# GoSlide Phase 2b — L1 Component Rendering Design Spec

> **Status: DRAFT**

## Decision Log

| # | Question | Decision |
|---|----------|----------|
| 1 | Render strategy | Client-side: Go outputs `data-type` + `data-params` divs, JS lazy-inits on slidechanged |
| 2 | Chart scope | All 5 types at once (bar, line, pie, radar, sparkline) |
| 3 | Mermaid timing | Render all mermaid at page load (static SVG); charts lazy init |
| 4 | Table sortable | Native JS ~30 lines, auto-detect number/string sort |
| 5 | Vendor strategy | Chart.js + Mermaid unconditionally embedded (binary ~6-7MB) |
| 6 | Frontend files | New `components.js` for component init; `runtime.js` unchanged |

---

## 1. Go Renderer — Placeholder Replacement

### New file: `internal/renderer/components.go`

```go
func renderComponents(html string, slide ir.Slide) string
```

For each `Component` in `slide.Components`, replaces `<!--goslide:component:N-->` with:

```html
<div class="goslide-component"
     data-type="chart:bar"
     data-params='{"title":"Yield","labels":["A","B"],"data":[96,93]}'
     data-raw=""
     data-comp-id="s1-c0">
</div>
```

Special handling:
- **mermaid**: `data-raw` contains the original mermaid source text, `data-params` is empty `{}`
- **Other types**: `data-params` contains JSON-marshaled `Component.Params`, `data-raw` is empty
- **JSON escaping**: `'` → `&#39;`, `<` → `&lt;`, `>` → `&gt;` in attribute values to prevent HTML breakage
- **`data-comp-id`**: `s{slideIndex}-c{compIndex}` for dedup tracking in JS

### Integration into `Render()`

In `renderer.go`, before `tmpl.Execute`, iterate over `pres.Slides` and replace placeholders in each slide's `BodyHTML` and `Regions[].HTML`:

```go
for i := range pres.Slides {
    slide := &pres.Slides[i]
    slide.BodyHTML = template.HTML(renderComponents(string(slide.BodyHTML), *slide))
    for j := range slide.Regions {
        slide.Regions[j].HTML = template.HTML(renderComponents(string(slide.Regions[j].HTML), *slide))
    }
}
```

Note: This mutates the slides before template execution. The original `Presentation` should be deep-copied or the mutation should be acceptable (render is the final step).

---

## 2. Frontend — `components.js`

### Architecture

Single IIFE (~150 lines) with:

1. **`initAllMermaid()`** — called once on `Reveal.on('ready')`, renders all `[data-type="mermaid"]` elements
2. **`initSlideComponents(slideEl)`** — called on `ready` + `slidechanged`, inits chart/table for current slide only
3. **`buildChartConfig(type, params)`** — translates GoSlide YAML params to Chart.js config
4. **`initTable(el)`** — builds `<table>` from params, attaches sort handlers if `sortable: true`
5. **`initialized`** map — tracks `data-comp-id` to prevent double-init

### Chart.js Config Mapping

| GoSlide param | Chart.js mapping |
|---|---|
| `title` | `options.plugins.title.display: true, text: title` |
| `labels` | `data.labels` |
| `data` (number[]) | `data.datasets[0].data` |
| `datasets` (object[]) | `data.datasets` with per-dataset `label`, `data`, `backgroundColor` |
| `unit` | `options.plugins.tooltip.callbacks.label` appends unit suffix |
| `color` | maps accent name to CSS variable value for `backgroundColor`/`borderColor` |
| `stacked` | `options.scales.x.stacked: true, y.stacked: true` |

### Sparkline Preset

When `type === 'sparkline'`:
```javascript
config.type = 'line';
config.options.scales = { x: { display: false }, y: { display: false } };
config.options.plugins.legend = { display: false };
config.options.plugins.title = { display: false };
config.options.elements = { point: { radius: 0 } };
config.options.responsive = true;
config.options.maintainAspectRatio = false;
```

### Chart Color Resolution

Charts need to read the current accent color from CSS. The `color` param in YAML maps to an accent name. Resolution:

```javascript
function resolveColor(colorName) {
    var root = document.documentElement;
    var varName = '--accent-' + (colorName || 'blue');
    return getComputedStyle(root).getPropertyValue(varName).trim() 
           || getComputedStyle(root).getPropertyValue('--slide-accent').trim();
}
```

For multi-dataset charts, each dataset can specify its own `color`. If not specified, use a built-in palette derived from the 8 accent colors.

### Mermaid Integration

```javascript
function initAllMermaid() {
    var els = document.querySelectorAll('.goslide-component[data-type="mermaid"]');
    if (els.length === 0) return;

    els.forEach(function(el) {
        var raw = el.getAttribute('data-raw');
        el.innerHTML = '<div class="mermaid">' + raw + '</div>';
    });

    var bg = getComputedStyle(document.querySelector('.reveal')).backgroundColor;
    var isDark = parseBrightness(bg) < 128;
    mermaid.initialize({ startOnLoad: false, theme: isDark ? 'dark' : 'default' });
    mermaid.run({ nodes: document.querySelectorAll('.mermaid') });
}

function parseBrightness(rgb) {
    var m = rgb.match(/\d+/g);
    if (!m) return 255;
    return (parseInt(m[0]) * 299 + parseInt(m[1]) * 587 + parseInt(m[2]) * 114) / 1000;
}
```

### Table Component

`initTable(el)` builds a table from `data-params`:

```javascript
function initTable(el) {
    var params = JSON.parse(el.getAttribute('data-params'));
    var table = document.createElement('table');
    
    // Build thead
    var thead = document.createElement('thead');
    var headerRow = document.createElement('tr');
    (params.columns || []).forEach(function(col) {
        var th = document.createElement('th');
        th.textContent = col;
        headerRow.appendChild(th);
    });
    thead.appendChild(headerRow);
    table.appendChild(thead);
    
    // Build tbody
    var tbody = document.createElement('tbody');
    (params.rows || []).forEach(function(row) {
        var tr = document.createElement('tr');
        row.forEach(function(cell) {
            var td = document.createElement('td');
            td.textContent = String(cell);
            tr.appendChild(td);
        });
        tbody.appendChild(tr);
    });
    table.appendChild(tbody);
    el.appendChild(table);
    
    if (params.sortable) makeSortable(table);
}
```

### Sortable Logic (~30 lines)

```javascript
function makeSortable(table) {
    table.querySelectorAll('th').forEach(function(th, colIdx) {
        th.style.cursor = 'pointer';
        th.addEventListener('click', function() { sortTable(table, colIdx, th); });
    });
}

function sortTable(table, colIdx, th) {
    var tbody = table.querySelector('tbody');
    var rows = Array.from(tbody.querySelectorAll('tr'));
    var asc = th.getAttribute('data-sort') !== 'asc';
    
    rows.sort(function(a, b) {
        var aText = a.cells[colIdx].textContent.trim();
        var bText = b.cells[colIdx].textContent.trim();
        var aNum = parseFloat(aText), bNum = parseFloat(bText);
        if (!isNaN(aNum) && !isNaN(bNum)) return asc ? aNum - bNum : bNum - aNum;
        return asc ? aText.localeCompare(bText) : bText.localeCompare(aText);
    });
    
    rows.forEach(function(row) { tbody.appendChild(row); });
    table.querySelectorAll('th').forEach(function(h) {
        h.removeAttribute('data-sort');
        h.textContent = h.textContent.replace(/ [▲▼]$/, '');
    });
    th.setAttribute('data-sort', asc ? 'asc' : 'desc');
    th.textContent += asc ? ' ▲' : ' ▼';
}
```

### Script Load Order in `slide.html`

```html
<script src="/static/chartjs/chart.min.js"></script>
<script src="/static/mermaid/mermaid.min.js"></script>
<script src="/static/reveal/reveal.js"></script>
<script src="/static/runtime.js"></script>
<script src="/static/components.js"></script>
<script>
  Reveal.initialize({...});
</script>
```

Chart.js and Mermaid load first (as libraries), then Reveal, then runtime (fragments + reload), then components (registers Reveal event listeners), then Reveal.initialize fires events.

---

## 3. Vendor Updates

### `scripts/vendor.sh` additions

```bash
CHARTJS_VER="4.4.7"
MERMAID_VER="11.4.1"

mkdir -p "$VENDOR_DIR/chartjs"
mkdir -p "$VENDOR_DIR/mermaid"

download "https://cdn.jsdelivr.net/npm/chart.js@${CHARTJS_VER}/dist/chart.umd.min.js" \
         "$VENDOR_DIR/chartjs/chart.min.js"
download "https://cdn.jsdelivr.net/npm/mermaid@${MERMAID_VER}/dist/mermaid.min.js" \
         "$VENDOR_DIR/mermaid/mermaid.min.js"
```

### `web/static/VERSIONS.md` additions

| Asset | Version | License | Source |
|-------|---------|---------|--------|
| Chart.js | 4.4.7 | MIT | https://github.com/chartjs/Chart.js |
| Mermaid | 11.4.1 | MIT | https://github.com/mermaid-js/mermaid |

---

## 4. CSS Additions (`tokens.css`)

```css
.goslide-component {
  margin: 1rem 0;
}
.goslide-component canvas {
  max-width: 100%;
}
.goslide-component table {
  width: 100%;
  border-collapse: collapse;
}
.goslide-component table th {
  cursor: pointer;
  user-select: none;
}
.goslide-component table th[data-sort]::after {
  margin-left: 0.3em;
}
.goslide-component .mermaid {
  display: flex;
  justify-content: center;
}
.goslide-component .mermaid svg {
  max-width: 100%;
  height: auto;
}
```

---

## 5. Testing Strategy

### Go Unit Tests

| File | Tests |
|------|-------|
| `renderer/components_test.go` | `renderComponents` basic replacement; mermaid uses `data-raw`; JSON special chars escaped; no components = unchanged HTML; multiple components in one slide; component in region HTML |

### Golden Tests

New fixture `chart-component.md`: slide with `~~~chart:bar` → verify output has `<div class="goslide-component" data-type="chart:bar" ...>` and no raw placeholder.

### Manual Checklist Additions

- chart:bar/line/pie/radar/sparkline rendering + tooltip
- chart accent color matches theme
- mermaid SVG rendering + dark theme auto-detect
- table sortable click + arrow indicator + number sort
- table non-sortable (sortable: false)
- component lazy init (chart only draws on slide visit)
- live reload re-renders components

---

## 6. Files Changed Summary

| Action | File |
|--------|------|
| Create | `internal/renderer/components.go` |
| Create | `internal/renderer/components_test.go` |
| Modify | `internal/renderer/renderer.go` — call `renderComponents` before template execute |
| Modify | `scripts/vendor.sh` — add Chart.js + Mermaid downloads |
| Modify | `web/static/VERSIONS.md` — add entries |
| Create | `web/static/components.js` — chart/mermaid/table init |
| Modify | `web/templates/slide.html` — add script tags |
| Modify | `web/themes/tokens.css` — add `.goslide-component` styles |
| Create | `internal/renderer/testdata/golden/chart-component.md` |
| Modify | `MANUAL_CHECKLIST.md` — add component test items |
| Modify | `examples/demo.md` — add chart/mermaid/table demo slides |
