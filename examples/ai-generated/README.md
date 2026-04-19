# AI-Generated Examples

Real outputs from `goslide generate` used to validate the end-to-end pipeline.

| File | Mode | Topic | Audience | Language | Phase |
|------|------|-------|----------|----------|-------|
| `k8s-en.md` | simple | Kubernetes | general | English | 6a (v1.2.0) |
| `k8s-zh-TW.md` | advanced (`prompt.md`) | Kubernetes | 高中學生 | 繁體中文 | 6a (v1.2.0) |
| `k8s-visual.md` | advanced (`prompt.md`) | Kubernetes architecture — diagram-heavy | Backend engineers | English | 6b (v1.3.0) |

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
