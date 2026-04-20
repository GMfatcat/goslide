# API + LLM Transformer Design

**Date:** 2026-04-20
**Status:** Design approved, ready for implementation plan
**Scope:** Phase 7a — LLM post-processing stage for the `api` component

---

## 1. Overview

The existing `api` component fetches JSON from a proxied endpoint and
renders it through a `render:` pipeline of items (`chart`, `table`,
`metric`, etc.). This feature adds a new render type `llm` that takes
the fetched data, passes it through a user-authored prompt, and displays
the LLM response inline in the slide.

This extends the AI story into production decks: instead of standalone
"chat with a bot", LLM output is a stage in an API pipeline — a place
for analyst commentary, summary bullets, or natural-language
interpretation of fetched data.

**Design principles:**

- **Cache-first** — same `(model, prompt, data)` triplet calls the LLM
  at most once; subsequent renders read cached content.
- **Click-to-call** — on cache miss in `serve` mode, the viewer sees a
  `Generate ✨` button; no auto-calling.
- **Build-lock** — `goslide build` bakes cached LLM results into the
  static HTML; the exported deck never calls an LLM at view time.
- **Reuse Phase 6a infra** — the OpenAI-compatible `generate.Client` is
  the HTTP transport; `goslide.yaml generate:` is the config source.

---

## 2. Author-facing syntax

### 2.1 New render type `llm`

```
~~~api
endpoint: /api/sales
refresh: 30s
render:
  - type: chart:bar
    label-key: quarter
    data-key: revenue
  - type: llm
    prompt: |
      Summarise these sales figures as a 3-bullet analyst commentary.
      Raw data:
      {{data}}
    model: openai/gpt-oss-120b:free   # optional; falls back to generate.model
    max_tokens: 1024                   # optional; default 1024
~~~
```

| Field | Required | Default | Notes |
|-------|----------|---------|-------|
| `type: llm` | yes | — | render type identifier |
| `prompt` | yes | — | template; the literal `{{data}}` is replaced with the API response JSON |
| `model` | no | `generate.model` | per-item model override |
| `max_tokens` | no | `1024` | cap on completion length |

### 2.2 Optional `fixture:` on the `api` component

Used during `goslide build` when no live API is reachable:

```
~~~api
endpoint: /api/sales
fixture: ./fixtures/sales.json
render:
  - type: chart:bar
  - type: llm
    prompt: ...
~~~
```

At build time, `fixture` file contents are the data input for every
`llm` render item on the component. At runtime (serve / host) the
fixture is ignored; live API is always used.

### 2.3 `goslide.yaml` — no new keys

The existing `generate:` section (added in Phase 6a) supplies
`base_url`, `model`, `api_key_env`, and `timeout`. No config changes.

### 2.4 CLI

- `goslide serve` / `goslide host` — unchanged; a new HTTP handler
  `POST /api/llm` is registered automatically.
- `goslide build` — new flag `--llm-refresh` (default `false`):
  - `false` → cache miss during build is a hard error; listing the
    affected slides and suggesting `goslide serve` to warm the cache,
    or the flag.
  - `true` → cache miss during build calls the LLM, writes the cache
    entry, and bakes the result.

---

## 3. Architecture

### 3.1 File layout

```
internal/
├── llm/                       # NEW package
│   ├── types.go               # Request, Result, CacheEntry
│   ├── transform.go           # Transform(ctx, req, cfg) (Result, error)
│   ├── cache.go               # DiskCache{Get, Put, Key}
│   ├── prompt.go              # Render(template, data) — {{data}} substitution
│   └── *_test.go
├── server/
│   ├── llm_handler.go         # POST /api/llm → llm.Transform
│   └── llm_handler_test.go
├── builder/
│   ├── llm_bake.go            # BakeLLM(pres, cfg, refresh) — bake-time
│   └── llm_bake_test.go
├── parser/
│   └── component.go           # allow type=llm as a render item
├── ir/
│   └── validate.go            # validate llm items require 'prompt'
├── generate/
│   └── client.go              # UNCHANGED; used as transport by llm.Transform
└── web/static/components.js   # new renderLLMItem + button UX
```

### 3.2 Package responsibilities

| Unit | Responsibility | Public surface |
|------|---------------|----------------|
| `llm/types.go` | Struct definitions | `Request{Model, Prompt, Data, MaxTokens}`, `Result{Content, Usage, FromCache}`, `CacheEntry{...}` |
| `llm/transform.go` | Orchestrator: key → cache → fallback Completer → cache put | `Transform(ctx, req, cfg) (Result, error)`; takes `Cache` and `Completer` interfaces for mocking |
| `llm/cache.go` | `.goslide-cache/<sha256>.json` read/write | `DiskCache{dir}`, `Get(key)`, `Put(key, entry)`, `Key(req)` |
| `llm/prompt.go` | `{{data}}` substitution | `Render(template string, data []byte) string` |
| `server/llm_handler.go` | HTTP route `POST /api/llm`; decode, call Transform, encode | `func(w, r)` registered on router |
| `builder/llm_bake.go` | Walk slides, resolve each `llm` render item (cache hit / refresh miss / fail), inject result into component's `data-llm-bakes` attribute | `BakeLLM(pres *ir.Presentation, cfg, refresh bool) error` |

