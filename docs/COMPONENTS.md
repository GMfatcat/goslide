# 📦 GoSlide Component Reference

Components are embedded in Markdown via fenced code blocks with custom language tags.

## Layer 1 — Declarative (Static Data)

### Chart

```markdown
~~~chart:bar
title: Yield by Line
labels: ["A", "B", "C"]
data: [96.2, 93.8, 97.1]
unit: "%"
color: teal
~~~
```

**Types:** `bar`, `line`, `pie`, `radar`, `sparkline`

**Parameters:**
| Param | Type | Description |
|-------|------|-------------|
| `title` | string | Chart title |
| `labels` | string[] | Axis labels / legend |
| `data` | number[] | Single dataset |
| `datasets` | object[] | Multi-series: `[{label, data, color}]` |
| `unit` | string | Tooltip value suffix |
| `color` | string | Accent color name |
| `stacked` | bool | Stack bars/areas |

### Mermaid

```markdown
~~~mermaid
graph TD
    A[Start] --> B[Process]
    B --> C{Decision}
    C -->|Yes| D[OK]
    C -->|No| E[NG]
~~~
```

Renders as SVG. Auto-detects dark/light theme.

### Table

```markdown
~~~table
columns: [Name, Role, Score]
rows:
  - ["Alice", "Engineer", 95]
  - ["Bob", "PM", 87]
sortable: true
~~~
```

Click column headers to sort (auto-detects number vs string).

---

## Layer 2 — Reactive (Interactive)

### Tabs + Panel

```markdown
~~~tabs
id: compare
labels: ["Plan A", "Plan B"]
~~~

~~~panel:compare-0
Plan A content here...
~~~

~~~panel:compare-1
Plan B content here...
~~~
```

### Slider

```markdown
~~~slider
id: threshold
label: Yield threshold
min: 80
max: 100
value: 95
step: 0.5
unit: "%"
~~~
```

Publishes value to reactive store. Read via `GoSlide.get('threshold')`.

### Toggle

```markdown
~~~toggle
id: show_details
label: Show details
default: false
~~~

~~~panel:show_details
Detail content shown when toggle is on.
~~~
```

---

## Layer 3 — API-Driven

### API Component

```markdown
~~~api
url: /api/endpoint
method: GET
refresh: 5s
layout: dashboard
render:
  - type: metric
    path: yield
    label: Yield
    unit: "%"
    color: green
  - type: chart:bar
    path: lines
    title: Line Status
    span: 2
~~~
```

**Render Types:**

| Type | Input | Output |
|------|-------|--------|
| `metric` | Single value | Large number card |
| `chart:*` | Array/object | Chart.js chart |
| `table` | Object array | Sortable table |
| `json` | Any | Formatted JSON viewer |
| `log` | String/string[] | Terminal-style text |
| `image` | URL/base64 | Rendered image |
| `markdown` | String | Plain text display |

**Layout:** `dashboard` arranges items in a 2-column grid. Use `span: 2` to make an item full-width.

**Path:** Dot notation to extract data from API response (e.g., `data.lines[0].yield`).

---

## Layer 4 — Escape Hatch

### Embed HTML

```markdown
~~~embed:html
<div id="demo">
  <button onclick="alert('Hello!')">Click Me</button>
</div>
<script>
  // Full JS execution
</script>
~~~
```

### Embed Iframe

```markdown
~~~embed:iframe
url: https://example.com
height: 500
~~~
```

---

## Expandable Cards

```markdown
<!-- layout: grid-cards -->

# Overview

~~~card
icon: 📷
color: blue
title: Module Name
desc: Short description
---
## Full Detail

Detailed markdown content shown in overlay.
Tables, lists, code blocks all supported.
~~~
```

Click card → modal overlay with detail content. Press Esc or click backdrop to close.
