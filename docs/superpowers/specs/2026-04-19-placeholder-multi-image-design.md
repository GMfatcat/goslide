# Placeholder Component + `image-grid` Layout Design

**Date:** 2026-04-19
**Status:** Design approved, ready for implementation plan
**Scope:** Phase 6b — image authoring ergonomics for manual + AI-generated slides

---

## 1. Overview

Two complementary features, shipped together:

- **`placeholder` component** — a styled dashed-rectangle "image stand-in"
  with icon + title + optional description. The author (or an LLM) emits a
  `placeholder` wherever a real image would belong but no URL is available.
  Solves the current gap where `goslide generate` avoids image slides
  entirely because it cannot invent URLs.

- **`image-grid` layout** — a CSS-grid slide layout that packs multiple
  cells (placeholders, real images, charts, or any component) into 2, 3,
  or 4 columns. Lets a single slide show an image gallery, architecture
  overview grid, or before/after comparison.

Both features are independent of each other: `placeholder` works in any
layout; `image-grid` accepts any cell content. Together, they enable
slides like a 2×2 grid of placeholder diagrams that the author later
replaces with real assets.

---

## 2. Syntax

### 2.1 `placeholder` component

```
~~~placeholder
hint: K8s cluster architecture
icon: 🗺️
aspect: 16:9
---
High-level control plane + worker node interaction
~~~
```

| Field | Required | Default | Notes |
|-------|----------|---------|-------|
| `hint` | yes | — | Title text; what the real image should depict |
| `icon` | no | `🖼️` | Single emoji; content-type cue |
| `aspect` | no | `16:9` | One of `16:9` / `4:3` / `1:1` / `3:4` / `9:16` |
| body | no | `Replace with actual content` | Text between `---` and closing fence |

### 2.2 `image-grid` layout

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
...
~~~

<!-- cell -->
~~~placeholder
hint: Trends
icon: 📈
~~~
```

- Reuses the existing `columns` slide meta (already used by `grid-cards`)
- Accepts `columns: 2`, `3`, or `4`; out-of-range values clamp to 2 with a
  warning
- `<!-- cell -->` is the only region marker (all cells are same type)
- Cells may contain any content: placeholder, Markdown image, chart,
  table, card, plain text, headings, lists
- Cells flow left-to-right, top-to-bottom; last row may be incomplete

### 2.3 Scope

`placeholder` may appear in **any** layout, not only `image-grid`
(default, two-column, image-left, section, etc.). A single `placeholder`
on a cover slide is a valid, common pattern.

---

## 3. Architecture

### 3.1 File changes

```
internal/
├── parser/
│   ├── component.go         # add "placeholder" to knownComponentPrefixes
│   └── slide_test.go        # add image-grid cell parse tests
├── ir/
│   ├── validate.go          # add "image-grid" to knownLayouts;
│   │                        # placeholder + aspect + columns validation
│   └── validate_test.go     # corresponding cases
├── renderer/
│   ├── components.go        # renderPlaceholder, image-grid wrapper
│   ├── components_test.go   # unit tests
│   └── testdata/golden/
│       ├── placeholder-basic.md + .html
│       ├── placeholder-custom.md + .html
│       └── image-grid-dark.md + .html
├── generate/
│   └── system_prompt.md     # new Placeholder section, Layouts entry, rule
└── web/css/ (existing file) # .gs-placeholder / .gs-image-grid rules
```

New tests only; no new implementation files. Reuses the existing
component and region-marker machinery (`isComponentFence`,
`extractComponents`, `<!-- name -->` region parsing).

### 3.2 Data model

If the existing codebase stores component params as `map[string]any` /
raw YAML, `placeholder` follows the same pattern — no new typed struct.
If there is a pattern of typed structs for `card` / `chart`, add:

```go
type Placeholder struct {
    Hint   string
    Icon   string // empty → renderer substitutes "🖼️"
    Aspect string // empty → renderer substitutes "16:9"
    Body   string // empty → renderer substitutes default subtitle
}
```

Implementation task will grep existing card/chart storage and follow
whichever convention is present; not re-engineering.

---

## 4. Rendering

### 4.1 Placeholder HTML

```html
<div class="gs-placeholder" data-aspect="16:9" style="aspect-ratio:16/9">
  <div class="gs-placeholder-icon">🗺️</div>
  <div class="gs-placeholder-hint">K8s cluster architecture</div>
  <div class="gs-placeholder-body">High-level control plane interaction</div>
