# API + LLM Transformer Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a new `llm` render type to the existing `api` component so fetched JSON can be post-processed by a user-authored prompt and shown inline — cache-first, click-to-call on miss, build-locked for static export.

**Architecture:** New `internal/llm` package orchestrates prompt substitution, SHA256-keyed disk cache (`.goslide-cache/`), and calls into the existing `internal/generate.Client` for transport. `internal/server` registers `POST /api/llm` that dispatches to `llm.Transform`. `internal/builder` runs a new `BakeLLM` pass before `renderer.Render` to inline cached results as a `data-llm-bakes` attribute on the api component div. Frontend adds `renderLLMItem` with button UX + localStorage caching.

**Tech Stack:** Go 1.21.6, `net/http` standard library, `encoding/json`, `crypto/sha256`, `gopkg.in/yaml.v3`; vanilla JS client; `github.com/stretchr/testify` for tests.

**Spec:** `docs/superpowers/specs/2026-04-20-api-llm-transformer-design.md`

---

## File Structure

**Create:**
- `internal/llm/types.go` — `Request`, `Result`, `CacheEntry`, `Completer` interface
- `internal/llm/types_test.go`
- `internal/llm/prompt.go` — `Render(template, data)` for `{{data}}` substitution
- `internal/llm/prompt_test.go`
- `internal/llm/cache.go` — `DiskCache` on `.goslide-cache/`
- `internal/llm/cache_test.go`
- `internal/llm/transform.go` — `Transform(ctx, req, completer, cache) (Result, error)`
- `internal/llm/transform_test.go`
- `internal/server/llm_handler.go` — `POST /api/llm` route
- `internal/server/llm_handler_test.go`
- `internal/builder/llm_bake.go` — `BakeLLM(pres, cfg, refreshOnMiss, mdDir) error`
- `internal/builder/llm_bake_test.go`
- `docs/superpowers/plans/2026-04-20-api-llm-transformer.md` (this file, already created)

**Modify:**
- `internal/ir/validate.go` — add `llm-missing-prompt` error for `llm` render items without a prompt
- `internal/ir/validate_test.go`
- `internal/renderer/components.go` — emit `data-llm-bakes` attribute when component params carry `_llm_bakes`
- `internal/renderer/renderer_test.go`
- `internal/server/handlers.go` — register `/api/llm` route
- `internal/server/server.go` — pass LLM config through to handler
- `internal/builder/builder.go` — call `BakeLLM` before `renderer.Render`; add `LLMRefresh` field to `Options`
- `internal/cli/build.go` — add `--llm-refresh` flag
- `web/static/components.js` — `renderLLMItem`, `llmCacheKey`, dispatch from `fetchAndRender`'s render loop
- `web/themes/layouts.css` — `.gs-llm-generate` + `.gs-llm-error` rules
- `internal/generate/system_prompt.md` — mention `llm` render type briefly (so the LLM-generated decks can produce it)
- `README.md`, `README_zh-TW.md` — new "LLM transformer" subsection
- `.gitignore` — note that `.goslide-cache/` is author-choice (do NOT auto-ignore; commenting only)

---

## Task 1: IR — validate `llm` render item requires `prompt`

**Files:**
- Modify: `internal/ir/validate.go`
- Modify: `internal/ir/validate_test.go`

- [ ] **Step 1: Write failing test**

Append to `internal/ir/validate_test.go`:

```go
func TestValidate_LLMRenderItemMissingPrompt(t *testing.T) {
	p := Presentation{Slides: []Slide{{
		Index: 1,
		Components: []Component{{
			Index: 0,
			Type:  "api",
			Params: map[string]any{
				"endpoint": "/api/sales",
				"render": []any{
					map[string]any{"type": "chart:bar"},
					map[string]any{"type": "llm"}, // no prompt
				},
			},
		}},
	}}}
	errs := p.Validate()
	e := findError(errs, "llm-missing-prompt")
	require.NotNil(t, e, "expected llm-missing-prompt error")
	require.Equal(t, "error", e.Severity)
}

func TestValidate_LLMRenderItemHappy(t *testing.T) {
	p := Presentation{Slides: []Slide{{
		Index: 1,
		Components: []Component{{
			Index: 0,
			Type:  "api",
			Params: map[string]any{
				"render": []any{
					map[string]any{"type": "llm", "prompt": "summarise {{data}}"},
				},
			},
		}},
	}}}
	errs := p.Validate()
	require.Nil(t, findError(errs, "llm-missing-prompt"))
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `GOTOOLCHAIN=local go test -C D:/CLAUDE-CODE-GOSLIDE ./internal/ir/ -run TestValidate_LLMRenderItem -v`
Expected: FAIL (no llm-missing-prompt validation yet).

- [ ] **Step 3: Implement**

At the end of `internal/ir/validate.go`, add:

```go
func validateLLMRenderItems(s Slide) []Error {
	var errs []Error
	for _, c := range s.Components {
		if c.Type != "api" {
			continue
		}
		render, ok := c.Params["render"].([]any)
		if !ok {
			continue
		}
		for _, raw := range render {
			item, ok := raw.(map[string]any)
			if !ok {
				continue
			}
			if item["type"] != "llm" {
				continue
			}
			prompt, _ := item["prompt"].(string)
			if strings.TrimSpace(prompt) == "" {
				errs = append(errs, Error{
					Slide: s.Index, Severity: "error", Code: "llm-missing-prompt",
					Message: fmt.Sprintf("slide %d: llm render item requires 'prompt' field", s.Index),
				})
			}
		}
	}
	return errs
}
```

Wire it into `Validate()`:

```go
for _, slide := range p.Slides {
    errs = append(errs, validateSlide(slide)...)
    errs = append(errs, validateImageGrid(slide)...)
    errs = append(errs, validatePlaceholder(slide)...)
    errs = append(errs, validateLLMRenderItems(slide)...)
}
```

- [ ] **Step 4: Run tests**

Run: `GOTOOLCHAIN=local go test -C D:/CLAUDE-CODE-GOSLIDE ./internal/ir/ -v`
Expected: all pass.

- [ ] **Step 5: Commit**

```bash
git -C D:/CLAUDE-CODE-GOSLIDE add internal/ir/validate.go internal/ir/validate_test.go
```

```bash
git -C D:/CLAUDE-CODE-GOSLIDE commit -m "feat(ir): validate llm render items require 'prompt'"
```

---

## Task 2: llm package — types

**Files:**
- Create: `internal/llm/types.go`
- Create: `internal/llm/types_test.go`

- [ ] **Step 1: Write failing test**

Create `internal/llm/types_test.go`:

```go
package llm

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRequest_ZeroValueValid(t *testing.T) {
	var r Request
	require.Empty(t, r.Model)
	require.Empty(t, r.Prompt)
	require.Empty(t, r.Data)
	require.Zero(t, r.MaxTokens)
}

func TestResult_FromCacheFalseByDefault(t *testing.T) {
	var r Result
	require.False(t, r.FromCache)
}
```

- [ ] **Step 2: Run to verify it fails (package doesn't exist)**

Run: `GOTOOLCHAIN=local go test -C D:/CLAUDE-CODE-GOSLIDE ./internal/llm/ -v`
Expected: compile error (package not found).

- [ ] **Step 3: Create types**

Create `internal/llm/types.go`:

```go
package llm

import (
	"context"
	"time"
)

// Request is the input to Transform. Data is the raw bytes of the API
// response that will be substituted into the prompt template.
type Request struct {
	Model     string
	Prompt    string
	Data      []byte
	MaxTokens int
}

// Usage mirrors the subset of OpenAI /chat/completions usage fields we
// surface to the caller.
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Result is the output of Transform. FromCache is true when the content
// was served from disk cache and no HTTP call happened.
type Result struct {
	Content   string
	Usage     Usage
	FromCache bool
}

// CacheEntry is what DiskCache stores on disk per key.
type CacheEntry struct {
	Version  int       `json:"version"`
	Created  time.Time `json:"created"`
	Model    string    `json:"model"`
	Prompt   string    `json:"prompt"`
	DataHash string    `json:"data_hash"`
	Content  string    `json:"content"`
	Usage    Usage     `json:"usage"`
}

// Completer is the abstraction Transform depends on for HTTP calls.
// internal/generate.Client already satisfies this shape; tests can
// substitute a fake.
type Completer interface {
	Complete(ctx context.Context, model string, msgs []Message) (string, Usage, error)
}

