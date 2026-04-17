# Phase 2b: L1 Component Rendering Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Render Chart.js charts (bar/line/pie/radar/sparkline), Mermaid diagrams, and sortable tables from component data extracted by Phase 2a's parser, replacing `<!--goslide:component:N-->` placeholders with interactive HTML.

**Architecture:** Go renderer replaces placeholders with `<div class="goslide-component" data-type="..." data-params="...">` elements. A new `components.js` on the frontend initializes Chart.js on slide change (lazy), Mermaid at page load (eager), and builds sortable tables from JSON params.

**Tech Stack:** Go 1.21.6, Chart.js 4.4.7 (UMD), Mermaid 11.4.1, vanilla JS for table sorting. No new Go dependencies.

**Shell rules (Windows):** Never chain commands with `&&`. Use SEPARATE Bash calls for `git add`, `git commit`, `go test`, `go build`. Use `GOTOOLCHAIN=local` prefix for ALL go commands. Use `-C` flag for go/git to specify directory.

---

## Task 1: Vendor — Download Chart.js + Mermaid

**Files:**
- Modify: `scripts/vendor.sh`
- Modify: `web/static/VERSIONS.md`

- [ ] **Step 1: Update vendor.sh**

Read `scripts/vendor.sh`. Add these lines BEFORE the checksum section (before `if [ "${1:-}" = "--update-checksums" ]`):

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

- [ ] **Step 2: Run vendor script**

Run: `bash D:/CLAUDE-CODE-GOSLIDE/scripts/vendor.sh --update-checksums`

If any URL fails, try alternative CDN URLs. Verify files exist:
- `ls -la D:/CLAUDE-CODE-GOSLIDE/web/static/chartjs/`
- `ls -la D:/CLAUDE-CODE-GOSLIDE/web/static/mermaid/`

- [ ] **Step 3: Update VERSIONS.md**

Read `web/static/VERSIONS.md`. Append to the table:

```markdown
| Chart.js | 4.4.7 | MIT | https://github.com/chartjs/Chart.js |
| Mermaid | 11.4.1 | MIT | https://github.com/mermaid-js/mermaid |
```

- [ ] **Step 4: Verify embed compiles**

Run: `GOTOOLCHAIN=local go build -C D:/CLAUDE-CODE-GOSLIDE ./web/`

- [ ] **Step 5: Commit**

```
git -C D:/CLAUDE-CODE-GOSLIDE add scripts/vendor.sh web/static/chartjs/ web/static/mermaid/ web/static/VERSIONS.md web/static/CHECKSUMS.sha256
```
```
git -C D:/CLAUDE-CODE-GOSLIDE commit -m "chore: vendor Chart.js 4.4.7 and Mermaid 11.4.1"
```

---

## Task 2: Go Renderer — Component Placeholder Replacement

**Files:**
- Create: `internal/renderer/components.go`
- Create: `internal/renderer/components_test.go`

- [ ] **Step 1: Write failing tests**

Create `internal/renderer/components_test.go`:

```go
package renderer

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/user/goslide/internal/ir"
)

func TestRenderComponents_BasicChart(t *testing.T) {
	slide := ir.Slide{
		Index: 1,
		Components: []ir.Component{
			{Index: 0, Type: "chart:bar", Raw: "title: Yield", Params: map[string]any{"title": "Yield", "data": []any{96.0, 93.0}}},
		},
	}
	html := "before<!--goslide:component:0-->after"
	result := renderComponents(html, slide)
	require.Contains(t, result, `data-type="chart:bar"`)
	require.Contains(t, result, `data-comp-id="s1-c0"`)
	require.Contains(t, result, `data-params=`)
	require.Contains(t, result, `"title":"Yield"`)
	require.Contains(t, result, "before")
	require.Contains(t, result, "after")
	require.NotContains(t, result, "<!--goslide:component:0-->")
}

func TestRenderComponents_Mermaid(t *testing.T) {
	slide := ir.Slide{
		Index: 2,
		Components: []ir.Component{
			{Index: 0, Type: "mermaid", Raw: "graph TD\n    A --> B"},
		},
	}
	html := "<!--goslide:component:0-->"
	result := renderComponents(html, slide)
	require.Contains(t, result, `data-type="mermaid"`)
	require.Contains(t, result, `data-raw="graph TD`)
	require.Contains(t, result, `data-comp-id="s2-c0"`)
}