### 3.3 Dependency direction

```
server  ──┐
          ├──► internal/llm.Transform ──► internal/generate.Client
builder ──┘         │
                    └──► internal/llm.DiskCache
```

No cycles. `internal/llm` depends on `internal/generate` only;
everything else depends on `internal/llm`.

---

## 4. Data flow

### 4.1 Runtime (`serve` / `host`)

```
Viewer opens slide
  ↓
components.js initApiComponent(el) → fetchAndRender
  ↓
API fetch through existing proxy → apiData (JSON)
  ↓
render loop encounters item type=llm:
  ├─ compute client-side key = sha256(model + prompt + apiData)
  ├─ check localStorage['goslide-llm:' + key]
  │    hit → renderMarkdown, done
  │    miss → show "✨ Generate" button
  │            on click:
  │              POST /api/llm {model, prompt, data, max_tokens}
  │              ↓
  │              llm_handler.go:
  │                ├─ same server-side key
  │                ├─ DiskCache.Get(key):
  │                │    hit → respond {content, from_cache: true, usage}
  │                │    miss:
  │                │      llm.Transform:
  │                │        prompt.Render(template, data)
  │                │        generate.Client.Complete(ctx, messages)
  │                │        DiskCache.Put(key, result)
  │                │      respond {content, from_cache: false, usage}
  │              ↓
  │              browser: localStorage set, renderMarkdown
```

- **Two layers of cache**: `localStorage` prevents repeated server hits
  from the same browser; server `DiskCache` is the source of truth and
  shared across viewers.

### 4.2 Build path (`goslide build`)

```
parser.Parse → ir.Presentation → validate
  ↓
builder.Build orchestrator:
  BakeLLM(pres, cfg, refreshOnMiss):
    for each slide:
      for each api component:
        resolve data source:
          if fixture field set → read ./fixtures/<file>.json
          elif runtime cache has an entry for the endpoint → use cached response
          else → skip this component's llm items; bake error if any
        for each llm render item:
          key = sha256(model + prompt + data)
          entry := DiskCache.Get(key)
          if entry == nil:
            if !refreshOnMiss → accumulate error
            else: llm.Transform → DiskCache.Put
          bakes[item_index] = entry.Content
        inject bakes into component as data-llm-bakes attribute
  if any error accumulated → fail with aggregate listing
  ↓
renderer.Render → HTML (LLM results inlined)
```

### 4.3 Cache key

```
key = sha256(
  model + "\x00" +
  prompt_template + "\x00" +
  canonical_json(data)
)
```

`canonical_json` is `json.Marshal` of an `interface{}` that was
`json.Unmarshal`'d from the raw response (so whitespace and key ordering
don't affect the key; same JSON value → same key).

Null byte separator prevents collisions between fields.

### 4.4 Cache entry format

`.goslide-cache/<hex-key>.json`:

```json
{
  "version": 1,
  "created": "2026-04-20T14:00:00Z",
  "model": "openai/gpt-oss-120b:free",
  "prompt": "Summarise ...",
  "data_hash": "sha256:...",
  "content": "## Key insights\n- ...",
  "usage": { "prompt_tokens": 120, "completion_tokens": 240, "total_tokens": 360 }
}
```

Human-readable, committable. The `prompt` and `data_hash` fields are
informational only (the filename hash is the actual key); they help
debugging without having to pipe through openssl.

### 4.5 Frontend integration

**Cache-hit path** (after localStorage hit OR after successful POST):
```javascript
container.innerHTML = renderMarkdown(content);
```

**Build-bake path** (static HTML): `BakeLLM` writes
`data-llm-bakes='{"2": "...", "4": "..."}'` on the component div
where keys are render-item indices. `components.js` checks
`el.dataset.llmBakes` first; if present, use directly and skip the
cache/button dance.

**CSS** (appended to `web/themes/layouts.css`):

```css
.gs-llm-generate {
  padding: 8px 16px;
  border: 1px dashed var(--slide-accent, #8888aa);
  border-radius: 6px;
  background: transparent;
  color: var(--slide-text);
  cursor: pointer;
  font-size: 0.9em;
}
.gs-llm-generate:hover { background: rgba(136,136,170,0.1); }
.gs-llm-generate:disabled { cursor: wait; opacity: 0.6; }
.gs-llm-error {
  color: #c66;
  font-size: 0.85em;
  font-style: italic;
}
```

---

## 5. Error handling

