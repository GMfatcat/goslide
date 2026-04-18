# Phase 3a: Config + API Proxy + Embed Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add `goslide.yaml` config file parsing with API proxy support, reverse proxy handlers for upstream API services, and `embed:html`/`embed:iframe` component rendering.

**Architecture:** New `internal/config` package loads optional `goslide.yaml` from the `.md` file's directory. New `internal/server/proxy.go` creates `httputil.ReverseProxy` handlers for each configured path prefix. `embed:html` uses `ContentHTML` (raw HTML injection), `embed:iframe` uses JS init in `components.js`.

**Tech Stack:** Go 1.21.6, `net/http/httputil` (ReverseProxy), `gopkg.in/yaml.v3`, `os.ExpandEnv`. No new dependencies.

**Shell rules (Windows):** Never chain commands with `&&`. Use SEPARATE Bash calls for `git add`, `git commit`, `go test`, `go build`. Use `GOTOOLCHAIN=local` prefix for ALL go commands. Use `-C` flag for go/git to specify directory.

---

## Task 1: Config Package

**Files:**
- Create: `internal/config/config.go`
- Create: `internal/config/config_test.go`

- [ ] **Step 1: Write failing tests**

Create `internal/config/config_test.go`:

```go
package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoad_NoConfigFile(t *testing.T) {
	cfg, err := Load(t.TempDir())
	require.NoError(t, err)
	require.NotNil(t, cfg)
	require.Empty(t, cfg.API.Proxy)
}

func TestLoad_ValidConfig(t *testing.T) {
	dir := t.TempDir()
	content := `
api:
  proxy:
    /api/test:
      target: http://localhost:9999
      headers:
        X-Custom: "hello"
`
	os.WriteFile(filepath.Join(dir, "goslide.yaml"), []byte(content), 0644)

	cfg, err := Load(dir)
	require.NoError(t, err)
	require.Len(t, cfg.API.Proxy, 1)

	proxy := cfg.API.Proxy["/api/test"]
	require.Equal(t, "http://localhost:9999", proxy.Target)
	require.Equal(t, "hello", proxy.Headers["X-Custom"])
}

func TestLoad_EnvVarExpansion(t *testing.T) {
	dir := t.TempDir()
	os.Setenv("GOSLIDE_TEST_TOKEN", "secret123")
	defer os.Unsetenv("GOSLIDE_TEST_TOKEN")

	content := `
api:
  proxy:
    /api/secure:
      target: http://localhost:8000
      headers:
        Authorization: "Bearer ${GOSLIDE_TEST_TOKEN}"
`
	os.WriteFile(filepath.Join(dir, "goslide.yaml"), []byte(content), 0644)

	cfg, err := Load(dir)
	require.NoError(t, err)
	require.Equal(t, "Bearer secret123", cfg.API.Proxy["/api/secure"].Headers["Authorization"])
}

func TestLoad_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "goslide.yaml"), []byte("  bad:\n    yaml\n"), 0644)

	_, err := Load(dir)
	require.Error(t, err)
}

func TestLoad_InvalidTargetURL(t *testing.T) {
	dir := t.TempDir()
	content := `
api:
  proxy:
    /api/bad:
      target: "://not-a-url"
`
	os.WriteFile(filepath.Join(dir, "goslide.yaml"), []byte(content), 0644)

	_, err := Load(dir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid target URL")
}

func TestLoad_MultipleProxies(t *testing.T) {
	dir := t.TempDir()
	content := `
api:
  proxy:
    /api/aoi:
      target: http://localhost:8000
    /api/ollama:
      target: http://localhost:11434
`
	os.WriteFile(filepath.Join(dir, "goslide.yaml"), []byte(content), 0644)

	cfg, err := Load(dir)
	require.NoError(t, err)
	require.Len(t, cfg.API.Proxy, 2)
}
```

- [ ] **Step 2: Run tests to verify failure**

Run: `GOTOOLCHAIN=local go test -C D:/CLAUDE-CODE-GOSLIDE/internal/config -v`

Expected: compilation error — package doesn't exist yet.

- [ ] **Step 3: Implement config.go**

Create `internal/config/config.go`:

```go
package config

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

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

func Load(dir string) (*Config, error) {
	path := filepath.Join(dir, "goslide.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse goslide.yaml: %w", err)
	}

	for prefix, proxy := range cfg.API.Proxy {
		if proxy.Target != "" {
			if _, err := url.Parse(proxy.Target); err != nil {
				return nil, fmt.Errorf("invalid target URL for %s: %w", prefix, err)
			}
			if proxy.Target[0] == ':' {
				return nil, fmt.Errorf("invalid target URL for %s: %s", prefix, proxy.Target)
			}
		}

		for k, v := range proxy.Headers {
			proxy.Headers[k] = os.ExpandEnv(v)
		}
		cfg.API.Proxy[prefix] = proxy
	}

	return &cfg, nil
}
```

- [ ] **Step 4: Run tests**

Run: `GOTOOLCHAIN=local go test -C D:/CLAUDE-CODE-GOSLIDE/internal/config -v`

Expected: all PASS.

- [ ] **Step 5: Commit**

```
git -C D:/CLAUDE-CODE-GOSLIDE add internal/config/config.go internal/config/config_test.go
```
```
git -C D:/CLAUDE-CODE-GOSLIDE commit -m "feat(config): add goslide.yaml config parsing with proxy targets and env var expansion"
```

---

## Task 2: API Proxy Handler

**Files:**
- Create: `internal/server/proxy.go`
- Create: `internal/server/proxy_test.go`

- [ ] **Step 1: Write failing tests**

Create `internal/server/proxy_test.go`:

```go
package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/user/goslide/internal/config"
)

func TestProxy_BasicForward(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("upstream:" + r.URL.Path))
	}))
	defer upstream.Close()

	mux := http.NewServeMux()
	setupProxy(mux, map[string]config.ProxyTarget{
		"/api/test": {Target: upstream.URL},
	})

	req := httptest.NewRequest("GET", "/api/test/status", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "upstream:/status", rec.Body.String())
}

func TestProxy_PrefixStripping(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("path:" + r.URL.Path))
	}))
	defer upstream.Close()

	mux := http.NewServeMux()
	setupProxy(mux, map[string]config.ProxyTarget{
		"/api/v1": {Target: upstream.URL},
	})

	req := httptest.NewRequest("GET", "/api/v1/users/123", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	require.Equal(t, "path:/users/123", rec.Body.String())
}

func TestProxy_HeaderInjection(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("auth:" + r.Header.Get("Authorization")))
	}))
	defer upstream.Close()

	mux := http.NewServeMux()
	setupProxy(mux, map[string]config.ProxyTarget{
		"/api/secure": {
			Target:  upstream.URL,
			Headers: map[string]string{"Authorization": "Bearer token123"},
		},
	})

	req := httptest.NewRequest("GET", "/api/secure/data", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	require.Equal(t, "auth:Bearer token123", rec.Body.String())
}

func TestProxy_MultipleRoutes(t *testing.T) {
	upstream1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("svc1"))
	}))
	defer upstream1.Close()

	upstream2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("svc2"))
	}))
	defer upstream2.Close()

	mux := http.NewServeMux()
	setupProxy(mux, map[string]config.ProxyTarget{
		"/api/one": {Target: upstream1.URL},
		"/api/two": {Target: upstream2.URL},
	})

	req1 := httptest.NewRequest("GET", "/api/one/x", nil)
	rec1 := httptest.NewRecorder()
	mux.ServeHTTP(rec1, req1)
	require.Equal(t, "svc1", rec1.Body.String())

	req2 := httptest.NewRequest("GET", "/api/two/y", nil)
	rec2 := httptest.NewRecorder()
	mux.ServeHTTP(rec2, req2)
	require.Equal(t, "svc2", rec2.Body.String())
}

func TestProxy_UpstreamDown(t *testing.T) {
	mux := http.NewServeMux()
	setupProxy(mux, map[string]config.ProxyTarget{
		"/api/down": {Target: "http://127.0.0.1:59999"},
	})

	req := httptest.NewRequest("GET", "/api/down/test", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadGateway, rec.Code)
}

func TestProxy_RootPath(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("root:" + r.URL.Path))
	}))
	defer upstream.Close()

	mux := http.NewServeMux()
	setupProxy(mux, map[string]config.ProxyTarget{
		"/api/test": {Target: upstream.URL},
	})

	req := httptest.NewRequest("GET", "/api/test/", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	body, _ := io.ReadAll(rec.Result().Body)
	require.Equal(t, "root:/", string(body))
}
```

