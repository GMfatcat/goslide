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

<!-- layout: title -->

# Thank You

Press **Esc** for overview mode.

Press **F** for fullscreen.
