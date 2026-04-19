# Placeholder Component + `image-grid` Layout Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a `~~~placeholder` component (styled image stand-in with icon + hint + body) and a new `image-grid` slide layout that packs multiple cells (placeholders, images, components) into 2/3/4 columns.

**Architecture:** Thin Go changes (register the new component prefix, register the new layout, support repeated region markers, run card-style body splitting for placeholder params) paired with client-side JS that renders the visual treatment, plus CSS rules for the layout grid. Reuses every existing extension point — no new files in `internal/`.

**Tech Stack:** Go 1.21.6, `net/http` (none needed here), `gopkg.in/yaml.v3`, `github.com/stretchr/testify`. Client-side: vanilla JS in `web/static/components.js`. CSS added to `web/themes/layouts.css`.

**Spec:** `docs/superpowers/specs/2026-04-19-placeholder-multi-image-design.md`

---

## File Structure

**Modify:**
- `internal/parser/component.go` — add `"placeholder"` to `knownComponentPrefixes`
- `internal/parser/slide.go` — add `"image-grid": {"cell"}` to `layoutRegions`; change internal region tracker to support repeated markers; add placeholder body-splitting analogous to card
- `internal/ir/validate.go` — add `"image-grid"` to `knownLayouts`; add validation for placeholder `hint`/`aspect` and image-grid `columns`/empty-cells
- `internal/ir/validate_test.go` — validation test cases
- `internal/parser/component_test.go` — placeholder fence parse tests
- `internal/parser/slide_test.go` — image-grid + cell parse tests
- `internal/generate/system_prompt.md` — new Layouts entry, Components subsection, Rule
- `web/static/components.js` — `initPlaceholder(el)` and dispatch in `initAllComponents`
- `web/themes/layouts.css` — `section[data-layout="image-grid"]` + `.goslide-component[data-type="placeholder"]` rules
- `internal/renderer/renderer_test.go` — placeholder data-attribute + image-grid data-columns assertions
- `PRD.md` — tick new feature, if relevant

**Create:**
- `internal/renderer/testdata/golden/placeholder-basic.md` + `.html`
- `internal/renderer/testdata/golden/image-grid-dark.md` + `.html`

---

## Task 1: IR — register `image-grid` layout

**Files:**
- Modify: `internal/ir/validate.go`
- Modify: `internal/ir/validate_test.go`

- [ ] **Step 1: Write failing test**

Append to `internal/ir/validate_test.go`:

```go
func TestValidate_ImageGridIsKnown(t *testing.T) {
	p := Presentation{Slides: []Slide{{Index: 1, Meta: SlideMeta{Layout: "image-grid"}}}}
	errs := p.Validate()
	require.Nil(t, findError(errs, "unknown-layout"))
	require.Nil(t, findError(errs, "future-layout"))
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `GOTOOLCHAIN=local go test -C . ./internal/ir/ -run TestValidate_ImageGridIsKnown -v`
Expected: FAIL (unknown-layout error is produced since `image-grid` isn't in `knownLayouts`).

- [ ] **Step 3: Implement**

Modify `internal/ir/validate.go`, add `"image-grid": true,` to the `knownLayouts` map:

```go
knownLayouts = map[string]bool{
    "default": true, "title": true, "section": true,
    "two-column": true, "code-preview": true,
    "three-column": true, "image-left": true, "image-right": true,
    "quote": true, "split-heading": true, "top-bottom": true, "blank": true,
    "grid-cards": true, "image-grid": true,
}
```

- [ ] **Step 4: Run the test**

Run: `GOTOOLCHAIN=local go test -C . ./internal/ir/ -v`
Expected: all pass.

- [ ] **Step 5: Commit**

```bash
git -C D:/CLAUDE-CODE-GOSLIDE add internal/ir/validate.go internal/ir/validate_test.go
```

```bash
git -C D:/CLAUDE-CODE-GOSLIDE commit -m "feat(ir): register image-grid as a known layout"
```

---

## Task 2: IR — `image-grid` columns + empty-cells validation

**Files:**
- Modify: `internal/ir/validate.go`
- Modify: `internal/ir/validate_test.go`

- [ ] **Step 1: Write failing tests**

Append to `internal/ir/validate_test.go`:

```go
func TestValidate_ImageGridColumnsOutOfRange(t *testing.T) {
	p := Presentation{Slides: []Slide{{
		Index: 1,
		Meta:  SlideMeta{Layout: "image-grid", Columns: 5},
		Regions: []Region{{Name: "cell", HTML: "x"}},
	}}}
	errs := p.Validate()
	e := findError(errs, "columns-out-of-range")
	require.NotNil(t, e, "expected columns-out-of-range warning")
	require.Equal(t, "warning", e.Severity)
}

func TestValidate_ImageGridEmpty(t *testing.T) {
	p := Presentation{Slides: []Slide{{
		Index: 1,
		Meta:  SlideMeta{Layout: "image-grid", Columns: 2},
	}}}
	errs := p.Validate()
	e := findError(errs, "image-grid-empty")
	require.NotNil(t, e)
	require.Equal(t, "warning", e.Severity)
}

