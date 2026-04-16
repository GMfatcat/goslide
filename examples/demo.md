---
title: GoSlide Demo
theme: default
accent: teal
transition: slide
slide-number: auto
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

<!-- layout: title -->

# Thank You

Press **Esc** for overview mode.

Press **F** for fullscreen.
