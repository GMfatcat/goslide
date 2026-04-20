# AI-Generated Examples

Real outputs used to validate GoSlide's AI pipelines end-to-end.

| File | Feature | Phase |
|------|---------|-------|
| `k8s-en.md` | `goslide generate` — simple mode | 6a (v1.2.0) |
| `k8s-zh-TW.md` | `goslide generate` — advanced (zh-TW HS audience) | 6a (v1.2.0) |
| `k8s-visual.md` | `goslide generate` — image placeholders + image-grid | 6b (v1.3.0) |
| `api-llm-sales/` | `api` component + `llm` render item (build-bake) | 7a (v1.4.0) |

**Phase 6a examples (`k8s-en.md`, `k8s-zh-TW.md`)**

- Model: `openai/gpt-oss-120b:free` on OpenRouter
- Date: 2026-04-19
- GoSlide version: post-v1.1.0, Phase 6a (`goslide generate`)
- Both outputs parsed cleanly on first pass.
- Known limitation: the `chart` component in `k8s-en.md` uses a triple-
  backtick fence ( ` ```chart ` ) instead of the required `~~~chart`. It
  renders as a plain code block rather than an interactive Chart.js. This
  was the LLM's interpretation of ambiguous wording in the system prompt;
  the Phase 6b prompt iteration (below) makes the `~~~` fence rule
  explicit, and the new `k8s-visual.md` uses the correct fence syntax.

**Phase 6b example (`k8s-visual.md`)**

- Model: `openai/gpt-oss-20b:free` on OpenRouter (chosen for smallest
  OSS-friendly model that still emits correct syntax)
- Date: 2026-04-19
- GoSlide version: Phase 6b (`placeholder` component + `image-grid` layout)
- 13 `placeholder` components + 4 `region-cell` divs (one 4-cell image-grid
  slide) rendered after `goslide build`.
- One warning: the LLM emitted `aspect: 3:2` on one slide, which falls
  back to 16:9 via the validator — the pipeline degrades gracefully.

**Phase 7a example (`api-llm-sales/`)**

- Model: `openai/gpt-oss-120b:free` on OpenRouter
- Date: 2026-04-20
- GoSlide version: Phase 7a (`api` component with new `type: llm` render item)
- Self-contained directory with `demo.md`, `sales.json` (fixture), `goslide.yaml`
  (generate config), and a committed `.goslide-cache/` entry. Running
  `goslide build demo.md` inside the directory reproduces the deck
  offline — no LLM call, no API key required, thanks to the build-lock
  model.
- To regenerate the cache entry against a live model, delete
  `.goslide-cache/` and run `OPENROUTER_API_KEY=... goslide build demo.md
  --llm-refresh`. The LLM received the fixture JSON via the `{{data}}`
  placeholder and returned a 2-bullet analyst summary naming the best and
  worst quarters with their revenue figures. Total tokens: 302.

## Model sweep (Phase 6b, 2026-04-19)

Same `prompt.md` fed to six free OpenRouter models after the Phase 6b
system-prompt iteration:

| Model | Rendered placeholders | Rendered cells | Notes |
|-------|----------------------|----------------|-------|
| `openai/gpt-oss-20b:free` | 13 | 4 | Cleanest; saved as `k8s-visual.md` |
| `openai/gpt-oss-120b:free` | 12 | 4 | Consistent with 20B variant |
| `arcee-ai/trinity-large-preview:free` | 9 | 8 | Three image-grid slides |
| `nvidia/nemotron-nano-9b-v2:free` | 2 | 0 | Too small; partial output |
| `nvidia/nemotron-3-super-120b-a12b:free` | 0 | 0 | Missed `hint` on 3 slides; validator correctly rejects |
| `nvidia/nemotron-3-nano-30b-a3b:free` | — | — | Parse failed |
| `meta-llama/llama-3.3-70b-instruct:free` | — | — | 429 rate-limited on the test run |
| `qwen/qwen3-next-80b-a3b-instruct:free` | — | — | 429 rate-limited |
| `z-ai/glm-4.5-air:free` | — | — | 429 rate-limited |

Broadly, 20B-class and larger models handle the full AI-facing subset
(components + layouts + placeholders) after two rounds of prompt tuning.
Sub-10B models produce partial output.

## Reproducing

See `scripts/test-generate-llm.ps1`. The script prompts for an OpenRouter
API key interactively so it stays out of files and shell history. Swap
the `model:` line in the generated `goslide.yaml` to try a different
provider.

## Caveat

These outputs are snapshots on one day against one provider. LLM quality
varies by model, provider rate-limiting, and prompt wording. See the
"experimental" note in the top-level README.