- [ ] **Step 2: Run tests to verify failure**

Run: `GOTOOLCHAIN=local go test -C D:/CLAUDE-CODE-GOSLIDE/internal/server -run TestProxy -v`

Expected: compilation error — `setupProxy` undefined.

- [ ] **Step 3: Implement proxy.go**

Create `internal/server/proxy.go`:

```go
package server

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/user/goslide/internal/config"
)

func setupProxy(mux *http.ServeMux, proxies map[string]config.ProxyTarget) {
	for prefix, target := range proxies {
		handler := makeProxyHandler(prefix, target)
		pattern := prefix
		if !strings.HasSuffix(pattern, "/") {
			pattern += "/"
		}
		mux.Handle(pattern, handler)
	}
}

func makeProxyHandler(prefix string, target config.ProxyTarget) http.Handler {
	targetURL, _ := url.Parse(target.Target)
	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = targetURL.Scheme
			req.URL.Host = targetURL.Host
			req.URL.Path = strings.TrimPrefix(req.URL.Path, prefix)
			if req.URL.Path == "" || req.URL.Path[0] != '/' {
				req.URL.Path = "/" + req.URL.Path
			}
			req.Host = targetURL.Host
			for k, v := range target.Headers {
				req.Header.Set(k, v)
			}
		},
	}
	return proxy
}
```

- [ ] **Step 4: Run proxy tests**

Run: `GOTOOLCHAIN=local go test -C D:/CLAUDE-CODE-GOSLIDE/internal/server -run TestProxy -v`

Expected: all PASS.

- [ ] **Step 5: Run all server tests**

Run: `GOTOOLCHAIN=local go test -C D:/CLAUDE-CODE-GOSLIDE/internal/server -v -race`

Expected: all PASS (proxy + broadcaster + handler + debounce tests).

- [ ] **Step 6: Commit**

```
git -C D:/CLAUDE-CODE-GOSLIDE add internal/server/proxy.go internal/server/proxy_test.go
```
```
git -C D:/CLAUDE-CODE-GOSLIDE commit -m "feat(server): add API reverse proxy with prefix stripping and header injection"
```

---

## Task 3: Integrate Config + Proxy into Server

**Files:**
- Modify: `internal/server/server.go`
- Modify: `internal/server/server_test.go`

- [ ] **Step 1: Add integration test**

Read `internal/server/server_test.go`. Append:

```go
func TestHandleProxy_WithConfig(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok","path":"` + r.URL.Path + `"}`))
	}))
	defer upstream.Close()

	dir := t.TempDir()
	mdPath := filepath.Join(dir, "test.md")
	os.WriteFile(mdPath, []byte("# Slide\n"), 0644)

	configContent := `
api:
  proxy:
    /api/test:
      target: ` + upstream.URL + `
`
	os.WriteFile(filepath.Join(dir, "goslide.yaml"), []byte(configContent), 0644)

	a, err := newApp(Options{File: mdPath, Port: 0, NoWatch: true, NoOpen: true})
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/api/test/hello", nil)
	rec := httptest.NewRecorder()
	a.mux.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), `"status":"ok"`)
}