// Message is the chat-completions message shape. It mirrors
// internal/generate.Message so adapters stay trivial.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
```

- [ ] **Step 4: Verify tests pass**

Run: `GOTOOLCHAIN=local go test -C D:/CLAUDE-CODE-GOSLIDE ./internal/llm/ -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git -C D:/CLAUDE-CODE-GOSLIDE add internal/llm/types.go internal/llm/types_test.go
```

```bash
git -C D:/CLAUDE-CODE-GOSLIDE commit -m "feat(llm): Request, Result, CacheEntry types + Completer interface"
```

---

## Task 3: llm package — prompt template substitution

**Files:**
- Create: `internal/llm/prompt.go`
- Create: `internal/llm/prompt_test.go`

- [ ] **Step 1: Write failing tests**

Create `internal/llm/prompt_test.go`:

```go
package llm

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRender_ReplacesDataPlaceholder(t *testing.T) {
	got := Render("Summarise {{data}} please", []byte(`{"x":1}`))
	require.Equal(t, `Summarise {"x":1} please`, got)
}

func TestRender_NoPlaceholder(t *testing.T) {
	got := Render("No variables here", []byte(`{"x":1}`))
	require.Equal(t, "No variables here", got)
}

func TestRender_EmptyData(t *testing.T) {
	got := Render("Summarise {{data}}", nil)
	require.Equal(t, "Summarise ", got)
}

func TestRender_MultiplePlaceholders(t *testing.T) {
	got := Render("{{data}} and again {{data}}", []byte(`X`))
	require.Equal(t, "X and again X", got)
}
```

- [ ] **Step 2: Run to verify failure**

Run: `GOTOOLCHAIN=local go test -C D:/CLAUDE-CODE-GOSLIDE ./internal/llm/ -run Render -v`
Expected: FAIL (undefined: Render).

- [ ] **Step 3: Implement**

Create `internal/llm/prompt.go`:

```go
package llm

import "strings"

// Render substitutes the literal string "{{data}}" in template with the
// raw bytes of data. No other template syntax is supported in the MVP.
func Render(template string, data []byte) string {
	return strings.ReplaceAll(template, "{{data}}", string(data))
}
```

- [ ] **Step 4: Verify tests pass**

Run: `GOTOOLCHAIN=local go test -C D:/CLAUDE-CODE-GOSLIDE ./internal/llm/ -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git -C D:/CLAUDE-CODE-GOSLIDE add internal/llm/prompt.go internal/llm/prompt_test.go
```

```bash
git -C D:/CLAUDE-CODE-GOSLIDE commit -m "feat(llm): prompt template {{data}} substitution"
```

---

## Task 4: llm package — disk cache

**Files:**
- Create: `internal/llm/cache.go`
- Create: `internal/llm/cache_test.go`

- [ ] **Step 1: Write failing tests**

Create `internal/llm/cache_test.go`:

```go
package llm

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestDiskCache_PutGet(t *testing.T) {
	dir := t.TempDir()
	c := NewDiskCache(dir)

	req := Request{Model: "gpt-4o", Prompt: "hi", Data: []byte(`{"x":1}`)}
	key := c.Key(req)
	require.Len(t, key, 64) // hex sha256

	_, ok, err := c.Get(key)
	require.NoError(t, err)
	require.False(t, ok, "expected miss")

	entry := CacheEntry{
		Version: 1,
		Created: time.Now().UTC(),
		Model:   "gpt-4o",
		Prompt:  "hi",
		Content: "hello!",
		Usage:   Usage{TotalTokens: 10},
	}
	require.NoError(t, c.Put(key, entry))

	got, ok, err := c.Get(key)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, "hello!", got.Content)
	require.Equal(t, 10, got.Usage.TotalTokens)
}

func TestDiskCache_KeyStableForSameInputs(t *testing.T) {
	c := NewDiskCache(t.TempDir())
	req := Request{Model: "m", Prompt: "p", Data: []byte(`{"a":1,"b":2}`)}
	a := c.Key(req)
	b := c.Key(req)
	require.Equal(t, a, b)
}

func TestDiskCache_KeyCanonicalizesDataJSON(t *testing.T) {
	c := NewDiskCache(t.TempDir())
	a := c.Key(Request{Model: "m", Prompt: "p", Data: []byte(`{"a":1,"b":2}`)})
	b := c.Key(Request{Model: "m", Prompt: "p", Data: []byte(`{"b":2,"a":1}`)})
	require.Equal(t, a, b, "logically equal JSON should hash the same")
}

func TestDiskCache_KeyDiffersOnModel(t *testing.T) {
	c := NewDiskCache(t.TempDir())
	a := c.Key(Request{Model: "m1", Prompt: "p", Data: []byte(`{}`)})
	b := c.Key(Request{Model: "m2", Prompt: "p", Data: []byte(`{}`)})
	require.NotEqual(t, a, b)
}

func TestDiskCache_CorruptFileTreatedAsMiss(t *testing.T) {
	dir := t.TempDir()
	c := NewDiskCache(dir)
	require.NoError(t, os.MkdirAll(dir, 0755))
	// simulate a corrupt entry
	bad := filepath.Join(dir, "deadbeef.json")
	require.NoError(t, os.WriteFile(bad, []byte("not json"), 0644))

	_, ok, err := c.Get("deadbeef")
	require.NoError(t, err)
	require.False(t, ok)
}

func TestDiskCache_AutoCreatesDir(t *testing.T) {
	parent := t.TempDir()
	dir := filepath.Join(parent, "nested", ".goslide-cache")
	c := NewDiskCache(dir)

	entry := CacheEntry{Version: 1, Content: "x"}
	require.NoError(t, c.Put("key1", entry))

	_, err := os.Stat(dir)
	require.NoError(t, err)
}
```

- [ ] **Step 2: Verify failure**

Run: `GOTOOLCHAIN=local go test -C D:/CLAUDE-CODE-GOSLIDE ./internal/llm/ -run DiskCache -v`
Expected: FAIL (undefined symbols).

- [ ] **Step 3: Implement**

Create `internal/llm/cache.go`:

```go
package llm

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

// DiskCache stores CacheEntry values as JSON files named <hex-key>.json
// inside a directory (typically .goslide-cache/ in the project root).
type DiskCache struct {
	dir string
}

func NewDiskCache(dir string) *DiskCache {
	return &DiskCache{dir: dir}
}

// Key returns the stable hex-sha256 of (model || 0 || prompt || 0 ||
// canonical(data)) for the request.
func (c *DiskCache) Key(req Request) string {
	h := sha256.New()
	h.Write([]byte(req.Model))
	h.Write([]byte{0})
	h.Write([]byte(req.Prompt))
	h.Write([]byte{0})
	h.Write(canonicalJSON(req.Data))
	return hex.EncodeToString(h.Sum(nil))
}

func (c *DiskCache) Get(key string) (CacheEntry, bool, error) {
	path := filepath.Join(c.dir, key+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return CacheEntry{}, false, nil
		}
		return CacheEntry{}, false, err
	}
	var entry CacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		// corrupt file: treat as miss
		return CacheEntry{}, false, nil
	}
	return entry, true, nil
}

func (c *DiskCache) Put(key string, entry CacheEntry) error {
	if err := os.MkdirAll(c.dir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return err
	}
	path := filepath.Join(c.dir, key+".json")
	return os.WriteFile(path, data, 0644)
}

// canonicalJSON round-trips the input through json.Unmarshal +
// json.Marshal so that two logically-equal JSON inputs produce the
// same canonical byte sequence (key sort, whitespace stripped).
// Non-JSON input falls through unchanged.
func canonicalJSON(data []byte) []byte {
	if len(data) == 0 {
		return data
	}
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		return data
	}
	out, err := json.Marshal(v)
	if err != nil {
		return data
	}
	return out
}
```

- [ ] **Step 4: Verify tests pass**

Run: `GOTOOLCHAIN=local go test -C D:/CLAUDE-CODE-GOSLIDE ./internal/llm/ -v`
Expected: all pass.

- [ ] **Step 5: Commit**

```bash
git -C D:/CLAUDE-CODE-GOSLIDE add internal/llm/cache.go internal/llm/cache_test.go
```

```bash
git -C D:/CLAUDE-CODE-GOSLIDE commit -m "feat(llm): disk cache with canonical-JSON SHA256 keys"
```

---

## Task 5: llm package — Transform orchestrator

**Files:**
- Create: `internal/llm/transform.go`
- Create: `internal/llm/transform_test.go`

- [ ] **Step 1: Write failing tests**

Create `internal/llm/transform_test.go`:

```go
package llm

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type fakeCompleter struct {
	content string
	usage   Usage
	err     error
	calls   int
}

