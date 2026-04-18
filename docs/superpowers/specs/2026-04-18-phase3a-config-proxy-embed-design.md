# GoSlide Phase 3a — Config + API Proxy + Embed Design Spec

> **Status: COMPLETED (2026-04-18)**

## Decision Log

| # | Question | Decision |
|---|----------|----------|
| 1 | Phase 3 decomposition | 3a (config+proxy+embed), 3b (api component+render types), 3c (dashboard layout); chat deferred |
| 2 | Testing strategy | httptest unit tests + mock-api server for manual testing |
| 3 | Config scope | Only `api.proxy` in Phase 3a; cache and presentation defaults deferred |
| 4 | Env var expansion | `os.ExpandEnv` only, no shell/subshell/default syntax |
| 5 | embed:html security | Direct injection, no sanitize; security scanning deferred |

---

## 1. Config Parsing

### New package: `internal/config/`

```go
type Config struct {
    API APIConfig `yaml:"api"`
}

type APIConfig struct {
    Proxy map[string]ProxyTarget `yaml:"proxy"`
}

type ProxyTarget struct {
    Target  string            `yaml:"target"`
    Headers map[string]string `yaml:"headers"`
}
```

### Load function

```go
func Load(dir string) (*Config, error)
```

1. Look for `goslide.yaml` in `dir` (the directory containing the `.md` file)
2. If not found → return empty `&Config{}`, `nil` error (config is optional)
3. `yaml.Unmarshal` the file
4. For each `ProxyTarget.Headers` value, apply `os.ExpandEnv` to expand `${VAR}` references
5. Validate each `ProxyTarget.Target` is a parseable URL (`url.Parse`)
6. Return config or error

### Integration

`server.Run(opts)` calls `config.Load(filepath.Dir(opts.File))` during startup. The resulting config is passed to `newApp`.

---

## 2. API Proxy

### New file: `internal/server/proxy.go`

```go
func setupProxy(mux *http.ServeMux, proxies map[string]config.ProxyTarget)
```

For each proxy entry, creates a `httputil.ReverseProxy` with a custom `Director`:

```go
func makeProxyHandler(prefix string, target config.ProxyTarget) http.Handler {
    targetURL, _ := url.Parse(target.Target)
    proxy := httputil.NewSingleHostReverseProxy(targetURL)

    origDirector := proxy.Director
    proxy.Director = func(req *http.Request) {
        origDirector(req)
        req.URL.Path = strings.TrimPrefix(req.URL.Path, prefix)
        if req.URL.Path == "" { req.URL.Path = "/" }
        req.Host = targetURL.Host
        for k, v := range target.Headers {
            req.Header.Set(k, v)
        }
    }
    return proxy
}
```

### Routing

Proxy handlers registered on `mux` with their path prefix (e.g., `/api/aoi/`). `http.ServeMux` longest-match ensures static routes (`/`, `/themes/`, `/static/`, `/fonts/`, `/ws`) take priority.

### Request flow

```
Browser: GET /api/aoi/status?line=A
  → ServeMux matches prefix /api/aoi/
  → Director strips prefix → /status?line=A
  → Proxy to http://192.168.1.100:8000/status?line=A
  → Injects Authorization header from config
  → Response proxied back to browser
```

### Error handling

- Invalid target URL at startup → fatal error (prevents `newApp` from returning)
- Upstream unreachable → `httputil.ReverseProxy` returns 502 Bad Gateway
- No extra logging unless `--verbose`

---

## 3. `embed:html` + `embed:iframe`

Both are already recognized component types from Phase 2a. Need frontend rendering.

### embed:html

**Go side:** In `parseSlide` component post-processing, add:

```go
} else if comp.Type == "embed:html" {
    comp.ContentHTML = comp.Raw
}
```

Raw HTML goes directly into ContentHTML (not markdown-rendered). Since ContentHTML is output inside the component `<div>` by the renderer, the browser parses and executes it including `<script>` tags.

**JS side:** No init needed — content is already in the DOM from server-side render.

### embed:iframe

**Go side:** No changes needed. `Params` contains `url` and `height` from YAML.

**JS side:** Add to `components.js`:

```javascript
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
```

In `initAllComponents`, add:
```javascript
else if (type === 'embed:iframe') initIframe(el);
```

### CSS

```css
.goslide-component iframe {
    max-width: 100%;
    background: var(--slide-card-bg);
}
```

---

## 4. Mock API Server

### `examples/mock-api/main.go`

Standalone Go HTTP server (~60 lines) providing test endpoints:
- `GET /status?line=X` → `{"yield": 96.2, "line": "X", "status": "running"}`
- `GET /metrics` → `{"lines": [{"name": "Line A", "yield": 96.2}, ...]}`

### `examples/goslide.yaml`

```yaml
api:
  proxy:
    /api/mock:
      target: http://localhost:9999
```

### Usage

Terminal 1: `go run examples/mock-api/main.go` (starts on :9999)
Terminal 2: `go run ./cmd/goslide serve examples/demo.md --no-open`

Browser: `http://localhost:3000` — slides with embed:html can fetch `/api/mock/status`.

---

## 5. Testing

### Unit tests

| Package | Tests |
|---------|-------|
| `config` | Load with valid config; Load without config file (returns empty); YAML syntax error; env var expansion in headers; invalid target URL |
| `server/proxy` | Prefix stripping; header injection; multiple proxy paths; upstream 502 |

Proxy tests use `httptest.NewServer` to simulate upstream — no external dependencies.

### Manual checklist

```markdown
## API Proxy (Phase 3a)
- [ ] goslide.yaml absent → starts normally without proxy
- [ ] goslide.yaml present → proxy routes work
- [ ] /api/mock/status → returns mock JSON
- [ ] Header injection → upstream receives configured headers
- [ ] env var expansion → ${TOKEN} replaced with actual env var value
- [ ] embed:html → custom HTML/JS executes in slide
- [ ] embed:iframe → iframe displays external page
- [ ] upstream unreachable → 502 response
```

---

## 6. Files Changed Summary

| Action | File |
|--------|------|
| Create | `internal/config/config.go` — Config struct + Load() |
| Create | `internal/config/config_test.go` |
| Create | `internal/server/proxy.go` — setupProxy + makeProxyHandler |
| Create | `internal/server/proxy_test.go` |
| Modify | `internal/server/server.go` — load config, call setupProxy |
| Modify | `internal/parser/slide.go` — embed:html ContentHTML handling |
| Modify | `web/static/components.js` — initIframe |
| Modify | `web/themes/tokens.css` — iframe CSS |
| Create | `examples/mock-api/main.go` |
| Create | `examples/goslide.yaml` |
| Modify | `examples/demo.md` — embed:html + embed:iframe demo slides |
| Modify | `MANUAL_CHECKLIST.md` — Phase 3a checklist |