func TestHandleProxy_NoConfig(t *testing.T) {
	dir := t.TempDir()
	mdPath := filepath.Join(dir, "test.md")
	os.WriteFile(mdPath, []byte("# Slide\n"), 0644)

	a, err := newApp(Options{File: mdPath, Port: 0, NoWatch: true, NoOpen: true})
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/api/anything", nil)
	rec := httptest.NewRecorder()
	a.mux.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
}
```

Add `"net/http/httptest"` to imports if not already there (it should be from existing tests).

- [ ] **Step 2: Update newApp to load config and setup proxy**

Read `internal/server/server.go`. Add `"path/filepath"` and `"github.com/user/goslide/internal/config"` to imports.

Replace the `newApp` function:

```go
func newApp(opts Options) (*app, error) {
	a := &app{
		opts:      opts,
		mux:       http.NewServeMux(),
		broadcast: newBroadcaster(),
	}
	a.setupRoutes()

	cfg, err := config.Load(filepath.Dir(opts.File))
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}
	if len(cfg.API.Proxy) > 0 {
		setupProxy(a.mux, cfg.API.Proxy)
	}

	if err := a.loadAndRender(); err != nil {
		return nil, err
	}
	return a, nil
}
```

- [ ] **Step 3: Run all tests**

Run: `GOTOOLCHAIN=local go test -C D:/CLAUDE-CODE-GOSLIDE ./internal/... -count=1 -race`

Expected: all PASS.

- [ ] **Step 4: Commit**

```
git -C D:/CLAUDE-CODE-GOSLIDE add internal/server/server.go internal/server/server_test.go
```
```
git -C D:/CLAUDE-CODE-GOSLIDE commit -m "feat(server): integrate config loading and proxy setup into newApp"
```

---

## Task 4: embed:html Parser Support

**Files:**
- Modify: `internal/parser/slide.go`
- Modify: `internal/parser/slide_test.go`

- [ ] **Step 1: Add test**

Read `internal/parser/slide_test.go`. Append:

```go
func TestParseSlide_EmbedHTML(t *testing.T) {
	raw := "# Custom\n\n~~~embed:html\n<div id=\"demo\">Hello</div>\n<script>document.getElementById('demo').textContent='World';</script>\n~~~\n"
	slide := parseSlide(1, raw, ir.Frontmatter{})
	require.Len(t, slide.Components, 1)
	require.Equal(t, "embed:html", slide.Components[0].Type)
	require.Contains(t, slide.Components[0].ContentHTML, "<div id=\"demo\">Hello</div>")
	require.Contains(t, slide.Components[0].ContentHTML, "<script>")
}

func TestParseSlide_EmbedIframe(t *testing.T) {
	raw := "~~~embed:iframe\nurl: http://example.com\nheight: 500\n~~~\n"
	slide := parseSlide(1, raw, ir.Frontmatter{})
	require.Len(t, slide.Components, 1)
	require.Equal(t, "embed:iframe", slide.Components[0].Type)
	require.Equal(t, "http://example.com", slide.Components[0].Params["url"])
	require.Equal(t, 500, slide.Components[0].Params["height"])
	require.Empty(t, slide.Components[0].ContentHTML)
}
```

- [ ] **Step 2: Add embed:html handling to slide.go**

Read `internal/parser/slide.go`. Find the component post-processing loop. Add after the card block:

```go
		} else if comp.Type == "embed:html" {
			components[i].ContentHTML = comp.Raw
		}
```

The full block becomes:
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
		} else if components[i].Type == "embed:html" {
			components[i].ContentHTML = components[i].Raw
		}
	}
```

Note: `embed:iframe` needs NO special handling — its YAML params are already parsed by `extractComponents`.

- [ ] **Step 3: Run tests**

Run: `GOTOOLCHAIN=local go test -C D:/CLAUDE-CODE-GOSLIDE/internal/parser -v`

Expected: all PASS.

- [ ] **Step 4: Commit**

```
git -C D:/CLAUDE-CODE-GOSLIDE add internal/parser/slide.go internal/parser/slide_test.go
```
```
git -C D:/CLAUDE-CODE-GOSLIDE commit -m "feat(parser): add embed:html ContentHTML support"
```

---

## Task 5: embed:iframe Frontend Init

**Files:**
- Modify: `web/static/components.js`
- Modify: `web/themes/tokens.css`

- [ ] **Step 1: Add initIframe to components.js**

Read `web/static/components.js`. Find the `initAllComponents` function. Add `initIframe` function before it:

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

Then in `initAllComponents`, find:

```javascript
      else if (type === 'table') initTable(el);
```

Add after it:

```javascript
      else if (type === 'embed:iframe') initIframe(el);
```

- [ ] **Step 2: Add iframe CSS to tokens.css**

Read `web/themes/tokens.css`. Append at the end:

```css

.goslide-component iframe {
  max-width: 100%;
  background: var(--slide-card-bg);
}
```

- [ ] **Step 3: Verify build**

Run: `GOTOOLCHAIN=local go build -C D:/CLAUDE-CODE-GOSLIDE ./...`

- [ ] **Step 4: Commit**

```
git -C D:/CLAUDE-CODE-GOSLIDE add web/static/components.js web/themes/tokens.css
```
```
git -C D:/CLAUDE-CODE-GOSLIDE commit -m "feat(frontend): add embed:iframe init and iframe CSS"
```

---

## Task 6: Mock API Server

**Files:**
- Create: `examples/mock-api/main.go`
- Create: `examples/goslide.yaml`

