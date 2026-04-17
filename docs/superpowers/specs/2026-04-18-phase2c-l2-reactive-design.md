# GoSlide Phase 2c — L2 Reactive Components Design Spec

> **Status: DRAFT**

## Decision Log

| # | Question | Decision |
|---|----------|----------|
| 1 | Scope | Minimal + toggle→panel visibility; slider/toggle publish $variable, chart binding deferred to Phase 3 |
| 2 | File structure | New `reactive.js` (store + L2 init), separate from `components.js` |
| 3 | Tabs/panel mechanism | Walk reactive store: tabs publish index via `GoSlide.set`, panels subscribe via `GoSlide.on` |
| 4 | Fragment interaction | Fully isolated — reactive and fragments don't interfere |
| 5 | UI implementation | Native HTML inputs (`<input type="range">`, `<input type="checkbox">`) + CSS styling |

---

## 1. Reactive Store

~15 lines of JS injected as `window.GoSlide`:

```javascript
window.GoSlide = window.GoSlide || {};
(function(G) {
  var store = {};
  var listeners = {};
  G.set = function(key, value) {
    store[key] = value;
    (listeners[key] || []).forEach(function(fn) { fn(value); });
  };
  G.get = function(key) { return store[key]; };
  G.on = function(key, fn) {
    (listeners[key] = listeners[key] || []).push(fn);
    if (key in store) fn(store[key]);
  };
})(window.GoSlide);
```

Key behavior: `GoSlide.on(key, fn)` immediately fires `fn` with current value if key already exists in store. This ensures late-initializing consumers get the producer's default value.

### Script load order

```html
<script src="/static/chartjs/chart.min.js"></script>
<script src="/static/mermaid/mermaid.min.js"></script>
<script src="/static/reveal/reveal.js"></script>
<script src="/static/runtime.js"></script>
<script src="/static/reactive.js"></script>
<script src="/static/components.js"></script>
<script>Reveal.initialize({...});</script>
```

`reactive.js` loads after `runtime.js` (needs Reveal events) and before `components.js` (Phase 3 charts may use `GoSlide.on`).

---

## 2. L2 Component Initialization

All L2 components init at page load in `Reveal.on('ready')`, same pattern as L1. No lazy init.

### Tabs

Renders a button bar. Each button click → `GoSlide.set(id, index)`. Subscribes to own ID to update active styling.

```javascript
function initTabs(el) {
  var params = JSON.parse(decodeAttr(el.getAttribute('data-params')));
  var id = params.id;
  var labels = params.labels || [];

  var bar = document.createElement('div');
  bar.className = 'goslide-tabs';
  labels.forEach(function(label, idx) {
    var btn = document.createElement('button');
    btn.textContent = label;
    btn.className = 'goslide-tab';
    btn.addEventListener('click', function() { GoSlide.set(id, idx); });
    bar.appendChild(btn);
  });
  el.appendChild(bar);

  GoSlide.on(id, function(value) {
    bar.querySelectorAll('.goslide-tab').forEach(function(b, i) {
      b.classList.toggle('active', i === value);
    });
  });

  GoSlide.set(id, 0);
}
```

### Panel

Panel `data-type` is `panel:compare-0` (tabs) or `panel:show_details` (toggle).

Parsing logic:
- If suffix after last `-` is numeric → tabs panel, subscribe to tabs ID with index matching
- Otherwise → toggle panel, subscribe to the full ID as boolean

```javascript
function initPanel(el) {
  var fullType = el.getAttribute('data-type');
  var parts = fullType.substring(6);
  var lastDash = parts.lastIndexOf('-');
  var suffix = parts.substring(lastDash + 1);

  if (!isNaN(parseInt(suffix)) && lastDash > 0) {
    var tabsId = parts.substring(0, lastDash);
    var panelIdx = parseInt(suffix);
    el.style.display = 'none';
    GoSlide.on(tabsId, function(value) {
      el.style.display = (value === panelIdx) ? '' : 'none';
    });
  } else {
    el.style.display = 'none';
    GoSlide.on(parts, function(value) {
      el.style.display = value ? '' : 'none';
    });
  }
}
```

### Slider

Renders `<label>` + `<input type="range">` + value display `<span>`. On input change → `GoSlide.set(id, value)`.

```javascript
function initSlider(el) {
  var params = JSON.parse(decodeAttr(el.getAttribute('data-params')));
  var id = params.id;

  var wrapper = document.createElement('div');
  wrapper.className = 'goslide-slider';

  var label = document.createElement('label');
  label.textContent = params.label || id;

  var input = document.createElement('input');
  input.type = 'range';
  input.min = params.min || 0;
  input.max = params.max || 100;
  input.step = params.step || 1;
  input.value = params.value || params.min || 0;

  var display = document.createElement('span');
  display.className = 'goslide-slider-value';
  display.textContent = input.value + (params.unit || '');

  input.addEventListener('input', function() {
    var val = parseFloat(input.value);
    display.textContent = val + (params.unit || '');
    GoSlide.set(id, val);
  });

  wrapper.appendChild(label);
  wrapper.appendChild(input);
  wrapper.appendChild(display);
  el.appendChild(wrapper);

  GoSlide.set(id, parseFloat(input.value));
}
```

