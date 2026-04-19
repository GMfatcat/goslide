# 🎉 GoSlide v1.2.0

## What's New

### 🤖 AI Slide Generation (experimental)

New `goslide generate` command produces a full GoSlide presentation by
calling any OpenAI-compatible LLM endpoint — OpenAI, OpenRouter, Ollama,
vllm, sglang, and others.

```bash
export OPENAI_API_KEY=sk-...
goslide generate "Introduction to Kubernetes"            # simple mode
goslide generate my-prompt.md -o talk.md                 # advanced mode
goslide generate --dump-prompt > system.txt              # inspect prompt
```

**Highlights**

- **Two modes.** Single-topic strings for quick drafts; `prompt.md` files
  with YAML frontmatter (`topic` / `audience` / `slides` / `theme` /
  `language`) plus free-text body for controlled generation.
- **Embedded system prompt.** Ships inside the binary, describes the
  AI-facing GoSlide subset (frontmatter, layouts, card/chart components).
  Dump it with `--dump-prompt` to feed any chat UI manually.
- **Heuristic auto-fix.** Four local rules recover the most common LLM
  syntax slips (unclosed code fences, missing frontmatter terminator,
  unquoted YAML values with colons, missing trailing newline). Fixes are
  reported transparently on stderr.
- **Safe defaults.** Refuses to overwrite existing output unless `--force`
  is passed. On unrecoverable parse failure, writes `<output>.raw.md` and
  `<output>.fixed.md` next to the target for diagnosis instead of
  clobbering your work.
- **API keys never touch disk.** Read from an environment variable named
  by `generate.api_key_env`.

**Configuration** (`goslide.yaml`):

```yaml
generate:
  base_url: https://api.openai.com/v1   # or https://openrouter.ai/api/v1, http://localhost:11434/v1, etc.
  model: gpt-4o
  api_key_env: OPENAI_API_KEY
  timeout: 120s
```

**Experimental notice.** Output quality depends on the chosen model and
prompt wording; semantic quality (flow, accuracy, style) is not guaranteed.
CLI flags and API may change. Review generated slides before presenting.

### ✅ Validated Examples

See [`examples/ai-generated/`](examples/ai-generated/) for real outputs
produced by `openai/gpt-oss-120b:free` on OpenRouter — English simple mode
and 繁體中文 advanced mode (high-school audience, 快餐廚房 metaphor). Both
parsed on first pass with no fixup needed. `scripts/test-generate-llm.ps1`
reproduces the test interactively.

### 📝 Documentation

- `PRD.md` §13 reflects the implemented state
- Both English and 繁體中文 READMEs get a new **AI slide generation**
  section with an experimental warning

## No Breaking Changes

v1.1.0 projects work unchanged. The `generate:` config section is optional
— absent when not using `goslide generate`.

## Full Changelog

See [v1.1.0...v1.2.0](https://github.com/GMfatcat/goslide/compare/v1.1.0...v1.2.0) for all changes.
