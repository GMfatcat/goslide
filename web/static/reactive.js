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
        if (typeof Reveal !== 'undefined') Reveal.layout();
      });
    } else {
      el.style.display = 'none';
      GoSlide.on(parts, function (value) {
        el.style.display = value ? '' : 'none';
        if (typeof Reveal !== 'undefined') Reveal.layout();
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