### Toggle

Renders hidden `<input type="checkbox">` + CSS switch span + label text. On change → `GoSlide.set(id, checked)`.

```javascript
function initToggle(el) {
  var params = JSON.parse(decodeAttr(el.getAttribute('data-params')));
  var id = params.id;

  var wrapper = document.createElement('div');
  wrapper.className = 'goslide-toggle';

  var label = document.createElement('label');
  label.className = 'goslide-toggle-label';

  var input = document.createElement('input');
  input.type = 'checkbox';
  input.checked = params.default === true;

  var switchSpan = document.createElement('span');
  switchSpan.className = 'goslide-toggle-switch';

  var text = document.createElement('span');
  text.textContent = params.label || id;

  input.addEventListener('change', function() {
    GoSlide.set(id, input.checked);
  });

  label.appendChild(input);
  label.appendChild(switchSpan);
  wrapper.appendChild(label);
  wrapper.appendChild(text);
  el.appendChild(wrapper);

  GoSlide.set(id, input.checked);
}
```

---

## 3. CSS Styling

All styles in `tokens.css`, using CSS variables for theme adaptation.

```css
/* Tabs */
.goslide-tabs {
  display: flex;
  gap: 0;
  border-bottom: 2px solid var(--slide-border, rgba(0,0,0,0.1));
  margin-bottom: 1rem;
}
.goslide-tab {
  padding: 0.5em 1.2em;
  border: none;
  background: none;
  color: var(--slide-muted);
  font-family: var(--font-sans);
  font-size: 0.8em;
  cursor: pointer;
  border-bottom: 2px solid transparent;
  margin-bottom: -2px;
  transition: color 0.2s, border-color 0.2s;
}
.goslide-tab:hover { color: var(--slide-text); }
.goslide-tab.active {
  color: var(--slide-accent);
  border-bottom-color: var(--slide-accent);
}

/* Slider */
.goslide-slider {
  display: flex;
  align-items: center;
  gap: 0.8em;
  margin: 0.8rem 0;
  font-size: 0.8em;
}
.goslide-slider label {
  color: var(--slide-text);
  white-space: nowrap;
}
.goslide-slider input[type="range"] {
  flex: 1;
  accent-color: var(--slide-accent);
  height: 6px;
}
.goslide-slider-value {
  color: var(--slide-accent);
  font-family: var(--font-mono);
  min-width: 3em;
  text-align: right;
}

/* Toggle */
.goslide-toggle {
  display: flex;
  align-items: center;
  gap: 0.8em;
  margin: 0.8rem 0;
  font-size: 0.8em;
}
.goslide-toggle-label {
  position: relative;
  display: inline-block;
  cursor: pointer;
}
.goslide-toggle-label input {
  opacity: 0;
  width: 0;
  height: 0;
  position: absolute;
}
.goslide-toggle-switch {
  display: inline-block;
  width: 2.8em;
  height: 1.5em;
  background: var(--slide-muted);
  border-radius: 1em;
  position: relative;
  transition: background 0.2s;
  vertical-align: middle;
}
.goslide-toggle-switch::after {
  content: '';
  position: absolute;
  top: 0.2em;
  left: 0.2em;
  width: 1.1em;
  height: 1.1em;
  background: white;
  border-radius: 50%;
  transition: transform 0.2s;
}
.goslide-toggle-label input:checked + .goslide-toggle-switch {
  background: var(--slide-accent);
}
.goslide-toggle-label input:checked + .goslide-toggle-switch::after {
  transform: translateX(1.3em);
}
```

---

## 4. Testing Strategy

### Go-side

No new Go tests needed. L2 components use the same `Component{Type, Params}` → `renderComponents()` pipeline as L1. The `data-type="tabs"`, `"slider"`, `"toggle"`, `"panel:xxx"` rendering is already covered by existing `renderComponents` tests.

### Manual Checklist additions

```markdown
## Reactive Components (Phase 2c)
- [ ] tabs — button bar displays, click switches active style
- [ ] tabs → panel — clicking tab shows/hides corresponding panel content
- [ ] slider — range input displays with label + live value
- [ ] slider accent — slider color matches accent
- [ ] slider value — dragging updates value display in real-time
- [ ] toggle — switch style, click toggles on/off
- [ ] toggle accent — on-state color matches accent
- [ ] toggle → panel — toggle on shows panel, off hides it
- [ ] dark theme — all L2 controls visible on dark theme
- [ ] console debug — browser console `GoSlide.get('id')` returns current value
- [ ] fragment coexistence — slide with both fragments + slider, no interference
```

---

## 5. Files Changed Summary

| Action | File |
|--------|------|
| Create | `web/static/reactive.js` — GoSlide store + tabs/slider/toggle/panel init |
| Modify | `web/templates/slide.html` — add `reactive.js` script tag |
| Modify | `web/themes/tokens.css` — add tabs/slider/toggle CSS |
| Modify | `examples/demo.md` — add tabs/slider/toggle demo slides |
| Modify | `MANUAL_CHECKLIST.md` — add reactive checklist |
