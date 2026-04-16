# Phase 2a: Component Parser + Layouts + Themes Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add component fence extraction with placeholder markers, 6 new slide layouts, and 3 new themes (corporate, minimal, hacker) to GoSlide.

**Architecture:** Pre-scan slide body text to extract `~~~componentType ... ~~~` fences into `[]Component` before goldmark rendering. Leave `<!--goslide:component:N-->` placeholders in body. Expand layout/theme whitelists. Change `ResolveAccent` to accept theme for per-theme default accents.

**Tech Stack:** Go 1.21.6, goldmark, gopkg.in/yaml.v3 (for component YAML params). No new dependencies.

**Shell rules (Windows):** Never chain commands with `&&`. Use SEPARATE Bash calls for `git add`, `git commit`, `go test`, `go build`. Use `GOTOOLCHAIN=local` prefix for ALL go commands. Use `-C` flag for go/git to specify directory.

---

## Task 1: IR Type — Add Component struct

**Files:**
- Modify: `internal/ir/presentation.go`

- [ ] **Step 1: Add Component type and update Slide struct**

Read `internal/ir/presentation.go`, then add `Component` type after `Region` and add `Components` field to `Slide`:

```go
type Component struct {
	Index  int
	Type   string
	Raw    string
	Params map[string]any
}
```

The `Slide` struct becomes:

```go
type Slide struct {
	Index      int
	Meta       SlideMeta
	RawBody    string
	BodyHTML   template.HTML
	Regions    []Region
	Components []Component
}
```

- [ ] **Step 2: Verify compilation**

Run: `GOTOOLCHAIN=local go build -C D:/CLAUDE-CODE-GOSLIDE ./internal/ir/`

- [ ] **Step 3: Run existing tests**

Run: `GOTOOLCHAIN=local go test -C D:/CLAUDE-CODE-GOSLIDE ./internal/... -count=1`
Expected: all PASS (Component field is zero-value compatible).

- [ ] **Step 4: Commit**

```
git -C D:/CLAUDE-CODE-GOSLIDE add internal/ir/presentation.go
```
```
git -C D:/CLAUDE-CODE-GOSLIDE commit -m "feat(ir): add Component type to Slide for component fence support"
```

---

## Task 2: Component Fence Extractor

**Files:**
- Create: `internal/parser/component.go`
- Create: `internal/parser/component_test.go`

- [ ] **Step 1: Write failing tests**

Create `internal/parser/component_test.go`:

```go
package parser

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExtractComponents_BasicChart(t *testing.T) {
	body := "# Title\n\n~~~chart:bar\ntitle: Yield\nlabels: [\"A\", \"B\"]\ndata: [96, 93]\n~~~\n\nMore text.\n"
	cleaned, comps := extractComponents(body)
	require.Len(t, comps, 1)
	require.Equal(t, "chart:bar", comps[0].Type)
	require.Equal(t, 0, comps[0].Index)
	require.Contains(t, comps[0].Raw, "title: Yield")
	require.Equal(t, "Yield", comps[0].Params["title"])
	require.Contains(t, cleaned, "<!--goslide:component:0-->")
	require.Contains(t, cleaned, "# Title")
	require.Contains(t, cleaned, "More text.")
	require.NotContains(t, cleaned, "~~~chart:bar")
}

func TestExtractComponents_Mermaid(t *testing.T) {
	body := "~~~mermaid\ngraph TD\n    A --> B\n~~~\n"
	cleaned, comps := extractComponents(body)
	require.Len(t, comps, 1)
	require.Equal(t, "mermaid", comps[0].Type)
	require.Contains(t, comps[0].Raw, "graph TD")
	require.Contains(t, cleaned, "<!--goslide:component:0-->")
}

func TestExtractComponents_MermaidParamsNil(t *testing.T) {
	body := "~~~mermaid\ngraph TD\n    A --> B\n~~~\n"
	_, comps := extractComponents(body)
	require.Nil(t, comps[0].Params)
}

func TestExtractComponents_MultipleComponents(t *testing.T) {
	body := "~~~chart:bar\ntitle: A\n~~~\n\nMiddle text.\n\n~~~chart:line\ntitle: B\n~~~\n"
	cleaned, comps := extractComponents(body)
	require.Len(t, comps, 2)
	require.Equal(t, "chart:bar", comps[0].Type)
	require.Equal(t, 0, comps[0].Index)
	require.Equal(t, "chart:line", comps[1].Type)
	require.Equal(t, 1, comps[1].Index)
	require.Contains(t, cleaned, "<!--goslide:component:0-->")
	require.Contains(t, cleaned, "<!--goslide:component:1-->")
	require.Contains(t, cleaned, "Middle text.")
}

func TestExtractComponents_NonComponentFencePassthrough(t *testing.T) {
	body := "~~~go\nfmt.Println(\"hello\")\n~~~\n"
	cleaned, comps := extractComponents(body)
	require.Len(t, comps, 0)
	require.Contains(t, cleaned, "~~~go")
	require.Contains(t, cleaned, "fmt.Println")
}

func TestExtractComponents_UnknownComponentPassthrough(t *testing.T) {
	body := "~~~unknown-widget\nfoo: bar\n~~~\n"
	cleaned, comps := extractComponents(body)
	require.Len(t, comps, 0)
	require.Contains(t, cleaned, "~~~unknown-widget")
}

func TestExtractComponents_NoComponents(t *testing.T) {
	body := "# Just markdown\n\nParagraph.\n"
	cleaned, comps := extractComponents(body)
	require.Len(t, comps, 0)
	require.Equal(t, body, cleaned)
}

func TestExtractComponents_ComponentWithColonType(t *testing.T) {
	body := "~~~embed:html\n<div>custom</div>\n~~~\n"
	cleaned, comps := extractComponents(body)
	require.Len(t, comps, 1)
	require.Equal(t, "embed:html", comps[0].Type)
	require.Contains(t, cleaned, "<!--goslide:component:0-->")
}

func TestExtractComponents_Table(t *testing.T) {
	body := "~~~table\ncolumns: [Name, Role]\nrows:\n  - [\"Alice\", \"Engineer\"]\nsortable: true\n~~~\n"
	_, comps := extractComponents(body)
	require.Len(t, comps, 1)
	require.Equal(t, "table", comps[0].Type)
	require.Equal(t, true, comps[0].Params["sortable"])
}

func TestExtractComponents_YAMLParseError(t *testing.T) {
	body := "~~~chart:bar\n  bad:\n    indent\n~~~\n"
	cleaned, comps := extractComponents(body)
	require.Len(t, comps, 1)
	require.Equal(t, "chart:bar", comps[0].Type)
	require.Nil(t, comps[0].Params)
	require.Contains(t, comps[0].Raw, "bad:")
	require.Contains(t, cleaned, "<!--goslide:component:0-->")
}

func TestExtractComponents_NestedFences(t *testing.T) {
	body := "~~~embed:html\n<pre>\n~~~\nfake fence\n~~~\n</pre>\n~~~\n"
	_, comps := extractComponents(body)
	require.Len(t, comps, 1)
	require.Equal(t, "embed:html", comps[0].Type)
}
```

- [ ] **Step 2: Run tests to verify failure**

Run: `GOTOOLCHAIN=local go test -C D:/CLAUDE-CODE-GOSLIDE/internal/parser -run TestExtractComponents -v`
Expected: compilation error — `extractComponents` undefined.

- [ ] **Step 3: Implement component.go**

Create `internal/parser/component.go`:

```go
package parser

import (
	"fmt"
	"strings"

	"github.com/user/goslide/internal/ir"
	"gopkg.in/yaml.v3"
)

var knownComponentPrefixes = map[string]bool{
	"chart": true, "mermaid": true, "table": true,
	"tabs": true, "panel": true, "slider": true,
	"toggle": true, "api": true, "embed": true, "card": true,
}

func isComponentFence(lang string) bool {
	if knownComponentPrefixes[lang] {
		return true
	}
	prefix := lang
	if idx := strings.Index(lang, ":"); idx != -1 {
		prefix = lang[:idx]
	}
	return knownComponentPrefixes[prefix]
}

func extractComponents(body string) (string, []ir.Component) {
	lines := strings.Split(body, "\n")
	var result []string
	var components []ir.Component

	i := 0
	for i < len(lines) {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "~~~") && len(trimmed) > 3 {
			lang := strings.TrimSpace(trimmed[3:])
			if isComponentFence(lang) {
				var contentLines []string
				i++
				for i < len(lines) {
					if strings.TrimSpace(lines[i]) == "~~~" {
						break
					}
					contentLines = append(contentLines, lines[i])
					i++
				}
				raw := strings.Join(contentLines, "\n")

				var params map[string]any
				if err := yaml.Unmarshal([]byte(raw), &params); err != nil {
					params = nil
				}

				comp := ir.Component{
					Index:  len(components),
					Type:   lang,
					Raw:    raw,
					Params: params,
				}
				components = append(components, comp)
				result = append(result, fmt.Sprintf("<!--goslide:component:%d-->", comp.Index))
				i++
				continue
			}
		}

		result = append(result, line)
		i++
	}

	cleaned := strings.Join(result, "\n")
	return cleaned, components
}
```