</div>
```

- Body is the fence body text; empty body → default
  `Replace with actual content`
- `aspect` value goes to inline `aspect-ratio` style (not CSS class)
  because it is runtime data
- Class prefix `gs-` avoids collision with reveal.js classes

### 4.2 image-grid HTML

```html
<div class="gs-image-grid" data-columns="2"
     style="display:grid;grid-template-columns:repeat(2,1fr);gap:16px">
  <div class="gs-cell"><!-- cell 1 content --></div>
  <div class="gs-cell"><!-- cell 2 content --></div>
  <div class="gs-cell"><!-- cell 3 content --></div>
  <div class="gs-cell"><!-- cell 4 content --></div>
</div>
```

- Gap fixed at 16px (YAGNI; open a parameter later if demanded)
- `grid-template-columns` is inline style (runtime value)

### 4.3 CSS (global, theme-agnostic)

Add to an appropriate existing global CSS file:

```css
.gs-placeholder {
  border: 2px dashed var(--gs-accent, #8888aa);
  border-radius: 8px;
  padding: clamp(16px, 3vw, 32px);
  background: color-mix(in srgb, var(--gs-accent, #8888aa) 8%, transparent);
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  text-align: center;
  width: 100%;
}
.gs-placeholder-icon {
  font-size: clamp(24px, 4vw, 48px);
  opacity: 0.6;
  margin-bottom: 8px;
}
.gs-placeholder-hint {
  font-weight: 600;
  font-size: clamp(14px, 1.8vw, 20px);
}
.gs-placeholder-body {
  font-size: 0.85em;
  opacity: 0.7;
  margin-top: 4px;
  font-style: italic;
}
.gs-image-grid { width: 100%; }
.gs-image-grid .gs-cell { min-width: 0; }
```

`--gs-accent` is the existing theme accent token; fallback `#8888aa`
keeps the mockup look when no theme is loaded. `color-mix` is widely
supported in modern Chromium/Firefox/Safari.

---

## 5. LLM System Prompt Updates

`internal/generate/system_prompt.md`:

**Add to Layouts:**

```markdown
- `image-grid` — grid of images, placeholders, or components. Use
  `columns: 2|3|4` and `<!-- cell -->` before each item. Cells may contain
  a `placeholder`, a Markdown image, a chart, or any other component.
```

**Add new Components subsection:**

````markdown
## Placeholder (for image-heavy slides without URLs)

When a slide should show a diagram, chart, photo, map, or screenshot but
you do not have an image URL, emit a `placeholder` component instead of
leaving the slide empty or skipping it:

```
~~~placeholder
hint: K8s cluster architecture
icon: 🗺️
aspect: 16:9
---
Control plane + worker node interaction
~~~
```

- `hint` (required): short title the author will later replace with a
  real image matching this description.
- `icon` (optional): a single emoji hinting at content type.
  Suggestions: 📊 charts, 🗺️ architecture diagrams, 📷 photos, 📈 trends,
  🖼️ generic image, 📐 schematics.
- `aspect` (optional): 16:9 (default) | 4:3 | 1:1 | 3:4 | 9:16.
- Body (between `---` and closing fence): optional subtitle/description.

Use `placeholder` freely wherever a real image would belong — a single
cover slide, an image-left/right region, or inside an `image-grid` cell.
````

**Add to Rules:**

```markdown
- When the slide is primarily visual (diagram, chart, screenshot) and you
  have no image URL, use `placeholder` with a descriptive `hint` and a
  fitting `icon`. Do not skip the slide and do not invent fake image URLs.
```

---

## 6. Validation

| Condition | Severity | Code | Message |
|-----------|----------|------|---------|
| `placeholder` without `hint` | error | `placeholder-missing-hint` | `placeholder component requires 'hint' field` |
| `placeholder` `aspect` not in whitelist | warning | `unknown-aspect` | `aspect %q not recognized (using 16:9)` |
| `image-grid` with `columns` not 2/3/4 | warning | `columns-out-of-range` | `columns %d out of range (2-4); clamping to 2` |
| `image-grid` with zero `<!-- cell -->` markers | warning | `image-grid-empty` | `image-grid has no cells` |

Valid aspect whitelist: `16:9`, `4:3`, `1:1`, `3:4`, `9:16`.

Adds `image-grid` to `knownLayouts`; does not register it in
`requiredRegions` (an empty grid is a warning, not an error — unlike
`two-column` which errors without both regions).

---

## 7. Testing

### 7.1 Parser

- `placeholder` fence is identified; YAML meta parsed (hint, icon,
  aspect)
- Missing `hint` still parses (caught by validator)
- `image-grid` + multiple `<!-- cell -->` markers split correctly
- Mixed cell content (placeholder + markdown image + chart) parses

### 7.2 IR validation

- Missing `hint` → error with code `placeholder-missing-hint`
- `aspect: 2:1` → warning with code `unknown-aspect`
- `columns: 5` → warning with code `columns-out-of-range`
- `image-grid` with no cells → warning with code `image-grid-empty`
- Fully valid slide → no errors, no warnings

### 7.3 Renderer (golden tests)

- `placeholder-basic.md` → default icon + default aspect + default body
- `placeholder-custom.md` → custom icon + custom aspect + custom body
- `image-grid-dark.md` → 2×2 mix of placeholder + Markdown image
- CSS class names present; inline `aspect-ratio` and
  `grid-template-columns` match input meta

### 7.4 Generate integration

- `goslide generate --dump-prompt` output contains the strings
  `placeholder`, `image-grid`, `hint:`, and `<!-- cell -->` (quick grep
  sanity check)
- No real LLM call in CI

### 7.5 Manual validation (not in CI)

Run `goslide generate "Kubernetes architecture"` through the same
OpenRouter model used for v1.2.0 validation; record one or two outputs
into `examples/ai-generated/` with brief notes on whether the LLM
proactively used `placeholder` / `image-grid`. No blocking criterion;
this informs prompt tuning if needed.

---

## 8. Out of Scope

- Custom gap, padding, or inter-cell spacing params on `image-grid`
  (fixed at 16px)
- `kind` keyword mapping to preset icons (chose direct emoji instead)
- Placeholder → image drag-and-drop in `goslide serve` UI
- Automatic image generation (text-to-image model integration)
- Per-cell rowspan/colspan in `image-grid`
- Fluid "masonry" layout (Pinterest-style)

---

## 9. Success Criteria

- Author writes a slide with `~~~placeholder` and it renders as the
  dashed-rectangle mockup (icon + hint + body) at the declared aspect
  ratio.
- `image-grid` layout with 2/3/4 columns renders a responsive CSS grid
  with mixed cell content.
- Validation warns on invalid `aspect`, out-of-range `columns`, and
  `image-grid` with no cells; errors on `placeholder` without `hint`.
- `--dump-prompt` output reflects the new Layout, Component, and Rule
  entries.
- All unit + golden tests pass; no new dependencies.

---

## 10. Release Note Snippet (draft for future v1.3.0 notes)

```
### 🖼️ Image Placeholder + image-grid Layout

New `placeholder` component renders a styled dashed rectangle with
an icon and descriptive text — drop it anywhere a real image would
go, then replace it with the actual asset later. Pairs with the new
`image-grid` layout for multi-image / multi-diagram slides. `goslide
generate` now produces image-centric slides instead of skipping
them when no image URL is available.
```
