# 🎉 GoSlide v1.3.0

## What's New

### 🖼️ Image Placeholder Component

A new `placeholder` component — a styled dashed-rectangle "image
stand-in" with an icon, title, and optional description. Drop one
wherever a real image will eventually go, then replace it with the
actual asset when ready.

```
~~~placeholder
hint: K8s cluster architecture
icon: 🗺️
aspect: 16:9
---
Control plane + worker node interaction
~~~
```

- `hint` (required) — title text describing what the image will show
- `icon` (optional) — single emoji cue (📊 charts, 🗺️ diagrams, 📷 photos, 📈 trends, 🖼️ generic)
- `aspect` (optional) — `16:9` (default), `4:3`, `1:1`, `3:4`, or `9:16`
- Body (between `---` and closing fence) — optional subtitle

Placeholders work in any layout: as a full-slide cover diagram, inside
an `image-left`/`image-right` region, or combined with the new
`image-grid` below.

### 🧩 `image-grid` Layout

A new CSS-grid slide layout that packs multiple cells (placeholders,
real images, charts, or any other component) into 2, 3, or 4 columns.

```
<!-- layout: image-grid -->
<!-- columns: 2 -->

<!-- cell -->
~~~placeholder
hint: Architecture
icon: 🗺️
~~~

<!-- cell -->
![Dashboard](./dashboard.png)

<!-- cell -->
~~~chart
type: bar
title: Sales
data:
  labels: [Q1, Q2, Q3]
  values: [10, 12, 15]
~~~

<!-- cell -->
~~~placeholder
hint: Trends
icon: 📈
~~~
```

`<!-- cell -->` before each item marks a new grid cell; cells can hold
any content.

### 🤖 Smarter `goslide generate`

The AI-generation command (Phase 6a) now knows about the new features.
The system prompt has been tightened so LLM output consistently uses the
correct fence and comment syntax:

- All component fences are `~~~` (triple tilde). Triple-backtick blocks
  are plain code and will not render as components.
- Per-slide layout settings use HTML comments
  (`<!-- layout: image-grid -->`, `<!-- columns: 2 -->`), never a
  YAML `---` block mid-document.
- Explicit "wrong vs right" examples and a new `image-grid` example in
  the prompt.

On OpenRouter free tier, 5 of 6 models tested after this iteration
produce valid, first-pass-parseable output. See
[`examples/ai-generated/k8s-visual.md`](examples/ai-generated/k8s-visual.md)
for a real generation (`openai/gpt-oss-20b:free`, 13 placeholders + one
4-cell image-grid slide).

### ✅ Validation

`goslide validate` / `goslide build` emit:

- **Error** `placeholder-missing-hint` when a placeholder lacks `hint`
- **Warning** `unknown-aspect` when `aspect` isn't on the whitelist (falls back to 16:9)
- **Warning** `columns-out-of-range` when `image-grid` columns fall outside 2-4
- **Warning** `image-grid-empty` when an image-grid layout contains no cells

## Compatibility

No breaking changes. v1.2.0 decks work unchanged. The internal region
parser was refactored from a name-keyed map to an ordered slice so that
repeatable markers (`<!-- cell -->`) produce distinct regions —
transparent to existing layouts (two-column, three-column, etc.) since
none of them used repeated markers.

## Full Changelog

See [v1.2.0...v1.3.0](https://github.com/GMfatcat/goslide/compare/v1.2.0...v1.3.0) for all changes.