- [ ] **Step 4: Run tests to verify pass**

Run: `GOTOOLCHAIN=local go test -C D:/CLAUDE-CODE-GOSLIDE/internal/parser -run TestExtractComponents -v`
Expected: all PASS.

- [ ] **Step 5: Run all parser tests**

Run: `GOTOOLCHAIN=local go test -C D:/CLAUDE-CODE-GOSLIDE/internal/parser -v`
Expected: all PASS.

- [ ] **Step 6: Commit**

```
git -C D:/CLAUDE-CODE-GOSLIDE add internal/parser/component.go internal/parser/component_test.go
```
```
git -C D:/CLAUDE-CODE-GOSLIDE commit -m "feat(parser): add component fence extractor with placeholder markers"
```

---

## Task 3: Integrate extractComponents into parseSlide

**Files:**
- Modify: `internal/parser/slide.go`
- Modify: `internal/parser/slide_test.go`

- [ ] **Step 1: Add tests for component integration**

Read `internal/parser/slide_test.go`, then append:

```go
func TestParseSlide_WithComponent(t *testing.T) {
	raw := "# Title\n\n~~~chart:bar\ntitle: Yield\ndata: [96, 93]\n~~~\n\nAfter chart.\n"
	slide := parseSlide(1, raw, ir.Frontmatter{})
	require.Len(t, slide.Components, 1)
	require.Equal(t, "chart:bar", slide.Components[0].Type)
	require.Contains(t, string(slide.BodyHTML), "<!--goslide:component:0-->")
	require.Contains(t, string(slide.BodyHTML), "After chart.")
	require.NotContains(t, string(slide.BodyHTML), "~~~chart:bar")
}

func TestParseSlide_ComponentInRegion(t *testing.T) {
	raw := "<!-- layout: two-column -->\n\n<!-- left -->\n\nText left.\n\n<!-- right -->\n\n~~~chart:pie\ntitle: Share\n~~~\n"
	slide := parseSlide(1, raw, ir.Frontmatter{})
	require.Len(t, slide.Components, 1)
	require.Len(t, slide.Regions, 2)
	require.Contains(t, string(slide.Regions[1].HTML), "<!--goslide:component:0-->")
}

func TestParseSlide_NonComponentFenceUntouched(t *testing.T) {
	raw := "# Code\n\n~~~go\nfmt.Println()\n~~~\n"
	slide := parseSlide(1, raw, ir.Frontmatter{})
	require.Len(t, slide.Components, 0)
	require.Contains(t, string(slide.BodyHTML), "<code")
}
```

- [ ] **Step 2: Update parseSlide to call extractComponents**

Read `internal/parser/slide.go`. Insert `extractComponents` call after metadata extraction and before region splitting. The key change is:

Replace the line:
```go
bodyLines := lines[bodyStart:]
bodyText := strings.Join(bodyLines, "\n")
```

With:
```go
bodyLines := lines[bodyStart:]
bodyText := strings.Join(bodyLines, "\n")
cleanedBody, components := extractComponents(bodyText)
bodyLines = strings.Split(cleanedBody, "\n")
```

And at the end, change the return to include `Components`:
```go
return ir.Slide{
    Index:      index,
    Meta:       meta,
    RawBody:    bodyText,
    BodyHTML:   bodyHTML,
    Regions:    regions,
    Components: components,
}
```

- [ ] **Step 3: Run all parser tests**

Run: `GOTOOLCHAIN=local go test -C D:/CLAUDE-CODE-GOSLIDE/internal/parser -v`
Expected: all PASS.

- [ ] **Step 4: Commit**

```
git -C D:/CLAUDE-CODE-GOSLIDE add internal/parser/slide.go internal/parser/slide_test.go
```
```
git -C D:/CLAUDE-CODE-GOSLIDE commit -m "feat(parser): integrate component extraction into parseSlide with placeholder markers"
```

---

## Task 4: Update Validation — Layout Whitelists + Component Rules

**Files:**
- Modify: `internal/ir/validate.go`
- Modify: `internal/ir/validate_test.go`

- [ ] **Step 1: Add new validation tests**

