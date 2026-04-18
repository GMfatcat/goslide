# Phase 2d: Expandable Cards Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add `grid-cards` layout with expandable card components that show a summary view and expand into a modal overlay with full detail content on click.

**Architecture:** Cards use the existing component pipeline (`extractComponents` → `renderComponents`). Parser splits card `Raw` at `---` into summary YAML (`Params`) and detail markdown (`ContentHTML`). JS renders summary view and opens a global overlay on click. Esc/backdrop/close-button dismiss the overlay.

**Tech Stack:** Go 1.21.6 (parser changes), vanilla JS (card init + overlay), CSS (grid layout + card + overlay styles). No new dependencies.

**Shell rules (Windows):** Never chain commands with `&&`. Use SEPARATE Bash calls for `git add`, `git commit`, `go test`, `go build`. Use `GOTOOLCHAIN=local` prefix for ALL go commands. Use `-C` flag for go/git to specify directory.

---

## Task 1: Go — SlideMeta.Columns + Card Parser + Validation

**Files:**
- Modify: `internal/ir/presentation.go`
- Modify: `internal/parser/slide.go`
- Modify: `internal/ir/validate.go`
- Modify: `internal/parser/slide_test.go`

- [ ] **Step 1: Add Columns to SlideMeta**

Read `internal/ir/presentation.go`. Add `Columns int` to `SlideMeta` after `SlideNumberHidden`:

```go
type SlideMeta struct {
	Title             string
	Layout            string
	Transition        string
	Fragments         bool
	FragmentStyle     string
	SlideNumberHidden bool
	Columns           int
}
```

- [ ] **Step 2: Add card split + columns parsing to slide.go**

Read `internal/parser/slide.go`. Make two changes:

**a)** In the component post-processing loop (after `extractComponents`), add card handling alongside the existing panel handling. Find:

```go
	for i := range components {
		if strings.HasPrefix(components[i].Type, "panel:") {
			components[i].ContentHTML = string(renderMarkdown(components[i].Raw))
		}
	}
```

Replace with:

```go
	for i := range components {
		if strings.HasPrefix(components[i].Type, "panel:") {
			components[i].ContentHTML = string(renderMarkdown(components[i].Raw))
		} else if components[i].Type == "card" {
			parts := strings.SplitN(components[i].Raw, "\n---\n", 2)
			if len(parts) == 2 {
				var summaryParams map[string]any
				if err := yaml.Unmarshal([]byte(parts[0]), &summaryParams); err == nil {
					components[i].Params = summaryParams
				}
				components[i].ContentHTML = string(renderMarkdown(parts[1]))
			}
		}
	}
```

Note: `yaml` is already imported (`"gopkg.in/yaml.v3"`). If not, add it.

**b)** In `buildSlideMeta`, add `columns` parsing. Find the `slide-number` handling block and add after it:

```go
	if v, ok := metaMap["columns"]; ok {
		n, err := strconv.Atoi(strings.TrimSpace(v))
		if err == nil {
			meta.Columns = n
		}
	}
```

Add `"strconv"` to imports if not present.

- [ ] **Step 3: Move grid-cards to knownLayouts**

Read `internal/ir/validate.go`. Find:

```go
futureLayouts = map[string]bool{"grid-cards": true}
```

Replace with:

```go
futureLayouts = map[string]bool{}
```

Find `knownLayouts` and add `"grid-cards": true`:

```go
knownLayouts = map[string]bool{
    "default": true, "title": true, "section": true,
    "two-column": true, "code-preview": true,
    "three-column": true, "image-left": true, "image-right": true,
    "quote": true, "split-heading": true, "top-bottom": true, "blank": true,
    "grid-cards": true,
}
```

- [ ] **Step 4: Add tests**

Read `internal/parser/slide_test.go`. Append:

