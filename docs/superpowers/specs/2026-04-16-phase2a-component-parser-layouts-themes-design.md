# GoSlide Phase 2a — Component Parser + Layouts + Themes Design Spec

> **Status: DRAFT**

## Decision Log

| # | Question | Decision |
|---|----------|----------|
| 1 | Component IR design | Unified `Component` struct with `Type`, `Raw`, `Params map[string]any`; typed parse in renderer layer |
| 2 | Parser integration | Pre-scan raw text before goldmark; extract component fences, leave placeholders |
| 3 | Position preservation | Placeholder `<!--goslide:component:N-->` in body text; renderer replaces with component HTML |
| 4 | Layout scope | 6 layouts now (three-column, image-left/right, quote, split-heading, top-bottom, blank); grid-cards deferred to Phase 2d |
| 5 | Theme accents | corporate=blue, minimal=blue, hacker=green |

---

## 1. Component IR

### New type in `internal/ir/presentation.go`

```go
type Component struct {
    Index  int            // position in slide (0-based)
    Type   string         // "chart:bar", "mermaid", "table", "tabs", etc.
    Raw    string         // original fence content (unparsed)
    Params map[string]any // yaml.Unmarshal result
}
```

### Slide struct change

```go
type Slide struct {
    Index      int
    Meta       SlideMeta
    RawBody    string
    BodyHTML   template.HTML // may contain <!--goslide:component:N--> placeholders
    Regions    []Region
    Components []Component   // NEW — ordered by appearance
}
```

---

## 2. Component Parser

### Extraction function

```go
func extractComponents(body string) (cleaned string, components []ir.Component)
```

**Algorithm:**
1. Line scan the body text
2. When encountering `~~~xxx` where xxx matches a known component prefix or contains `:`, begin collecting
3. Collect lines until closing `~~~`
4. YAML unmarshal collected content into `map[string]any`
5. Replace the fence block with `<!--goslide:component:N-->`
6. Non-component fences (e.g., `~~~go`, `~~~bash`) pass through untouched to goldmark

**Known component prefixes** (registered but not rendered in Phase 2a):
`chart`, `mermaid`, `table`, `tabs`, `panel`, `slider`, `toggle`, `api`, `embed`, `card`

### Updated parseSlide flow

```
raw slide text
  → extractComponents() → cleaned text + []Component
  → metadata extraction (unchanged)
  → region splitting (unchanged; placeholders stay inside regions)
  → goldmark render (unchanged; placeholders are HTML comments, preserved)
  → Slide { BodyHTML(with placeholders), Components, Regions, ... }
```

### Validation changes

- `future-component` warning removed for known component types (they are now parsed)
- Unknown component types (not in the known list) still emit `future-component` warning and fall back to code block (no extraction)

---

## 3. Six New Layouts

### Region definitions

| Layout | Regions | CSS behavior |
|--------|---------|-------------|
| `three-column` | `col1`, `col2`, `col3` | `grid-template-columns: 1fr 1fr 1fr` |
| `image-left` | `image`, `text` | 2-col grid, left image full-width, right text |
| `image-right` | `text`, `image` | 2-col grid, left text, right image full-width |
| `quote` | none (single flow) | centered, large italic blockquote, attribution in last paragraph |
| `split-heading` | `heading`, `body` | 2-col grid (2fr/3fr), left large heading, right body |
| `top-bottom` | `top`, `bottom` | 2-row grid, top visual, bottom text |
| `blank` | none (single flow) | no padding, no default styles |

### Parser changes

`layoutRegions` map additions:

```go
var layoutRegions = map[string][]string{
    "two-column":    {"left", "right"},
    "code-preview":  {"code", "preview"},
    "three-column":  {"col1", "col2", "col3"},
    "image-left":    {"image", "text"},
    "image-right":   {"text", "image"},
    "split-heading": {"heading", "body"},
    "top-bottom":    {"top", "bottom"},
}
```

