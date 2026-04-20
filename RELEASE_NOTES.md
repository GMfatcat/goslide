# 🎉 GoSlide v1.4.0

## What's New

### 🧠 LLM transformer inside `api` components (experimental)

The `api` component accepts a new render item of type `llm`. Fetched
JSON is substituted into a user-authored prompt via `{{data}}` and the
model's reply renders inline alongside the chart, table, or metric.

```
~~~api
endpoint: /api/sales
fixture: sales.json           # optional; used by goslide build
render:
  - type: chart:bar
    label-key: quarter
    data-key: revenue
  - type: llm
    prompt: |
      Write 2 analyst bullets on these numbers:
      {{data}}
~~~
```

**Control model** (three layers, not configurable — that's the point):

- **Cache-first** — identical `(model, prompt, data)` triples call the
  LLM at most once. Results land in `.goslide-cache/<sha256>.json`
  (human-readable, commit-safe). Canonical JSON keying means logically
  equal data (different key order, whitespace) hits the same entry.
- **Click-to-call** — `goslide serve` shows a `Generate ✨` button on
  cache miss. Page load never triggers an LLM call automatically.
  Localstorage caches per-browser.
- **Build-lock** — `goslide build` inlines cached results as a
  `data-llm-bakes` attribute on the api component. The exported HTML
  never contacts an LLM at view time.

```bash
# Warm cache via the dev loop:
goslide serve talk.md        # click Generate to populate .goslide-cache/

# Export to static HTML (reads cache only, zero network):
goslide build talk.md

# Or refresh cache during build (the one place we call LLM non-interactively):
goslide build talk.md --llm-refresh
```

Cache miss during `goslide build` is a hard error by default, listing
every affected `slide / component / render-item`. Pass `--llm-refresh`
to opt in to filling the cache during the build.

**Offline build with a fixture file** — `fixture: ./sales.json` on the
api component lets `goslide build` read static data instead of calling
a live endpoint. Great for committing a reproducible snapshot.

**Reuses existing `generate:` config** — no new YAML keys; the same
OpenAI-compatible endpoint that powers `goslide generate` powers the
LLM transformer. Works with any compatible provider (OpenAI,
OpenRouter, Ollama, vllm, sglang, etc.).

### ✅ Validation

The IR validator rejects `llm` render items missing `prompt` with code
`llm-missing-prompt` — `goslide build` / `goslide serve` won't start
against a broken deck.

### 📦 Validated example

[`examples/ai-generated/api-llm-sales/`](examples/ai-generated/api-llm-sales/)
is a fully self-contained directory. Clone this repo, `cd` into the
directory, run `goslide build demo.md`, and you'll see an LLM-written
analyst summary render in the output HTML without ever touching an LLM
— the committed `.goslide-cache/` entry satisfies the bake. Real-LLM
regeneration requires `OPENROUTER_API_KEY` and `--llm-refresh`.

## Compatibility

No breaking changes. v1.3.0 decks build unchanged; the new `llm` render
type is additive and only exercises the new code path when present.
The internal change to `builder.Build` (moving `config.Load` earlier so
the bake can mutate the IR before render) is invisible to callers.

## Out of scope (future work)

- Streaming LLM responses (SSE). Current MVP buffers the full reply.
- JSONPath-style `{{field.x}}` expressions in the prompt. Current MVP
  injects the whole response as `{{data}}`.
- `goslide generate` emitting `llm` render items. Manual-author-only in
  v1.4.0.

## Full Changelog

See [v1.3.0...v1.4.0](https://github.com/GMfatcat/goslide/compare/v1.3.0...v1.4.0) for all changes.