func TestRenderComponents_MultipleComponents(t *testing.T) {
	slide := ir.Slide{
		Index: 1,
		Components: []ir.Component{
			{Index: 0, Type: "chart:bar", Params: map[string]any{"title": "A"}},
			{Index: 1, Type: "chart:line", Params: map[string]any{"title": "B"}},
		},
	}
	html := "<!--goslide:component:0-->middle<!--goslide:component:1-->"
	result := renderComponents(html, slide)
	require.Contains(t, result, `data-comp-id="s1-c0"`)
	require.Contains(t, result, `data-comp-id="s1-c1"`)
	require.Contains(t, result, "middle")
}

func TestRenderComponents_NoComponents(t *testing.T) {
	slide := ir.Slide{Index: 1}
	html := "<p>no components here</p>"
	result := renderComponents(html, slide)
	require.Equal(t, html, result)
}

func TestRenderComponents_HTMLEscape(t *testing.T) {
	slide := ir.Slide{
		Index: 1,
		Components: []ir.Component{
			{Index: 0, Type: "chart:bar", Params: map[string]any{"title": "A < B & C's"}},
		},
	}
	html := "<!--goslide:component:0-->"
	result := renderComponents(html, slide)
	require.NotContains(t, result, `A < B`)
	require.Contains(t, result, `A &lt; B`)
}

func TestRenderComponents_Table(t *testing.T) {
	slide := ir.Slide{
		Index: 1,
		Components: []ir.Component{
			{Index: 0, Type: "table", Params: map[string]any{
				"columns":  []any{"Name", "Role"},
				"rows":     []any{[]any{"Alice", "Engineer"}},
				"sortable": true,
			}},
		},
	}
	html := "<!--goslide:component:0-->"
	result := renderComponents(html, slide)
	require.Contains(t, result, `data-type="table"`)
	require.Contains(t, result, `"sortable":true`)
}
```

- [ ] **Step 2: Run tests to verify failure**

Run: `GOTOOLCHAIN=local go test -C D:/CLAUDE-CODE-GOSLIDE/internal/renderer -run TestRenderComponents -v`

Expected: compilation error — `renderComponents` undefined.

- [ ] **Step 3: Implement components.go**

Create `internal/renderer/components.go`:

```go
package renderer

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/user/goslide/internal/ir"
)

func renderComponents(html string, slide ir.Slide) string {
	if len(slide.Components) == 0 {
		return html
	}

	for _, comp := range slide.Components {
		placeholder := fmt.Sprintf("<!--goslide:component:%d-->", comp.Index)
		replacement := buildComponentDiv(slide.Index, comp)
		html = strings.Replace(html, placeholder, replacement, 1)
	}

	return html
}

func buildComponentDiv(slideIndex int, comp ir.Component) string {
	compID := fmt.Sprintf("s%d-c%d", slideIndex, comp.Index)

	var paramsAttr string
	var rawAttr string

	if comp.Type == "mermaid" {
		paramsAttr = "{}"
		rawAttr = escapeAttr(comp.Raw)
	} else {
		paramsJSON, err := json.Marshal(comp.Params)
		if err != nil {
			paramsJSON = []byte("{}")
		}
		paramsAttr = escapeAttr(string(paramsJSON))
		rawAttr = ""
	}

	return fmt.Sprintf(
		`<div class="goslide-component" data-type="%s" data-params="%s" data-raw="%s" data-comp-id="%s"></div>`,
		escapeAttr(comp.Type),
		paramsAttr,
		rawAttr,
		compID,
	)
}

