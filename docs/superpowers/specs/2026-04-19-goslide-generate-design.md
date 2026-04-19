# `goslide generate` — LLM-Driven Slide Generation

**Date:** 2026-04-19
**Status:** Design approved, ready for implementation plan
**Scope:** Phase 6a — AI slide generation (PRD §13)

---

## 1. Overview

Add a new CLI subcommand `goslide generate` that produces a valid GoSlide `.md`
file by calling an OpenAI-compatible LLM endpoint. One implementation covers
cloud OpenAI, OpenRouter, local vllm, sglang, and Ollama (all expose the
`/v1/chat/completions` contract).

Two usage modes:

- **Simple:** `goslide generate "topic"` — one-shot from a topic string.
- **Advanced:** `goslide generate prompt.md` — YAML frontmatter + free-text
  body for audience/slides/theme/language/notes control.

A `--dump-prompt` flag prints the embedded system prompt so users can feed it
to any LLM chat UI manually (PRD §13.2 external-use case).

---

## 2. CLI Surface

```
goslide generate "topic"                     # simple mode
goslide generate prompt.md                    # advanced mode (arg is a file)
goslide generate --dump-prompt                # print system prompt, exit

Flags:
  -o, --output <file>       Output path (default: talk.md)
  -f, --force               Overwrite existing output file
      --model <name>        Override goslide.yaml generate.model
      --base-url <url>      Override goslide.yaml generate.base_url
      --api-key-env <var>   Override env var name (default: OPENAI_API_KEY)
      --dump-prompt         Print embedded system prompt to stdout, exit 0
```

Mode detection: if the positional arg resolves to an existing file path →
advanced mode; otherwise simple mode treating the arg as topic text.

**Overwrite behaviour:** default refuses to overwrite an existing output file.
`-f/--force` overrides. `talk.md` is often edited by hand after generation, so
silent overwrite is unsafe.

---

## 3. Config (`goslide.yaml`)

New top-level section:

```yaml
generate:
  base_url: https://api.openai.com/v1   # or https://openrouter.ai/api/v1, http://localhost:11434/v1 (Ollama), etc.
  model: gpt-4o
  api_key_env: OPENAI_API_KEY
  timeout: 120s                          # optional, default 120s
```

**Precedence** (highest wins): CLI flag > `goslide.yaml` > default.

**API key** is read only from the environment variable named by
`api_key_env`. Never stored in YAML or any file written by GoSlide. If the
variable is unset, generation fails with an actionable error.

---

## 4. Architecture

### 4.1 File layout

```
internal/
├── cli/
│   └── generate.go              # Cobra command, flag parsing, dispatch
└── generate/
    ├── generate.go              # Run(ctx, Options) error — orchestrator
    ├── options.go               # Options struct + resolve(yaml, flags, env)
    ├── client.go                # OpenAI-compatible HTTP client
    ├── prompt.go                # Build system + user messages
    ├── fixup.go                 # Heuristic fix rules + FixReport
    ├── system_prompt.md         # //go:embed — built-in system prompt
    └── *_test.go
```

Flat single-package layout, consistent with existing `internal/` (parser,
renderer, etc.). No `provider/` subpackage — only one provider today
(OpenAI-compatible); abstraction can be introduced later if a non-compatible
provider is added.

### 4.2 Unit responsibilities

| File | Responsibility | Public surface |
|------|----------------|----------------|
| `cli/generate.go` | Parse args/flags, load config, dispatch `--dump-prompt`, call `generate.Run` | `newGenerateCmd() *cobra.Command` |
| `generate.go` | Orchestrator: build prompt → call client → parse → fixup → write | `Run(ctx, Options) error` |
| `options.go` | Options struct; merge YAML + flags + env; validate required fields | `type Options struct{...}`; `Resolve(...)` |
| `client.go` | POST `/chat/completions`; non-streaming; timeout honoured | `type Client`; `Complete(ctx, msgs) (string, Usage, error)` |
| `prompt.go` | Build `[]Message` from simple topic or advanced prompt.md | `BuildUserMessage(Input) (string, error)` |
| `fixup.go` | Apply heuristic rules; return fixed markdown + report | `Try(md string, parseErr error) (fixed string, report FixReport)` |
| `system_prompt.md` | Embedded prompt describing AI-facing spec subset (§6) | `//go:embed` |

