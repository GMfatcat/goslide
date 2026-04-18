# Phase 3b: API Component + Render Types Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add `api` component that fetches data from proxied API endpoints and renders results using 7 render types (metric, chart:*, table, json, log, image, markdown), with client-side polling support.

**Architecture:** All logic is client-side JS in `components.js`. The `api` component reads `data-params` (url, method, refresh, render[]), fetches via the existing proxy, extracts data using dot-notation path, and dispatches to render type functions. Polling via `setInterval` with pause/resume on slide change.

**Tech Stack:** Vanilla JS (fetch API, setInterval). Reuses existing Chart.js + table builder from Phase 2b. No new dependencies. No Go-side changes.

**Shell rules (Windows):** Never chain commands with `&&`. Use SEPARATE Bash calls for `git add`, `git commit`, `go test`, `go build`. Use `GOTOOLCHAIN=local` prefix for ALL go commands. Use `-C` flag for go/git to specify directory.

---

## Task 1: API Component JS — Init + Path Extraction + 7 Render Types + Polling

**Files:**
- Modify: `web/static/components.js`

- [ ] **Step 1: Add api component code to components.js**

Read `web/static/components.js`. Add the following code BEFORE the `initAllComponents` function.

First, add the path extraction utility:

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

  function parseRefresh(s) {
    if (!s) return 0;
    var m = String(s).match(/^(\d+)(s|ms)?$/);
    if (!m) return 0;
    var val = parseInt(m[1]);
    if (m[2] === 'ms') return val;
    return val * 1000;
  }
```

Then add the 7 render type functions:

```javascript
  function renderMetric(container, item, data) {
    var div = document.createElement('div');
    div.className = 'goslide-metric';
    var value = document.createElement('div');
    value.className = 'goslide-metric-value';
    value.textContent = data + (item.unit || '');
    if (item.color) value.style.color = resolveColor(item.color);
    var label = document.createElement('div');
    label.className = 'goslide-metric-label';
    label.textContent = item.label || '';
    div.appendChild(value);
    div.appendChild(label);
    container.appendChild(div);
  }

  function renderApiChart(container, item, data) {
    var chartType = item.type.split(':')[1] || 'bar';
    var params = { title: item.title, color: item.color, unit: item.unit };
    if (Array.isArray(data) && data.length > 0 && typeof data[0] === 'object') {
      var keys = Object.keys(data[0]);
      var labelKey = null, dataKey = null;
      for (var k = 0; k < keys.length; k++) {
        if (typeof data[0][keys[k]] === 'string' && !labelKey) labelKey = keys[k];
        if (typeof data[0][keys[k]] === 'number' && !dataKey) dataKey = keys[k];
      }
      params.labels = data.map(function(d) { return d[labelKey || keys[0]]; });
      params.data = data.map(function(d) { return d[dataKey || keys[1]]; });
    } else if (Array.isArray(data)) {
      params.data = data;
      params.labels = item.labels || data.map(function(_, i) { return '' + i; });
    } else {
      params.data = [data];
      params.labels = [item.label || ''];
    }
    var canvas = document.createElement('canvas');
    container.appendChild(canvas);
    var config = buildChartConfig(chartType, params);
    new Chart(canvas, config);
  }

  function renderApiTable(container, item, data) {
    if (!Array.isArray(data) || data.length === 0) return;
    var columns = item.columns || Object.keys(data[0]);
    var table = document.createElement('table');
    var thead = document.createElement('thead');
    var headerRow = document.createElement('tr');
    columns.forEach(function(col) {
      var th = document.createElement('th');
      th.textContent = col;
      headerRow.appendChild(th);
    });
    thead.appendChild(headerRow);
    table.appendChild(thead);
    var tbody = document.createElement('tbody');
    data.forEach(function(obj) {
      var tr = document.createElement('tr');
      columns.forEach(function(col) {
        var td = document.createElement('td');
        td.textContent = obj[col] != null ? String(obj[col]) : '';
        tr.appendChild(td);
      });
      tbody.appendChild(tr);
    });
    table.appendChild(tbody);
    container.appendChild(table);
  }

  function renderJSON(container, item, data) {
    var pre = document.createElement('pre');
    pre.className = 'goslide-json';
    pre.textContent = JSON.stringify(data, null, 2);
    container.appendChild(pre);
  }

  function renderLog(container, item, data) {
    var pre = document.createElement('pre');
    pre.className = 'goslide-log';
    pre.textContent = Array.isArray(data) ? data.join('\n') : String(data);
    container.appendChild(pre);
  }

  function renderApiImage(container, item, data) {
    var img = document.createElement('img');
    var src = String(data);
    if (src.startsWith('data:') || src.startsWith('http')) {
      img.src = src;
    } else {
      img.src = 'data:image/png;base64,' + src;
    }
    img.style.maxWidth = '100%';
    img.style.borderRadius = '0.5rem';
    container.appendChild(img);
  }

  function renderMarkdownRaw(container, item, data) {
    var pre = document.createElement('pre');
    pre.className = 'goslide-markdown-raw';
    pre.textContent = String(data);
    container.appendChild(pre);
  }

  function renderItem(container, item, data) {
    if (data === undefined) return;
    var type = item.type || '';
    if (type === 'metric') renderMetric(container, item, data);
    else if (type.indexOf('chart') === 0) renderApiChart(container, item, data);
    else if (type === 'table') renderApiTable(container, item, data);
    else if (type === 'json') renderJSON(container, item, data);
    else if (type === 'log') renderLog(container, item, data);
    else if (type === 'image') renderApiImage(container, item, data);
    else if (type === 'markdown') renderMarkdownRaw(container, item, data);
  }
