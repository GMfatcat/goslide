(function () {
  'use strict';
  var isStatic = document.body.dataset.mode === 'static';

  var saved = sessionStorage.getItem('goslide:indices');
  if (saved && window.Reveal) {
    Reveal.on('ready', function () {
      try {
        var idx = JSON.parse(saved);
        Reveal.slide(idx.h || 0, idx.v || 0, idx.f || -1);
      } catch (e) { /* ignore */ }
      sessionStorage.removeItem('goslide:indices');
    });
  }

  document.querySelectorAll('section[data-fragments="true"]').forEach(function (section) {
    var style = section.getAttribute('data-fragment-style') || '';
    var items = section.querySelectorAll('li');
    for (var i = 1; i < items.length; i++) {
      items[i].classList.add('fragment');
      if (style) items[i].classList.add(style);
    }
  });

  var toast = document.getElementById('goslide-toast');

  if (!isStatic) {
    var proto = location.protocol === 'https:' ? 'wss:' : 'ws:';
    var ws = new WebSocket(proto + '//' + location.host + '/ws');

    ws.addEventListener('message', function (ev) {
      try {
        var msg = JSON.parse(ev.data);
        if (msg.type === 'reload') {
          var indices = Reveal.getIndices();
          sessionStorage.setItem('goslide:indices', JSON.stringify(indices));
          location.reload();
        } else if (msg.type === 'error') {
          if (toast) {
            toast.textContent = msg.message;
            toast.hidden = false;
          }
        } else if (msg.type === 'ok') {
          if (toast) toast.hidden = true;
        }
      } catch (e) { /* ignore non-JSON */ }
    });

    ws.addEventListener('close', function () {
      setTimeout(function () {
        if (toast) {
          toast.textContent = 'goslide: server disconnected';
          toast.hidden = false;
        }
      }, 200);
    });

    // Presenter slide tracking
    var isPresenter = location.search.indexOf('role=presenter') !== -1;

    if (isPresenter) {
      function broadcastSlide() {
        if (ws.readyState === WebSocket.OPEN) {
          ws.send(JSON.stringify({
            type: 'presenter-slide',
            h: Reveal.getIndices().h,
            total: Reveal.getTotalSlides()
          }));
        }
      }
      Reveal.on('ready', broadcastSlide);
      Reveal.on('slidechanged', broadcastSlide);
    }

    // Viewer: show presenter indicator
    var presenterIndicator = null;

    ws.addEventListener('message', function (ev) {
      try {
        var msg = JSON.parse(ev.data);
        if (msg.type === 'presenter-slide' && !isPresenter) {
          if (!presenterIndicator) {
            presenterIndicator = document.createElement('div');
            presenterIndicator.id = 'goslide-presenter-indicator';
            var text = document.createElement('span');
            var btn = document.createElement('button');
            btn.textContent = 'Jump';
            btn.addEventListener('click', function () {
              Reveal.slide(presenterIndicator._presenterH || 0);
            });
            presenterIndicator.appendChild(text);
            presenterIndicator.appendChild(btn);
            document.body.appendChild(presenterIndicator);
          }
          presenterIndicator._presenterH = msg.h;
          presenterIndicator.querySelector('span').textContent =
            'Presenter: ' + (msg.h + 1) + '/' + msg.total + ' ';
          presenterIndicator.hidden = false;
        }
      } catch (e) { /* ignore */ }
    });
  }

  // Page number indicator
  var pageNumEl = document.getElementById('goslide-page-num');
  var slideNumberMode = document.body.getAttribute('data-slide-number') || '';
  var slideNumberFormat = document.body.getAttribute('data-slide-number-format') || 'total';
  var autoHideLayouts = ['title', 'section'];

  function updatePageNum() {
    if (!pageNumEl || !slideNumberMode || slideNumberMode === 'false') {
      if (pageNumEl) pageNumEl.hidden = true;
      return;
    }
    var indices = Reveal.getIndices();
    var current = indices.h + 1;
    if (slideNumberFormat === 'current') {
      pageNumEl.textContent = current;
    } else {
      pageNumEl.textContent = current + ' / ' + Reveal.getTotalSlides();
    }

    var slide = Reveal.getCurrentSlide();
    var isHidden = slide && slide.getAttribute('data-slide-number-hidden') === 'true';

    if (slideNumberMode === 'auto') {
      var layout = slide ? (slide.getAttribute('data-layout') || '') : '';
      pageNumEl.hidden = isHidden || autoHideLayouts.indexOf(layout) !== -1;
    } else {
      pageNumEl.hidden = isHidden;
    }
  }

  Reveal.on('ready', updatePageNum);
  Reveal.on('slidechanged', updatePageNum);
})();