| Condition | Behaviour | Exit / status |
|-----------|-----------|---------------|
| `generate:` section missing in `goslide.yaml` | startup warning; runtime `/api/llm` returns 500 with clear message | HTTP 500 |
| API-key env var unset | same pattern | HTTP 500 |
| LLM timeout | `/api/llm` returns 504 + retry hint | HTTP 504 |
| LLM HTTP 4xx/5xx | pass upstream status; body contains OpenAI-shape `error.message` | HTTP 4xx/5xx |
| LLM returns empty content | HTTP 502 `llm returned empty content` | 502 |
| Cache file corrupt (bad JSON) | log warn; treat as miss | — |
| Cache write fails (disk full) | log error; return result anyway | — |
| `goslide build` cache miss with `--llm-refresh=false` | accumulate per-slide, exit non-zero with full list | exit 1 |
| IR: `llm` render item missing `prompt` | validator error, code `llm-missing-prompt` | validation error |
| Browser: network failure on `/api/llm` | `<div class="gs-llm-error">` with message; button re-enabled for retry | — |

---

## 6. Testing

### 6.1 Unit

- `llm/prompt_test.go` — `{{data}}` substitution: simple string, nested
  JSON, empty data, template without placeholder, multiple occurrences.
- `llm/cache_test.go` — `Key` stability; `Put`/`Get` round-trip; corrupt
  file → miss; non-existent `.goslide-cache/` auto-created; different
  keys don't interfere.
- `llm/transform_test.go` — with fake `Completer`: cache hit path (no
  Completer call, `FromCache: true`); cache miss path (Completer called,
  cache written, `FromCache: false`); Completer error → Transform error,
  no cache write; empty content → error.

### 6.2 Server

- `server/llm_handler_test.go` — `httptest` + fake Transform:
  - valid POST → 200 + correct JSON
  - malformed JSON body → 400
  - Transform error → 502
  - `from_cache: true` flag propagates

### 6.3 Builder

- `builder/llm_bake_test.go`:
  - cache hit → HTML has `data-llm-bakes`
  - cache miss + `refreshOnMiss=false` → error with slide/item indices
  - cache miss + `refreshOnMiss=true` → Transform called, cache written,
    HTML baked
  - `fixture: ./x.json` mechanism: file read and used as data input
  - neither fixture nor runtime cache → error for that component

### 6.4 IR validation

- `llm` render item without `prompt` → error `llm-missing-prompt`
- Valid llm item → no errors

### 6.5 Parser

- `~~~api` with `render: [{type: llm, ...}]` → component.Params contains
  the llm item with prompt/model fields

### 6.6 Excluded from CI

- No real LLM calls; fake `Completer` in all tests.
- No real browsers; JS side is hand-verified post-merge (§6.7).

### 6.7 Manual verification (post-merge)

- `goslide serve` a sample deck with one api component + llm render
  item; click button; verify LLM content appears, reload keeps it
  (localStorage hit), server restart keeps it (disk hit).
- `goslide build --llm-refresh` on the same deck; verify output HTML
  renders LLM content in a browser with no network.

---

## 7. Out of Scope

- Streaming responses (SSE). The MVP buffers the full response.
  Streaming fits the future Feature A (scripted reveal) and will share
  the transport layer then.
- `{{field.x}}` JSONPath expressions in prompts. The MVP injects the
  whole response as `{{data}}`. If token cost becomes a real concern
  for users, revisit.
- Per-session or per-user budget caps. Cache + click-to-call + build-
  lock are enough gatekeeping for the MVP.
- Audit log / dashboard of LLM calls.
- Preset/library of reusable prompts.
- Retry with backoff on LLM 5xx. Users can click Generate again.

---

## 8. Success Criteria

- Author writes an `api` component with an `llm` render item; `goslide
  serve` shows a Generate button; clicking it streams (buffered) the
  LLM response and caches it; reload doesn't re-call.
- `goslide build` on a deck with warm cache produces a static HTML that
  renders LLM content with no network access.
- `goslide build` on a cold cache fails with a clear message naming the
  affected slides; adding `--llm-refresh` recovers.
- `/api/llm` is a single POST route that both serve and host modes
  register; the handler shares cache across viewers.
- All unit + integration tests pass; no real LLM calls in CI; no new
  external dependencies beyond what Phase 6a already introduced.

---

## 9. Release Note Snippet (draft for v1.4.0)

```
### 🧠 LLM transformer in API components (experimental)

Add a new render type `llm` to any `api` component. The LLM receives the
fetched data via the `{{data}}` placeholder in a user-authored prompt
and returns a summary / commentary / interpretation that renders inline.

Cache-first: the same prompt+data pair calls the model at most once.
Click-to-call: cache misses in `goslide serve` show a button instead of
auto-calling. Build-lock: `goslide build` bakes cached results into
static HTML so the exported deck never reaches out.

Reuses the Phase 6a `generate:` config — no new settings needed.
```
