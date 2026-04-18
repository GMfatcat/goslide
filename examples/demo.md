---
title: GoSlide Demo
theme: default
accent: teal
transition: slide
slide-number: auto
slide-number-format: total
---

# GoSlide

Markdown-driven interactive presentations.

Built with Go + Reveal.js.

---

<!-- layout: title -->

# Welcome to GoSlide

A single binary, offline-first presentation tool.

---

<!-- layout: section -->

# Features

---

<!-- layout: two-column -->

# Go vs Others

<!-- left -->

## GoSlide

- Single binary
- Offline-first
- Live reload
- CJK support

<!-- right -->

## Alternatives

- Marp (Node.js)
- Slidev (Vue)
- reveal-md (Node.js)

---

<!-- layout: code-preview -->

# Code Example

<!-- code -->

```go
package main

import "fmt"

func main() {
    fmt.Println("Hello, GoSlide!")
}
```

<!-- preview -->

## Output

```
Hello, GoSlide!
```

A simple Go program rendered in a code-preview layout.

---

<!-- fragments: true -->
<!-- fragment-style: highlight-current -->

# Key Points

- Markdown is the source of truth
- Layouts via HTML comments
- Themes with accent colors
- Progressive reveal with fragments
- Live reload keeps your slide position

---

# 中文支援

GoSlide 內建 Noto Sans TC 字型。

支援繁體中文、日文漢字等 CJK 字元。

**粗體**、*斜體*、`程式碼` 都沒問題。

---

<!-- layout: quote -->

> The best way to predict the future is to invent it.
>
> — Alan Kay

---

<!-- layout: split-heading -->

<!-- heading -->

## System Design

<!-- body -->

GoSlide uses a pipeline architecture:

1. **Parse** — Markdown to IR
2. **Validate** — Check whitelists
3. **Render** — IR to HTML

Each stage is independently testable.

---

<!-- layout: three-column -->

# Three Options

<!-- col1 -->

## Plan A

- Low cost
- Quick start
- Limited scale

<!-- col2 -->

## Plan B

- Medium cost
- Balanced approach
- Good scale

<!-- col3 -->

## Plan C

- High investment
- Full features
- Enterprise scale

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

---

# Tabs Demo

~~~tabs
id: compare
labels: ["Plan A", "Plan B", "Plan C"]
~~~

~~~panel:compare-0
## Plan A

Low cost, quick start, limited scalability.
Best for small teams and prototypes.
~~~

~~~panel:compare-1
## Plan B

Medium investment, balanced approach.
Suitable for most production workloads.
~~~

~~~panel:compare-2
## Plan C

Full investment, enterprise features.
Maximum scalability and support.
~~~

---

# Interactive Controls

~~~slider
id: threshold
label: Yield threshold
min: 80
max: 100
value: 95
step: 0.5
unit: "%"
~~~

~~~toggle
id: show_details
label: Show details
default: false
~~~

~~~panel:show_details
### Detail View

When the toggle is on, this panel becomes visible.
The slider value can be read via `GoSlide.get('threshold')` in the browser console.
~~~

---

<!-- layout: grid-cards -->

# System Overview

~~~card
icon: 📷
color: blue
title: Image Capture
desc: High-speed camera acquisition
---
## Image Capture

| Spec | Value |
|------|-------|
| Speed | 120 fps |
| Resolution | 5MP |

The capture module interfaces with industrial cameras.
~~~

~~~card
icon: 🔍
color: teal
title: Defect Detection
desc: SegFormer + YOLO pipeline
---
## Defect Detection

Accuracy: 98.7% | Latency: 12ms

The pipeline combines SegFormer for semantic segmentation
with YOLO for discrete defect detection.
~~~

~~~card
icon: 📊
color: purple
title: Analytics
desc: Real-time yield monitoring
---
## Analytics Dashboard

- Real-time yield tracking
- Defect distribution analysis
- Trend visualization
- Alert thresholds
~~~

~~~card
icon: 🔧
color: amber
title: Maintenance
desc: Predictive maintenance alerts
---
## Maintenance System

Predictive maintenance using vibration and temperature sensors.
Mean time between failures: 2,400 hours.
~~~

---

<!-- layout: title -->

# Thank You

Press **Esc** for overview mode.

Press **F** for fullscreen.