Read `internal/ir/validate_test.go`, then append:

```go
func TestValidate_ThreeColumnLayout_Valid(t *testing.T) {
	p := &Presentation{
		Slides: []Slide{{Index: 1, Meta: SlideMeta{Layout: "three-column"},
			Regions: []Region{{Name: "col1", HTML: "a"}, {Name: "col2", HTML: "b"}, {Name: "col3", HTML: "c"}}}},
	}
	errs := p.Validate()
	e := findError(errs, "future-layout")
	require.Nil(t, e)
	e = findError(errs, "missing-region")
	require.Nil(t, e)
}

func TestValidate_ThreeColumnMissingRegion(t *testing.T) {
	p := &Presentation{
		Slides: []Slide{{Index: 1, Meta: SlideMeta{Layout: "three-column"},
			Regions: []Region{{Name: "col1", HTML: "a"}}}},
	}
	errs := p.Validate()
	e := findError(errs, "missing-region")
	require.NotNil(t, e)
}

func TestValidate_QuoteLayout_Valid(t *testing.T) {
	p := &Presentation{
		Slides: []Slide{{Index: 1, Meta: SlideMeta{Layout: "quote"}}},
	}
	errs := p.Validate()
	e := findError(errs, "future-layout")
	require.Nil(t, e)
}

func TestValidate_BlankLayout_Valid(t *testing.T) {
	p := &Presentation{
		Slides: []Slide{{Index: 1, Meta: SlideMeta{Layout: "blank"}}},
	}
	errs := p.Validate()
	e := findError(errs, "future-layout")
	require.Nil(t, e)
}

func TestValidate_GridCardsStillFuture(t *testing.T) {
	p := &Presentation{
		Slides: []Slide{{Index: 1, Meta: SlideMeta{Layout: "grid-cards"}}},
	}
	errs := p.Validate()
	e := findError(errs, "future-layout")
	require.NotNil(t, e)
	require.Equal(t, "warning", e.Severity)
}

func TestValidate_KnownComponentNoWarning(t *testing.T) {
	p := &Presentation{
		Slides: []Slide{{Index: 1, Meta: SlideMeta{Layout: "default"},
			Components: []Component{{Index: 0, Type: "chart:bar"}},
			RawBody:    "# Title\n"}},
	}
	errs := p.Validate()
	e := findError(errs, "future-component")
	require.Nil(t, e)
}

func TestValidate_ImageLeftRequiredRegions(t *testing.T) {
	p := &Presentation{
		Slides: []Slide{{Index: 1, Meta: SlideMeta{Layout: "image-left"},
			Regions: []Region{{Name: "image", HTML: "img"}}}},
	}
	errs := p.Validate()
	e := findError(errs, "missing-region")
	require.NotNil(t, e)
	require.Contains(t, e.Message, "text")
}

func TestValidate_TopBottomComplete(t *testing.T) {
	p := &Presentation{
		Slides: []Slide{{Index: 1, Meta: SlideMeta{Layout: "top-bottom"},
			Regions: []Region{{Name: "top", HTML: "t"}, {Name: "bottom", HTML: "b"}}}},
	}
	errs := p.Validate()
	e := findError(errs, "missing-region")
	require.Nil(t, e)
}
```

- [ ] **Step 2: Update validate.go**

Read `internal/ir/validate.go` and make these changes:

**a)** Replace `phase1Layouts` and `futureLayouts`:

```go
knownLayouts = map[string]bool{
    "default": true, "title": true, "section": true,
    "two-column": true, "code-preview": true,
    "three-column": true, "image-left": true, "image-right": true,
    "quote": true, "split-heading": true, "top-bottom": true, "blank": true,
}
futureLayouts = map[string]bool{"grid-cards": true}
```

**b)** Expand `requiredRegions`:

```go
requiredRegions = map[string][]string{
    "two-column":    {"left", "right"},
    "code-preview":  {"code", "preview"},
    "three-column":  {"col1", "col2", "col3"},
    "image-left":    {"image", "text"},
    "image-right":   {"text", "image"},
    "split-heading": {"heading", "body"},
    "top-bottom":    {"top", "bottom"},
}
```

**c)** In `validateSlide`, replace all references to `phase1Layouts` with `knownLayouts`.

**d)** In `closestLayout`, replace `phase1Layouts` with `knownLayouts`.

**e)** The `future-component` check on `RawBody` should now skip components that were already extracted (they are in `Slide.Components`). Since `extractComponents` removes known component fences from the body text, the `RawBody` field still has the original text. Change the future-component check to skip if the slide has `Components`:

Replace the future-component block:
```go
if matches := futureComponentRe.FindAllString(s.RawBody, -1); len(matches) > 0 {
```
With:
```go
if len(s.Components) == 0 {
    if matches := futureComponentRe.FindAllString(s.RawBody, -1); len(matches) > 0 {
```
And add closing brace.

- [ ] **Step 3: Run tests**

Run: `GOTOOLCHAIN=local go test -C D:/CLAUDE-CODE-GOSLIDE/internal/ir -v`
Expected: all PASS.

- [ ] **Step 4: Run all tests**

Run: `GOTOOLCHAIN=local go test -C D:/CLAUDE-CODE-GOSLIDE ./internal/... -count=1`
Expected: all PASS.

- [ ] **Step 5: Commit**

```
git -C D:/CLAUDE-CODE-GOSLIDE add internal/ir/validate.go internal/ir/validate_test.go
```
```
git -C D:/CLAUDE-CODE-GOSLIDE commit -m "feat(ir): update layout whitelists and component validation for Phase 2a"
```

---

## Task 5: Theme Package — Add 3 Themes + Change ResolveAccent

**Files:**
- Modify: `internal/theme/theme.go`
- Modify: `internal/theme/theme_test.go`

- [ ] **Step 1: Add failing tests**

Read `internal/theme/theme_test.go`, then append:

```go
func TestResolveTheme_Corporate(t *testing.T) {
	require.Equal(t, "corporate", ResolveTheme("corporate"))
}

func TestResolveTheme_Minimal(t *testing.T) {
	require.Equal(t, "minimal", ResolveTheme("minimal"))
}

func TestResolveTheme_Hacker(t *testing.T) {
	require.Equal(t, "hacker", ResolveTheme("hacker"))
}

func TestResolveAccent_DefaultForHacker(t *testing.T) {
	require.Equal(t, "green", ResolveAccent("", "hacker"))
}

func TestResolveAccent_DefaultForCorporate(t *testing.T) {
	require.Equal(t, "blue", ResolveAccent("", "corporate"))
}

func TestResolveAccent_ExplicitOverridesThemeDefault(t *testing.T) {
	require.Equal(t, "coral", ResolveAccent("coral", "hacker"))
}

func TestResolveAccent_EmptyTheme(t *testing.T) {
	require.Equal(t, "blue", ResolveAccent("", ""))
}
```

- [ ] **Step 2: Run tests to verify failure**

Run: `GOTOOLCHAIN=local go test -C D:/CLAUDE-CODE-GOSLIDE/internal/theme -v`
Expected: FAIL — `ResolveAccent` signature changed (now takes 2 args).

- [ ] **Step 3: Update theme.go**

Replace entire content of `internal/theme/theme.go`:

```go
package theme

var validThemes = map[string]bool{
	"default": true, "dark": true,
	"corporate": true, "minimal": true, "hacker": true,
}

var themeDefaultAccents = map[string]string{
	"default": "blue", "dark": "blue", "corporate": "blue",
	"minimal": "blue", "hacker": "green",
}

func ResolveTheme(name string) string {
	if name == "" || !validThemes[name] {
		return "default"
	}
	return name
}

func ResolveAccent(accent, themeName string) string {
	if accent != "" {
		return accent
	}
	if def, ok := themeDefaultAccents[themeName]; ok {
		return def
	}
	return "blue"
}

func ThemeCSSPath(name string) string {
	return "themes/" + name + ".css"
}
```

- [ ] **Step 4: Fix existing tests**

Read `internal/theme/theme_test.go`. The existing `TestResolveAccent_Default` and `TestResolveAccent_AllValid` tests call `ResolveAccent` with 1 arg. Update them:

Change `TestResolveAccent_Default`:
```go
func TestResolveAccent_Default(t *testing.T) {
	name := ResolveAccent("", "")
	require.Equal(t, "blue", name)
}
```

Change `TestResolveAccent_AllValid`:
```go
func TestResolveAccent_AllValid(t *testing.T) {
	accents := []string{"blue", "teal", "purple", "coral", "amber", "green", "red", "pink"}
	for _, a := range accents {
		require.Equal(t, a, ResolveAccent(a, "default"))
	}
}
```

- [ ] **Step 5: Run theme tests**

Run: `GOTOOLCHAIN=local go test -C D:/CLAUDE-CODE-GOSLIDE/internal/theme -v`
Expected: all PASS.