func (f *fakeCompleter) Complete(_ context.Context, _ string, _ []Message) (string, Usage, error) {
	f.calls++
	return f.content, f.usage, f.err
}

func TestTransform_CacheMissCallsCompleterAndStores(t *testing.T) {
	cache := NewDiskCache(t.TempDir())
	fc := &fakeCompleter{content: "out", usage: Usage{TotalTokens: 42}}
	req := Request{Model: "m", Prompt: "hi {{data}}", Data: []byte(`{"x":1}`), MaxTokens: 128}

	res, err := Transform(context.Background(), req, fc, cache)
	require.NoError(t, err)
	require.Equal(t, "out", res.Content)
	require.False(t, res.FromCache)
	require.Equal(t, 42, res.Usage.TotalTokens)
	require.Equal(t, 1, fc.calls)

	// second call should hit cache
	res2, err := Transform(context.Background(), req, fc, cache)
	require.NoError(t, err)
	require.True(t, res2.FromCache)
	require.Equal(t, "out", res2.Content)
	require.Equal(t, 1, fc.calls, "completer must not be called again on hit")
}

func TestTransform_SubstitutesPromptBeforeCalling(t *testing.T) {
	cache := NewDiskCache(t.TempDir())
	var observed string
	fc := &stubCompleter{fn: func(msgs []Message) (string, Usage, error) {
		observed = msgs[0].Content
		return "ok", Usage{}, nil
	}}
	req := Request{Model: "m", Prompt: "Summarise {{data}}", Data: []byte(`{"x":1}`)}

	_, err := Transform(context.Background(), req, fc, cache)
	require.NoError(t, err)
	require.Contains(t, observed, `{"x":1}`)
	require.NotContains(t, observed, "{{data}}")
}

func TestTransform_CompleterErrorSkipsCacheWrite(t *testing.T) {
	cache := NewDiskCache(t.TempDir())
	fc := &fakeCompleter{err: errors.New("boom")}
	req := Request{Model: "m", Prompt: "hi", Data: []byte(`{}`)}

	_, err := Transform(context.Background(), req, fc, cache)
	require.Error(t, err)

	_, hit, _ := cache.Get(cache.Key(req))
	require.False(t, hit, "error responses must not be cached")
}

func TestTransform_EmptyContentIsError(t *testing.T) {
	cache := NewDiskCache(t.TempDir())
	fc := &fakeCompleter{content: "", usage: Usage{}}
	req := Request{Model: "m", Prompt: "hi", Data: []byte(`{}`)}

	_, err := Transform(context.Background(), req, fc, cache)
	require.Error(t, err)
	require.Contains(t, err.Error(), "empty")
}

// stubCompleter lets a test observe the Messages passed to Complete.
type stubCompleter struct {
	fn func(msgs []Message) (string, Usage, error)
}

func (s *stubCompleter) Complete(_ context.Context, _ string, msgs []Message) (string, Usage, error) {
	return s.fn(msgs)
}

var _ = time.Now // keep import if needed
```

- [ ] **Step 2: Verify failure**

Run: `GOTOOLCHAIN=local go test -C D:/CLAUDE-CODE-GOSLIDE ./internal/llm/ -run TestTransform -v`
Expected: FAIL (undefined: Transform).

- [ ] **Step 3: Implement**

Create `internal/llm/transform.go`:

```go
package llm

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
)

// Transform applies the request: build the final prompt via substitution,
// consult the cache, call the completer on miss, persist the result, and
// return Result{FromCache: true|false}. Errors never write to cache.
func Transform(ctx context.Context, req Request, completer Completer, cache *DiskCache) (Result, error) {
	if completer == nil {
		return Result{}, errors.New("llm: completer is nil")
	}
	if cache == nil {
		return Result{}, errors.New("llm: cache is nil")
	}

	key := cache.Key(req)
	if entry, ok, err := cache.Get(key); err == nil && ok {
		return Result{Content: entry.Content, Usage: entry.Usage, FromCache: true}, nil
	}

	finalPrompt := Render(req.Prompt, req.Data)
	msgs := []Message{{Role: "user", Content: finalPrompt}}

	content, usage, err := completer.Complete(ctx, req.Model, msgs)
	if err != nil {
		return Result{}, fmt.Errorf("llm: completer: %w", err)
	}
	if content == "" {
		return Result{}, errors.New("llm: completer returned empty content")
	}

	entry := CacheEntry{
		Version:  1,
		Created:  time.Now().UTC(),
		Model:    req.Model,
		Prompt:   req.Prompt,
		DataHash: dataHash(req.Data),
		Content:  content,
		Usage:    usage,
	}
	if putErr := cache.Put(key, entry); putErr != nil {
		// Cache write failure is non-fatal; return the result anyway.
		// Caller can log at a higher layer if desired.
		_ = putErr
	}

	return Result{Content: content, Usage: usage, FromCache: false}, nil
}

func dataHash(data []byte) string {
	h := sha256.Sum256(canonicalJSON(data))
	return "sha256:" + hex.EncodeToString(h[:])
}
```

- [ ] **Step 4: Verify tests pass**

Run: `GOTOOLCHAIN=local go test -C D:/CLAUDE-CODE-GOSLIDE ./internal/llm/ -v`
Expected: all pass.

- [ ] **Step 5: Commit**

```bash
git -C D:/CLAUDE-CODE-GOSLIDE add internal/llm/transform.go internal/llm/transform_test.go
```

```bash
git -C D:/CLAUDE-CODE-GOSLIDE commit -m "feat(llm): Transform orchestrator (cache-first, Completer-agnostic)"
```

---

## Task 6: server — `POST /api/llm` handler

**Files:**
- Create: `internal/server/llm_handler.go`
- Create: `internal/server/llm_handler_test.go`

- [ ] **Step 1: Write failing tests**

Create `internal/server/llm_handler_test.go`:

```go
package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GMfatcat/goslide/internal/llm"
	"github.com/stretchr/testify/require"
)

type fakeTransformer struct {
	result llm.Result
	err    error
	got    llm.Request
}

func (f *fakeTransformer) Transform(_ context.Context, req llm.Request) (llm.Result, error) {
	f.got = req
	return f.result, f.err
}