```go
func TestParseSlide_CardWithDetail(t *testing.T) {
	raw := "<!-- layout: grid-cards -->\n\n~~~card\ntitle: Test Card\ndesc: A description\n---\n## Detail\n\nDetail content.\n~~~\n"
	slide := parseSlide(1, raw, ir.Frontmatter{})
	require.Len(t, slide.Components, 1)
	require.Equal(t, "card", slide.Components[0].Type)
	require.Equal(t, "Test Card", slide.Components[0].Params["title"])
	require.Contains(t, slide.Components[0].ContentHTML, "Detail content.")
	require.Contains(t, slide.Components[0].ContentHTML, "<h2>Detail</h2>")
}

func TestParseSlide_CardWithoutDetail(t *testing.T) {
	raw := "~~~card\ntitle: Simple\ndesc: No detail\n~~~\n"
	slide := parseSlide(1, raw, ir.Frontmatter{})
	require.Len(t, slide.Components, 1)
	require.Equal(t, "Simple", slide.Components[0].Params["title"])
	require.Empty(t, slide.Components[0].ContentHTML)
}

func TestParseSlide_Columns(t *testing.T) {
	raw := "<!-- layout: grid-cards -->\n<!-- columns: 3 -->\n\n# Title\n"
	slide := parseSlide(1, raw, ir.Frontmatter{})
	require.Equal(t, "grid-cards", slide.Meta.Layout)
	require.Equal(t, 3, slide.Meta.Columns)
}
```

- [ ] **Step 5: Run tests**

Run: `GOTOOLCHAIN=local go test -C D:/CLAUDE-CODE-GOSLIDE/internal/parser -v`

Run: `GOTOOLCHAIN=local go test -C D:/CLAUDE-CODE-GOSLIDE ./internal/... -count=1`

- [ ] **Step 6: Commit**

```
git -C D:/CLAUDE-CODE-GOSLIDE add internal/ir/presentation.go internal/parser/slide.go internal/ir/validate.go internal/parser/slide_test.go
```
```
git -C D:/CLAUDE-CODE-GOSLIDE commit -m "feat: add card component parsing, columns meta, and grid-cards to known layouts"
```

---

## Task 2: Template — Add data-columns attribute

**Files:**
- Modify: `web/templates/slide.html`

- [ ] **Step 1: Add data-columns to section tag**

Read `web/templates/slide.html`. Find the `<section` tag with all the data attributes. Add `data-columns` after the existing conditionals. Find:

```
        {{- if .Meta.SlideNumberHidden}} data-slide-number-hidden="true"{{end}}>
```

Replace with:

```
        {{- if .Meta.SlideNumberHidden}} data-slide-number-hidden="true"{{end}}
        {{- if .Meta.Columns}} data-columns="{{.Meta.Columns}}"{{end}}>
```

- [ ] **Step 2: Verify build + update golden files**

Run: `GOTOOLCHAIN=local go build -C D:/CLAUDE-CODE-GOSLIDE ./...`

Run: `GOTOOLCHAIN=local go test -C D:/CLAUDE-CODE-GOSLIDE/internal/renderer -run TestGolden -v -args -update`

Run: `GOTOOLCHAIN=local go test -C D:/CLAUDE-CODE-GOSLIDE ./internal/... -count=1`

- [ ] **Step 3: Commit**

```
git -C D:/CLAUDE-CODE-GOSLIDE add web/templates/slide.html internal/renderer/testdata/golden/
```
```
git -C D:/CLAUDE-CODE-GOSLIDE commit -m "feat: add data-columns attribute to slide section template"
```

---

## Task 3: Frontend — Card Init + Overlay in reactive.js

**Files:**
- Modify: `web/static/reactive.js`

- [ ] **Step 1: Add card + overlay code to reactive.js**

Read `web/static/reactive.js`. Add the card and overlay functions BEFORE the `initAllL2` function. Insert before `// --- Init all L2 components ---`:

```javascript
  // --- Card ---
  var overlay = null;
  var overlayEscHandler = null;

  function initCard(el) {
    var params = JSON.parse(decodeAttr(el.getAttribute('data-params')));
    var detailHTML = el.innerHTML;
    el.innerHTML = '';

    var summary = document.createElement('div');
    summary.className = 'goslide-card-summary';

    if (params.icon) {
      var icon = document.createElement('div');
      icon.className = 'goslide-card-icon';
      icon.textContent = params.icon;
      summary.appendChild(icon);
    }

    var title = document.createElement('div');
    title.className = 'goslide-card-title';
    title.textContent = params.title || '';
    summary.appendChild(title);

    if (params.desc) {
      var desc = document.createElement('div');
      desc.className = 'goslide-card-desc';
      desc.textContent = params.desc;
      summary.appendChild(desc);
    }

    el.appendChild(summary);
    el._detailHTML = detailHTML;

    summary.addEventListener('click', function () {
      openCardOverlay(el);
    });
  }

  function openCardOverlay(cardEl) {
    if (overlay) closeCardOverlay();

    overlay = document.createElement('div');
    overlay.className = 'goslide-card-overlay';

    var panel = document.createElement('div');
    panel.className = 'goslide-card-panel';
    panel.innerHTML = cardEl._detailHTML;

    var closeBtn = document.createElement('button');
    closeBtn.className = 'goslide-card-close';
    closeBtn.textContent = '\u2715';
    closeBtn.addEventListener('click', closeCardOverlay);

    panel.insertBefore(closeBtn, panel.firstChild);
    overlay.appendChild(panel);
    document.body.appendChild(overlay);

    overlay.addEventListener('click', function (e) {
      if (e.target === overlay) closeCardOverlay();
    });

    overlayEscHandler = function (e) {
      if (e.key === 'Escape') {
        e.stopPropagation();
        e.preventDefault();
        closeCardOverlay();
      }
    };
    document.addEventListener('keydown', overlayEscHandler, true);

    Reveal.configure({ keyboard: false });

    requestAnimationFrame(function () { overlay.classList.add('active'); });
  }

  function closeCardOverlay() {
    if (!overlay) return;
    overlay.classList.remove('active');
    if (overlayEscHandler) {
      document.removeEventListener('keydown', overlayEscHandler, true);
      overlayEscHandler = null;
    }
    Reveal.configure({ keyboard: { 13: 'next', 8: 'prev' } });
    setTimeout(function () {
      if (overlay && overlay.parentNode) overlay.parentNode.removeChild(overlay);
      overlay = null;
    }, 200);
  }
```

Then update `initAllL2` to include card. Find:

```javascript
      else if (type.indexOf('panel:') === 0) initPanel(el);
```

Add after it:

```javascript
      else if (type === 'card') initCard(el);
```

- [ ] **Step 2: Verify build**

Run: `GOTOOLCHAIN=local go build -C D:/CLAUDE-CODE-GOSLIDE ./...`

- [ ] **Step 3: Commit**

```
git -C D:/CLAUDE-CODE-GOSLIDE add web/static/reactive.js
```
```
git -C D:/CLAUDE-CODE-GOSLIDE commit -m "feat(frontend): add card component init and overlay expand/collapse"
```

---

## Task 4: CSS — Grid-cards Layout + Card + Overlay Styles

**Files:**
- Modify: `web/themes/layouts.css`
- Modify: `web/themes/tokens.css`

- [ ] **Step 1: Add grid-cards layout to layouts.css**

Read `web/themes/layouts.css`. Append at the end:

```css

/* Layout: grid-cards */
section[data-layout="grid-cards"] {
  display: flex;
  flex-wrap: wrap;
  gap: 1.5rem;
  justify-content: center;
  align-content: start;
  padding-top: 1rem;
}
section[data-layout="grid-cards"] > h1,
section[data-layout="grid-cards"] > h2 {
  width: 100%;
  text-align: center;
  flex-shrink: 0;
}
section[data-layout="grid-cards"] .goslide-component[data-type="card"] {
  width: calc(50% - 1rem);
}
section[data-layout="grid-cards"][data-columns="3"] .goslide-component[data-type="card"] {
  width: calc(33.33% - 1rem);
}
section[data-layout="grid-cards"][data-columns="4"] .goslide-component[data-type="card"] {
  width: calc(25% - 1.2rem);
}
```

- [ ] **Step 2: Add card + overlay CSS to tokens.css**

Read `web/themes/tokens.css`. Append at the end:

```css

/* Card summary */
.goslide-card-summary {
  background: var(--slide-card-bg);
  border-radius: 0.75rem;
  padding: 1.2rem;
  cursor: pointer;
  transition: transform 0.15s, box-shadow 0.15s;
  border: 1px solid var(--slide-border, rgba(0,0,0,0.1));
  text-align: left;
}
.goslide-card-summary:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0,0,0,0.15);
}
.goslide-card-icon {
  font-size: 1.5em;
  margin-bottom: 0.4em;
}
.goslide-card-title {
  font-size: 0.9em;
  font-weight: 700;
  color: var(--slide-heading);
  margin-bottom: 0.3em;
}
.goslide-card-desc {
  font-size: 0.7em;
  color: var(--slide-muted);
  line-height: 1.4;
}

/* Card overlay */
.goslide-card-overlay {
  position: fixed;
  top: 0; left: 0; right: 0; bottom: 0;
  background: rgba(0, 0, 0, 0.6);
  z-index: 1000;
  display: flex;
  justify-content: center;
  align-items: center;
  opacity: 0;
  transition: opacity 0.2s;
}
.goslide-card-overlay.active {
  opacity: 1;
}
.goslide-card-panel {
  background: var(--slide-bg, #ffffff);
  color: var(--slide-text, #1a1a1a);
  border-radius: 1rem;
  padding: 2rem 2.5rem;
  max-width: 85%;
  max-height: 85%;
  overflow-y: auto;
  position: relative;
  font-family: var(--font-sans);
  font-size: 16px;
  line-height: 1.6;
  box-shadow: 0 8px 32px rgba(0,0,0,0.3);
}
.goslide-card-panel h1, .goslide-card-panel h2, .goslide-card-panel h3 {
  color: var(--slide-heading);
}
.goslide-card-panel code {
  font-family: var(--font-mono);
}
.goslide-card-panel pre {
  background: var(--slide-code-bg);
  padding: 1rem;
  border-radius: 0.5rem;
  overflow-x: auto;
}
.goslide-card-close {
  position: absolute;
  top: 0.8rem;
  right: 1rem;
  background: none;
  border: none;
  font-size: 1.5rem;
  color: var(--slide-muted);
  cursor: pointer;
  line-height: 1;
  padding: 0.2em;
}
.goslide-card-close:hover {
  color: var(--slide-text);
}
```

- [ ] **Step 3: Verify build**

Run: `GOTOOLCHAIN=local go build -C D:/CLAUDE-CODE-GOSLIDE ./...`

- [ ] **Step 4: Commit**

```
git -C D:/CLAUDE-CODE-GOSLIDE add web/themes/layouts.css web/themes/tokens.css
```
```
git -C D:/CLAUDE-CODE-GOSLIDE commit -m "feat: add grid-cards layout and card/overlay CSS styles"
```

---

## Task 5: Demo + Manual Checklist

**Files:**
- Modify: `examples/demo.md`
- Modify: `MANUAL_CHECKLIST.md`

- [ ] **Step 1: Add grid-cards demo slide**

Read `examples/demo.md`. Find the "Thank You" slide. INSERT before it:

```markdown

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
```

- [ ] **Step 2: Update MANUAL_CHECKLIST.md**

Read `MANUAL_CHECKLIST.md`. Append:

```markdown

## Expandable Cards (Phase 2d)
- [ ] grid-cards layout — cards in 2-column grid
- [ ] card summary — icon, title, desc displayed
- [ ] card hover — slight lift effect
- [ ] card click → overlay with detail content
- [ ] overlay close — click ✕ button
- [ ] overlay close — click backdrop
- [ ] overlay close — Esc key
- [ ] Esc precedence — closes overlay, not Reveal.js overview
- [ ] keyboard lock — arrows don't navigate while overlay open
- [ ] overlay dark theme — panel bg matches theme
- [ ] detail content — markdown rendered (headings, tables, lists)
```

- [ ] **Step 3: Run all tests + build**

Run: `GOTOOLCHAIN=local go test -C D:/CLAUDE-CODE-GOSLIDE ./... -count=1 -race`

Run: `GOTOOLCHAIN=local go build -C D:/CLAUDE-CODE-GOSLIDE -ldflags "-X main.version=0.4.0" -o D:/CLAUDE-CODE-GOSLIDE/goslide.exe ./cmd/goslide`

- [ ] **Step 4: Commit**

```
git -C D:/CLAUDE-CODE-GOSLIDE add examples/demo.md MANUAL_CHECKLIST.md
```
```
git -C D:/CLAUDE-CODE-GOSLIDE commit -m "feat: add expandable cards demo and Phase 2d manual checklist"
```