- [ ] **Step 6: Commit**

```
git -C D:/CLAUDE-CODE-GOSLIDE add internal/theme/theme.go internal/theme/theme_test.go
```
```
git -C D:/CLAUDE-CODE-GOSLIDE commit -m "feat(theme): add corporate, minimal, hacker themes with per-theme default accents"
```

---

## Task 6: Update Renderer — Pass Theme to ResolveAccent

**Files:**
- Modify: `internal/renderer/renderer.go`
- Modify: `internal/renderer/renderer_test.go`

- [ ] **Step 1: Update renderer.go**

Read `internal/renderer/renderer.go`. Change the `Accent` line in the `data` struct initialization:

From:
```go
Accent:      theme.ResolveAccent(pres.Meta.Accent),
```
To:
```go
Accent:      theme.ResolveAccent(pres.Meta.Accent, data.Theme),
```

Note: `data.Theme` is already set on the line above, so use that resolved value. The full block becomes:

```go
resolvedTheme := theme.ResolveTheme(pres.Meta.Theme)
data := templateData{
    Title:             pres.Meta.Title,
    Theme:             resolvedTheme,
    Accent:            theme.ResolveAccent(pres.Meta.Accent, resolvedTheme),
    Transition:        resolveTransition(pres.Meta.Transition),
    SlideNumber:       pres.Meta.SlideNumber,
    SlideNumberFormat: resolveSlideNumberFormat(pres.Meta.SlideNumberFormat),
    Slides:            pres.Slides,
}
```

- [ ] **Step 2: Add test for hacker default accent**

Read `internal/renderer/renderer_test.go`, then append:

```go
func TestRender_HackerThemeDefaultAccent(t *testing.T) {
	pres := &ir.Presentation{
		Meta:   ir.Frontmatter{Title: "Hack", Theme: "hacker"},
		Slides: []ir.Slide{{Index: 1, Meta: ir.SlideMeta{Layout: "default"}, BodyHTML: "<p>x</p>"}},
	}
	html, err := Render(pres)
	require.NoError(t, err)
	require.Contains(t, html, `data-accent="green"`)
	require.Contains(t, html, `href="/themes/hacker.css"`)
}
```

- [ ] **Step 3: Run renderer tests**

Run: `GOTOOLCHAIN=local go test -C D:/CLAUDE-CODE-GOSLIDE/internal/renderer -v`
Expected: all PASS.

- [ ] **Step 4: Run all tests**

Run: `GOTOOLCHAIN=local go test -C D:/CLAUDE-CODE-GOSLIDE ./internal/... -count=1`
Expected: all PASS.

- [ ] **Step 5: Commit**

```
git -C D:/CLAUDE-CODE-GOSLIDE add internal/renderer/renderer.go internal/renderer/renderer_test.go
```
```
git -C D:/CLAUDE-CODE-GOSLIDE commit -m "feat(renderer): pass theme to ResolveAccent for per-theme default accents"
```

---

## Task 7: Update Validation — Sync validThemes

**Files:**
- Modify: `internal/ir/validate.go`

- [ ] **Step 1: Update validThemes in validate.go**

Read `internal/ir/validate.go`. Change:

```go
validThemes = map[string]bool{"default": true, "dark": true}
```

To:

```go
validThemes = map[string]bool{"default": true, "dark": true, "corporate": true, "minimal": true, "hacker": true}
```

- [ ] **Step 2: Run all tests**

Run: `GOTOOLCHAIN=local go test -C D:/CLAUDE-CODE-GOSLIDE ./internal/... -count=1`
Expected: all PASS.

- [ ] **Step 3: Commit**

```
git -C D:/CLAUDE-CODE-GOSLIDE add internal/ir/validate.go
```
```
git -C D:/CLAUDE-CODE-GOSLIDE commit -m "feat(ir): sync validThemes with new corporate, minimal, hacker themes"
```

---

## Task 8: Layout CSS — 6 New Layouts

**Files:**
- Modify: `web/themes/layouts.css`
- Modify: `internal/parser/slide.go` (layoutRegions map)

- [ ] **Step 1: Update layoutRegions in slide.go**

Read `internal/parser/slide.go`. Replace the `layoutRegions` map:

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

- [ ] **Step 2: Append 6 new layout CSS rules to layouts.css**

Read `web/themes/layouts.css`, then append after the existing code-preview rules:

```css

/* Layout: three-column */
section[data-layout="three-column"] .slide-body {
  display: grid;
  grid-template-columns: 1fr 1fr 1fr;
  gap: 1.5rem;
  width: 100%;
}
section[data-layout="three-column"] .region-col1,
section[data-layout="three-column"] .region-col2,
section[data-layout="three-column"] .region-col3 { text-align: left; }

/* Layout: image-left */
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

/* Layout: image-right */
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

/* Layout: quote */
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

/* Layout: split-heading */
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

/* Layout: top-bottom */
section[data-layout="top-bottom"] .slide-body {
  display: grid;
  grid-template-rows: 1fr auto;
  gap: 1.5rem;
  width: 100%;
  height: 100%;
}
section[data-layout="top-bottom"] .region-top { text-align: center; }
section[data-layout="top-bottom"] .region-bottom { text-align: left; }

/* Layout: blank */
section[data-layout="blank"] {
  padding: 0;
}
```

- [ ] **Step 3: Run all tests**

Run: `GOTOOLCHAIN=local go test -C D:/CLAUDE-CODE-GOSLIDE ./internal/... -count=1`
Expected: all PASS.

- [ ] **Step 4: Verify build**

Run: `GOTOOLCHAIN=local go build -C D:/CLAUDE-CODE-GOSLIDE ./...`

- [ ] **Step 5: Commit**

```
git -C D:/CLAUDE-CODE-GOSLIDE add internal/parser/slide.go web/themes/layouts.css
```
```
git -C D:/CLAUDE-CODE-GOSLIDE commit -m "feat: add 6 new slide layouts (three-column, image-left/right, quote, split-heading, top-bottom, blank)"
```

---

## Task 9: Theme CSS Files — corporate, minimal, hacker

**Files:**
- Create: `web/themes/corporate.css`
- Create: `web/themes/minimal.css`
- Create: `web/themes/hacker.css`

- [ ] **Step 1: Create corporate.css**

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

.reveal {
  background: var(--slide-bg);
  color: var(--slide-text);
  font-family: var(--font-sans);
}
.reveal h1, .reveal h2, .reveal h3, .reveal h4 {
  color: var(--slide-heading);
  font-family: var(--font-sans);
}
.reveal a { color: var(--slide-accent); }
.reveal a:hover { color: var(--slide-accent); filter: brightness(1.2); }
.reveal pre {
  background: var(--slide-code-bg);
  border-radius: 0.5rem;
  padding: 1rem;
  width: 100%;
  box-sizing: border-box;
}
.reveal code {
  font-family: var(--font-mono);
  color: var(--slide-code-text);
}
.reveal pre code {
  background: none;
  font-size: 0.85em;
  line-height: 1.5;
}
.reveal blockquote {
  border-left: 4px solid var(--slide-accent);
  padding-left: 1rem;
  color: var(--slide-muted);
}
.reveal table th {
  border-bottom: 2px solid var(--slide-accent);
}
.reveal table td {
  border-bottom: 1px solid var(--slide-border);
}
.reveal .controls { color: var(--slide-accent); }
.reveal .progress { color: var(--slide-accent); }
.reveal .slide-number { color: var(--slide-muted); }
#goslide-page-num { color: var(--slide-muted); }
```

- [ ] **Step 2: Create minimal.css**

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

.reveal {
  background: var(--slide-bg);
  color: var(--slide-text);
  font-family: var(--font-sans);
}
.reveal h1 { font-size: 2.2em; }
.reveal h1, .reveal h2, .reveal h3, .reveal h4 {
  color: var(--slide-heading);
  font-family: var(--font-sans);
}
.reveal a { color: var(--slide-accent); }
.reveal a:hover { color: var(--slide-accent); filter: brightness(1.2); }
.reveal pre {
  background: var(--slide-code-bg);
  border-radius: 0.5rem;
  padding: 1rem;
  width: 100%;
  box-sizing: border-box;
}
.reveal code {
  font-family: var(--font-mono);
  color: var(--slide-code-text);
}
.reveal pre code {
  background: none;
  font-size: 0.85em;
  line-height: 1.5;
}
.reveal blockquote {
  border-left: 4px solid var(--slide-accent);
  padding-left: 1rem;
  color: var(--slide-muted);
}
.reveal table th {
  border-bottom: 2px solid var(--slide-accent);
}
.reveal table td {
  border-bottom: 1px solid var(--slide-border);
}
.reveal .controls { color: var(--slide-accent); }
.reveal .progress { color: var(--slide-accent); }
.reveal .slide-number { color: var(--slide-muted); }
#goslide-page-num { color: var(--slide-muted); }
```