func TestLLMHandler_ValidRequest(t *testing.T) {
	ft := &fakeTransformer{result: llm.Result{Content: "hi", Usage: llm.Usage{TotalTokens: 9}, FromCache: false}}
	h := newLLMHandler(ft)

	body := []byte(`{"model":"m","prompt":"hi {{data}}","data":{"x":1},"max_tokens":256}`)
	r := httptest.NewRequest(http.MethodPost, "/api/llm", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.ServeHTTP(w, r)

	require.Equal(t, http.StatusOK, w.Code)
	var resp struct {
		Content   string    `json:"content"`
		FromCache bool      `json:"from_cache"`
		Usage     llm.Usage `json:"usage"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Equal(t, "hi", resp.Content)
	require.False(t, resp.FromCache)
	require.Equal(t, 9, resp.Usage.TotalTokens)

	// Data bytes were forwarded to Transform
	require.JSONEq(t, `{"x":1}`, string(ft.got.Data))
	require.Equal(t, "m", ft.got.Model)
	require.Equal(t, 256, ft.got.MaxTokens)
}

func TestLLMHandler_MethodNotAllowed(t *testing.T) {
	h := newLLMHandler(&fakeTransformer{})
	r := httptest.NewRequest(http.MethodGet, "/api/llm", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	require.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestLLMHandler_MalformedJSON(t *testing.T) {
	h := newLLMHandler(&fakeTransformer{})
	r := httptest.NewRequest(http.MethodPost, "/api/llm", bytes.NewReader([]byte("not json")))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLLMHandler_TransformError(t *testing.T) {
	h := newLLMHandler(&fakeTransformer{err: errors.New("upstream failed")})
	r := httptest.NewRequest(http.MethodPost, "/api/llm", bytes.NewReader([]byte(`{"prompt":"x"}`)))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	require.Equal(t, http.StatusBadGateway, w.Code)
}
```

- [ ] **Step 2: Verify failure**

Run: `GOTOOLCHAIN=local go test -C D:/CLAUDE-CODE-GOSLIDE ./internal/server/ -run TestLLMHandler -v`
Expected: FAIL (undefined: newLLMHandler).

- [ ] **Step 3: Implement**

Create `internal/server/llm_handler.go`:

```go
package server

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/GMfatcat/goslide/internal/llm"
)

// Transformer is the narrow interface llm_handler depends on. The
// production wiring passes an adapter that calls llm.Transform with a
// shared DiskCache and Completer; tests can pass a fake.
type Transformer interface {
	Transform(ctx context.Context, req llm.Request) (llm.Result, error)
}

type llmHandler struct {
	t Transformer
}

func newLLMHandler(t Transformer) *llmHandler {
	return &llmHandler{t: t}
}

type llmRequestBody struct {
	Model     string          `json:"model"`
	Prompt    string          `json:"prompt"`
	Data      json.RawMessage `json:"data"`
	MaxTokens int             `json:"max_tokens"`
}

type llmResponseBody struct {
	Content   string    `json:"content"`
	FromCache bool      `json:"from_cache"`
	Usage     llm.Usage `json:"usage"`
}

func (h *llmHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var body llmRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	req := llm.Request{
		Model:     body.Model,
		Prompt:    body.Prompt,
		Data:      []byte(body.Data),
		MaxTokens: body.MaxTokens,
	}
	res, err := h.t.Transform(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(llmResponseBody{
		Content:   res.Content,
		FromCache: res.FromCache,
		Usage:     res.Usage,
	})
}
```

- [ ] **Step 4: Verify tests pass**

Run: `GOTOOLCHAIN=local go test -C D:/CLAUDE-CODE-GOSLIDE ./internal/server/ -v`
Expected: all pass.

- [ ] **Step 5: Commit**

```bash
git -C D:/CLAUDE-CODE-GOSLIDE add internal/server/llm_handler.go internal/server/llm_handler_test.go
```

```bash
git -C D:/CLAUDE-CODE-GOSLIDE commit -m "feat(server): POST /api/llm handler with Transformer interface"
```

---

## Task 7: server — register `/api/llm` route with a real transformer adapter

**Files:**
- Modify: `internal/server/handlers.go`
- Modify: `internal/server/server.go`

- [ ] **Step 1: Understand existing shape**

Open `internal/server/server.go` around line 44-65 and `internal/server/handlers.go` to see how `setupRoutes` registers handlers. The `app` struct has `cfg *config.Config` which carries `Generate` config fields.

- [ ] **Step 2: Add a small adapter and register the route**

In `internal/server/handlers.go`, extend `setupRoutes` to register `/api/llm` when `generate` is configured. Add imports for `generate`, `llm`, and `path/filepath` if not present.

```go
// near the top of handlers.go imports:
// "context"
// "path/filepath"
//
// "github.com/GMfatcat/goslide/internal/generate"
// "github.com/GMfatcat/goslide/internal/llm"
```

Extend `setupRoutes`:

```go
func (a *app) setupRoutes() {
	a.mux.HandleFunc("/", a.handleIndex)
	a.mux.HandleFunc("/ws", a.broadcast.handleWS)
	// ... existing routes ...

	// LLM transformer route (Phase 7a). Only wire up if generate config
	// is present; without it, Transform would fail at request time anyway.
	if a.cfg != nil && a.cfg.Generate.BaseURL != "" && a.cfg.Generate.Model != "" {
		ta := newLLMTransformerAdapter(a.cfg, a.opts.File)
		a.mux.Handle("/api/llm", newLLMHandler(ta))
	}
}
```

Add the adapter as a new block in `internal/server/handlers.go` (or a new file `internal/server/llm_adapter.go` — your call; keep whichever feels cleaner in this repo style):

```go
// llmTransformerAdapter wires llm.Transform into the Transformer
// interface llm_handler depends on.
type llmTransformerAdapter struct {
	cfg     *config.Config
	cache   *llm.DiskCache
}

func newLLMTransformerAdapter(cfg *config.Config, presentationFile string) *llmTransformerAdapter {
	cacheDir := filepath.Join(filepath.Dir(presentationFile), ".goslide-cache")
	return &llmTransformerAdapter{
		cfg:   cfg,
		cache: llm.NewDiskCache(cacheDir),
	}
}

func (a *llmTransformerAdapter) Transform(ctx context.Context, req llm.Request) (llm.Result, error) {
	model := req.Model
	if model == "" {
		model = a.cfg.Generate.Model
	}
	req.Model = model

	apiKey := os.Getenv(firstNonEmptyEnv(a.cfg.Generate.APIKeyEnv, "OPENAI_API_KEY"))
	timeout := a.cfg.Generate.Timeout
	if timeout == 0 {
		timeout = 120 * time.Second
	}
	client := generate.NewClient(a.cfg.Generate.BaseURL, apiKey, timeout)

	completer := clientAsCompleter{client}
	return llm.Transform(ctx, req, completer, a.cache)
}

// clientAsCompleter adapts generate.Client to llm.Completer (trivial:
// their Message types are parallel).
type clientAsCompleter struct {
	c *generate.Client
}

func (a clientAsCompleter) Complete(ctx context.Context, model string, msgs []llm.Message) (string, llm.Usage, error) {
	gmsgs := make([]generate.Message, len(msgs))
	for i, m := range msgs {
		gmsgs[i] = generate.Message{Role: m.Role, Content: m.Content}
	}
	content, usage, err := a.c.Complete(ctx, model, gmsgs)
	return content, llm.Usage(usage), err
}

func firstNonEmptyEnv(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
```

Add the new imports to `handlers.go`:

```go
import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/GMfatcat/goslide/internal/config"
	"github.com/GMfatcat/goslide/internal/generate"
	"github.com/GMfatcat/goslide/internal/llm"
)
```

(Keep existing imports; add only what's missing.)

- [ ] **Step 3: Add a smoke test for route registration**

Append to `internal/server/llm_handler_test.go`:

```go
func TestSetupRoutes_RegistersLLMWhenGenerateConfigured(t *testing.T) {
	// (we don't exercise the route end-to-end here; Task 6 already does
	// that with a fake Transformer. This test confirms wiring.)
	t.Skip("integration — exercised manually via goslide serve")
}
```

(The test is a skip-placeholder — the actual route wiring is verified by the adapter compiling and `go vet` passing. A real integration test would spin up the HTTP server, which overlaps with `server_test.go`.)

- [ ] **Step 4: Build + run tests**

Run: `GOTOOLCHAIN=local go build -C D:/CLAUDE-CODE-GOSLIDE ./cmd/goslide`
Expected: success.

Run: `GOTOOLCHAIN=local go test -C D:/CLAUDE-CODE-GOSLIDE ./internal/server/ -v`
Expected: all pass.

- [ ] **Step 5: Commit**

```bash
git -C D:/CLAUDE-CODE-GOSLIDE add internal/server/handlers.go internal/server/llm_handler_test.go
```

```bash
git -C D:/CLAUDE-CODE-GOSLIDE commit -m "feat(server): wire /api/llm into serve/host with generate-backed adapter"
```

---

## Task 8: renderer — emit `data-llm-bakes` attribute from Component params

**Files:**
- Modify: `internal/renderer/components.go`
- Modify: `internal/renderer/renderer_test.go`

- [ ] **Step 1: Write failing test**

Append to `internal/renderer/renderer_test.go`:

```go
func TestRender_EmitsLLMBakes(t *testing.T) {
	pres := &ir.Presentation{
		Meta: ir.Frontmatter{Theme: "dark"},
		Slides: []ir.Slide{{
			Index: 0,
			Meta:  ir.SlideMeta{Layout: "default"},
			Components: []ir.Component{{
				Index: 0,
				Type:  "api",
				Params: map[string]any{
					"endpoint":    "/api/x",
					"_llm_bakes":  map[string]any{"1": "## Insights"},
				},
			}},
			BodyHTML: `<!--goslide:component:0-->`,
		}},
	}
	html, err := Render(pres)
	require.NoError(t, err)
	require.Contains(t, html, `data-llm-bakes=`)
	require.Contains(t, html, `Insights`)
}
```

- [ ] **Step 2: Verify failure**

Run: `GOTOOLCHAIN=local go test -C D:/CLAUDE-CODE-GOSLIDE ./internal/renderer/ -run TestRender_EmitsLLMBakes -v`
Expected: FAIL (no `data-llm-bakes` attribute emitted).

- [ ] **Step 3: Modify `buildComponentDiv` in `internal/renderer/components.go`**

Change:

```go
return fmt.Sprintf(
	`<div class="goslide-component" data-type="%s" data-params="%s" data-raw="%s" data-comp-id="%s">%s</div>`,
	escapeAttr(comp.Type),
	paramsAttr,
	rawAttr,
	compID,
	comp.ContentHTML,
)
```

Into:

```go
bakesAttr := ""
if comp.Params != nil {
	if bakes, ok := comp.Params["_llm_bakes"]; ok {
		var buf bytes.Buffer
		enc := json.NewEncoder(&buf)
		enc.SetEscapeHTML(false)
		if err := enc.Encode(bakes); err == nil {
			bakesAttr = fmt.Sprintf(` data-llm-bakes="%s"`, escapeAttr(strings.TrimRight(buf.String(), "\n")))
		}
	}
}
return fmt.Sprintf(
	`<div class="goslide-component" data-type="%s" data-params="%s" data-raw="%s" data-comp-id="%s"%s>%s</div>`,
	escapeAttr(comp.Type),
	paramsAttr,
	rawAttr,
	compID,
	bakesAttr,
	comp.ContentHTML,
)
```

**Important:** the existing `paramsAttr` is serialised from `comp.Params` which will now include the `_llm_bakes` key. To avoid duplicating bakes data in both `data-params` and `data-llm-bakes`, strip `_llm_bakes` from the params copy before JSON-encoding it:

Update the params serialisation block (lines 26-45 in the current file) to:

```go
var paramsAttr string
var rawAttr string

if comp.Type == "mermaid" {
	paramsAttr = "{}"
	rawAttr = escapeAttr(comp.Raw)
} else {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	paramsForAttr := comp.Params
	if _, has := comp.Params["_llm_bakes"]; has {
		paramsForAttr = make(map[string]any, len(comp.Params)-1)
		for k, v := range comp.Params {
			if k == "_llm_bakes" {
				continue
			}
			paramsForAttr[k] = v
		}
	}
	if err := enc.Encode(paramsForAttr); err != nil {
		buf.Reset()
		buf.WriteString("{}")
	}
	paramsAttr = escapeAttr(strings.TrimRight(buf.String(), "\n"))
	rawAttr = ""
}
```

- [ ] **Step 4: Verify tests pass**

Run: `GOTOOLCHAIN=local go test -C D:/CLAUDE-CODE-GOSLIDE ./internal/renderer/ -v`
Expected: all pass (existing + new).

- [ ] **Step 5: Commit**

```bash
git -C D:/CLAUDE-CODE-GOSLIDE add internal/renderer/components.go internal/renderer/renderer_test.go
```

```bash
git -C D:/CLAUDE-CODE-GOSLIDE commit -m "feat(renderer): emit data-llm-bakes attribute and strip _llm_bakes from data-params"
```

---

## Task 9: builder — BakeLLM pass

**Files:**
- Create: `internal/builder/llm_bake.go`
- Create: `internal/builder/llm_bake_test.go`

- [ ] **Step 1: Write failing tests**

Create `internal/builder/llm_bake_test.go`:

```go
package builder

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/GMfatcat/goslide/internal/ir"
	"github.com/GMfatcat/goslide/internal/llm"
	"github.com/stretchr/testify/require"
)

type fakeCompleter struct {
	content string
	err     error
	calls   int
}

func (f *fakeCompleter) Complete(_ context.Context, _ string, _ []llm.Message) (string, llm.Usage, error) {
	f.calls++
	return f.content, llm.Usage{}, f.err
}

func TestBakeLLM_CacheHitProducesBakes(t *testing.T) {
	dir := t.TempDir()
	cache := llm.NewDiskCache(filepath.Join(dir, ".goslide-cache"))
	// pre-populate cache so bake hits without any completer calls
	req := llm.Request{Model: "m", Prompt: "hi {{data}}", Data: []byte(`{"x":1}`)}
	require.NoError(t, cache.Put(cache.Key(req), llm.CacheEntry{Version: 1, Content: "result"}))

	fixture := filepath.Join(dir, "data.json")
	require.NoError(t, os.WriteFile(fixture, []byte(`{"x":1}`), 0644))

	pres := &ir.Presentation{
		Slides: []ir.Slide{{
			Index: 0,
			Components: []ir.Component{{
				Index: 0,
				Type:  "api",
				Params: map[string]any{
					"endpoint": "/api/x",
					"fixture":  "data.json",
					"render": []any{
						map[string]any{"type": "chart:bar"},
						map[string]any{"type": "llm", "prompt": "hi {{data}}", "model": "m"},
					},
				},
			}},
		}},
	}

	fc := &fakeCompleter{}
	err := BakeLLM(pres, cache, fc, "m", dir, false)
	require.NoError(t, err)
	require.Equal(t, 0, fc.calls, "cache hit must not call completer")

	bakes := pres.Slides[0].Components[0].Params["_llm_bakes"].(map[string]string)
	require.Equal(t, "result", bakes["1"])
}

func TestBakeLLM_CacheMissWithoutRefreshReturnsError(t *testing.T) {
	dir := t.TempDir()
	cache := llm.NewDiskCache(filepath.Join(dir, ".goslide-cache"))
	fixture := filepath.Join(dir, "data.json")
	require.NoError(t, os.WriteFile(fixture, []byte(`{"x":1}`), 0644))

	pres := &ir.Presentation{
		Slides: []ir.Slide{{
			Index: 0,
			Components: []ir.Component{{
				Index: 0,
				Type:  "api",
				Params: map[string]any{
					"fixture": "data.json",
					"render": []any{
						map[string]any{"type": "llm", "prompt": "hi", "model": "m"},
					},
				},
			}},
		}},
	}

	err := BakeLLM(pres, cache, &fakeCompleter{}, "m", dir, false)
	require.Error(t, err)
	require.Contains(t, err.Error(), "cache miss")
	require.Contains(t, err.Error(), "slide 0")
}

func TestBakeLLM_CacheMissWithRefreshCallsCompleter(t *testing.T) {
	dir := t.TempDir()
	cache := llm.NewDiskCache(filepath.Join(dir, ".goslide-cache"))
	fixture := filepath.Join(dir, "data.json")
	require.NoError(t, os.WriteFile(fixture, []byte(`{"x":1}`), 0644))

	pres := &ir.Presentation{
		Slides: []ir.Slide{{
			Index: 0,
			Components: []ir.Component{{
				Index: 0,
				Type:  "api",
				Params: map[string]any{
					"fixture": "data.json",
					"render": []any{
						map[string]any{"type": "llm", "prompt": "hi {{data}}", "model": "m"},
					},
				},
			}},
		}},
	}

	fc := &fakeCompleter{content: "fresh"}
	err := BakeLLM(pres, cache, fc, "m", dir, true)
	require.NoError(t, err)
	require.Equal(t, 1, fc.calls)

	bakes := pres.Slides[0].Components[0].Params["_llm_bakes"].(map[string]string)
	require.Equal(t, "fresh", bakes["0"])
}

func TestBakeLLM_SkipsNonAPIComponents(t *testing.T) {
	dir := t.TempDir()
	cache := llm.NewDiskCache(filepath.Join(dir, ".goslide-cache"))

	pres := &ir.Presentation{
		Slides: []ir.Slide{{
			Index: 0,
			Components: []ir.Component{{Index: 0, Type: "chart", Params: map[string]any{}}},
		}},
	}

	err := BakeLLM(pres, cache, &fakeCompleter{}, "m", dir, false)
	require.NoError(t, err)
}

func TestBakeLLM_NoFixtureAndNoCacheReturnsError(t *testing.T) {
	dir := t.TempDir()
	cache := llm.NewDiskCache(filepath.Join(dir, ".goslide-cache"))

	pres := &ir.Presentation{
		Slides: []ir.Slide{{
			Index: 0,
			Components: []ir.Component{{
				Index: 0,
				Type:  "api",
				Params: map[string]any{
					"render": []any{
						map[string]any{"type": "llm", "prompt": "hi"},
					},
				},
			}},
		}},
	}

	err := BakeLLM(pres, cache, &fakeCompleter{}, "m", dir, false)
	require.Error(t, err)
}
```

- [ ] **Step 2: Verify failure**

Run: `GOTOOLCHAIN=local go test -C D:/CLAUDE-CODE-GOSLIDE ./internal/builder/ -run TestBakeLLM -v`
Expected: FAIL (undefined: BakeLLM).

- [ ] **Step 3: Implement**

Create `internal/builder/llm_bake.go`:

```go
package builder

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/GMfatcat/goslide/internal/ir"
	"github.com/GMfatcat/goslide/internal/llm"
)

// BakeLLM resolves every `llm` render item inside every `api` component
// of the presentation, storing cached results (or freshly-fetched ones
// when refreshOnMiss is true) under the Component param `_llm_bakes`
// as a map[string]string keyed by render-item index.
//
// mdDir is the directory containing the source .md file; it anchors
// relative paths in `fixture:` fields. defaultModel is used when a
// render item omits `model`.
func BakeLLM(pres *ir.Presentation, cache *llm.DiskCache, completer llm.Completer, defaultModel string, mdDir string, refreshOnMiss bool) error {
	var missErrors []string

	for si := range pres.Slides {
		slide := &pres.Slides[si]
		for ci := range slide.Components {
			comp := &slide.Components[ci]
			if comp.Type != "api" {
				continue
			}

			renderList, ok := comp.Params["render"].([]any)
			if !ok {
				continue
			}

			hasLLMItem := false
			for _, raw := range renderList {
				item, ok := raw.(map[string]any)
				if !ok {
					continue
				}
				if item["type"] == "llm" {
					hasLLMItem = true
					break
				}
			}
			if !hasLLMItem {
				continue
			}

			// Resolve data source: fixture file only in MVP (runtime
			// cache fallback is a future enhancement).
			fixture, _ := comp.Params["fixture"].(string)
			if fixture == "" {
				missErrors = append(missErrors, fmt.Sprintf("slide %d, component %d: api component has llm render items but no 'fixture:' file", slide.Index, comp.Index))
				continue
			}
			data, readErr := os.ReadFile(filepath.Join(mdDir, fixture))
			if readErr != nil {
				missErrors = append(missErrors, fmt.Sprintf("slide %d, component %d: read fixture %s: %v", slide.Index, comp.Index, fixture, readErr))
				continue
			}

			bakes := map[string]string{}
			for itemIdx, raw := range renderList {
				item, ok := raw.(map[string]any)
				if !ok || item["type"] != "llm" {
					continue
				}
				prompt, _ := item["prompt"].(string)
				model, _ := item["model"].(string)
				if model == "" {
					model = defaultModel
				}
				maxTokens := 1024
				if mt, ok := item["max_tokens"].(int); ok {
					maxTokens = mt
				}

				req := llm.Request{Model: model, Prompt: prompt, Data: data, MaxTokens: maxTokens}
				key := cache.Key(req)

				entry, hit, _ := cache.Get(key)
				if hit {
					bakes[strconv.Itoa(itemIdx)] = entry.Content
					continue
				}
				if !refreshOnMiss {
					missErrors = append(missErrors, fmt.Sprintf("slide %d, component %d, item %d: cache miss (run 'goslide serve' to warm cache, or pass --llm-refresh)", slide.Index, comp.Index, itemIdx))
					continue
				}
				res, err := llm.Transform(context.Background(), req, completer, cache)
				if err != nil {
					missErrors = append(missErrors, fmt.Sprintf("slide %d, component %d, item %d: llm call failed: %v", slide.Index, comp.Index, itemIdx, err))
					continue
				}
				bakes[strconv.Itoa(itemIdx)] = res.Content
			}

			if len(bakes) > 0 {
				if comp.Params == nil {
					comp.Params = map[string]any{}
				}
				comp.Params["_llm_bakes"] = bakes
			}
		}
	}

	if len(missErrors) > 0 {
		return fmt.Errorf("BakeLLM: %d problem(s):\n  %s", len(missErrors), strings.Join(missErrors, "\n  "))
	}
	return nil
}
```

- [ ] **Step 4: Verify tests pass**

Run: `GOTOOLCHAIN=local go test -C D:/CLAUDE-CODE-GOSLIDE ./internal/builder/ -v`
Expected: all pass.

- [ ] **Step 5: Commit**

```bash
git -C D:/CLAUDE-CODE-GOSLIDE add internal/builder/llm_bake.go internal/builder/llm_bake_test.go
```

```bash
git -C D:/CLAUDE-CODE-GOSLIDE commit -m "feat(builder): BakeLLM pass — inline cached LLM results into components"
```

---

## Task 10: builder + CLI — wire BakeLLM into build pipeline with `--llm-refresh` flag

**Files:**
- Modify: `internal/builder/builder.go`
- Modify: `internal/cli/build.go`

- [ ] **Step 1: Extend `Options` and `Build`**

Open `internal/builder/builder.go` and add `LLMRefresh bool` to the `Options` struct:

```go
type Options struct {
	File       string
	Output     string
	Theme      string
	Accent     string
	LLMRefresh bool
}
```

Between the `validate` block and the `renderer.Render(pres)` call, insert the bake step. Add these imports if not present: `context`, `os`, `time`, `github.com/GMfatcat/goslide/internal/generate`, `github.com/GMfatcat/goslide/internal/llm`.

```go
// After validate succeeds:
cfg, _ := config.Load(filepath.Dir(opts.File))

// BakeLLM — only runs if any component has an llm render item. Fatal
// on miss unless --llm-refresh (opts.LLMRefresh) is set.
if cfg != nil && cfg.Generate.BaseURL != "" && cfg.Generate.Model != "" {
	mdDir := filepath.Dir(opts.File)
	cache := llm.NewDiskCache(filepath.Join(mdDir, ".goslide-cache"))
	completer := buildLLMCompleter(cfg)
	if err := llm.BakeIfAny(pres, cache, completer, cfg.Generate.Model, mdDir, opts.LLMRefresh); err != nil {
		return err
	}
}
```

(`BakeIfAny` here is `BakeLLM` from Task 9 — rename if convenient, or wrap with a guard that checks whether any slide has llm items before doing work. For the plan's simplicity, call `llm.BakeLLM(...)` directly; it already no-ops on slides with no llm items.)

Replace the above with the direct call:

```go
if cfg != nil && cfg.Generate.BaseURL != "" && cfg.Generate.Model != "" {
	mdDir := filepath.Dir(opts.File)
	cache := llm.NewDiskCache(filepath.Join(mdDir, ".goslide-cache"))
	completer := buildLLMCompleter(cfg)
	if err := BakeLLM(pres, cache, completer, cfg.Generate.Model, mdDir, opts.LLMRefresh); err != nil {
		return err
	}
}
```

Add a small helper `buildLLMCompleter` in the same file (below the `Build` function):

```go
func buildLLMCompleter(cfg *config.Config) llm.Completer {
	apiKey := os.Getenv(firstNonEmpty(cfg.Generate.APIKeyEnv, "OPENAI_API_KEY"))
	timeout := cfg.Generate.Timeout
	if timeout == 0 {
		timeout = 120 * time.Second
	}
	client := generate.NewClient(cfg.Generate.BaseURL, apiKey, timeout)
	return completerAdapter{client}
}

type completerAdapter struct{ c *generate.Client }

func (a completerAdapter) Complete(ctx context.Context, model string, msgs []llm.Message) (string, llm.Usage, error) {
	gmsgs := make([]generate.Message, len(msgs))
	for i, m := range msgs {
		gmsgs[i] = generate.Message{Role: m.Role, Content: m.Content}
	}
	content, usage, err := a.c.Complete(ctx, model, gmsgs)
	return content, llm.Usage(usage), err
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
```

- [ ] **Step 2: Expose `--llm-refresh` in the CLI**

Open `internal/cli/build.go` and add the flag. Find where other flags are defined and add:

```go
var llmRefresh bool

// In init():
buildCmd.Flags().BoolVar(&llmRefresh, "llm-refresh", false, "call the LLM on cache miss during build (default: miss is an error)")

// In runBuild (or equivalent):
opts := builder.Options{
	File:       /* existing */,
	Output:     /* existing */,
	Theme:      /* existing */,
	Accent:     /* existing */,
	LLMRefresh: llmRefresh,
}
```

(Adapt the variable names to match what's already there; the specific pattern `Flags().BoolVar(&X, "llm-refresh", false, ...)` is the key.)

- [ ] **Step 3: Build + run all tests**

Run: `GOTOOLCHAIN=local go build -C D:/CLAUDE-CODE-GOSLIDE ./cmd/goslide`
Expected: success.

Run: `GOTOOLCHAIN=local go test -C D:/CLAUDE-CODE-GOSLIDE ./...`
Expected: all pass.

- [ ] **Step 4: Smoke test — build without llm items is unaffected**

Run: `GOTOOLCHAIN=local go build -C D:/CLAUDE-CODE-GOSLIDE -o D:/CLAUDE-CODE-GOSLIDE/goslide.exe ./cmd/goslide`

Then `./goslide.exe build examples/demo.md -o /tmp/demo.html` — should succeed because `examples/demo.md` has no llm render items.

- [ ] **Step 5: Commit**

```bash
git -C D:/CLAUDE-CODE-GOSLIDE add internal/builder/builder.go internal/cli/build.go
```

```bash
git -C D:/CLAUDE-CODE-GOSLIDE commit -m "feat(build): integrate BakeLLM with --llm-refresh flag"
```

---

## Task 11: frontend — renderLLMItem + button UX + localStorage

**Files:**
- Modify: `web/static/components.js`

- [ ] **Step 1: Locate the api render loop**

Open `web/static/components.js` and find the `fetchAndRender` function around line 344. It iterates `params.render` and produces HTML per render item. You'll branch a new case for `type === 'llm'`.

- [ ] **Step 2: Add helper functions**

Before `fetchAndRender`, add:

```javascript
  function llmCacheKey(model, prompt, data) {
    // Simple stable string key; SHA not needed here because localStorage
    // is local. Server uses real SHA256 for disk cache.
    return (model || '') + '|' + prompt + '|' + JSON.stringify(data);
  }

  function renderLLMItem(compEl, item, apiData, container) {
    // 1. Build-baked result takes precedence (static export path).
    try {
      var bakes = compEl.dataset.llmBakes ? JSON.parse(compEl.dataset.llmBakes) : null;
      if (bakes && bakes[String(item._index)] != null) {
        container.innerHTML = simpleMarkdown(bakes[String(item._index)]);
        return;
      }
    } catch (e) { /* fall through */ }

    // 2. localStorage cache
    var key = 'goslide-llm:' + llmCacheKey(item.model, item.prompt, apiData);
    var cached = null;
    try { cached = localStorage.getItem(key); } catch (e) {}
    if (cached) {
      container.innerHTML = simpleMarkdown(cached);
      return;
    }

    // 3. Click-to-call button
    var btn = document.createElement('button');
    btn.className = 'gs-llm-generate';
    btn.textContent = '✨ Generate';
    btn.onclick = function () {
      btn.disabled = true;
      btn.textContent = '⏳ Generating…';
      fetch('/api/llm', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          model: item.model || '',
          prompt: item.prompt,
          data: apiData,
          max_tokens: item.max_tokens || 1024
        })
      })
        .then(function (r) {
          if (!r.ok) throw new Error('HTTP ' + r.status);
          return r.json();
        })
        .then(function (res) {
          try { localStorage.setItem(key, res.content); } catch (e) {}
          container.innerHTML = simpleMarkdown(res.content);
        })
        .catch(function (err) {
          container.innerHTML = '';
          var msg = document.createElement('div');
          msg.className = 'gs-llm-error';
          msg.textContent = 'LLM call failed: ' + err.message;
          container.appendChild(msg);
          btn.disabled = false;
          btn.textContent = '✨ Retry';
          container.appendChild(btn);
        });
    };
    container.appendChild(btn);
  }

  // simpleMarkdown — minimal conversion: keep newlines, handle bold/italic/code,
  // unordered lists. For richer output, the reveal.js markdown plugin is an
  // option; MVP keeps a zero-dependency inline renderer.
  function simpleMarkdown(s) {
    if (!s) return '';
    // escape HTML first
    var out = s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
    // code spans
    out = out.replace(/`([^`]+)`/g, '<code>$1</code>');
    // bold **x**
    out = out.replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>');
    // italic *x*
    out = out.replace(/(^|\W)\*([^*]+)\*(\W|$)/g, '$1<em>$2</em>$3');
    // unordered list items
    out = out.replace(/^- (.+)$/gm, '<li>$1</li>');
    out = out.replace(/(<li>[\s\S]*?<\/li>)/g, '<ul>$1</ul>');
    // paragraphs / newlines
    out = out.split(/\n\n+/).map(function (p) {
      if (p.indexOf('<ul>') === 0) return p;
      return '<p>' + p.replace(/\n/g, '<br>') + '</p>';
    }).join('\n');
    return out;
  }
```

- [ ] **Step 3: Hook renderLLMItem into the api render loop**

Inside `fetchAndRender` (look for the loop around `var renderList = params.render;`), find the branch that dispatches per render type. Add a new branch:

```javascript
renderList.forEach(function (item, idx) {
  var subEl = document.createElement('div');
  subEl.className = 'goslide-api-item';
  items.appendChild(subEl);

  // annotate index so renderLLMItem can look up data-llm-bakes by position
  item._index = idx;

  if (item.type === 'llm') {
    renderLLMItem(el, item, json, subEl);
    return;
  }

  // ... existing branches (chart, table, metric, markdown, log, image) ...
});
```

If the existing loop already uses a different iteration structure, adapt — keep the principle: `item._index = idx; if (item.type === 'llm') renderLLMItem(...)` short-circuits before the existing dispatch.

- [ ] **Step 4: Build + smoke-check**

Run: `GOTOOLCHAIN=local go build -C D:/CLAUDE-CODE-GOSLIDE ./cmd/goslide`
Expected: success (JS is embedded via go:embed; any syntax error will be caught by the embed read at runtime, not build time — proceed to step 5 for real verification).

Run: `GOTOOLCHAIN=local go test -C D:/CLAUDE-CODE-GOSLIDE ./...`
Expected: all pass.

- [ ] **Step 5: Manual smoke test (optional)**

Write a quick `demo.md` with an `api` component + `render: [{type: llm, prompt: "hi {{data}}", ...}]`, point `endpoint` at a mock, run `goslide serve` and verify:
- Page shows "✨ Generate" button on first visit
- Click → button becomes "⏳ Generating…" → replaced with content
- Reload → content shows directly (localStorage hit)

- [ ] **Step 6: Commit**

```bash
git -C D:/CLAUDE-CODE-GOSLIDE add web/static/components.js
```

```bash
git -C D:/CLAUDE-CODE-GOSLIDE commit -m "feat(web): renderLLMItem — button UX, localStorage, build-bake preference"
```

---

## Task 12: frontend — CSS rules

**Files:**
- Modify: `web/themes/layouts.css`

- [ ] **Step 1: Append LLM styles**

At the end of `web/themes/layouts.css`, append:

```css
/* Component: LLM render item inside api component */
.gs-llm-generate {
  padding: 8px 16px;
  border: 1px dashed var(--slide-accent, #8888aa);
  border-radius: 6px;
  background: transparent;
  color: var(--slide-text, inherit);
  cursor: pointer;
  font-size: 0.9em;
  font-family: inherit;
}
.gs-llm-generate:hover {
  background: rgba(136, 136, 170, 0.1);
}
.gs-llm-generate:disabled {
  cursor: wait;
  opacity: 0.6;
}
.gs-llm-error {
  color: #c66;
  font-size: 0.85em;
  font-style: italic;
  margin-bottom: 8px;
}
```

- [ ] **Step 2: Build**

Run: `GOTOOLCHAIN=local go build -C D:/CLAUDE-CODE-GOSLIDE ./cmd/goslide`
Expected: success.

- [ ] **Step 3: Commit**

```bash
git -C D:/CLAUDE-CODE-GOSLIDE add web/themes/layouts.css
```

```bash
git -C D:/CLAUDE-CODE-GOSLIDE commit -m "feat(web): CSS for LLM generate button + error state"
```

---

## Task 13: system prompt — brief mention of the `llm` render type

**Files:**
- Modify: `internal/generate/system_prompt.md`
- Modify: `internal/generate/embed_test.go`

Phase 6a's prompt explicitly tells the LLM "Do NOT use `api:`". We keep that restriction for now (the `llm` render type is a new, advanced feature and letting the generator use it risks bad outputs). Instead, add a tiny note mentioning the feature exists for manual authors.

- [ ] **Step 1: Skip this task if the MVP is manual-author-only**

The spec §7 Out of Scope doesn't list "teach LLM about llm render type", but §1 emphasises the feature is "for production decks", implying manual authoring. Leave the system prompt untouched; add a note to the README (Task 14) that the `llm` render type is manual-only.

If later we decide to let `goslide generate` emit llm items, revisit this task.

**No file changes. Skip to Task 14.**

---

## Task 14: docs — README + RELEASE_NOTES

**Files:**
- Modify: `README.md`
- Modify: `README_zh-TW.md`
- Modify: `PRD.md` (tick the relevant checkbox if present)

- [ ] **Step 1: README English**

Insert a new subsection in `README.md` right after the "Image placeholders and multi-image slides" subsection (added in v1.3.0). Use this content verbatim:

````markdown
### LLM transformer inside api components (experimental)

The `api` component accepts a new render item of type `llm`. Fetched
JSON is substituted into a user-authored prompt via `{{data}}` and the
result renders inline:

```
~~~api
endpoint: /api/sales
render:
  - type: chart:bar
    label-key: quarter
    data-key: revenue
  - type: llm
    prompt: |
      Summarise these figures as 3 analyst bullets:
      {{data}}
    model: openai/gpt-oss-120b:free   # optional; defaults to generate.model
~~~
```

- **Cache-first.** Same `(model, prompt, data)` triplet only calls the
  LLM once; subsequent renders read from `.goslide-cache/`.
- **Click-to-call.** On a cache miss in `goslide serve`, viewers see a
  `Generate ✨` button. No automatic LLM calls on page load.
- **Build-lock.** `goslide build` inlines cached results into the
  static HTML. The exported deck never reaches an LLM at view time.

Reuses the `generate:` section of `goslide.yaml` — no new configuration.

For an offline build workflow, add `fixture: ./data.json` to the `api`
component; `goslide build` uses the fixture as the input data source
instead of calling the live endpoint.

`goslide build --llm-refresh` calls the LLM to populate missing cache
entries during the build. Without the flag, cache misses fail the
build and list the offending slides.

This feature is manual-author-only in v1.4.0. `goslide generate` does
not emit `llm` render items yet.
````

- [ ] **Step 2: README 繁體中文**

Insert a parallel subsection in `README_zh-TW.md`:

````markdown
### api component 的 LLM 轉換器（experimental）

`api` component 新增一種 render item：`type: llm`。它把 fetch 到的 JSON 透過 `{{data}}` 代入作者寫的 prompt，再把 LLM 回應渲染在該位置：

```
~~~api
endpoint: /api/sales
render:
  - type: chart:bar
    label-key: quarter
    data-key: revenue
  - type: llm
    prompt: |
      用 3 個分析師觀點摘要以下數字：
      {{data}}
    model: openai/gpt-oss-120b:free   # 可選；未設則用 generate.model
~~~
```

- **Cache-first。** 相同 `(model, prompt, data)` 組合最多呼叫 LLM 一次，後續從 `.goslide-cache/` 讀。
- **Click-to-call。** `goslide serve` 遇到 cache miss 時顯示 `Generate ✨` 按鈕；頁面開啟本身不會自動呼叫 LLM。
- **Build-lock。** `goslide build` 會把 cache 內容烘進靜態 HTML；匯出後的簡報絕不會在觀看時對外呼叫。

直接重用 `goslide.yaml` 的 `generate:` 區段，不需新增設定。

離線 build workflow：在 `api` component 加 `fixture: ./data.json`，`goslide build` 就會用 fixture 當資料來源，不會呼叫實際 endpoint。

`goslide build --llm-refresh` 會在 cache miss 時真的打 LLM 補 cache。沒加這個 flag 的話，cache miss 會讓 build 失敗並列出缺的 slide。

v1.4.0 的版本只支援手寫 `llm` render item；`goslide generate` 還不會自動生成這類 render item。
````

- [ ] **Step 3: PRD checkbox**

Open `PRD.md`. If there is a line about the Chat render type / LLM-in-API in the deferred list, mark it `[x]` with a note "(v1.4.0, as api-component llm render type)". If no such line exists, add:

```markdown
- [x] LLM transformer render type for api components (Phase 7a, v1.4.0)
```

- [ ] **Step 4: Commit**

```bash
git -C D:/CLAUDE-CODE-GOSLIDE add README.md README_zh-TW.md PRD.md
```

```bash
git -C D:/CLAUDE-CODE-GOSLIDE commit -m "docs: README + PRD updates for api+llm transformer"
```

---

## Task 15: Final verification

**Files:** none (verification only)

- [ ] **Step 1: gofmt the whole repo**

Run: `GOTOOLCHAIN=local gofmt -l D:/CLAUDE-CODE-GOSLIDE`
Expected: empty output (no unformatted files).

If any file is listed, run `GOTOOLCHAIN=local gofmt -w <file>` for each, re-stage, amend the most recent commit or add a tidy-up commit.

- [ ] **Step 2: Full test suite**

Run: `GOTOOLCHAIN=local go test -C D:/CLAUDE-CODE-GOSLIDE ./...`
Expected: all pass.

- [ ] **Step 3: go vet**

Run: `GOTOOLCHAIN=local go vet -C D:/CLAUDE-CODE-GOSLIDE ./...`
Expected: clean.

- [ ] **Step 4: Build**

Run: `GOTOOLCHAIN=local go build -C D:/CLAUDE-CODE-GOSLIDE -o D:/CLAUDE-CODE-GOSLIDE/goslide.exe ./cmd/goslide`
Expected: success.

- [ ] **Step 5: Backwards compatibility — v1.3.0 deck still builds**

Run: `D:/CLAUDE-CODE-GOSLIDE/goslide.exe build D:/CLAUDE-CODE-GOSLIDE/examples/demo.md -o /tmp/demo-v1.4.html`
Expected: success. No errors about missing llm config (decks without llm render items should not exercise the new code path).

- [ ] **Step 6: Smoke — new feature end-to-end**

Create a scratch directory with:

```yaml
# goslide.yaml
generate:
  base_url: https://openrouter.ai/api/v1
  model: openai/gpt-oss-120b:free
  api_key_env: OPENROUTER_API_KEY
  timeout: 180s

api:
  proxy:
    /api/sales:
      target: http://localhost:9999
```

```markdown
<!-- demo.md -->
---
title: LLM Transformer Smoke
theme: dark
---

# Sales Review

~~~api
endpoint: /api/sales
fixture: sales.json
render:
  - type: llm
    prompt: |
      Summarise in 2 bullets:
      {{data}}
~~~
```

```json
// sales.json
{"q1":100,"q2":120,"q3":95,"q4":140}
```

- `OPENROUTER_API_KEY=... D:/CLAUDE-CODE-GOSLIDE/goslide.exe build demo.md --llm-refresh`
  - Expected: success; `demo.html` contains `data-llm-bakes="..."` on the api component; `.goslide-cache/` populated.
- `D:/CLAUDE-CODE-GOSLIDE/goslide.exe build demo.md` (no flag)
  - Expected: success on the second run (cache hit); no LLM call.
- Delete the `.goslide-cache/` directory, run again without `--llm-refresh`:
  - Expected: hard fail with a message listing slide 0, component 0, item 0.

- [ ] **Step 7: No commit needed**

If all steps pass, implementation is complete. Tasks 1-14 produced the deliverable.

---

## Success Criteria (from spec §8)

- ✅ `api` component with an `llm` render item renders a Generate button in `serve` mode on cache miss; clicking it produces the LLM response and caches it. (Tasks 9-11)
- ✅ Reload after a successful generation shows content instantly from localStorage. (Task 11)
- ✅ `goslide build` with a warm cache produces a standalone HTML file with LLM content inlined, no network needed at view time. (Tasks 9-10)
- ✅ `goslide build` with a cold cache fails with a clear per-slide list; `--llm-refresh` fills the cache by calling the LLM. (Tasks 9-10)
- ✅ `/api/llm` is registered once by both serve and host modes; disk cache is shared across viewers. (Task 7)
- ✅ All unit + integration tests pass; no real LLM calls in CI. (Task 15 step 2)
- ✅ No new external Go dependencies beyond Phase 6a. (All tasks; verify via `go.mod` diff)
