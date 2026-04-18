# GoSlide Phase 2d — Expandable Cards Design Spec

> **Status: COMPLETED (2026-04-18)**
>
> Key learnings:
> - Reveal.js center:true sets flex-direction:column on all sections — grid-cards needs !important override
> - goldmark requires extension.Table for GFM table rendering
> - .goslide-component base max-width/margin conflicts with grid layout — card elements need explicit override
> - Card overlay polish (table spacing, emoji) deferred to later phase

## Decision Log

| # | Question | Decision |
|---|----------|----------|
| 1 | Card fence separator | `---` inside card fence separates summary YAML from detail markdown |
| 2 | Esc key conflict | Overlay captures Esc first (capture: true + stopPropagation), Reveal.js gets it after close |
| 3 | Implementation approach | grid-cards is a CSS layout; card is a component in the standard pipeline |
| 4 | Default expand | grid-cards layout defaults to expand: true |

---

## 1. Card Component Data Flow

### Parser

`extractComponents` already recognizes `card` type. Card `Raw` contains `---`-separated summary + detail:

```
icon: clock
color: blue
title: Image capture
desc: High-speed camera acquisition
---
## Image capture detail

Full markdown content here...
```

In `parseSlide`, after `extractComponents`, card components get special handling:

```go
if comp.Type == "card" {
    parts := strings.SplitN(comp.Raw, "\n---\n", 2)
    if len(parts) == 2 {
        // Re-parse summary-only YAML for clean Params
        yaml.Unmarshal([]byte(parts[0]), &comp.Params)
        // Render detail markdown to ContentHTML
        comp.ContentHTML = string(renderMarkdown(parts[1]))
    }
}
```

This reuses the `ContentHTML` field added in Phase 2c for panels.

### Renderer

`buildComponentDiv` already outputs `ContentHTML` inside the div (Phase 2c). Card output:

```html
<div class="goslide-component" data-type="card"
     data-params='{"icon":"clock","color":"blue","title":"Image capture","desc":"High-speed camera acquisition"}'
     data-comp-id="s1-c0">
  <h2>Image capture detail</h2>
  <p>Full markdown content here...</p>
</div>
```

### SlideMeta — Columns

Add `Columns int` to `SlideMeta`. Parser reads `<!-- columns: 3 -->` from slide comments.

```go
if v, ok := metaMap["columns"]; ok {
    n, err := strconv.Atoi(strings.TrimSpace(v))
    if err == nil { meta.Columns = n }
}
```

Template outputs `data-columns` on section:

```html
<section data-layout="{{.Meta.Layout}}"
  {{- if .Meta.Columns}} data-columns="{{.Meta.Columns}}"{{end}}
  ...>
```

---

## 2. Frontend — Card Init + Overlay

### initCard(el)

Reads `data-params` for summary, saves `innerHTML` as detail HTML, renders summary view with icon/title/desc. Click → opens overlay.

```javascript
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

  summary.style.cursor = 'pointer';
  summary.addEventListener('click', function() {
    openCardOverlay(el);
  });
}
```

### Overlay

Single global overlay div. Opening disables Reveal.js keyboard. Closing re-enables.

```javascript
var overlay = null;
var overlayEscHandler = null;

function openCardOverlay(cardEl) {
  if (overlay) closeCardOverlay();

  overlay = document.createElement('div');
  overlay.className = 'goslide-card-overlay';

  var panel = document.createElement('div');
  panel.className = 'goslide-card-panel';
  panel.innerHTML = cardEl._detailHTML;

  var closeBtn = document.createElement('button');
  closeBtn.className = 'goslide-card-close';
  closeBtn.textContent = '✕';
  closeBtn.addEventListener('click', closeCardOverlay);

  panel.insertBefore(closeBtn, panel.firstChild);
  overlay.appendChild(panel);
  document.body.appendChild(overlay);

  overlay.addEventListener('click', function(e) {
    if (e.target === overlay) closeCardOverlay();
  });

  overlayEscHandler = function(e) {
    if (e.key === 'Escape') {
      e.stopPropagation();
      e.preventDefault();
      closeCardOverlay();
    }
  };
  document.addEventListener('keydown', overlayEscHandler, true);

  Reveal.configure({ keyboard: false });

  requestAnimationFrame(function() { overlay.classList.add('active'); });
}

function closeCardOverlay() {
  if (!overlay) return;
  overlay.classList.remove('active');
  if (overlayEscHandler) {
    document.removeEventListener('keydown', overlayEscHandler, true);
    overlayEscHandler = null;
  }
  Reveal.configure({ keyboard: { 13: 'next', 8: 'prev' } });
  setTimeout(function() {
    if (overlay && overlay.parentNode) overlay.parentNode.removeChild(overlay);
    overlay = null;
  }, 200);
}
```

### Integration

Card init and overlay functions go in `reactive.js`. In `initAllL2`, add:

```javascript
else if (type === 'card') initCard(el);
```

---

## 3. CSS

### grid-cards Layout

```css
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

### Card Summary

```css
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
```

### Overlay

```css
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

---

## 4. Validation Changes

- Move `grid-cards` from `futureLayouts` to `knownLayouts`
- `futureLayouts` becomes empty map (keep for forward compatibility)

---

## 5. Testing

### Go Unit Tests

- Card `---` split: summary YAML parsed into Params, detail markdown rendered to ContentHTML
- Card without `---`: entire Raw as summary, no ContentHTML
- `columns` parsing in `buildSlideMeta`

### Manual Checklist

```markdown
## Expandable Cards (Phase 2d)
- [ ] grid-cards layout — cards in 2-column grid
- [ ] columns: 3 — cards in 3-column grid
- [ ] card summary — icon, title, desc displayed
- [ ] card hover — slight lift effect
- [ ] card click → overlay with detail content
- [ ] overlay close — ✕ button
- [ ] overlay close — click backdrop
- [ ] overlay close — Esc key
- [ ] Esc precedence — closes overlay, not Reveal.js overview
- [ ] keyboard lock — arrows don't navigate while overlay open
- [ ] overlay dark theme — panel bg matches theme
- [ ] detail content — markdown rendered (headings, code, lists)
```

---

## 6. Files Changed Summary

| Action | File |
|--------|------|
| Modify | `internal/ir/presentation.go` — `SlideMeta` add `Columns int` |
| Modify | `internal/parser/slide.go` — card `---` split + ContentHTML, `columns` parsing |
| Modify | `internal/ir/validate.go` — move `grid-cards` to `knownLayouts` |
| Modify | `web/templates/slide.html` — add `data-columns` attribute |
| Modify | `web/static/reactive.js` — add `initCard`, overlay functions, card in `initAllL2` |
| Modify | `web/themes/layouts.css` — add grid-cards layout CSS |
| Modify | `web/themes/tokens.css` — add card summary + overlay CSS |
| Modify | `examples/demo.md` — add grid-cards demo slide |
| Modify | `MANUAL_CHECKLIST.md` — add expandable cards checklist |
