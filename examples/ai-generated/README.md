# AI-Generated Examples

Real outputs from `goslide generate` used to validate the end-to-end pipeline.

| File | Mode | Topic | Audience | Language |
|------|------|-------|----------|----------|
| `k8s-en.md` | simple | Kubernetes | general | English |
| `k8s-zh-TW.md` | advanced (`prompt.md`) | Kubernetes | 高中學生 | 繁體中文 |

**Generation details**

- Model: `openai/gpt-oss-120b:free` on OpenRouter
- Date: 2026-04-19
- GoSlide version: post-v1.1.0, Phase 6a (`goslide generate`)
- Both outputs parsed cleanly on first pass — the heuristic fixup ran but
  made no changes.

**Reproducing**

See `scripts/test-generate-llm.ps1`. You need an OpenRouter API key exported
as `OPENROUTER_API_KEY`. The script prompts for the key interactively so it
stays out of files and shell history.

**Caveat**

These outputs are representative of one model on one day. LLM output quality
varies by model, provider availability, and prompt wording. See the
"experimental" note in the top-level README.