- [ ] **Step 1: Create mock API server**

Create `examples/mock-api/main.go`:

```go
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"yield":  96.2,
			"line":   r.URL.Query().Get("line"),
			"status": "running",
		})
	})

	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"lines": []map[string]any{
				{"name": "Line A", "yield": 96.2},
				{"name": "Line B", "yield": 93.8},
				{"name": "Line C", "yield": 97.1},
			},
		})
	})

	fmt.Println("Mock API running on http://localhost:9999")
	http.ListenAndServe(":9999", nil)
}
```

- [ ] **Step 2: Create examples/goslide.yaml**

Create `examples/goslide.yaml`:

```yaml
api:
  proxy:
    /api/mock:
      target: http://localhost:9999
```

- [ ] **Step 3: Verify mock server builds**

Run: `GOTOOLCHAIN=local go build -C D:/CLAUDE-CODE-GOSLIDE/examples/mock-api -o D:/CLAUDE-CODE-GOSLIDE/examples/mock-api/mock-api.exe`

- [ ] **Step 4: Commit**

```
git -C D:/CLAUDE-CODE-GOSLIDE add examples/mock-api/main.go examples/goslide.yaml
```
```
git -C D:/CLAUDE-CODE-GOSLIDE commit -m "feat: add mock API server and example goslide.yaml config"
```

---

## Task 7: Golden Tests + Demo + Checklist

**Files:**
- Modify: `examples/demo.md`
- Modify: `MANUAL_CHECKLIST.md`
- Modify: golden test files (regenerate)

- [ ] **Step 1: Regenerate golden files**

Run: `GOTOOLCHAIN=local go test -C D:/CLAUDE-CODE-GOSLIDE/internal/renderer -run TestGolden -v -args -update`

Run: `GOTOOLCHAIN=local go test -C D:/CLAUDE-CODE-GOSLIDE/internal/renderer -run TestGolden -v`

Expected: all PASS.

- [ ] **Step 2: Add embed demo slides to demo.md**

Read `examples/demo.md`. Find the "Thank You" slide. INSERT before it:

```markdown

---

# Embed HTML Demo

~~~embed:html
<div style="text-align:center; padding:1rem;">
  <button onclick="this.textContent='Clicked!'" style="padding:0.5rem 1.5rem; font-size:1rem; cursor:pointer; border-radius:0.5rem; border:1px solid #ccc;">
    Click Me
  </button>
  <p id="embed-time" style="margin-top:1rem; color:gray;"></p>
  <script>
    document.getElementById('embed-time').textContent = 'Rendered at: ' + new Date().toLocaleTimeString();
  </script>
</div>
~~~

---

# Embed Iframe Demo

~~~embed:iframe
url: https://example.com
height: 400
~~~
```

- [ ] **Step 3: Update MANUAL_CHECKLIST.md**

Read `MANUAL_CHECKLIST.md`. Append:

```markdown

## API Proxy (Phase 3a)
- [ ] goslide.yaml absent → starts normally without proxy
- [ ] goslide.yaml present → proxy routes work
- [ ] /api/mock/status → returns mock JSON (requires mock-api running)
- [ ] Header injection → upstream receives configured headers
- [ ] env var expansion → ${TOKEN} replaced with actual env var value
- [ ] embed:html → custom HTML/JS executes in slide (click button works)
- [ ] embed:iframe → iframe displays external page
- [ ] upstream unreachable → 502 response
```

- [ ] **Step 4: Run final full test**

Run: `GOTOOLCHAIN=local go test -C D:/CLAUDE-CODE-GOSLIDE ./... -count=1 -race`

Expected: all PASS.

- [ ] **Step 5: Build binary**

Run: `GOTOOLCHAIN=local go build -C D:/CLAUDE-CODE-GOSLIDE -ldflags "-X main.version=0.5.0" -o D:/CLAUDE-CODE-GOSLIDE/goslide.exe ./cmd/goslide`

- [ ] **Step 6: Commit**

```
git -C D:/CLAUDE-CODE-GOSLIDE add internal/renderer/testdata/golden/ examples/demo.md MANUAL_CHECKLIST.md
```
```
git -C D:/CLAUDE-CODE-GOSLIDE commit -m "feat: add embed demos, Phase 3a checklist, and regenerate golden files"
```