func escapeAttr(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	s = strings.ReplaceAll(s, "'", "&#39;")
	return s
}
```

- [ ] **Step 4: Run tests to verify pass**

Run: `GOTOOLCHAIN=local go test -C D:/CLAUDE-CODE-GOSLIDE/internal/renderer -run TestRenderComponents -v`

Expected: all PASS.

- [ ] **Step 5: Commit**

```
git -C D:/CLAUDE-CODE-GOSLIDE add internal/renderer/components.go internal/renderer/components_test.go
```
```
git -C D:/CLAUDE-CODE-GOSLIDE commit -m "feat(renderer): add component placeholder replacement with data attributes"
```

---

## Task 3: Integrate renderComponents into Render()

**Files:**
- Modify: `internal/renderer/renderer.go`
- Modify: `internal/renderer/renderer_test.go`

- [ ] **Step 1: Add integration test**

Read `internal/renderer/renderer_test.go`. Append:

```go
func TestRender_WithChartComponent(t *testing.T) {
	pres := &ir.Presentation{
		Meta: ir.Frontmatter{Title: "Chart", Theme: "default"},
		Slides: []ir.Slide{
			{
				Index:    1,
				Meta:     ir.SlideMeta{Layout: "default"},
				BodyHTML: "<h1>Dashboard</h1>\n<!--goslide:component:0-->\n",
				Components: []ir.Component{
					{Index: 0, Type: "chart:bar", Params: map[string]any{"title": "Yield"}},
				},
			},
		},
	}
	html, err := Render(pres)
	require.NoError(t, err)
	require.Contains(t, html, `data-type="chart:bar"`)
	require.Contains(t, html, `data-comp-id="s1-c0"`)
	require.NotContains(t, html, "<!--goslide:component:0-->")
}

func TestRender_WithMermaidComponent(t *testing.T) {
	pres := &ir.Presentation{
		Meta: ir.Frontmatter{Title: "Mermaid", Theme: "default"},
		Slides: []ir.Slide{
			{
				Index:    1,
				Meta:     ir.SlideMeta{Layout: "default"},
				BodyHTML: "<!--goslide:component:0-->",
				Components: []ir.Component{
					{Index: 0, Type: "mermaid", Raw: "graph TD\n    A --> B"},
				},
			},
		},
	}
	html, err := Render(pres)
	require.NoError(t, err)
	require.Contains(t, html, `data-type="mermaid"`)
	require.Contains(t, html, `data-raw="graph TD`)
}
```

- [ ] **Step 2: Update Render() to call renderComponents**

Read `internal/renderer/renderer.go`. Before the `tmpl.Execute` call, add placeholder replacement. Insert this block after `data.Title` fallback and before `var buf bytes.Buffer`:

```go
	for i := range data.Slides {
		s := &data.Slides[i]
		s.BodyHTML = template.HTML(renderComponents(string(s.BodyHTML), *s))
		for j := range s.Regions {
			s.Regions[j].HTML = template.HTML(renderComponents(string(s.Regions[j].HTML), *s))
		}
	}
