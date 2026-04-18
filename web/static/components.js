(function () {
  'use strict';

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

  function initAllComponents() {
    document.querySelectorAll('.goslide-component').forEach(function (el) {
      var id = el.getAttribute('data-comp-id');
      if (initialized[id]) return;
      var type = el.getAttribute('data-type');
      if (type.indexOf('chart') === 0) initChart(el);
      else if (type === 'table') initTable(el);
      else if (type === 'embed:iframe') initIframe(el);
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
})();