### 4.3 External dependencies

- `net/http` standard library — no new SDK dependency; a single JSON POST is
  simple enough to implement directly.
- Reuse `internal/parser` for sanity-check parsing of the LLM output.
- Reuse `internal/config` (add the `generate:` section struct).

---

## 5. Data Flow

### 5.1 Happy path

```
CLI args + goslide.yaml + env
        ↓
  options.Resolve → Options{BaseURL, Model, APIKey, Timeout, Input, Output, Force}
        ↓
  prompt.BuildUserMessage(Input) → user content
        ↓
  Messages = [system (embedded), user]
        ↓
  client.Complete(ctx, Messages) → raw markdown
        ↓
  parser.Parse(raw) — sanity check
        ↓
  ok  → write Output (refuse if exists unless --force); print ✓ + usage
  fail → see §5.3
```

### 5.2 Input modes

**Simple mode** (`Input{Topic: "..."}`):
```
user = template.Render("Generate a GoSlide presentation about: {Topic}.
                        Target 10–15 slides.")
```

**Advanced mode** (`prompt.md`): parse YAML frontmatter + body into:
```go
type Input struct {
    Topic    string   // required
    Audience string   // optional
    Slides   int      // optional, default 10–15
    Theme    string   // optional, one of the 22 built-in themes
    Language string   // optional, default "en"
    Notes    string   // the markdown body — free-text requirements
}
```
Rendered into a structured user message with sections: Topic, Audience,
Slides, Theme, Language, Additional notes.

### 5.3 Parse-failure path with heuristic auto-fix

```
parser.Parse(raw) fails
        ↓
  fixup.Try(raw, parseErr) → (fixed, report)
        ↓
  parser.Parse(fixed)
        ├─ ok  → write Output; print transparent report (see §5.4)
        └─ fail → write <output>.raw.md    (original LLM output)
                  write <output>.fixed.md  (after fixup attempt)
                  stderr: both parse errors + both file paths
                  exit 1
```

### 5.4 Heuristic fix rules (MVP)

1. **fence-close** — unclosed ` ``` `: scan fence pairs; append closing fence.
2. **fm-terminator** — frontmatter opened with `---` but no closing `---`.
3. **separator-in-fence** — `---` inside a fenced block misread as separator;
   the parser handles this already, but fixup skips such lines when scanning
   for other issues.
4. **fm-unquoted-colon** — YAML value containing `:` without quotes.
5. **trailing-newline** — missing final `\n`.

Example transparent report on success:
```
⚠ Generated markdown had 2 issue(s); auto-fix applied:
  [fence-close]   line 87: unclosed ``` block → appended closing fence
  [fm-terminator] line 4:  frontmatter missing terminator → inserted '---'
  Original parse error: line 87: unexpected EOF in code block
  ✓ Re-parse succeeded. Written to talk.md
```

### 5.5 Progress UX

- Before API call: `Generating… (model=gpt-4o, base=https://api.openai.com/v1)`
  with a simple spinner.
- On response: print token usage if returned (`prompt=320 completion=1240 total=1560`).
- Parse + fixup are fast; no spinner needed.
- MVP is non-streaming. Streaming (for the future `chat` render type) is out
  of scope.

---

## 6. Generation Scope (AI-facing spec subset)

The embedded system prompt describes this subset only. Advanced features
remain manual (PRD §13.3).

**In scope:**
- YAML frontmatter: `title`, `theme`, `layout`, `language`
- `---` slide separators
- Standard Markdown: headings, paragraphs, bullet/numbered lists, tables,
  inline code, code blocks, images
- Layouts: `default`, `two-column`, `dashboard`
- Components: `card` (with metadata frontmatter), `chart` (static data only)