```

Note: `data.Slides` is a copy from `pres.Slides` (value semantics in the `templateData` struct), so this mutation is safe.

- [ ] **Step 3: Run tests**

Run: `GOTOOLCHAIN=local go test -C D:/CLAUDE-CODE-GOSLIDE/internal/renderer -v`

Expected: all PASS.

- [ ] **Step 4: Commit**

```
git -C D:/CLAUDE-CODE-GOSLIDE add internal/renderer/renderer.go internal/renderer/renderer_test.go
```
```
git -C D:/CLAUDE-CODE-GOSLIDE commit -m "feat(renderer): integrate component placeholder replacement into Render pipeline"
```

---

## Task 4: Frontend — components.js

**Files:**
- Create: `web/static/components.js`

- [ ] **Step 1: Create components.js**

Create `web/static/components.js`:

```javascript
(function () {
  'use strict';

  var initialized = {};

  function resolveColor(colorName) {
    var root = document.documentElement;
    if (!colorName) {
      return getComputedStyle(root).getPropertyValue('--slide-accent').trim();
    }
    var val = getComputedStyle(root).getPropertyValue('--accent-' + colorName).trim();
    return val || getComputedStyle(root).getPropertyValue('--slide-accent').trim();
  }

  var defaultPalette = ['blue', 'teal', 'coral', 'purple', 'amber', 'green', 'red', 'pink'];

  function getDatasetColors(count, userColor) {
    if (userColor) return Array(count).fill(resolveColor(userColor));
    var colors = [];
    for (var i = 0; i < count; i++) {
      colors.push(resolveColor(defaultPalette[i % defaultPalette.length]));
    }
    return colors;
  }

  function buildChartConfig(type, params) {
    var chartType = type;
    var opts = {
      responsive: true,
      maintainAspectRatio: true,
      plugins: {
        title: { display: false },
        legend: { display: true }
      }
    };

    if (chartType === 'sparkline') {
      chartType = 'line';
      opts.scales = { x: { display: false }, y: { display: false } };
      opts.plugins.legend = { display: false };
      opts.plugins.title = { display: false };
      opts.elements = { point: { radius: 0 }, line: { borderWidth: 2 } };
      opts.maintainAspectRatio = false;
    }

    if (params.title) {
      opts.plugins.title = { display: true, text: params.title };
    }

    if (params.stacked) {
      opts.scales = opts.scales || {};
      opts.scales.x = opts.scales.x || {};
      opts.scales.y = opts.scales.y || {};
      opts.scales.x.stacked = true;
      opts.scales.y.stacked = true;
    }

    if (params.unit) {
      var unit = params.unit;
      opts.plugins.tooltip = {
        callbacks: {
          label: function (ctx) {
            return ctx.dataset.label + ': ' + ctx.parsed.y + unit;
          }
        }
      };
    }

    var datasets;
    if (params.datasets) {
      datasets = params.datasets.map(function (ds, idx) {
        var color = resolveColor(ds.color || defaultPalette[idx % defaultPalette.length]);
        return {
          label: ds.label || '',
          data: ds.data || [],
          backgroundColor: color,
          borderColor: color,
          borderWidth: 1
        };
      });
    } else {
      var data = params.data || [];
      var color = resolveColor(params.color);
      var bgColors;
      if (chartType === 'pie' || chartType === 'radar') {
        bgColors = getDatasetColors(data.length, null);
      } else {
        bgColors = color;
      }
      datasets = [{
        label: params.title || '',
        data: data,
        backgroundColor: bgColors,
        borderColor: chartType === 'line' ? color : bgColors,
        borderWidth: chartType === 'line' ? 2 : 1,
        fill: chartType === 'line' ? false : undefined
      }];
    }

    return {
      type: chartType,
      data: {
        labels: params.labels || [],
        datasets: datasets
      },
      options: opts
    };
  }

  function initChart(el) {
    var fullType = el.getAttribute('data-type');
    var chartType = fullType.split(':')[1] || 'bar';
    var params = JSON.parse(decodeAttr(el.getAttribute('data-params')));
    var canvas = document.createElement('canvas');
    if (chartType === 'sparkline') {
      canvas.style.height = '60px';
    }
    el.appendChild(canvas);
    var config = buildChartConfig(chartType, params);
    new Chart(canvas, config);
  }

  function initTable(el) {
    var params = JSON.parse(decodeAttr(el.getAttribute('data-params')));
    var table = document.createElement('table');

    var thead = document.createElement('thead');
    var headerRow = document.createElement('tr');
    (params.columns || []).forEach(function (col) {
      var th = document.createElement('th');
      th.textContent = col;
      headerRow.appendChild(th);
    });
    thead.appendChild(headerRow);
    table.appendChild(thead);

    var tbody = document.createElement('tbody');
    (params.rows || []).forEach(function (row) {
      var tr = document.createElement('tr');
      (Array.isArray(row) ? row : []).forEach(function (cell) {
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

  function makeSortable(table) {
    table.querySelectorAll('th').forEach(function (th, colIdx) {
      th.style.cursor = 'pointer';
      th.addEventListener('click', function () { sortTable(table, colIdx, th); });
    });
  }

  function sortTable(table, colIdx, th) {
    var tbody = table.querySelector('tbody');
    var rows = Array.from(tbody.querySelectorAll('tr'));
    var asc = th.getAttribute('data-sort') !== 'asc';

    rows.sort(function (a, b) {
      var aText = a.cells[colIdx].textContent.trim();
      var bText = b.cells[colIdx].textContent.trim();
      var aNum = parseFloat(aText), bNum = parseFloat(bText);
      if (!isNaN(aNum) && !isNaN(bNum)) return asc ? aNum - bNum : bNum - aNum;
      return asc ? aText.localeCompare(bText) : bText.localeCompare(aText);
    });

    rows.forEach(function (row) { tbody.appendChild(row); });
    table.querySelectorAll('th').forEach(function (h) {
      h.removeAttribute('data-sort');
      h.textContent = h.textContent.replace(/ [▲▼]$/, '');
    });
    th.setAttribute('data-sort', asc ? 'asc' : 'desc');
    th.textContent += asc ? ' ▲' : ' ▼';
  }

  function initSlideComponents(slide) {
    if (!slide) return;
    var comps = slide.querySelectorAll('.goslide-component');
    comps.forEach(function (el) {
      var id = el.getAttribute('data-comp-id');
      if (initialized[id]) return;
      var type = el.getAttribute('data-type');
      if (type.indexOf('chart') === 0) initChart(el);
      else if (type === 'table') initTable(el);
      initialized[id] = true;
    });
  }

  function parseBrightness(rgb) {
    var m = rgb.match(/\d+/g);
    if (!m) return 255;
    return (parseInt(m[0]) * 299 + parseInt(m[1]) * 587 + parseInt(m[2]) * 114) / 1000;
  }

  function initAllMermaid() {
    var els = document.querySelectorAll('.goslide-component[data-type="mermaid"]');
    if (els.length === 0) return;

    els.forEach(function (el) {
      var raw = decodeAttr(el.getAttribute('data-raw'));
      el.innerHTML = '<div class="mermaid">' + raw + '</div>';
    });

    var bg = getComputedStyle(document.querySelector('.reveal')).backgroundColor;
    var isDark = parseBrightness(bg) < 128;
    mermaid.initialize({ startOnLoad: false, theme: isDark ? 'dark' : 'default' });
    mermaid.run({ nodes: document.querySelectorAll('.mermaid') });
  }

  function decodeAttr(s) {
    if (!s) return '';
    return s.replace(/&quot;/g, '"').replace(/&#39;/g, "'").replace(/&lt;/g, '<').replace(/&gt;/g, '>').replace(/&amp;/g, '&');
  }

  Reveal.on('ready', function () {
    initAllMermaid();
    initSlideComponents(Reveal.getCurrentSlide());
  });
  Reveal.on('slidechanged', function (ev) {
    initSlideComponents(ev.currentSlide);
  });
})();
```

- [ ] **Step 2: Verify build**

Run: `GOTOOLCHAIN=local go build -C D:/CLAUDE-CODE-GOSLIDE ./web/`

- [ ] **Step 3: Commit**

```
git -C D:/CLAUDE-CODE-GOSLIDE add web/static/components.js
```
```
git -C D:/CLAUDE-CODE-GOSLIDE commit -m "feat(frontend): add components.js for chart, mermaid, and table rendering"
```

---

## Task 5: Template + CSS Updates

**Files:**
- Modify: `web/templates/slide.html`
- Modify: `web/themes/tokens.css`

- [ ] **Step 1: Update slide.html — add script tags**

Read `web/templates/slide.html`. Replace the script section at the bottom:

From:
```html
  <script src="/static/reveal/reveal.js"></script>
  <script src="/static/runtime.js"></script>
  <script>
```

To:
```html
  <script src="/static/chartjs/chart.min.js"></script>
  <script src="/static/mermaid/mermaid.min.js"></script>
  <script src="/static/reveal/reveal.js"></script>
  <script src="/static/runtime.js"></script>
  <script src="/static/components.js"></script>
  <script>
```

- [ ] **Step 2: Add component CSS to tokens.css**

Read `web/themes/tokens.css`. Append at the end:

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
  font-size: 0.75em;
}
.goslide-component table th {
  padding: 0.5em 0.75em;
  border-bottom: 2px solid var(--slide-accent);
  user-select: none;
}
.goslide-component table td {
  padding: 0.4em 0.75em;
  border-bottom: 1px solid var(--slide-border, rgba(0,0,0,0.1));
}
.goslide-component table th[data-sort] {
  color: var(--slide-accent);
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

- [ ] **Step 3: Verify build**

Run: `GOTOOLCHAIN=local go build -C D:/CLAUDE-CODE-GOSLIDE ./...`

- [ ] **Step 4: Commit**

```
git -C D:/CLAUDE-CODE-GOSLIDE add web/templates/slide.html web/themes/tokens.css
```
```
git -C D:/CLAUDE-CODE-GOSLIDE commit -m "feat: add Chart.js/Mermaid script tags and component CSS styles"
```

---

## Task 6: Golden Tests + Update Existing

**Files:**
- Create: `internal/renderer/testdata/golden/chart-component.md`
- Modify: existing golden .html files (regenerate)

- [ ] **Step 1: Create chart-component golden test input**

Create `internal/renderer/testdata/golden/chart-component.md`:

```markdown
---
title: Chart Test
theme: default
---

# Dashboard

~~~chart:bar
title: Yield by Line
labels: ["Line A", "Line B", "Line C"]
data: [96.2, 93.8, 97.1]
unit: "%"
color: teal
~~~

Some text after the chart.
```

- [ ] **Step 2: Regenerate all golden files**

Run: `GOTOOLCHAIN=local go test -C D:/CLAUDE-CODE-GOSLIDE/internal/renderer -run TestGolden -v -args -update`

- [ ] **Step 3: Verify golden tests pass without -update**

Run: `GOTOOLCHAIN=local go test -C D:/CLAUDE-CODE-GOSLIDE/internal/renderer -run TestGolden -v`

Expected: all PASS (6 subtests now).

- [ ] **Step 4: Run all tests**

Run: `GOTOOLCHAIN=local go test -C D:/CLAUDE-CODE-GOSLIDE ./internal/... -count=1 -race`

Expected: all PASS.

- [ ] **Step 5: Commit**

```
git -C D:/CLAUDE-CODE-GOSLIDE add internal/renderer/testdata/golden/
```
```
git -C D:/CLAUDE-CODE-GOSLIDE commit -m "test(renderer): add chart-component golden test and regenerate snapshots"
```

---

## Task 7: Demo + Manual Checklist

**Files:**
- Modify: `examples/demo.md`
- Modify: `MANUAL_CHECKLIST.md`

- [ ] **Step 1: Add component demo slides to demo.md**

Read `examples/demo.md`. Insert these slides BEFORE the "Thank You" slide (before the last `<!-- layout: title -->` block):

```markdown

---

# Chart Demo

~~~chart:bar
title: Yield by Production Line
labels: ["Line A", "Line B", "Line C", "Line D"]
data: [96.2, 93.8, 97.1, 91.5]
unit: "%"
color: teal
~~~

---

# Line Chart

~~~chart:line
title: Monthly Trend
labels: ["Jan", "Feb", "Mar", "Apr", "May", "Jun"]
data: [65, 72, 68, 85, 79, 92]
color: blue
~~~

---

# Pie Chart

~~~chart:pie
title: Market Share
labels: ["Product A", "Product B", "Product C"]
data: [45, 35, 20]
~~~

---

# Mermaid Diagram

~~~mermaid
graph TD
    A[Image Capture] --> B[Preprocessing]
    B --> C[Model Inference]
    C --> D{Pass?}
    D -->|Yes| E[OK]
    D -->|No| F[NG]
~~~

---

# Sortable Table

~~~table
columns: [Name, Role, Score]
rows:
  - ["Alice", "Engineer", 95]
  - ["Bob", "PM", 87]
  - ["Carol", "Lead", 92]
  - ["Dave", "Designer", 78]
sortable: true
~~~
```

- [ ] **Step 2: Update MANUAL_CHECKLIST.md**

Read `MANUAL_CHECKLIST.md`. Append after the existing checklist sections:

```markdown

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
```

- [ ] **Step 3: Run final full test**

Run: `GOTOOLCHAIN=local go test -C D:/CLAUDE-CODE-GOSLIDE ./... -count=1 -race`

Expected: all PASS.

- [ ] **Step 4: Build binary**

Run: `GOTOOLCHAIN=local go build -C D:/CLAUDE-CODE-GOSLIDE -ldflags "-X main.version=0.2.0" -o D:/CLAUDE-CODE-GOSLIDE/goslide.exe ./cmd/goslide`

- [ ] **Step 5: Commit**

```
git -C D:/CLAUDE-CODE-GOSLIDE add examples/demo.md MANUAL_CHECKLIST.md
```
```
git -C D:/CLAUDE-CODE-GOSLIDE commit -m "feat: add chart/mermaid/table demo slides and Phase 2b manual checklist"
```