- [ ] **Step 3: Create hacker.css**

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

.reveal {
  background: var(--slide-bg);
  color: var(--slide-text);
  font-family: var(--font-mono);
}
.reveal h1, .reveal h2, .reveal h3, .reveal h4 {
  color: var(--slide-heading);
  font-family: var(--font-mono);
  text-shadow: 0 0 8px rgba(0, 255, 0, 0.3);
}
.reveal a { color: var(--slide-accent); }
.reveal a:hover { color: var(--slide-accent); filter: brightness(1.3); }
.reveal pre {
  background: var(--slide-code-bg);
  border-radius: 0.25rem;
  padding: 1rem;
  width: 100%;
  box-sizing: border-box;
  border: 1px solid var(--slide-border);
}
.reveal code {
  font-family: var(--font-mono);
  color: var(--slide-code-text);
}
.reveal pre code {
  background: none;
  font-size: 0.85em;
  line-height: 1.5;
}
.reveal blockquote {
  border-left: 4px solid var(--slide-accent);
  padding-left: 1rem;
  color: var(--slide-muted);
}
.reveal table th {
  border-bottom: 2px solid var(--slide-accent);
}
.reveal table td {
  border-bottom: 1px solid var(--slide-border);
}
.reveal .controls { color: var(--slide-accent); }
.reveal .progress { color: var(--slide-accent); }
.reveal .slide-number { color: var(--slide-muted); }
#goslide-page-num { color: rgba(0, 255, 0, 0.5); }
```

- [ ] **Step 4: Verify build**

Run: `GOTOOLCHAIN=local go build -C D:/CLAUDE-CODE-GOSLIDE ./...`

- [ ] **Step 5: Commit**

```
git -C D:/CLAUDE-CODE-GOSLIDE add web/themes/corporate.css web/themes/minimal.css web/themes/hacker.css
```
```
git -C D:/CLAUDE-CODE-GOSLIDE commit -m "feat: add corporate, minimal, and hacker theme CSS files"
```

---

## Task 10: Golden Tests + Demo Update

**Files:**
- Create: `internal/renderer/testdata/golden/three-column.md`
- Create: `internal/renderer/testdata/golden/quote.md`
- Modify: `internal/renderer/golden_test.go` (no changes needed — it auto-discovers *.md)
- Modify: `examples/demo.md`

- [ ] **Step 1: Create three-column golden test**

Create `internal/renderer/testdata/golden/three-column.md`:

```markdown
---
title: Three Column
theme: corporate
accent: teal
---

<!-- layout: three-column -->

# Comparison

<!-- col1 -->

## Option A

- Fast
- Simple

<!-- col2 -->

## Option B

- Flexible
- Popular

<!-- col3 -->

## Option C

- Robust
- Scalable
```

- [ ] **Step 2: Create quote golden test**

Create `internal/renderer/testdata/golden/quote.md`:

```markdown
---
title: Quote Test
theme: minimal
---

<!-- layout: quote -->

> The best way to predict the future is to invent it.
>
> — Alan Kay
```

- [ ] **Step 3: Generate golden files**

Run: `GOTOOLCHAIN=local go test -C D:/CLAUDE-CODE-GOSLIDE/internal/renderer -run TestGolden -v -args -update`

- [ ] **Step 4: Verify golden tests pass without -update**

Run: `GOTOOLCHAIN=local go test -C D:/CLAUDE-CODE-GOSLIDE/internal/renderer -run TestGolden -v`
Expected: all PASS (5 subtests now).

- [ ] **Step 5: Update demo.md with new layouts and themes**

Read `examples/demo.md`, then add new slides showcasing the new layouts. Append before the "Thank You" slide:

```markdown

---

<!-- layout: quote -->

> The best way to predict the future is to invent it.
>
> — Alan Kay

---

<!-- layout: split-heading -->

# Architecture

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
```

- [ ] **Step 6: Run all tests**

Run: `GOTOOLCHAIN=local go test -C D:/CLAUDE-CODE-GOSLIDE ./... -count=1 -race`
Expected: all PASS.

- [ ] **Step 7: Commit**

```
git -C D:/CLAUDE-CODE-GOSLIDE add internal/renderer/testdata/golden/ examples/demo.md
```
```
git -C D:/CLAUDE-CODE-GOSLIDE commit -m "test: add golden tests for new layouts and update demo with Phase 2a features"
```
