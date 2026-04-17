# Phase 2c: L2 Reactive Components Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add tabs/panel switching, slider, toggle controls, and a reactive event bus (`GoSlide.set/get/on`) so L2 components can communicate on the same slide.

**Architecture:** A new `reactive.js` file provides the `GoSlide` store (pub/sub event bus) and initializes all L2 components at page load. Tabs publish their active index, panels subscribe to show/hide. Toggle publishes boolean, controlling panel visibility. Slider publishes numeric value. No Go-side changes needed — L2 uses the same `Component{Type, Params}` → `data-type`/`data-params` pipeline from Phase 2a/2b.

**Tech Stack:** Vanilla JS only. Native `<input type="range">` and `<input type="checkbox">` with CSS styling. No new dependencies.

**Shell rules (Windows):** Never chain commands with `&&`. Use SEPARATE Bash calls for `git add`, `git commit`, `go test`, `go build`. Use `GOTOOLCHAIN=local` prefix for ALL go commands. Use `-C` flag for go/git to specify directory.

---

## Task 1: Create `reactive.js` — Store + L2 Component Init

**Files:**
- Create: `web/static/reactive.js`

- [ ] **Step 1: Create reactive.js**

Create `web/static/reactive.js` with this exact content:

```javascript
(function () {
  'use strict';

  // Reactive store
  var store = {};
  var listeners = {};
  window.GoSlide = window.GoSlide || {};
  GoSlide.set = function (key, value) {
    store[key] = value;
    (listeners[key] || []).forEach(function (fn) { fn(value); });
  };
  GoSlide.get = function (key) { return store[key]; };
  GoSlide.on = function (key, fn) {
    (listeners[key] = listeners[key] || []).push(fn);
    if (key in store) fn(store[key]);
  };

  function decodeAttr(s) {
    if (!s) return '';
    return s.replace(/&quot;/g, '"').replace(/&#39;/g, "'").replace(/&lt;/g, '<').replace(/&gt;/g, '>').replace(/&amp;/g, '&');
  }

  // --- Tabs ---
  function initTabs(el) {
    var params = JSON.parse(decodeAttr(el.getAttribute('data-params')));
    var id = params.id;
    var labels = params.labels || [];

    var bar = document.createElement('div');
    bar.className = 'goslide-tabs';
    labels.forEach(function (label, idx) {
      var btn = document.createElement('button');
      btn.textContent = label;
      btn.className = 'goslide-tab';
      btn.addEventListener('click', function () { GoSlide.set(id, idx); });
      bar.appendChild(btn);
    });
    el.appendChild(bar);

    GoSlide.on(id, function (value) {
      bar.querySelectorAll('.goslide-tab').forEach(function (b, i) {
        b.classList.toggle('active', i === value);
      });
    });

    GoSlide.set(id, 0);
  }

  // --- Panel ---
  function initPanel(el) {
    var fullType = el.getAttribute('data-type');
    var parts = fullType.substring(6);
    var lastDash = parts.lastIndexOf('-');
    var suffix = parts.substring(lastDash + 1);

    if (!isNaN(parseInt(suffix)) && lastDash > 0) {
      var tabsId = parts.substring(0, lastDash);
      var panelIdx = parseInt(suffix);
      el.style.display = 'none';
      GoSlide.on(tabsId, function (value) {
        el.style.display = (value === panelIdx) ? '' : 'none';
      });
    } else {
      el.style.display = 'none';
      GoSlide.on(parts, function (value) {
        el.style.display = value ? '' : 'none';
      });
    }
  }

  // --- Slider ---
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

    input.addEventListener('input', function () {
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

  // --- Toggle ---
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

    input.addEventListener('change', function () {
      GoSlide.set(id, input.checked);
    });

    label.appendChild(input);
    label.appendChild(switchSpan);
    wrapper.appendChild(label);
    wrapper.appendChild(text);
    el.appendChild(wrapper);

    GoSlide.set(id, input.checked);
  }

  // --- Init all L2 components ---
  function initAllL2() {
    document.querySelectorAll('.goslide-component').forEach(function (el) {
      var type = el.getAttribute('data-type');
      if (type === 'tabs') initTabs(el);
      else if (type === 'slider') initSlider(el);
      else if (type === 'toggle') initToggle(el);
      else if (type.indexOf('panel:') === 0) initPanel(el);
    });
  }

  Reveal.on('ready', function () {
    initAllL2();
  });
})();
```

- [ ] **Step 2: Verify build**

Run: `GOTOOLCHAIN=local go build -C D:/CLAUDE-CODE-GOSLIDE ./web/`

- [ ] **Step 3: Commit**

```
git -C D:/CLAUDE-CODE-GOSLIDE add web/static/reactive.js
```
```
git -C D:/CLAUDE-CODE-GOSLIDE commit -m "feat(frontend): add reactive.js with GoSlide store and L2 component init"
```

---

## Task 2: Update `slide.html` — Add `reactive.js` Script Tag

**Files:**
- Modify: `web/templates/slide.html`

- [ ] **Step 1: Add reactive.js script tag**

Read `web/templates/slide.html`. Find:

```html
  <script src="/static/runtime.js"></script>
  <script src="/static/components.js"></script>
```

Replace with:

```html
  <script src="/static/runtime.js"></script>
  <script src="/static/reactive.js"></script>
  <script src="/static/components.js"></script>
```

- [ ] **Step 2: Verify build**

Run: `GOTOOLCHAIN=local go build -C D:/CLAUDE-CODE-GOSLIDE ./...`

- [ ] **Step 3: Run all tests**

Run: `GOTOOLCHAIN=local go test -C D:/CLAUDE-CODE-GOSLIDE ./internal/... -count=1`

Expected: all PASS (no Go changes, just template).

- [ ] **Step 4: Commit**

```
git -C D:/CLAUDE-CODE-GOSLIDE add web/templates/slide.html
```
```
git -C D:/CLAUDE-CODE-GOSLIDE commit -m "feat: add reactive.js script tag to slide template"
```

---

## Task 3: Add L2 Component CSS to `tokens.css`

**Files:**
- Modify: `web/themes/tokens.css`

- [ ] **Step 1: Append L2 CSS**

Read `web/themes/tokens.css`. Append at the end:

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

- [ ] **Step 2: Verify build**

Run: `GOTOOLCHAIN=local go build -C D:/CLAUDE-CODE-GOSLIDE ./...`

- [ ] **Step 3: Commit**

```
git -C D:/CLAUDE-CODE-GOSLIDE add web/themes/tokens.css
```
```
git -C D:/CLAUDE-CODE-GOSLIDE commit -m "feat: add tabs, slider, and toggle CSS styles"
```

---

## Task 4: Update Golden Tests

**Files:**
- Modify: golden `.html` files (regenerate due to template change)

- [ ] **Step 1: Regenerate golden files**

Run: `GOTOOLCHAIN=local go test -C D:/CLAUDE-CODE-GOSLIDE/internal/renderer -run TestGolden -v -args -update`

Expected: all subtests PASS.

- [ ] **Step 2: Verify golden tests pass without -update**

Run: `GOTOOLCHAIN=local go test -C D:/CLAUDE-CODE-GOSLIDE/internal/renderer -run TestGolden -v`

Expected: all PASS.

- [ ] **Step 3: Run all tests**

Run: `GOTOOLCHAIN=local go test -C D:/CLAUDE-CODE-GOSLIDE ./internal/... -count=1 -race`

Expected: all PASS.

- [ ] **Step 4: Commit**

```
git -C D:/CLAUDE-CODE-GOSLIDE add internal/renderer/testdata/golden/
```
```
git -C D:/CLAUDE-CODE-GOSLIDE commit -m "test: regenerate golden files after reactive.js template addition"
```

---

## Task 5: Demo + Manual Checklist

**Files:**
- Modify: `examples/demo.md`
- Modify: `MANUAL_CHECKLIST.md`

- [ ] **Step 1: Add reactive demo slides to demo.md**

Read `examples/demo.md`. Find the last slide (the "Thank You" slide with `<!-- layout: title -->`). INSERT these new slides BEFORE that "Thank You" slide:

```markdown

---

# Tabs Demo

~~~tabs
id: compare
labels: ["Plan A", "Plan B", "Plan C"]
~~~

~~~panel:compare-0
## Plan A

Low cost, quick start, limited scalability.
Best for small teams and prototypes.
~~~

~~~panel:compare-1
## Plan B

Medium investment, balanced approach.
Suitable for most production workloads.
~~~

~~~panel:compare-2
## Plan C

Full investment, enterprise features.
Maximum scalability and support.
~~~

---

# Interactive Controls

~~~slider
id: threshold
label: Yield threshold
min: 80
max: 100
value: 95
step: 0.5
unit: "%"
~~~

~~~toggle
id: show_details
label: Show details
default: false
~~~

~~~panel:show_details
### Detail View

When the toggle is on, this panel becomes visible.
The slider value can be read via `GoSlide.get('threshold')` in the browser console.
~~~
```

- [ ] **Step 2: Update MANUAL_CHECKLIST.md**

Read `MANUAL_CHECKLIST.md`. Append at the end:

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
- [ ] console debug — browser console `GoSlide.get('threshold')` returns current value
- [ ] fragment coexistence — slide with both fragments + slider, no interference
```

- [ ] **Step 3: Run final full test**

Run: `GOTOOLCHAIN=local go test -C D:/CLAUDE-CODE-GOSLIDE ./... -count=1 -race`

Expected: all PASS.

- [ ] **Step 4: Build binary**

Run: `GOTOOLCHAIN=local go build -C D:/CLAUDE-CODE-GOSLIDE -ldflags "-X main.version=0.3.0" -o D:/CLAUDE-CODE-GOSLIDE/goslide.exe ./cmd/goslide`

- [ ] **Step 5: Commit**

```
git -C D:/CLAUDE-CODE-GOSLIDE add examples/demo.md MANUAL_CHECKLIST.md
```
```
git -C D:/CLAUDE-CODE-GOSLIDE commit -m "feat: add tabs/slider/toggle demo slides and Phase 2c manual checklist"
```