func TestValidate_ImageGridHappy(t *testing.T) {
	p := Presentation{Slides: []Slide{{
		Index: 1,
		Meta:  SlideMeta{Layout: "image-grid", Columns: 2},
		Regions: []Region{
			{Name: "cell", HTML: "a"},
			{Name: "cell", HTML: "b"},
		},
	}}}
	errs := p.Validate()
	require.Nil(t, findError(errs, "columns-out-of-range"))
	require.Nil(t, findError(errs, "image-grid-empty"))
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `GOTOOLCHAIN=local go test -C . ./internal/ir/ -run TestValidate_ImageGrid -v`
Expected: FAIL (no image-grid validation implemented yet).

- [ ] **Step 3: Implement**

Inside `Validate()` in `internal/ir/validate.go`, find the `for _, slide := range p.Slides` loop (around line 79). After the existing `validateSlide(slide)` call, add a helper call. Add at the end of `validate.go`:

```go
func validateImageGrid(s Slide) []Error {
	if s.Meta.Layout != "image-grid" {
		return nil
	}
	var errs []Error
	if s.Meta.Columns != 0 && (s.Meta.Columns < 2 || s.Meta.Columns > 4) {
		errs = append(errs, Error{
			Slide: s.Index, Severity: "warning", Code: "columns-out-of-range",
			Message: fmt.Sprintf("slide %d: columns %d out of range (2-4); clamping to 2", s.Index, s.Meta.Columns),
		})
	}
	cellCount := 0
	for _, r := range s.Regions {
		if r.Name == "cell" {
			cellCount++
		}
	}
	if cellCount == 0 {
		errs = append(errs, Error{
			Slide: s.Index, Severity: "warning", Code: "image-grid-empty",
			Message: fmt.Sprintf("slide %d: image-grid layout has no cells", s.Index),
		})
	}
	return errs
}
```

And wire it in. In `Validate()`, change the loop:

```go
for _, slide := range p.Slides {
    errs = append(errs, validateSlide(slide)...)
    errs = append(errs, validateImageGrid(slide)...)
}
```

- [ ] **Step 4: Run tests**

Run: `GOTOOLCHAIN=local go test -C . ./internal/ir/ -v`
Expected: all pass.

- [ ] **Step 5: Commit**

```bash
git -C D:/CLAUDE-CODE-GOSLIDE add internal/ir/validate.go internal/ir/validate_test.go
```

```bash
git -C D:/CLAUDE-CODE-GOSLIDE commit -m "feat(ir): validate image-grid columns and empty cells"
```

---

## Task 3: IR — placeholder validation (hint required, aspect whitelist)

**Files:**
- Modify: `internal/ir/validate.go`
- Modify: `internal/ir/validate_test.go`

- [ ] **Step 1: Write failing tests**

Append to `internal/ir/validate_test.go`:

```go
func TestValidate_PlaceholderMissingHint(t *testing.T) {
	p := Presentation{Slides: []Slide{{
		Index: 1,
		Components: []Component{{Index: 0, Type: "placeholder", Params: map[string]any{"icon": "🗺️"}}},
	}}}
	errs := p.Validate()
	e := findError(errs, "placeholder-missing-hint")
	require.NotNil(t, e)
	require.Equal(t, "error", e.Severity)
}

func TestValidate_PlaceholderUnknownAspect(t *testing.T) {
	p := Presentation{Slides: []Slide{{
		Index: 1,
		Components: []Component{{Index: 0, Type: "placeholder", Params: map[string]any{"hint": "x", "aspect": "2:1"}}},
	}}}
	errs := p.Validate()
	e := findError(errs, "unknown-aspect")
	require.NotNil(t, e)
	require.Equal(t, "warning", e.Severity)
}

func TestValidate_PlaceholderHappy(t *testing.T) {
	p := Presentation{Slides: []Slide{{
		Index: 1,
		Components: []Component{{Index: 0, Type: "placeholder", Params: map[string]any{"hint": "K8s", "aspect": "16:9", "icon": "🗺️"}}},
	}}}
	errs := p.Validate()
	require.Nil(t, findError(errs, "placeholder-missing-hint"))
	require.Nil(t, findError(errs, "unknown-aspect"))
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `GOTOOLCHAIN=local go test -C . ./internal/ir/ -run TestValidate_Placeholder -v`
Expected: FAIL.

- [ ] **Step 3: Implement**

At the end of `internal/ir/validate.go`, add:

```go
var validAspects = map[string]bool{
	"16:9": true, "4:3": true, "1:1": true, "3:4": true, "9:16": true,
}

func validatePlaceholder(s Slide) []Error {
	var errs []Error
	for _, c := range s.Components {
		if c.Type != "placeholder" {
			continue
		}
		hint, _ := c.Params["hint"].(string)
		if strings.TrimSpace(hint) == "" {
			errs = append(errs, Error{
				Slide: s.Index, Severity: "error", Code: "placeholder-missing-hint",
				Message: fmt.Sprintf("slide %d: placeholder component requires 'hint' field", s.Index),
			})
		}
		if aspect, ok := c.Params["aspect"].(string); ok && aspect != "" && !validAspects[aspect] {
			errs = append(errs, Error{
				Slide: s.Index, Severity: "warning", Code: "unknown-aspect",
				Message: fmt.Sprintf("slide %d: aspect %q not recognized (using 16:9)", s.Index, aspect),
			})
		}
	}
	return errs
}
```

Add `"strings"` to the imports if not already there (it should be).

Wire it in the same `Validate()` loop:

```go
for _, slide := range p.Slides {
    errs = append(errs, validateSlide(slide)...)
    errs = append(errs, validateImageGrid(slide)...)
    errs = append(errs, validatePlaceholder(slide)...)
}
```

- [ ] **Step 4: Run tests**

Run: `GOTOOLCHAIN=local go test -C . ./internal/ir/ -v`
Expected: all pass.

- [ ] **Step 5: Commit**

```bash
git -C D:/CLAUDE-CODE-GOSLIDE add internal/ir/validate.go internal/ir/validate_test.go
```

```bash
git -C D:/CLAUDE-CODE-GOSLIDE commit -m "feat(ir): validate placeholder hint + aspect whitelist"
```

---

## Task 4: Parser — register `placeholder` component

**Files:**
- Modify: `internal/parser/component.go`
- Modify: `internal/parser/component_test.go`

- [ ] **Step 1: Write failing test**

Append to `internal/parser/component_test.go`:

```go
func TestExtractComponents_Placeholder(t *testing.T) {
	body := "~~~placeholder\nhint: K8s architecture\nicon: 🗺️\naspect: 16:9\n~~~\n"
	_, comps := extractComponents(body)
	require.Len(t, comps, 1)
	require.Equal(t, "placeholder", comps[0].Type)
	require.Equal(t, "K8s architecture", comps[0].Params["hint"])
	require.Equal(t, "🗺️", comps[0].Params["icon"])
	require.Equal(t, "16:9", comps[0].Params["aspect"])
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `GOTOOLCHAIN=local go test -C . ./internal/parser/ -run TestExtractComponents_Placeholder -v`
Expected: FAIL (the `~~~placeholder` fence isn't recognised, so the block is kept as raw text and `comps` is empty).

- [ ] **Step 3: Implement**

Modify `internal/parser/component.go`, add `"placeholder": true,` to `knownComponentPrefixes`:

```go
var knownComponentPrefixes = map[string]bool{
    "chart": true, "mermaid": true, "table": true,
    "tabs": true, "panel": true, "slider": true,
    "toggle": true, "api": true, "embed": true, "card": true,
    "placeholder": true,
}
```

- [ ] **Step 4: Run the tests**

Run: `GOTOOLCHAIN=local go test -C . ./internal/parser/ -v`
Expected: all pass.

- [ ] **Step 5: Commit**

```bash
git -C D:/CLAUDE-CODE-GOSLIDE add internal/parser/component.go internal/parser/component_test.go
```

```bash
git -C D:/CLAUDE-CODE-GOSLIDE commit -m "feat(parser): recognize ~~~placeholder component"
```

---

## Task 5: Parser — placeholder body split (params vs description)

**Files:**
- Modify: `internal/parser/slide.go`
- Modify: `internal/parser/component_test.go`

Context: For `card`, `slide.go` splits the fence body at the first `---\n` into YAML params (before) and Markdown body (after), storing the rendered body in `ContentHTML`. Placeholder needs the same so the optional subtitle text is available to the client JS.

- [ ] **Step 1: Write failing test**

Append to `internal/parser/component_test.go`:

```go
func TestExtractComponents_PlaceholderWithBody(t *testing.T) {
	// Actual splitting happens in slide.parseSlide, not extractComponents.
	// Here we just verify the raw body is captured; Task 5 verifies the split.
	body := "~~~placeholder\nhint: Cluster\n---\nsubtitle text\n~~~\n"
	_, comps := extractComponents(body)
	require.Len(t, comps, 1)
	require.Contains(t, comps[0].Raw, "subtitle text")
}
```

Add a slide-level test in `internal/parser/slide_test.go`:

```go
func TestParseSlide_PlaceholderBodySplit(t *testing.T) {
	raw := "~~~placeholder\nhint: K8s\nicon: 🗺️\n---\nControl plane detail\n~~~\n"
	slide := parseSlide(0, raw, ir.Frontmatter{})
	require.Len(t, slide.Components, 1)
	c := slide.Components[0]
	require.Equal(t, "placeholder", c.Type)
	require.Equal(t, "K8s", c.Params["hint"])
	require.Contains(t, string(c.ContentHTML), "Control plane detail")
}
```

- [ ] **Step 2: Run tests**

Run: `GOTOOLCHAIN=local go test -C . ./internal/parser/ -run 'TestExtractComponents_PlaceholderWithBody|TestParseSlide_PlaceholderBodySplit' -v`
Expected: the first test passes (captured in Raw), the second fails (body not split yet).

- [ ] **Step 3: Implement**

In `internal/parser/slide.go`, locate the existing card body-split block (around lines 70-80 in the current file, the loop that iterates `components` after `extractComponents`). It looks like:

```go
for i := range components {
    if strings.HasPrefix(components[i].Type, "panel:") {
        // ...
    } else if strings.HasPrefix(components[i].Type, "card") {
        parts := strings.SplitN(components[i].Raw, "---\n", 2)
        if len(parts) == 2 {
            var summaryParams map[string]any
            if err := yaml.Unmarshal([]byte(parts[0]), &summaryParams); err == nil {
                components[i].Params = summaryParams
            }
            components[i].ContentHTML = string(renderMarkdown(parts[1]))
        }
    } else if components[i].Type == "embed:html" {
        components[i].ContentHTML = components[i].Raw
    }
}
```

Add a new branch for placeholder **before** the existing `embed:html` branch:

```go
} else if components[i].Type == "placeholder" {
    parts := strings.SplitN(components[i].Raw, "---\n", 2)
    if len(parts) == 2 {
        var ph map[string]any
        if err := yaml.Unmarshal([]byte(parts[0]), &ph); err == nil {
            components[i].Params = ph
        }
        components[i].ContentHTML = string(renderMarkdown(parts[1]))
    }
    // If no `---` body separator, Params is already set by extractComponents
    // and ContentHTML remains empty (renderer substitutes default text).
} else if components[i].Type == "embed:html" {
```

- [ ] **Step 4: Run tests**

Run: `GOTOOLCHAIN=local go test -C . ./internal/parser/ -v`
Expected: all pass.

- [ ] **Step 5: Commit**

```bash
git -C D:/CLAUDE-CODE-GOSLIDE add internal/parser/slide.go internal/parser/component_test.go internal/parser/slide_test.go
```

```bash
git -C D:/CLAUDE-CODE-GOSLIDE commit -m "feat(parser): split placeholder body into params + description HTML"
```

---

## Task 6: Parser — support repeated region markers for `image-grid`

**Files:**
- Modify: `internal/parser/slide.go`
- Modify: `internal/parser/slide_test.go`

Context: The parser currently dedupes regions into a `map[string][]string` keyed by name. `image-grid` needs multiple `<!-- cell -->` markers to each create a distinct `ir.Region`. Solution: drop the map, use an ordered slice of `(name, lines)` pairs — each marker hit always starts a new entry.

- [ ] **Step 1: Write failing test**

Append to `internal/parser/slide_test.go`:

```go
func TestParseSlide_ImageGridCells(t *testing.T) {
	raw := "<!-- layout: image-grid -->\n<!-- columns: 2 -->\n\n<!-- cell -->\n\n~~~placeholder\nhint: A\n~~~\n\n<!-- cell -->\n\n~~~placeholder\nhint: B\n~~~\n\n<!-- cell -->\n\nplain text\n\n<!-- cell -->\n\n![alt](./img.png)\n"
	slide := parseSlide(0, raw, ir.Frontmatter{})
	require.Equal(t, "image-grid", slide.Meta.Layout)
	require.Equal(t, 2, slide.Meta.Columns)
	require.Len(t, slide.Regions, 4, "four cells expected, got %d", len(slide.Regions))
	for _, r := range slide.Regions {
		require.Equal(t, "cell", r.Name)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `GOTOOLCHAIN=local go test -C . ./internal/parser/ -run TestParseSlide_ImageGridCells -v`
Expected: FAIL — no cells because `image-grid` isn't in `layoutRegions`, and even if it were, repeated `<!-- cell -->` would all collapse to one region.

- [ ] **Step 3: Implement — step 1: add the layout**

In `internal/parser/slide.go`, add `"image-grid": {"cell"},` to `layoutRegions`:

```go
var layoutRegions = map[string][]string{
    "two-column":    {"left", "right"},
    "code-preview":  {"code", "preview"},
    "three-column":  {"col1", "col2", "col3"},
    "image-left":    {"image", "text"},
    "image-right":   {"text", "image"},
    "split-heading": {"heading", "body"},
    "top-bottom":    {"top", "bottom"},
    "image-grid":    {"cell"},
}
```

- [ ] **Step 4: Implement — step 2: rewrite region tracker to allow repeats**

In the same file, locate the region-collection block (currently uses `regionContent map[string][]string` + `regionOrder []string`). Replace with an ordered slice of entries:

```go
type regionCursor struct {
    name  string
    lines []string
}
var cursors []regionCursor
currentIdx := -1

for _, line := range bodyLines {
    trimmed := strings.TrimSpace(line)
    if m := regionRe.FindStringSubmatch(trimmed); m != nil {
        name := strings.ToLower(m[1])
        if validSet[name] {
            cursors = append(cursors, regionCursor{name: name})
            currentIdx = len(cursors) - 1
            continue
        }
    }
    if currentIdx == -1 {
        mainLines = append(mainLines, line)
    } else {
        cursors[currentIdx].lines = append(cursors[currentIdx].lines, line)
    }
}

// then:
for _, c := range cursors {
    regions = append(regions, ir.Region{
        Name: c.name,
        HTML: renderMarkdown(strings.Join(c.lines, "\n")),
    })
}
```

Remove the now-unused `regionContent`, `regionOrder`, and `currentRegion` variables.

- [ ] **Step 5: Run all parser tests**

Run: `GOTOOLCHAIN=local go test -C . ./internal/parser/ -v`
Expected: all pass (existing region tests AND new image-grid test).

If any existing test relied on duplicate-region-name deduplication, it should still work because legitimate layouts (`two-column`, etc.) never emit the same marker twice. If a test breaks, fix the test to reflect the new ordered-slice semantics (each marker = new region).

- [ ] **Step 6: Commit**

```bash
git -C D:/CLAUDE-CODE-GOSLIDE add internal/parser/slide.go internal/parser/slide_test.go
```

```bash
git -C D:/CLAUDE-CODE-GOSLIDE commit -m "feat(parser): image-grid layout + repeatable cell markers"
```

---

## Task 7: Renderer — verify placeholder + image-grid render correctly

**Files:**
- Modify: `internal/renderer/renderer_test.go`

The Go renderer already handles these cases without code changes: placeholder is routed through the generic `buildComponentDiv` (emits `data-type="placeholder"` + `data-params=...`), and image-grid regions are rendered by the existing template loop (`<div class="region-cell">...</div>` repeated per cell). We only need to add assertions.

- [ ] **Step 1: Write failing tests**

Append to `internal/renderer/renderer_test.go`:

```go
func TestRender_Placeholder(t *testing.T) {
	pres := &ir.Presentation{
		Meta: ir.Frontmatter{Theme: "dark"},
		Slides: []ir.Slide{{
			Index: 0,
			Meta:  ir.SlideMeta{Layout: "default"},
			Components: []ir.Component{{
				Index:  0,
				Type:   "placeholder",
				Params: map[string]any{"hint": "K8s architecture", "icon": "🗺️", "aspect": "16:9"},
			}},
			BodyHTML: `<!--goslide:component:0-->`,
		}},
	}
	html, err := Render(pres)
	require.NoError(t, err)
	require.Contains(t, html, `data-type="placeholder"`)
	require.Contains(t, html, `K8s architecture`)
}

func TestRender_ImageGrid(t *testing.T) {
	pres := &ir.Presentation{
		Meta: ir.Frontmatter{Theme: "dark"},
		Slides: []ir.Slide{{
			Index: 0,
			Meta:  ir.SlideMeta{Layout: "image-grid", Columns: 2},
			Regions: []ir.Region{
				{Name: "cell", HTML: "<p>A</p>"},
				{Name: "cell", HTML: "<p>B</p>"},
			},
		}},
	}
	html, err := Render(pres)
	require.NoError(t, err)
	require.Contains(t, html, `data-layout="image-grid"`)
	require.Contains(t, html, `data-columns="2"`)
	// two separate region-cell divs
	require.Equal(t, 2, strings.Count(html, `class="region-cell"`))
}
```

Add `"strings"` import if missing.

- [ ] **Step 2: Run tests**

Run: `GOTOOLCHAIN=local go test -C . ./internal/renderer/ -run 'TestRender_Placeholder|TestRender_ImageGrid' -v`
Expected: PASS on first try — no code change required.

If a test fails, the reason must be a pre-existing gap in the renderer; fix it minimally there.

- [ ] **Step 3: Commit**

```bash
git -C D:/CLAUDE-CODE-GOSLIDE add internal/renderer/renderer_test.go
```

```bash
git -C D:/CLAUDE-CODE-GOSLIDE commit -m "test(renderer): cover placeholder + image-grid emission"
```

---

## Task 8: Golden test — placeholder-basic

**Files:**
- Create: `internal/renderer/testdata/golden/placeholder-basic.md`
- Create: `internal/renderer/testdata/golden/placeholder-basic.html`
- Modify: `internal/renderer/golden_test.go` (add fixture to the test loop if it's list-driven)

- [ ] **Step 1: Inspect the golden test harness**

Read `internal/renderer/golden_test.go`. It iterates fixtures in `testdata/golden/` — confirm how fixtures are paired (typically `<name>.md` → `<name>.html`). Adapt the rest of this task to match.

- [ ] **Step 2: Create the Markdown input**

Create `internal/renderer/testdata/golden/placeholder-basic.md`:

````markdown
---
title: Placeholder Demo
theme: dark
---

# Placeholder Demo

~~~placeholder
hint: K8s architecture
icon: 🗺️
aspect: 16:9
---
Control plane + worker interaction
~~~
````

- [ ] **Step 3: Run the golden test once with UPDATE mode to capture the output**

Run: `GOTOOLCHAIN=local UPDATE_GOLDEN=1 go test -C . ./internal/renderer/ -run Golden -v`
(If the golden harness uses a different update flag, inspect `golden_test.go` and use that.)

Expected: test writes `placeholder-basic.html` next to the `.md`.

- [ ] **Step 4: Inspect the generated .html**

Read `internal/renderer/testdata/golden/placeholder-basic.html` and verify it contains:
- `<section ... data-layout="default">` (or similar)
- `<div class="goslide-component" data-type="placeholder" data-params="..."`
- The `K8s architecture` hint appears inside `data-params`
- Body HTML `Control plane + worker interaction` is rendered

If anything looks wrong, fix the underlying code — do not hand-edit the golden.

- [ ] **Step 5: Run golden test without UPDATE**

Run: `GOTOOLCHAIN=local go test -C . ./internal/renderer/ -run Golden -v`
Expected: all golden comparisons pass.

- [ ] **Step 6: Commit**

```bash
git -C D:/CLAUDE-CODE-GOSLIDE add internal/renderer/testdata/golden/placeholder-basic.md internal/renderer/testdata/golden/placeholder-basic.html
```

```bash
git -C D:/CLAUDE-CODE-GOSLIDE commit -m "test(renderer): golden fixture for placeholder-basic"
```

---

## Task 9: Golden test — image-grid-dark

**Files:**
- Create: `internal/renderer/testdata/golden/image-grid-dark.md`
- Create: `internal/renderer/testdata/golden/image-grid-dark.html`

- [ ] **Step 1: Create the Markdown input**

Create `internal/renderer/testdata/golden/image-grid-dark.md`:

````markdown
---
title: Image Grid Demo
theme: dark
---

<!-- layout: image-grid -->
<!-- columns: 2 -->

<!-- cell -->

~~~placeholder
hint: Architecture diagram
icon: 🗺️
~~~

<!-- cell -->

~~~placeholder
hint: Performance chart
icon: 📊
~~~

<!-- cell -->

Plain text cell.

<!-- cell -->

![Fallback](./img.png)
````

- [ ] **Step 2: Generate golden**

Run: `GOTOOLCHAIN=local UPDATE_GOLDEN=1 go test -C . ./internal/renderer/ -run Golden -v`
(Use whatever update flag the harness defines.)

- [ ] **Step 3: Inspect the generated .html**

Read `internal/renderer/testdata/golden/image-grid-dark.html` and verify:
- `<section ... data-layout="image-grid" data-columns="2">`
- Exactly **four** `<div class="region-cell">` elements
- Two cells contain `data-type="placeholder"` with the respective hints
- One cell contains the text "Plain text cell."
- One cell contains the Markdown image as `<img src="./img.png"`

If any of these is missing, the parser or renderer has a bug — fix it and regenerate.

- [ ] **Step 4: Run golden test without UPDATE**

Run: `GOTOOLCHAIN=local go test -C . ./internal/renderer/ -run Golden -v`
Expected: all pass.

- [ ] **Step 5: Commit**

```bash
git -C D:/CLAUDE-CODE-GOSLIDE add internal/renderer/testdata/golden/image-grid-dark.md internal/renderer/testdata/golden/image-grid-dark.html
```

```bash
git -C D:/CLAUDE-CODE-GOSLIDE commit -m "test(renderer): golden fixture for image-grid-dark"
```

---

## Task 10: Client JS — initPlaceholder

**Files:**
- Modify: `web/static/components.js`

Context: The Go renderer emits `<div class="goslide-component" data-type="placeholder" data-params='{"hint":"...","icon":"...","aspect":"..."}' data-raw="" data-comp-id="s0-c0">{BODY_HTML}</div>`. The client JS must reshape this into the visual placeholder on page load.

- [ ] **Step 1: Read the dispatch site**

Open `web/static/components.js` around line 390 (`initAllComponents`) to see the existing dispatch pattern (chart / table / embed:iframe / api). You'll add a `placeholder` branch.

- [ ] **Step 2: Add `initPlaceholder` function**

Insert this function anywhere among the other `init*` functions (above `initAllComponents`):

```javascript
function initPlaceholder(el) {
    var params = {};
    try { params = JSON.parse(el.getAttribute('data-params') || '{}'); } catch (e) {}
    var hint = params.hint || 'Image placeholder';
    var icon = params.icon || '🖼️';
    var aspect = params.aspect || '16:9';
    var bodyHTML = el.innerHTML; // rendered Markdown from the fence body, may be empty
    var body = bodyHTML && bodyHTML.trim() ? bodyHTML : '<em>Replace with actual content</em>';

    var aspectParts = aspect.split(':');
    var aspectCSS = (aspectParts.length === 2) ? (aspectParts[0] + '/' + aspectParts[1]) : '16/9';

    el.classList.add('gs-placeholder');
    el.setAttribute('data-aspect', aspect);
    el.style.aspectRatio = aspectCSS;
    el.innerHTML =
        '<div class="gs-placeholder-icon">' + icon + '</div>' +
        '<div class="gs-placeholder-hint">' + escapeText(hint) + '</div>' +
        '<div class="gs-placeholder-body">' + body + '</div>';
}

function escapeText(s) {
    var div = document.createElement('div');
    div.textContent = s;
    return div.innerHTML;
}
```

- [ ] **Step 3: Wire into the dispatcher**

In `initAllComponents`, add a branch:

```javascript
else if (type === 'placeholder') initPlaceholder(el);
```

Place it after the `embed:iframe` branch and before the `api` branch — order doesn't matter functionally.

- [ ] **Step 4: Smoke test**

Build the binary:
```bash
GOTOOLCHAIN=local go build -C . ./cmd/goslide
```

Create a quick sanity file in a scratch directory and run `goslide serve` manually — open the page and confirm the placeholder shows the dashed box with icon + hint + body. This step is a visual check; commit only after passing.

Note: since there is no automated JS test harness in this repo, the golden test covers the HTML-structure side and the manual check covers the visual side.

- [ ] **Step 5: Commit**

```bash
git -C D:/CLAUDE-CODE-GOSLIDE add web/static/components.js
```

```bash
git -C D:/CLAUDE-CODE-GOSLIDE commit -m "feat(web): client-side placeholder renderer"
```

---

## Task 11: CSS — placeholder + image-grid styles

**Files:**
- Modify: `web/themes/layouts.css`

- [ ] **Step 1: Append image-grid layout rules**

At the end of `web/themes/layouts.css`, append:

```css
/* Layout: image-grid */
section[data-layout="image-grid"] .slide-body {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 1rem;
  width: 100%;
  align-items: stretch;
}
section[data-layout="image-grid"][data-columns="3"] .slide-body {
  grid-template-columns: repeat(3, 1fr);
}
section[data-layout="image-grid"][data-columns="4"] .slide-body {
  grid-template-columns: repeat(4, 1fr);
}
section[data-layout="image-grid"] .region-cell {
  min-width: 0;
  display: flex;
  align-items: center;
  justify-content: center;
}
section[data-layout="image-grid"] .region-cell img {
  max-width: 100%;
  height: auto;
  object-fit: cover;
}
```

- [ ] **Step 2: Append placeholder component rules**

At the end of the same file, append:

```css
/* Component: placeholder */
.goslide-component[data-type="placeholder"].gs-placeholder {
  border: 2px dashed var(--slide-accent, #8888aa);
  border-radius: 8px;
  padding: clamp(16px, 3vw, 32px);
  background: color-mix(in srgb, var(--slide-accent, #8888aa) 8%, transparent);
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
```

Note: `--slide-accent` is the existing accent token used elsewhere in the codebase (observed in `components.js`), not `--gs-accent`. Using the existing token means placeholder borders pick up the current theme's accent.

- [ ] **Step 3: Smoke test visually**

Build, serve a sample deck containing both a single-placeholder slide and an image-grid slide, open in the browser and confirm:
- Dashed border renders with the theme's accent colour
- Grid shows 2 columns by default, 3 or 4 when specified
- Aspect ratio of placeholders matches the declared ratio

- [ ] **Step 4: Commit**

```bash
git -C D:/CLAUDE-CODE-GOSLIDE add web/themes/layouts.css
```

```bash
git -C D:/CLAUDE-CODE-GOSLIDE commit -m "feat(web): CSS for placeholder component and image-grid layout"
```

---

## Task 12: Update LLM system prompt

**Files:**
- Modify: `internal/generate/system_prompt.md`
- Modify: `internal/generate/embed_test.go`

- [ ] **Step 1: Write failing keyword test**

Append to `internal/generate/embed_test.go`:

```go
func TestSystemPrompt_PlaceholderAndImageGrid(t *testing.T) {
	p := SystemPrompt()
	for _, kw := range []string{"placeholder", "image-grid", "hint:", "<!-- cell -->"} {
		require.Truef(t, strings.Contains(p, kw), "system prompt missing keyword %q", kw)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `GOTOOLCHAIN=local go test -C . ./internal/generate/ -run TestSystemPrompt_PlaceholderAndImageGrid -v`
Expected: FAIL.

- [ ] **Step 3: Update the system prompt**

Open `internal/generate/system_prompt.md`. Locate the `# Layouts` section and add this bullet (alphabetical / next to `dashboard` is fine):

```markdown
- `image-grid` — grid of images, placeholders, or components. Use
  `columns: 2|3|4` and `<!-- cell -->` before each item. Cells may contain
  a `placeholder`, a Markdown image, a chart, or any other component.
```

Add this new section after the `## Chart` subsection in `# Components`:

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

Add this bullet to the `# Rules` section:

```markdown
- When the slide is primarily visual (diagram, chart, screenshot) and you
  have no image URL, use `placeholder` with a descriptive `hint` and a
  fitting `icon`. Do not skip the slide and do not invent fake image URLs.
```

- [ ] **Step 4: Run test**

Run: `GOTOOLCHAIN=local go test -C . ./internal/generate/ -v`
Expected: all pass.

- [ ] **Step 5: Smoke test dump-prompt**

Run: `./goslide.exe generate --dump-prompt | head -60`
(Rebuild first if needed: `GOTOOLCHAIN=local go build -C . ./cmd/goslide`)
Expected: output includes the new Layouts bullet and the Placeholder section.

- [ ] **Step 6: Commit**

```bash
git -C D:/CLAUDE-CODE-GOSLIDE add internal/generate/system_prompt.md internal/generate/embed_test.go
```

```bash
git -C D:/CLAUDE-CODE-GOSLIDE commit -m "feat(generate): teach LLM about placeholder + image-grid"
```

---

## Task 13: Docs — README + PRD

**Files:**
- Modify: `README.md`
- Modify: `README_zh-TW.md`
- Modify: `PRD.md`

- [ ] **Step 1: PRD — update the Phase 5 / 6 checklist**

Open `PRD.md`, find the deferred-list / Phase 5 checkbox block. Add (or tick, if it already exists) a line for multi-image/placeholder:

```markdown
- [x] Multi-image layout + placeholder component (Phase 6b, v1.3.0)
```

If no such checkbox exists, add one under the Phase 5 "Deferred" list (position matches project style).

- [ ] **Step 2: README — new subsection**

In `README.md`, insert this subsection near the other layout/component docs (a sensible place is right after the AI slide generation section added in v1.2.0):

````markdown
### Image placeholders and multi-image slides

When you don't have an image URL yet — either while drafting, or when a
presentation is produced by `goslide generate` — use the `placeholder`
component:

```
~~~placeholder
hint: K8s architecture
icon: 🗺️
aspect: 16:9
---
Control plane + worker node interaction
~~~
```

Combine several placeholders (or real images, charts, cards) with the
`image-grid` layout:

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
~~~placeholder
hint: Trends
icon: 📈
~~~

<!-- cell -->
~~~chart
type: bar
...
~~~
```

Columns accept `2`, `3`, or `4`. Each `<!-- cell -->` marks a new cell;
cells may hold any content.
````

- [ ] **Step 3: README_zh-TW — parallel section**

Add a Traditional Chinese translation of the above section in the
matching location in `README_zh-TW.md`. Keep code blocks unchanged.

Suggested content:

````markdown
### 圖片佔位符與多圖投影片

尚未準備好實際圖檔（或用 `goslide generate` 生成時），可用 `placeholder` component：

```
~~~placeholder
hint: K8s 架構圖
icon: 🗺️
aspect: 16:9
---
Control plane 與 worker 互動示意
~~~
```

搭配 `image-grid` layout 可在同一張投影片並排多個 placeholder、真實圖片、圖表、卡片：

```
<!-- layout: image-grid -->
<!-- columns: 2 -->

<!-- cell -->
~~~placeholder
hint: 架構圖
icon: 🗺️
~~~

<!-- cell -->
![Dashboard](./dashboard.png)

<!-- cell -->
~~~placeholder
hint: 趨勢分析
icon: 📈
~~~

<!-- cell -->
~~~chart
type: bar
...
~~~
```

`columns` 可設 `2`、`3` 或 `4`。每個 `<!-- cell -->` 代表一格；格內可放任何內容。
````

- [ ] **Step 4: Commit**

```bash
git -C D:/CLAUDE-CODE-GOSLIDE add README.md README_zh-TW.md PRD.md
```

```bash
git -C D:/CLAUDE-CODE-GOSLIDE commit -m "docs: README + PRD updates for placeholder + image-grid"
```

---

## Task 14: Final verification

**Files:** none (verification only)

- [ ] **Step 1: Full test suite**

Run: `GOTOOLCHAIN=local go test -C . ./...`
Expected: PASS across all packages.

- [ ] **Step 2: Build**

Run: `GOTOOLCHAIN=local go build -C . ./cmd/goslide`
Expected: success.

- [ ] **Step 3: Validate the system prompt end-to-end**

Run: `./goslide.exe generate --dump-prompt`
Expected: output contains the new Layouts bullet, Placeholder section, and the new Rules bullet.

- [ ] **Step 4: Manual smoke test — serve mode**

Create a scratch `demo.md` containing a single slide with `~~~placeholder` plus a separate slide using `layout: image-grid` with four `<!-- cell -->` markers. Run `./goslide.exe serve demo.md`, open the browser, and verify:
- The placeholder slide renders a dashed box with the icon + hint + body text
- The grid slide shows four cells in 2 columns with the expected mix of placeholders and images
- Aspect ratio behaves when switching `aspect: 16:9` vs `aspect: 1:1`

This is the visual acceptance check — if anything looks broken, diagnose the responsible layer (CSS vs JS vs HTML) before proceeding.

- [ ] **Step 5: Manual LLM validation (optional, not required for release prep)**

Re-run `scripts/test-generate-llm.ps1` or an equivalent on a topic with a strong visual component (e.g. "Kubernetes architecture") and verify the output uses `placeholder` and/or `image-grid`. Copy a representative result to `examples/ai-generated/k8s-visual.md` if the result is interesting.

- [ ] **Step 6: No commit needed**

If all previous steps passed, implementation is complete. Tasks 1-13 produced the deliverable.

---

## Success Criteria (from spec §9)

- ✅ `~~~placeholder` renders a dashed-rectangle with icon + hint + body at the declared aspect ratio (Tasks 8, 10, 11, 14.4)
- ✅ `image-grid` with 2/3/4 columns renders a CSS grid of mixed cell content (Tasks 6, 9, 11, 14.4)
- ✅ Validation warns on invalid `aspect`, out-of-range `columns`, and empty `image-grid`; errors on missing `hint` (Tasks 2, 3)
- ✅ `--dump-prompt` output reflects the new Layouts, Components, and Rules entries (Task 12.5, Task 14.3)
- ✅ All unit + golden tests pass; no new dependencies (Task 14.1)