**Out of scope (manual authoring only):**
- `api` components / proxy bindings
- Reactive variables (`{{var}}`)
- `embed:html` / `embed:iframe`
- Per-slide speaker notes (`<!-- notes: -->`)

This scope keeps the system prompt manageable while producing output that
"looks like a real presentation" rather than a bare outline.

---

## 7. Error Handling

| Condition | Behaviour | Exit |
|-----------|-----------|------|
| `generate:` section missing in `goslide.yaml` | "Add a `generate:` section to goslide.yaml. See docs." | 1 |
| API-key env var unset (and not `--dump-prompt`) | "Environment variable `<NAME>` is not set." | 1 |
| `base_url` DNS / connection failure | Raw error + "Check base_url and network." | 1 |
| HTTP 4xx from LLM | Print status + `error.message` field from OpenAI-shape JSON | 1 |
| HTTP 5xx from LLM | Print status + "Retry manually." (no auto-retry in MVP) | 1 |
| Request timeout | "Timed out after Ns. Increase `generate.timeout`." | 1 |
| Empty `content` in response | Treated as parse failure → fixup path | 1 |
| `prompt.md` read / frontmatter invalid | Error pointing at offending line | 1 |
| Output file exists without `--force` | "File `<path>` exists; use --force to overwrite." | 1 |

---

## 8. Testing

### 8.1 Unit tests

- **`options_test.go`** — YAML + flag + env precedence; missing-required errors.
- **`prompt_test.go`** — simple-mode template output; advanced-mode
  frontmatter parsing (including missing-optional fields); invalid frontmatter.
- **`fixup_test.go`** — one table-driven case per rule: bad input →
  expected fixed output + `FixReport` contents.
- **`client_test.go`** — `httptest.Server` mocking OpenAI: success, 401,
  429, timeout, malformed JSON.

### 8.2 Integration test (`generate_test.go`)

- `httptest.Server` stands in for the LLM; returns fixture markdown.
- Full `Run(ctx, Options)` executes; verify output file contents.
- Failure path: mock returns broken markdown → verify `.raw.md` and
  `.fixed.md` are written and exit is non-zero.
- `--force` / no-clobber behaviour.

### 8.3 CLI test (`cli/generate_test.go`)

- `--dump-prompt` produces non-empty output containing keywords
  (e.g. `layout:`, `theme:`).
- Flag parsing: `--model`, `--output`, `--force` correctly mapped to Options.

### 8.4 Excluded from CI

- No real LLM calls (cost, flakiness, credentials).
- System-prompt quality is validated by humans: `--dump-prompt` piped into a
  chat UI, review generated output manually. No automated eval in MVP.

### 8.5 Fixtures

- `testdata/responses/*.md` — sample LLM outputs (well-formed, missing fence,
  bad YAML, empty content).
- `testdata/prompts/*.md` — sample advanced-mode prompt files.

---

## 9. Out of Scope (explicit non-goals)

- Multi-provider abstraction (Anthropic/Gemini native APIs) — YAGNI until a
  non-OpenAI-compatible provider is actually needed.
- LLM-based auto-repair (re-calling the LLM with the parser error) — MVP
  uses local heuristics only.
- Streaming responses — deferred to the `chat` render type feature.
- Outline-then-slides two-phase workflow — simple/advanced modes cover the
  needs; can revisit if prompt-length issues emerge.
- PDF / single-file export integration — the `build` command handles that
  independently on the generated `.md`.
- Automated quality evaluation of generated slides.

---

## 10. Success Criteria

- `goslide generate "topic"` produces a `talk.md` that `goslide serve` renders
  without errors, for a handful of diverse topics (verified manually).
- `goslide generate prompt.md` honours `topic`, `audience`, `slides`, `theme`,
  and `language` fields in the generated output.
- `--dump-prompt` output, pasted into ChatGPT/Claude with a topic, produces
  comparable-quality slides (manual spot check).
- Heuristic fixup recovers at least the five listed error patterns
  (unit-tested).
- All unit + integration tests pass; no real LLM calls in CI.