### Validation changes

- `phase1Layouts` renamed to `knownLayouts`, merged with new layouts
- `futureLayouts` reduced to only `grid-cards`
- `requiredRegions` expanded to include all region-bearing layouts

```go
knownLayouts = map[string]bool{
    "default": true, "title": true, "section": true,
    "two-column": true, "code-preview": true,
    "three-column": true, "image-left": true, "image-right": true,
    "quote": true, "split-heading": true, "top-bottom": true, "blank": true,
}
futureLayouts = map[string]bool{"grid-cards": true}
```

### CSS additions to `layouts.css`

```css
/* three-column */
section[data-layout="three-column"] .slide-body {
  display: grid;
  grid-template-columns: 1fr 1fr 1fr;
  gap: 1.5rem;
  width: 100%;
}
section[data-layout="three-column"] .region-col1,
section[data-layout="three-column"] .region-col2,
section[data-layout="three-column"] .region-col3 { text-align: left; }

/* image-left */
section[data-layout="image-left"] .slide-body {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 2rem;
  width: 100%;
  align-items: center;
}
section[data-layout="image-left"] .region-image img {
  width: 100%; height: auto; border-radius: 0.5rem;
}
section[data-layout="image-left"] .region-text { text-align: left; }

/* image-right */
section[data-layout="image-right"] .slide-body {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 2rem;
  width: 100%;
  align-items: center;
}
section[data-layout="image-right"] .region-image img {
  width: 100%; height: auto; border-radius: 0.5rem;
}
section[data-layout="image-right"] .region-text { text-align: left; }

/* quote */
section[data-layout="quote"] {
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  text-align: center;
}
section[data-layout="quote"] blockquote {
  font-size: 1.3em;
  border-left: none;
  font-style: italic;
  max-width: 80%;
}
section[data-layout="quote"] blockquote p:last-child {
  font-size: 0.7em;
  font-style: normal;
  color: var(--slide-muted);
  margin-top: 1rem;
}

/* split-heading */
section[data-layout="split-heading"] .slide-body {
  display: grid;
  grid-template-columns: 2fr 3fr;
  gap: 2rem;
  width: 100%;
  align-items: start;
}
section[data-layout="split-heading"] .region-heading h1,
section[data-layout="split-heading"] .region-heading h2 {
  font-size: 2em;
  line-height: 1.2;
}
section[data-layout="split-heading"] .region-body { text-align: left; }

/* top-bottom */
section[data-layout="top-bottom"] .slide-body {
  display: grid;
  grid-template-rows: 1fr auto;
  gap: 1.5rem;
  width: 100%;
  height: 100%;
}
section[data-layout="top-bottom"] .region-top { text-align: center; }
section[data-layout="top-bottom"] .region-bottom { text-align: left; }

/* blank */
section[data-layout="blank"] {
  padding: 0;
}
```

---

## 4. Three New Themes

All themes follow the same CSS variable structure as `default.css`/`dark.css`.

### `corporate.css`

Formal, structured, muted palette. Default accent: blue.

```css
:root {
  --slide-bg:        #f5f5f0;
  --slide-text:      #2d2d2d;
  --slide-heading:   #1a1a1a;
  --slide-code-bg:   #e8e8e3;
  --slide-code-text: #2d2d2d;
  --slide-border:    rgba(0, 0, 0, 0.12);
  --slide-muted:     #777777;
  --slide-card-bg:   #eaeae5;
}
```

### `minimal.css`

Maximum whitespace, large type, sparse. Default accent: blue.

```css
:root {
  --slide-bg:        #ffffff;
  --slide-text:      #333333;
  --slide-heading:   #111111;
  --slide-code-bg:   #fafafa;
  --slide-code-text: #333333;
  --slide-border:    rgba(0, 0, 0, 0.06);
  --slide-muted:     #999999;
  --slide-card-bg:   #fafafa;
}
```

