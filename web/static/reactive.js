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

  // --- Init all L2 components ---
  function initAllL2() {
    document.querySelectorAll('.goslide-component').forEach(function (el) {
      var type = el.getAttribute('data-type');
      if (type === 'tabs') initTabs(el);
      else if (type === 'slider') initSlider(el);
      else if (type === 'toggle') initToggle(el);
      else if (type.indexOf('panel:') === 0) initPanel(el);
      else if (type === 'card') initCard(el);
    });
  }

  Reveal.on('ready', function () {
    initAllL2();
  });
})();
