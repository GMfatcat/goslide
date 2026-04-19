(function () {
  'use strict';
  var isStatic = document.body.dataset.mode === 'static';

  var initialized = {};

  function resolveColor(colorName) {
    var root = document.documentElement;
    if (!colorName) {
      return getComputedStyle(root).getPropertyValue('--slide-accent').trim();
    }
    var val = getComputedStyle(root).getPropertyValue('--accent-' + colorName).trim();
    return val || getComputedStyle(root).getPropertyValue('--slide-accent').trim();
  }

  var defaultPalette = ['blue', 'teal', 'coral', 'purple', 'amber', 'green', 'red', 'pink'];

  function getDatasetColors(count, userColor) {
    if (userColor) return Array(count).fill(resolveColor(userColor));
    var colors = [];
    for (var i = 0; i < count; i++) {
      colors.push(resolveColor(defaultPalette[i % defaultPalette.length]));
    }
    return colors;
  }

  function getThemeColors() {
    var style = getComputedStyle(document.documentElement);
    return {
      text: style.getPropertyValue('--slide-text').trim() || '#666666',
      border: style.getPropertyValue('--slide-border').trim() || 'rgba(0,0,0,0.1)',
      muted: style.getPropertyValue('--slide-muted').trim() || '#999999'
    };
  }

  function buildChartConfig(type, params) {
    var chartType = type;
    var tc = getThemeColors();
    var opts = {
      responsive: true,
      color: tc.text,
      plugins: {
        title: { display: false, color: tc.text },
        legend: { display: true, labels: { color: tc.text } }
      },
      scales: {
        x: { ticks: { color: tc.muted }, grid: { color: tc.border } },
        y: { ticks: { color: tc.muted }, grid: { color: tc.border } }
      }
    };

    if (chartType === 'sparkline') {
      chartType = 'line';
      opts.scales.x.display = false;
      opts.scales.y.display = false;
      opts.plugins.legend = { display: false };
      opts.plugins.title = { display: false };
      opts.elements = { point: { radius: 0 }, line: { borderWidth: 2 } };
      opts.maintainAspectRatio = false;
    }

    if (params.title) {
      opts.plugins.title = { display: true, text: params.title, color: tc.text };
    }

    if (params.stacked) {
      opts.scales = opts.scales || {};
      opts.scales.x = opts.scales.x || {};
      opts.scales.y = opts.scales.y || {};
      opts.scales.x.stacked = true;
      opts.scales.y.stacked = true;
    }

    if (params.unit) {
      var unit = params.unit;
      opts.plugins.tooltip = {
        callbacks: {
          label: function (ctx) {
            var label = ctx.dataset.label || '';
            var value = ctx.parsed.y !== undefined ? ctx.parsed.y : ctx.parsed;
            return label + ': ' + value + unit;
          }
        }
      };
    }

    var datasets;
    if (params.datasets) {
      datasets = params.datasets.map(function (ds, idx) {
        var color = resolveColor(ds.color || defaultPalette[idx % defaultPalette.length]);
        return {
          label: ds.label || '',
          data: ds.data || [],
          backgroundColor: color,
          borderColor: color,
          borderWidth: 1
        };
      });
    } else {
      var data = params.data || [];
      var color = resolveColor(params.color);
      var bgColors;
      if (chartType === 'pie' || chartType === 'radar') {
        bgColors = getDatasetColors(data.length, null);
      } else {
        bgColors = color;
      }
      datasets = [{
        label: params.title || '',
        data: data,
        backgroundColor: bgColors,
        borderColor: chartType === 'line' ? color : bgColors,
        borderWidth: chartType === 'line' ? 2 : 1,
        fill: chartType === 'line' ? false : undefined
      }];
    }

    return {
      type: chartType,
      data: {
        labels: params.labels || [],
        datasets: datasets
      },
      options: opts
    };
  }

  function initChart(el) {
    var fullType = el.getAttribute('data-type');
    var chartType = fullType.split(':')[1] || 'bar';
    var params = JSON.parse(decodeAttr(el.getAttribute('data-params')));

    var canvas = document.createElement('canvas');
    el.appendChild(canvas);

    var config = buildChartConfig(chartType, params);
    var chart = new Chart(canvas, config);
    el._chart = chart;
  }

  function initTable(el) {
    var params = JSON.parse(decodeAttr(el.getAttribute('data-params')));
    var table = document.createElement('table');

    var thead = document.createElement('thead');
    var headerRow = document.createElement('tr');
    (params.columns || []).forEach(function (col) {
      var th = document.createElement('th');
      th.textContent = col;
      headerRow.appendChild(th);
    });
    thead.appendChild(headerRow);
    table.appendChild(thead);

    var tbody = document.createElement('tbody');
    (params.rows || []).forEach(function (row) {
      var tr = document.createElement('tr');
      (Array.isArray(row) ? row : []).forEach(function (cell) {
        var td = document.createElement('td');
        td.textContent = String(cell);
        tr.appendChild(td);
      });
      tbody.appendChild(tr);
    });
    table.appendChild(tbody);
    el.appendChild(table);

    if (params.sortable) makeSortable(table);
  }

  function makeSortable(table) {
    table.querySelectorAll('th').forEach(function (th, colIdx) {
      th.style.cursor = 'pointer';
      th.addEventListener('click', function () { sortTable(table, colIdx, th); });
    });
  }

  function sortTable(table, colIdx, th) {
    var tbody = table.querySelector('tbody');
    var rows = Array.from(tbody.querySelectorAll('tr'));
    var asc = th.getAttribute('data-sort') !== 'asc';

    rows.sort(function (a, b) {
      var aText = a.cells[colIdx].textContent.trim();
      var bText = b.cells[colIdx].textContent.trim();
      var aNum = parseFloat(aText), bNum = parseFloat(bText);
      if (!isNaN(aNum) && !isNaN(bNum)) return asc ? aNum - bNum : bNum - aNum;
      return asc ? aText.localeCompare(bText) : bText.localeCompare(aText);
    });

    rows.forEach(function (row) { tbody.appendChild(row); });
    table.querySelectorAll('th').forEach(function (h) {
      h.removeAttribute('data-sort');
      h.textContent = h.textContent.replace(/ [▲▼]$/, '');
    });
    th.setAttribute('data-sort', asc ? 'asc' : 'desc');
    th.textContent += asc ? ' ▲' : ' ▼';
  }

  function initIframe(el) {
    var params = JSON.parse(decodeAttr(el.getAttribute('data-params')));
    var iframe = document.createElement('iframe');
    iframe.src = params.url || '';
    iframe.style.width = '100%';
    iframe.style.height = (params.height || 400) + 'px';
    iframe.style.border = 'none';
    iframe.style.borderRadius = '0.5rem';
    el.appendChild(iframe);
  }

  function extractPath(obj, path) {
    if (!path) return obj;
    var parts = path.replace(/\[(\d+)\]/g, '.$1').split('.');
    var current = obj;
    for (var i = 0; i < parts.length; i++) {
      if (current == null) return undefined;
      current = current[parts[i]];
    }
    return current;
  }

  function parseRefresh(s) {
    if (!s) return 0;
    var m = String(s).match(/^(\d+)(s|ms)?$/);
    if (!m) return 0;
    var val = parseInt(m[1]);
    if (m[2] === 'ms') return val;
    return val * 1000;
  }

  function renderMetric(container, item, data) {
    var div = document.createElement('div');
    div.className = 'goslide-metric';
    var value = document.createElement('div');
    value.className = 'goslide-metric-value';
    value.textContent = data + (item.unit || '');
    if (item.color) value.style.color = resolveColor(item.color);
    var label = document.createElement('div');
    label.className = 'goslide-metric-label';
    label.textContent = item.label || '';
    div.appendChild(value);
    div.appendChild(label);
    container.appendChild(div);
  }

  function renderApiChart(container, item, data) {
    var chartType = item.type.split(':')[1] || 'bar';
    var params = { title: item.title, color: item.color, unit: item.unit };
    if (Array.isArray(data) && data.length > 0 && typeof data[0] === 'object') {
      var keys = Object.keys(data[0]);
      var labelKey = null, dataKey = null;
      for (var k = 0; k < keys.length; k++) {
        if (typeof data[0][keys[k]] === 'string' && !labelKey) labelKey = keys[k];
        if (typeof data[0][keys[k]] === 'number' && !dataKey) dataKey = keys[k];
      }
      params.labels = data.map(function(d) { return d[labelKey || keys[0]]; });
      params.data = data.map(function(d) { return d[dataKey || keys[1]]; });
    } else if (Array.isArray(data)) {
      params.data = data;
      params.labels = item.labels || data.map(function(_, i) { return '' + i; });
    } else {
      params.data = [data];
      params.labels = [item.label || ''];
    }
    var canvas = document.createElement('canvas');
    container.appendChild(canvas);
    var config = buildChartConfig(chartType, params);
    new Chart(canvas, config);
  }

  function renderApiTable(container, item, data) {
    if (!Array.isArray(data) || data.length === 0) return;
    var columns = item.columns || Object.keys(data[0]);
    var table = document.createElement('table');
    var thead = document.createElement('thead');
    var headerRow = document.createElement('tr');
    columns.forEach(function(col) {
      var th = document.createElement('th');
      th.textContent = col;
      headerRow.appendChild(th);
    });
    thead.appendChild(headerRow);
    table.appendChild(thead);
    var tbody = document.createElement('tbody');
    data.forEach(function(obj) {
      var tr = document.createElement('tr');
      columns.forEach(function(col) {
        var td = document.createElement('td');
        td.textContent = obj[col] != null ? String(obj[col]) : '';
        tr.appendChild(td);
      });
      tbody.appendChild(tr);
    });
    table.appendChild(tbody);
    container.appendChild(table);
  }

  function renderJSON(container, item, data) {
    var pre = document.createElement('pre');
    pre.className = 'goslide-json';
    pre.textContent = JSON.stringify(data, null, 2);
    container.appendChild(pre);
  }

  function renderLog(container, item, data) {
    var pre = document.createElement('pre');
    pre.className = 'goslide-log';
    pre.textContent = Array.isArray(data) ? data.join('\n') : String(data);
    container.appendChild(pre);
  }

  function renderApiImage(container, item, data) {
    var img = document.createElement('img');
    var src = String(data);
    if (src.startsWith('data:') || src.startsWith('http')) {
      img.src = src;
    } else {
      img.src = 'data:image/png;base64,' + src;
    }
    img.style.maxWidth = '100%';
    img.style.borderRadius = '0.5rem';
    container.appendChild(img);
  }

  function renderMarkdownRaw(container, item, data) {
    var pre = document.createElement('pre');
    pre.className = 'goslide-markdown-raw';
    pre.textContent = String(data);
    container.appendChild(pre);
  }

  function renderItem(container, item, data) {
    if (data === undefined) return;
    var type = item.type || '';
    if (type === 'metric') renderMetric(container, item, data);
    else if (type.indexOf('chart') === 0) renderApiChart(container, item, data);
    else if (type === 'table') renderApiTable(container, item, data);
    else if (type === 'json') renderJSON(container, item, data);
    else if (type === 'log') renderLog(container, item, data);
    else if (type === 'image') renderApiImage(container, item, data);
    else if (type === 'markdown') renderMarkdownRaw(container, item, data);
  }

  function fetchAndRender(el, params) {
    var url = params.url;
    var opts = { method: (params.method || 'GET').toUpperCase() };
    if (params.body) {
      opts.body = JSON.stringify(params.body);
      opts.headers = { 'Content-Type': 'application/json' };
    }
    fetch(url, opts)
      .then(function(r) { return r.json(); })
      .then(function(json) {
        el.innerHTML = '';
        var items = document.createElement('div');
        items.className = params.layout === 'dashboard' ? 'goslide-api-dashboard' : 'goslide-api-items';
        var renderList = params.render;
        if (!Array.isArray(renderList)) {
          renderList = [renderList || { type: 'json' }];
        }
        renderList.forEach(function(item) {
          var data = extractPath(json, item.path);
          var wrapper = document.createElement('div');
          wrapper.className = 'goslide-api-item';
          if (item.span && item.span > 1) {
            wrapper.style.gridColumn = 'span ' + item.span;
          }
          renderItem(wrapper, item, data);
          items.appendChild(wrapper);
        });
        el.appendChild(items);
      })
      .catch(function(err) {
        el.innerHTML = '<pre class="goslide-api-error">API error: ' + err.message + '</pre>';
      });
  }

  function initApiComponent(el) {
    var params = JSON.parse(decodeAttr(el.getAttribute('data-params')));
    var refreshMs = parseRefresh(params.refresh);

    el._fetchFn = function() { fetchAndRender(el, params); };
    el._refreshMs = refreshMs;
    el._fetchFn();

    if (refreshMs > 0 && el.closest('section') === Reveal.getCurrentSlide()) {
      el._pollInterval = setInterval(el._fetchFn, refreshMs);
    }
  }

  function escapeText(s) {
    var div = document.createElement('div');
    div.textContent = s;
    return div.innerHTML;
  }

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

  function initAllComponents() {
    document.querySelectorAll('.goslide-component').forEach(function (el) {
      var id = el.getAttribute('data-comp-id');
      if (initialized[id]) return;
      var type = el.getAttribute('data-type');
      if (type.indexOf('chart') === 0) initChart(el);
      else if (type === 'table') initTable(el);
      else if (type === 'embed:iframe') initIframe(el);
      else if (type === 'placeholder') initPlaceholder(el);
      else if (type === 'api') {
        if (isStatic) {
          el.innerHTML = '<div style="color:var(--slide-muted);font-size:0.75em;text-align:center;padding:1rem;">API data requires goslide serve</div>';
        } else {
          initApiComponent(el);
        }
        initialized[id] = true;
        return;
      }
      initialized[id] = true;
    });
  }

  function parseBrightness(rgb) {
    var m = rgb.match(/\d+/g);
    if (!m) return 255;
    return (parseInt(m[0]) * 299 + parseInt(m[1]) * 587 + parseInt(m[2]) * 114) / 1000;
  }

  function initAllMermaid() {
    var els = document.querySelectorAll('.goslide-component[data-type="mermaid"]');
    if (els.length === 0) return;

    var bg = getComputedStyle(document.querySelector('.reveal')).backgroundColor;
    var isDark = parseBrightness(bg) < 128;
    mermaid.initialize({ startOnLoad: false, theme: isDark ? 'dark' : 'default' });

    els.forEach(function (el, idx) {
      var raw = decodeAttr(el.getAttribute('data-raw'));
      var id = 'goslide-mermaid-' + idx;
      mermaid.render(id, raw).then(function (result) {
        el.innerHTML = '<div class="mermaid">' + result.svg + '</div>';
      }).catch(function (err) {
        el.innerHTML = '<pre style="color:red;">Mermaid error: ' + err.message + '</pre>';
        console.error('Mermaid render error:', err);
      });
    });
  }

  function decodeAttr(s) {
    if (!s) return '';
    return s.replace(/&quot;/g, '"').replace(/&#39;/g, "'").replace(/&lt;/g, '<').replace(/&gt;/g, '>').replace(/&amp;/g, '&');
  }

  Reveal.on('ready', function () {
    initAllMermaid();
    initAllComponents();
  });
  if (!isStatic) {
    Reveal.on('slidechanged', function() {
      document.querySelectorAll('.goslide-component[data-type="api"]').forEach(function(el) {
        var isVisible = el.closest('section') === Reveal.getCurrentSlide();
        if (!isVisible && el._pollInterval) {
          clearInterval(el._pollInterval);
          el._pollInterval = null;
        } else if (isVisible && el._refreshMs && !el._pollInterval) {
          el._fetchFn();
          el._pollInterval = setInterval(el._fetchFn, el._refreshMs);
        }
      });
    });
  }
})();