```

Then add the api component init and polling:

```javascript
  function fetchAndRender(el, params) {
    var url = params.url;
    var opts = { method: (params.method || 'GET').toUpperCase() };
    if (params.body) {
      opts.body = JSON.stringify(params.body);
      opts.headers = { 'Content-Type': 'application/json' };
    }
    fetch(url, opts)
      .then(function(r) { return r.json(); })
      .then(function(json) {
        el.innerHTML = '';
        var items = document.createElement('div');
        items.className = 'goslide-api-items';
        var renderList = params.render;
        if (!Array.isArray(renderList)) {
          renderList = [renderList || { type: 'json' }];
        }
        renderList.forEach(function(item) {
          var data = extractPath(json, item.path);
          renderItem(items, item, data);
        });
        el.appendChild(items);
      })
      .catch(function(err) {
        el.innerHTML = '<pre class="goslide-api-error">API error: ' + err.message + '</pre>';
      });
  }

  function initApiComponent(el) {
    var params = JSON.parse(decodeAttr(el.getAttribute('data-params')));
    var refreshMs = parseRefresh(params.refresh);

    el._fetchFn = function() { fetchAndRender(el, params); };
    el._refreshMs = refreshMs;
    el._fetchFn();

    if (refreshMs > 0 && el.closest('section') === Reveal.getCurrentSlide()) {
      el._pollInterval = setInterval(el._fetchFn, refreshMs);
    }
  }
```

Then in the `initAllComponents` function, add after the `embed:iframe` line:

```javascript
      else if (type === 'api') initApiComponent(el);
```

Finally, add polling pause/resume. Find the existing `Reveal.on('ready'` block at the bottom and ADD a new slidechanged handler AFTER it:

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

- [ ] **Step 2: Verify build**

Run: `GOTOOLCHAIN=local go build -C D:/CLAUDE-CODE-GOSLIDE ./...`

- [ ] **Step 3: Commit**

```
git -C D:/CLAUDE-CODE-GOSLIDE add web/static/components.js
```
```
git -C D:/CLAUDE-CODE-GOSLIDE commit -m "feat(frontend): add api component with 7 render types, path extraction, and polling"
```

---

## Task 2: CSS for Render Types

**Files:**
- Modify: `web/themes/tokens.css`

- [ ] **Step 1: Append render type CSS**

Read `web/themes/tokens.css`. Append at the end:

```css

/* API component */
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

/* Metric */
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

/* JSON viewer */
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

/* Log viewer */
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

/* Markdown raw */
.goslide-markdown-raw {
  background: var(--slide-card-bg);
  font-family: var(--font-mono);
  font-size: 0.7em;
  padding: 1rem;
  border-radius: 0.5rem;
  white-space: pre-wrap;
}
```

- [ ] **Step 2: Verify build**

Run: `GOTOOLCHAIN=local go build -C D:/CLAUDE-CODE-GOSLIDE ./...`

- [ ] **Step 3: Commit**

```
git -C D:/CLAUDE-CODE-GOSLIDE add web/themes/tokens.css
```
```
git -C D:/CLAUDE-CODE-GOSLIDE commit -m "feat: add CSS for api component render types (metric, json, log, markdown)"
```

---

## Task 3: Mock API Expansion + Demo + Checklist

**Files:**
- Modify: `examples/mock-api/main.go`
- Modify: `examples/demo.md`
- Modify: `MANUAL_CHECKLIST.md`

- [ ] **Step 1: Add /log endpoint to mock API**

Read `examples/mock-api/main.go`. Add before the `fmt.Println` line:

```go
	http.HandleFunc("/log", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"entries": []string{
				"[2026-04-18 10:00:01] System startup",
				"[2026-04-18 10:00:02] Camera initialized",
				"[2026-04-18 10:00:03] Model loaded: SegFormer v2",
				"[2026-04-18 10:00:05] First inspection completed",
				"[2026-04-18 10:00:08] Yield: 96.2%",
			},
		})
	})
```

- [ ] **Step 2: Add api demo slides to demo.md**

Read `examples/demo.md`. Find the "Thank You" slide. INSERT before it:

```markdown

---

# API Dashboard

~~~api
url: /api/mock/metrics
render:
  - type: metric
    path: lines[0].yield
    label: Line A Yield
    unit: "%"
    color: teal
  - type: metric
    path: lines[1].yield
    label: Line B Yield
    unit: "%"
    color: coral
  - type: chart:bar
    path: lines
    title: Yield by Line
    color: blue
~~~

---

# API Data Views

~~~api
url: /api/mock/metrics
render:
  - type: table
    path: lines
    title: Line Data
  - type: json
    path: lines[0]
    title: Raw JSON
~~~

---

# System Log

~~~api
url: /api/mock/log
render:
  - type: log
    path: entries
~~~
```

- [ ] **Step 3: Update MANUAL_CHECKLIST.md**

Read the file. Append:

```markdown

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
```

- [ ] **Step 4: Run all tests + build**

Run: `GOTOOLCHAIN=local go test -C D:/CLAUDE-CODE-GOSLIDE ./... -count=1 -race`

Run: `GOTOOLCHAIN=local go build -C D:/CLAUDE-CODE-GOSLIDE -ldflags "-X main.version=0.6.0" -o D:/CLAUDE-CODE-GOSLIDE/goslide.exe ./cmd/goslide`

- [ ] **Step 5: Commit**

```
git -C D:/CLAUDE-CODE-GOSLIDE add examples/mock-api/main.go examples/demo.md MANUAL_CHECKLIST.md
```
```
git -C D:/CLAUDE-CODE-GOSLIDE commit -m "feat: add api component demo slides, mock /log endpoint, and Phase 3b checklist"
```