Overrides: `h1` font-size bumped to 2.2em.

### `hacker.css`

Terminal aesthetic, monospace font, green on black. Default accent: green.

```css
:root {
  --slide-bg:        #0a0a0a;
  --slide-text:      #00ff00;
  --slide-heading:   #00ff00;
  --slide-code-bg:   #0d0d0d;
  --slide-code-text: #00ff00;
  --slide-border:    rgba(0, 255, 0, 0.15);
  --slide-muted:     #007700;
  --slide-card-bg:   #111111;
}
```

Overrides: `.reveal` font-family set to `var(--font-mono)`, optional text-shadow glow.

### Theme code changes

**`internal/theme/theme.go`:**
- `validThemes` expanded: add `corporate`, `minimal`, `hacker`
- `ResolveAccent` signature changed to `ResolveAccent(accent, theme string) string`
- Per-theme default accents via `themeDefaultAccents` map

```go
var themeDefaultAccents = map[string]string{
    "default": "blue", "dark": "blue", "corporate": "blue",
    "minimal": "blue", "hacker": "green",
}

func ResolveAccent(accent, theme string) string {
    if accent != "" {
        return accent
    }
    if def, ok := themeDefaultAccents[theme]; ok {
        return def
    }
    return "blue"
}
```

**`internal/ir/validate.go`:** `validThemes` synced with theme package.

**`internal/renderer/renderer.go`:** `theme.ResolveAccent` call updated to pass theme name.

**Each theme CSS includes:**
- All `.reveal` element styles (same as default.css/dark.css)
- `.reveal .controls { color: var(--slide-accent); }`
- `.reveal .progress { color: var(--slide-accent); }`
- `#goslide-page-num` with theme-appropriate color

---

## 5. Testing Strategy

### Unit tests

| Package | New tests |
|---------|-----------|
| `parser` | `extractComponents`: basic extraction, nested fences, unknown component fallback, YAML parse error, multiple components in one slide, component inside region |
| `parser` | `parseSlide` with components: placeholder position, component count, params populated |
| `ir/validate` | known layouts no longer warn; `grid-cards` still warns as future; unknown layout still warns; new required regions validated |
| `theme` | `ResolveAccent` with theme-specific defaults; 5 themes valid; `ThemeCSSPath` for new themes |

### Golden tests

- Add `three-column.md`, `quote.md` golden test inputs
- Update `all-typos.md` to test new validation (future-layout for grid-cards only)

### Manual checklist additions

- All 6 new layouts render correctly
- 3 new themes display properly with correct default accents
- Component fence extraction: verify `~~~chart:bar` doesn't render as code block (shows placeholder or nothing in Phase 2a)

---

## 6. Files Changed Summary

| Action | File |
|--------|------|
| Modify | `internal/ir/presentation.go` — add `Component` type, add `Components` field to `Slide` |
| Modify | `internal/parser/slide.go` — add `extractComponents()`, update `parseSlide` flow |
| Create | `internal/parser/component.go` — `extractComponents()` implementation |
| Create | `internal/parser/component_test.go` — extraction tests |
| Modify | `internal/parser/slide_test.go` — tests with components |
| Modify | `internal/ir/validate.go` — update layout whitelists, remove future-component for known types |
| Modify | `internal/ir/validate_test.go` — new layout validation tests |
| Modify | `internal/theme/theme.go` — add themes, change `ResolveAccent` signature |
| Modify | `internal/theme/theme_test.go` — tests for new themes + accent defaults |
| Modify | `internal/renderer/renderer.go` — pass theme to `ResolveAccent` |
| Modify | `web/themes/layouts.css` — 6 new layout CSS rules |
| Create | `web/themes/corporate.css` |
| Create | `web/themes/minimal.css` |
| Create | `web/themes/hacker.css` |
| Modify | golden test files — update + add new fixtures |
