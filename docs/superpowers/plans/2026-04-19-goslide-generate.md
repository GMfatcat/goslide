# `goslide generate` Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a `goslide generate` CLI command that produces valid GoSlide `.md` files by calling an OpenAI-compatible LLM endpoint, with simple/advanced prompt modes and heuristic auto-fix.

**Architecture:** Single flat `internal/generate/` package orchestrating prompt building → HTTP call → parse sanity check → heuristic fixup → file write. Plus `internal/cli/generate.go` for Cobra wiring and `generate:` section added to `internal/config`.

**Tech Stack:** Go 1.21.6, `net/http` (std lib), `gopkg.in/yaml.v3`, `github.com/spf13/cobra`, `github.com/stretchr/testify`. Reuses existing `internal/parser` for sanity check.

**Spec:** `docs/superpowers/specs/2026-04-19-goslide-generate-design.md`

---

## File Structure

**Create:**
- `internal/generate/options.go` — `Options` struct, `Resolve(cfg, flags, env)`
- `internal/generate/options_test.go`
- `internal/generate/prompt.go` — `Input`, `BuildUserMessage`, advanced-mode frontmatter parse
- `internal/generate/prompt_test.go`
- `internal/generate/client.go` — `Client`, `Complete(ctx, msgs)`
- `internal/generate/client_test.go`
- `internal/generate/fixup.go` — heuristic rules, `Try`, `FixReport`
- `internal/generate/fixup_test.go`
- `internal/generate/generate.go` — `Run(ctx, Options) error`
- `internal/generate/generate_test.go` — end-to-end with httptest
- `internal/generate/system_prompt.md` — embedded system prompt
- `internal/generate/embed.go` — `//go:embed` bridge
- `internal/cli/generate.go`
- `internal/cli/generate_test.go`

**Modify:**
- `internal/config/config.go` — add `Generate GenerateConfig` field
- `internal/config/config_test.go` — add test for new section

**Testdata:**
- `internal/generate/testdata/responses/good.md`
- `internal/generate/testdata/responses/unclosed_fence.md`
- `internal/generate/testdata/responses/missing_fm_terminator.md`
- `internal/generate/testdata/responses/broken_unfixable.md`
- `internal/generate/testdata/prompts/full.md`
- `internal/generate/testdata/prompts/minimal.md`

---

## Task 1: Extend `Config` with `generate:` section

**Files:**
- Modify: `internal/config/config.go`
- Modify: `internal/config/config_test.go`

- [ ] **Step 1: Write the failing test**

Append to `internal/config/config_test.go`:

```go
func TestLoad_GenerateSection(t *testing.T) {
	dir := t.TempDir()
	content := `
generate:
  base_url: https://api.openai.com/v1
  model: gpt-4o
  api_key_env: OPENAI_API_KEY
  timeout: 90s
`
	os.WriteFile(filepath.Join(dir, "goslide.yaml"), []byte(content), 0644)

	cfg, err := Load(dir)
	require.NoError(t, err)
	require.Equal(t, "https://api.openai.com/v1", cfg.Generate.BaseURL)
	require.Equal(t, "gpt-4o", cfg.Generate.Model)
	require.Equal(t, "OPENAI_API_KEY", cfg.Generate.APIKeyEnv)
	require.Equal(t, 90*time.Second, cfg.Generate.Timeout)
}
```

Add `"time"` to the imports if not present.

- [ ] **Step 2: Run test to verify it fails**

Run: `go test -C . ./internal/config/ -run TestLoad_GenerateSection -v`
Expected: FAIL (Generate field does not exist).

- [ ] **Step 3: Implement**

Modify `internal/config/config.go`:

```go
package config

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Theme    ThemeConfig    `yaml:"theme"`
	API      APIConfig      `yaml:"api"`
	Generate GenerateConfig `yaml:"generate"`
}

type ThemeConfig struct {
	Overrides map[string]string `yaml:"overrides"`
}

type APIConfig struct {
	Proxy map[string]ProxyTarget `yaml:"proxy"`
}

type ProxyTarget struct {
	Target  string            `yaml:"target"`
	Headers map[string]string `yaml:"headers"`
}

type GenerateConfig struct {
	BaseURL   string        `yaml:"base_url"`
	Model     string        `yaml:"model"`
	APIKeyEnv string        `yaml:"api_key_env"`
	Timeout   time.Duration `yaml:"timeout"`
}
```

`time.Duration` unmarshals strings like `"90s"`, `"2m"` natively via yaml.v3.

- [ ] **Step 4: Run the test to verify it passes**

Run: `go test -C . ./internal/config/ -v`
Expected: all pass (existing + new).

- [ ] **Step 5: Commit**

```bash
git add internal/config/config.go internal/config/config_test.go
git commit -m "feat(config): add generate section (base_url, model, api_key_env, timeout)"
```

---

## Task 2: Embed system prompt file

**Files:**
- Create: `internal/generate/system_prompt.md`
- Create: `internal/generate/embed.go`
- Create: `internal/generate/embed_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/generate/embed_test.go`:

```go
package generate

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSystemPrompt_NotEmpty(t *testing.T) {
	p := SystemPrompt()
	require.NotEmpty(t, p)
}

func TestSystemPrompt_ContainsCoreKeywords(t *testing.T) {
	p := SystemPrompt()
	for _, kw := range []string{"layout:", "theme:", "two-column", "dashboard", "---"} {
		require.Truef(t, strings.Contains(p, kw), "system prompt missing keyword %q", kw)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test -C . ./internal/generate/ -v`
Expected: compile error (package does not exist).

- [ ] **Step 3: Create the system prompt content**

Create `internal/generate/system_prompt.md`:

```markdown
You generate GoSlide presentations as a single Markdown file.

GoSlide is a Markdown-driven slide system built on Reveal.js. Output only the
Markdown document — no preamble, no closing commentary, no code fences around
the whole document.

# File structure

- A `---` line on its own separates slides.
- Each slide MAY start with a YAML frontmatter block delimited by `---` lines
  (the opening `---` is the slide separator).
- Standard Markdown is used for content: headings, paragraphs, bullet and
  numbered lists, tables, inline code, fenced code blocks, images.

# Frontmatter fields

```yaml
---
title: Slide title        # optional
theme: dark               # optional; one of the built-in themes
layout: two-column        # optional; see Layouts below
language: en              # optional
---
```

Omit fields you do not need. The very first slide typically sets `theme`.

# Layouts

- `default` — single column (omit `layout:` for this).
- `two-column` — left/right regions split by `<!-- col -->` on its own line.
- `dashboard` — grid of cards/charts; one component per cell.

# Components

## Card

````
```card
---
title: Card title
icon: "📊"        # optional emoji
---
Body text in Markdown. Supports **bold**, *italics*, lists, links.
```
````

## Chart (static data only)

````
```chart
type: bar                 # bar | line | pie
title: Sales by quarter
data:
  labels: [Q1, Q2, Q3, Q4]
  values: [12, 19, 7, 15]
```
````

# Rules

- Produce 8–15 slides unless the user asks for a different count.
- The first slide is a title slide (H1 + subtitle paragraph).
- The last slide is either a summary or a Q&A prompt.
- Do NOT use `api:`, reactive variables `{{var}}`, `embed:html`, or
  `embed:iframe`. Those are manual-only features.
- Keep each slide focused: one idea per slide, ≤6 bullets, ≤60 words of body.
- Use the user's requested language; default to English.
- Return ONLY the Markdown document. No JSON, no wrapping fences, no prose
  around it.
```

- [ ] **Step 4: Create the embed bridge**

Create `internal/generate/embed.go`:

```go
package generate

import _ "embed"

//go:embed system_prompt.md
var systemPromptRaw string

// SystemPrompt returns the embedded GoSlide system prompt.
func SystemPrompt() string {
	return systemPromptRaw
}
```

- [ ] **Step 5: Run the test to verify it passes**

Run: `go test -C . ./internal/generate/ -v`
Expected: both tests pass.

- [ ] **Step 6: Commit**

```bash
git add internal/generate/system_prompt.md internal/generate/embed.go internal/generate/embed_test.go
git commit -m "feat(generate): embed GoSlide system prompt"
```

---

## Task 3: Prompt builder — simple mode

**Files:**
- Create: `internal/generate/prompt.go`
- Create: `internal/generate/prompt_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/generate/prompt_test.go`:

```go
package generate

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildUserMessage_SimpleMode(t *testing.T) {
	in := Input{Topic: "Introduction to Kubernetes"}
	msg, err := BuildUserMessage(in)
	require.NoError(t, err)
	require.Contains(t, msg, "Introduction to Kubernetes")
	require.Contains(t, strings.ToLower(msg), "topic")
}

func TestBuildUserMessage_EmptyInputFails(t *testing.T) {
	_, err := BuildUserMessage(Input{})
	require.Error(t, err)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test -C . ./internal/generate/ -run BuildUserMessage -v`
Expected: FAIL (undefined: Input, BuildUserMessage).

- [ ] **Step 3: Implement**

Create `internal/generate/prompt.go`:

```go
package generate

import (
	"errors"
	"fmt"
	"strings"
)

// Input describes what the user wants generated. Topic is required; every
// other field is optional and used in advanced mode.
type Input struct {
	Topic    string
	Audience string
	Slides   int
	Theme    string
	Language string
	Notes    string // free-text body from prompt.md
}

// BuildUserMessage composes the user-role message sent to the LLM.
func BuildUserMessage(in Input) (string, error) {
	if strings.TrimSpace(in.Topic) == "" {
		return "", errors.New("topic is required")
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Topic: %s\n", in.Topic)

	if in.Audience != "" {
		fmt.Fprintf(&b, "Audience: %s\n", in.Audience)
	}
	if in.Slides > 0 {
		fmt.Fprintf(&b, "Slides: %d\n", in.Slides)
	} else {
		b.WriteString("Slides: 10-15\n")
	}
	if in.Theme != "" {
		fmt.Fprintf(&b, "Theme: %s\n", in.Theme)
	}
	if in.Language != "" {
		fmt.Fprintf(&b, "Language: %s\n", in.Language)
	} else {
		b.WriteString("Language: en\n")
	}
	if strings.TrimSpace(in.Notes) != "" {
		b.WriteString("\nAdditional notes:\n")
		b.WriteString(strings.TrimSpace(in.Notes))
		b.WriteString("\n")
	}

	b.WriteString("\nGenerate the full GoSlide Markdown now.\n")
	return b.String(), nil
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test -C . ./internal/generate/ -run BuildUserMessage -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/generate/prompt.go internal/generate/prompt_test.go
git commit -m "feat(generate): add Input + simple-mode BuildUserMessage"
```

---

## Task 4: Prompt builder — advanced mode (parse prompt.md)

**Files:**
- Modify: `internal/generate/prompt.go`
- Modify: `internal/generate/prompt_test.go`
- Create: `internal/generate/testdata/prompts/full.md`
- Create: `internal/generate/testdata/prompts/minimal.md`

- [ ] **Step 1: Create testdata fixtures**

Create `internal/generate/testdata/prompts/full.md`:

```markdown
---
topic: Kubernetes Architecture
audience: Backend engineers
slides: 15
theme: dark
language: en
---
Emphasize Pod/Service/Ingress. End with a Q&A slide. Include one chart
showing request routing.
```

Create `internal/generate/testdata/prompts/minimal.md`:

```markdown
---
topic: Quarterly review
---
```

- [ ] **Step 2: Write failing tests**

Append to `internal/generate/prompt_test.go`:

```go
func TestParsePromptFile_Full(t *testing.T) {
	in, err := ParsePromptFile("testdata/prompts/full.md")
	require.NoError(t, err)
	require.Equal(t, "Kubernetes Architecture", in.Topic)
	require.Equal(t, "Backend engineers", in.Audience)
	require.Equal(t, 15, in.Slides)
	require.Equal(t, "dark", in.Theme)
	require.Equal(t, "en", in.Language)
	require.Contains(t, in.Notes, "Pod/Service/Ingress")
}

func TestParsePromptFile_Minimal(t *testing.T) {
	in, err := ParsePromptFile("testdata/prompts/minimal.md")
	require.NoError(t, err)
	require.Equal(t, "Quarterly review", in.Topic)
	require.Empty(t, in.Audience)
	require.Zero(t, in.Slides)
	require.Empty(t, strings.TrimSpace(in.Notes))
}

func TestParsePromptFile_MissingTopic(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "bad.md")
	require.NoError(t, os.WriteFile(p, []byte("---\naudience: eng\n---\nbody\n"), 0644))

	_, err := ParsePromptFile(p)
	require.Error(t, err)
	require.Contains(t, err.Error(), "topic")
}

func TestParsePromptFile_NoFrontmatter(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "bad.md")
	require.NoError(t, os.WriteFile(p, []byte("just body\n"), 0644))

	_, err := ParsePromptFile(p)
	require.Error(t, err)
}
```

Add imports for `"os"`, `"path/filepath"`.

- [ ] **Step 3: Run tests to verify they fail**

Run: `go test -C . ./internal/generate/ -run ParsePromptFile -v`
Expected: FAIL (undefined: ParsePromptFile).

- [ ] **Step 4: Implement**

Append to `internal/generate/prompt.go`:

```go
import (
	// existing imports...
	"bufio"
	"bytes"
	"os"

	"gopkg.in/yaml.v3"
)

// ParsePromptFile reads a prompt.md file (YAML frontmatter + Markdown body)
// into an Input. The `topic` frontmatter field is required.
func ParsePromptFile(path string) (Input, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Input{}, fmt.Errorf("read prompt file: %w", err)
	}

	fmBytes, body, err := splitFrontmatter(data)
	if err != nil {
		return Input{}, err
	}

	var raw struct {
		Topic    string `yaml:"topic"`
		Audience string `yaml:"audience"`
		Slides   int    `yaml:"slides"`
		Theme    string `yaml:"theme"`
		Language string `yaml:"language"`
	}
	if err := yaml.Unmarshal(fmBytes, &raw); err != nil {
		return Input{}, fmt.Errorf("parse frontmatter: %w", err)
	}
	if strings.TrimSpace(raw.Topic) == "" {
		return Input{}, errors.New("prompt.md frontmatter: `topic` is required")
	}

	return Input{
		Topic:    raw.Topic,
		Audience: raw.Audience,
		Slides:   raw.Slides,
		Theme:    raw.Theme,
		Language: raw.Language,
		Notes:    string(body),
	}, nil
}

// splitFrontmatter returns (yaml, body, error). The file MUST start with a
// line that is exactly "---"; the frontmatter ends at the next line that is
// exactly "---". Everything after is the body.
func splitFrontmatter(data []byte) (fm []byte, body []byte, err error) {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)

	var fmBuf, bodyBuf bytes.Buffer
	state := 0 // 0=expect opening ---, 1=inside fm, 2=body
	line := 0
	for scanner.Scan() {
		line++
		t := scanner.Text()
		switch state {
		case 0:
			if strings.TrimSpace(t) != "---" {
				return nil, nil, fmt.Errorf("prompt.md must start with '---' (line 1)")
			}
			state = 1
		case 1:
			if strings.TrimSpace(t) == "---" {
				state = 2
				continue
			}
			fmBuf.WriteString(t)
			fmBuf.WriteByte('\n')
		case 2:
			bodyBuf.WriteString(t)
			bodyBuf.WriteByte('\n')
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, nil, err
	}
	if state != 2 {
		return nil, nil, errors.New("prompt.md frontmatter missing closing '---'")
	}
	return fmBuf.Bytes(), bodyBuf.Bytes(), nil
}
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `go test -C . ./internal/generate/ -v`
Expected: all pass.

- [ ] **Step 6: Commit**

```bash
git add internal/generate/prompt.go internal/generate/prompt_test.go internal/generate/testdata/prompts/
git commit -m "feat(generate): parse prompt.md frontmatter for advanced mode"
```

---

## Task 5: `Options` + `Resolve`

**Files:**
- Create: `internal/generate/options.go`
- Create: `internal/generate/options_test.go`

- [ ] **Step 1: Write failing tests**

Create `internal/generate/options_test.go`:

```go
package generate

import (
	"testing"
	"time"

	"github.com/GMfatcat/goslide/internal/config"
	"github.com/stretchr/testify/require"
)

func TestResolve_UsesYAMLDefaults(t *testing.T) {
	cfg := &config.Config{Generate: config.GenerateConfig{
		BaseURL:   "https://api.openai.com/v1",
		Model:     "gpt-4o",
		APIKeyEnv: "OPENAI_API_KEY",
		Timeout:   60 * time.Second,
	}}
	t.Setenv("OPENAI_API_KEY", "sk-test")

	opts, err := Resolve(cfg, Flags{}, Input{Topic: "t"})
	require.NoError(t, err)
	require.Equal(t, "https://api.openai.com/v1", opts.BaseURL)
	require.Equal(t, "gpt-4o", opts.Model)
	require.Equal(t, "sk-test", opts.APIKey)
	require.Equal(t, 60*time.Second, opts.Timeout)
}

func TestResolve_FlagsOverrideYAML(t *testing.T) {
	cfg := &config.Config{Generate: config.GenerateConfig{
		BaseURL:   "https://api.openai.com/v1",
		Model:     "gpt-4o",
		APIKeyEnv: "OPENAI_API_KEY",
	}}
	t.Setenv("OPENAI_API_KEY", "sk-test")

	opts, err := Resolve(cfg, Flags{
		BaseURL: "http://localhost:11434/v1",
		Model:   "llama3",
	}, Input{Topic: "t"})
	require.NoError(t, err)
	require.Equal(t, "http://localhost:11434/v1", opts.BaseURL)
	require.Equal(t, "llama3", opts.Model)
}

func TestResolve_APIKeyEnvOverride(t *testing.T) {
	cfg := &config.Config{Generate: config.GenerateConfig{
		BaseURL:   "x",
		Model:     "y",
		APIKeyEnv: "OPENAI_API_KEY",
	}}
	t.Setenv("CUSTOM_KEY", "sk-custom")

	opts, err := Resolve(cfg, Flags{APIKeyEnv: "CUSTOM_KEY"}, Input{Topic: "t"})
	require.NoError(t, err)
	require.Equal(t, "sk-custom", opts.APIKey)
}

func TestResolve_MissingAPIKey(t *testing.T) {
	cfg := &config.Config{Generate: config.GenerateConfig{
		BaseURL: "x", Model: "y", APIKeyEnv: "NOPE_NOT_SET_12345",
	}}
	_, err := Resolve(cfg, Flags{}, Input{Topic: "t"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "NOPE_NOT_SET_12345")
}

func TestResolve_MissingBaseURL(t *testing.T) {
	cfg := &config.Config{Generate: config.GenerateConfig{Model: "y", APIKeyEnv: "OPENAI_API_KEY"}}
	t.Setenv("OPENAI_API_KEY", "k")
	_, err := Resolve(cfg, Flags{}, Input{Topic: "t"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "base_url")
}

func TestResolve_MissingModel(t *testing.T) {
	cfg := &config.Config{Generate: config.GenerateConfig{BaseURL: "x", APIKeyEnv: "OPENAI_API_KEY"}}
	t.Setenv("OPENAI_API_KEY", "k")
	_, err := Resolve(cfg, Flags{}, Input{Topic: "t"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "model")
}

func TestResolve_DefaultTimeout(t *testing.T) {
	cfg := &config.Config{Generate: config.GenerateConfig{BaseURL: "x", Model: "y", APIKeyEnv: "OPENAI_API_KEY"}}
	t.Setenv("OPENAI_API_KEY", "k")
	opts, err := Resolve(cfg, Flags{}, Input{Topic: "t"})
	require.NoError(t, err)
	require.Equal(t, 120*time.Second, opts.Timeout)
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test -C . ./internal/generate/ -run Resolve -v`
Expected: FAIL (undefined symbols).

- [ ] **Step 3: Implement**

Create `internal/generate/options.go`:

```go
package generate

import (
	"fmt"
	"os"
	"time"

	"github.com/GMfatcat/goslide/internal/config"
)

// Flags holds values that come from CLI flags (empty means "not provided").
type Flags struct {
	BaseURL   string
	Model     string
	APIKeyEnv string
	Output    string
	Force     bool
}

// Options is the fully-resolved configuration used by Run.
type Options struct {
	BaseURL string
	Model   string
	APIKey  string
	Timeout time.Duration

	Input  Input
	Output string
	Force  bool
}

const defaultTimeout = 120 * time.Second

// Resolve merges YAML config, CLI flags, and environment variables.
// Precedence: flags > yaml > defaults. Returns an error if required fields
// are missing or the API key env var is unset.
func Resolve(cfg *config.Config, flags Flags, in Input) (Options, error) {
	g := cfg.Generate

	baseURL := firstNonEmpty(flags.BaseURL, g.BaseURL)
	model := firstNonEmpty(flags.Model, g.Model)
	apiKeyEnv := firstNonEmpty(flags.APIKeyEnv, g.APIKeyEnv, "OPENAI_API_KEY")
	output := firstNonEmpty(flags.Output, "talk.md")
	timeout := g.Timeout
	if timeout <= 0 {
		timeout = defaultTimeout
	}

	if baseURL == "" {
		return Options{}, fmt.Errorf("generate.base_url is not set (use --base-url or add it to goslide.yaml)")
	}
	if model == "" {
		return Options{}, fmt.Errorf("generate.model is not set (use --model or add it to goslide.yaml)")
	}

	apiKey := os.Getenv(apiKeyEnv)
	if apiKey == "" {
		return Options{}, fmt.Errorf("environment variable %s is not set", apiKeyEnv)
	}

	return Options{
		BaseURL: baseURL,
		Model:   model,
		APIKey:  apiKey,
		Timeout: timeout,
		Input:   in,
		Output:  output,
		Force:   flags.Force,
	}, nil
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

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test -C . ./internal/generate/ -v`
Expected: all pass.

- [ ] **Step 5: Commit**

```bash
git add internal/generate/options.go internal/generate/options_test.go
git commit -m "feat(generate): add Options + Resolve(cfg, flags, env)"
```

---

## Task 6: OpenAI-compatible HTTP client

**Files:**
- Create: `internal/generate/client.go`
- Create: `internal/generate/client_test.go`

- [ ] **Step 1: Write failing tests**

Create `internal/generate/client_test.go`:

```go
package generate

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestClient_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/chat/completions", r.URL.Path)
		require.Equal(t, "Bearer sk-test", r.Header.Get("Authorization"))

		body, _ := io.ReadAll(r.Body)
		var req struct {
			Model    string `json:"model"`
			Messages []struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"messages"`
		}
		require.NoError(t, json.Unmarshal(body, &req))
		require.Equal(t, "gpt-4o", req.Model)
		require.Len(t, req.Messages, 2)
		require.Equal(t, "system", req.Messages[0].Role)
		require.Equal(t, "user", req.Messages[1].Role)

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"choices": [{"message": {"content": "# slide\n"}}],
			"usage": {"prompt_tokens": 10, "completion_tokens": 20, "total_tokens": 30}
		}`))
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "sk-test", 5*time.Second)
	content, usage, err := c.Complete(context.Background(), "gpt-4o", []Message{
		{Role: "system", Content: "sys"},
		{Role: "user", Content: "hi"},
	})
	require.NoError(t, err)
	require.Equal(t, "# slide\n", content)
	require.Equal(t, 30, usage.TotalTokens)
}

func TestClient_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
		w.Write([]byte(`{"error": {"message": "invalid API key"}}`))
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "bad", time.Second)
	_, _, err := c.Complete(context.Background(), "gpt-4o", []Message{{Role: "user", Content: "x"}})
	require.Error(t, err)
	require.Contains(t, err.Error(), "401")
	require.Contains(t, err.Error(), "invalid API key")
}

func TestClient_EmptyChoices(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"choices": []}`))
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "k", time.Second)
	_, _, err := c.Complete(context.Background(), "gpt-4o", []Message{{Role: "user", Content: "x"}})
	require.Error(t, err)
	require.Contains(t, strings.ToLower(err.Error()), "empty")
}

func TestClient_Timeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "k", 50*time.Millisecond)
	_, _, err := c.Complete(context.Background(), "gpt-4o", []Message{{Role: "user", Content: "x"}})
	require.Error(t, err)
}

func TestClient_TrimsTrailingSlashInBaseURL(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Write([]byte(`{"choices":[{"message":{"content":"ok"}}]}`))
	}))
	defer srv.Close()

	c := NewClient(srv.URL+"/", "k", time.Second)
	_, _, err := c.Complete(context.Background(), "m", []Message{{Role: "user", Content: "x"}})
	require.NoError(t, err)
	require.Equal(t, "/chat/completions", gotPath)
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test -C . ./internal/generate/ -run TestClient -v`
Expected: FAIL (undefined symbols).

- [ ] **Step 3: Implement**

Create `internal/generate/client.go`:

```go
package generate

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Message is a single chat-completions message.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Usage is the token accounting returned by the API (may be zero if absent).
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Completer is the abstraction Run depends on. Client implements it; tests
// can substitute a fake.
type Completer interface {
	Complete(ctx context.Context, model string, msgs []Message) (string, Usage, error)
}

// Client is an OpenAI-compatible chat-completions client.
type Client struct {
	baseURL string
	apiKey  string
	http    *http.Client
}

// NewClient builds a Client. baseURL should be the API root that exposes
// /chat/completions (e.g. https://api.openai.com/v1).
func NewClient(baseURL, apiKey string, timeout time.Duration) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		apiKey:  apiKey,
		http:    &http.Client{Timeout: timeout},
	}
}

type chatReq struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

type chatResp struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
	Usage Usage `json:"usage"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// Complete posts to /chat/completions and returns the content of the first
// choice plus token usage.
func (c *Client) Complete(ctx context.Context, model string, msgs []Message) (string, Usage, error) {
	body, err := json.Marshal(chatReq{Model: model, Messages: msgs, Stream: false})
	if err != nil {
		return "", Usage{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", Usage{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return "", Usage{}, fmt.Errorf("llm request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		msg := strings.TrimSpace(string(respBody))
		var parsed chatResp
		if json.Unmarshal(respBody, &parsed) == nil && parsed.Error != nil && parsed.Error.Message != "" {
			msg = parsed.Error.Message
		}
		return "", Usage{}, fmt.Errorf("llm http %d: %s", resp.StatusCode, msg)
	}

	var parsed chatResp
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", Usage{}, fmt.Errorf("decode llm response: %w", err)
	}
	if len(parsed.Choices) == 0 {
		return "", Usage{}, fmt.Errorf("llm returned empty choices")
	}
	return parsed.Choices[0].Message.Content, parsed.Usage, nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test -C . ./internal/generate/ -v`
Expected: all pass.

- [ ] **Step 5: Commit**

```bash
git add internal/generate/client.go internal/generate/client_test.go
git commit -m "feat(generate): OpenAI-compatible chat.completions client"
```

---

## Task 7: Heuristic fixup — scaffolding + `FixReport`

**Files:**
- Create: `internal/generate/fixup.go`
- Create: `internal/generate/fixup_test.go`

- [ ] **Step 1: Write failing test for the scaffold**

Create `internal/generate/fixup_test.go`:

```go
package generate

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTry_NoChangesForValidMarkdown(t *testing.T) {
	md := "---\ntitle: ok\n---\n\n# Hello\n\nBody.\n"
	fixed, report := Try(md, nil)
	require.Equal(t, md, fixed)
	require.Empty(t, report.Fixes)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test -C . ./internal/generate/ -run TestTry -v`
Expected: FAIL (undefined: Try, FixReport).

- [ ] **Step 3: Implement scaffold**

Create `internal/generate/fixup.go`:

```go
package generate

import "strings"

// FixReport describes which heuristic fixes were applied.
type FixReport struct {
	Fixes []Fix
}

// Fix is one applied heuristic.
type Fix struct {
	Rule        string // e.g. "fence-close"
	Line        int    // 1-based line number where the issue was detected
	Description string // short human-readable summary of what was done
}

// Try applies heuristic fixes in a fixed order and returns the possibly
// modified markdown plus a report. parseErr may be nil (the caller will
// decide whether to re-parse); the current implementation ignores it.
func Try(md string, parseErr error) (string, FixReport) {
	report := FixReport{}
	lines := strings.Split(md, "\n")

	// Rules are applied in order; each mutates `lines` and appends to report.
	applyFenceClose(&lines, &report)
	applyFrontmatterTerminator(&lines, &report)
	applyFrontmatterUnquotedColon(&lines, &report)
	applyTrailingNewline(&lines, &report)

	return strings.Join(lines, "\n"), report
}

// --- rule stubs (filled in by later tasks) ---

func applyFenceClose(lines *[]string, report *FixReport)              {}
func applyFrontmatterTerminator(lines *[]string, report *FixReport)   {}
func applyFrontmatterUnquotedColon(lines *[]string, report *FixReport) {}
func applyTrailingNewline(lines *[]string, report *FixReport)         {}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test -C . ./internal/generate/ -run TestTry -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/generate/fixup.go internal/generate/fixup_test.go
git commit -m "feat(generate): fixup scaffolding (Try, FixReport)"
```

---

## Task 8: Fixup rule — unclosed code fence

**Files:**
- Modify: `internal/generate/fixup.go`
- Modify: `internal/generate/fixup_test.go`

- [ ] **Step 1: Write failing test**

Append to `internal/generate/fixup_test.go`:

```go
func TestApplyFenceClose_AppendsClosingFence(t *testing.T) {
	md := "# Title\n\n```go\nfunc main() {}\n"
	fixed, report := Try(md, nil)

	require.Contains(t, fixed, "```\n")
	require.Len(t, report.Fixes, 1)
	require.Equal(t, "fence-close", report.Fixes[0].Rule)
	require.Equal(t, 3, report.Fixes[0].Line) // opening fence was on line 3
}

func TestApplyFenceClose_NoopWhenBalanced(t *testing.T) {
	md := "```\ncode\n```\n"
	fixed, report := Try(md, nil)
	require.Equal(t, md, fixed)
	require.Empty(t, report.Fixes)
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test -C . ./internal/generate/ -run ApplyFenceClose -v`
Expected: FAIL (body does not contain expected strings).

- [ ] **Step 3: Implement the rule**

Replace the `applyFenceClose` stub in `internal/generate/fixup.go`:

```go
func applyFenceClose(lines *[]string, report *FixReport) {
	openLine := -1
	for i, ln := range *lines {
		t := strings.TrimSpace(ln)
		if strings.HasPrefix(t, "```") {
			if openLine == -1 {
				openLine = i
			} else {
				openLine = -1
			}
		}
	}
	if openLine == -1 {
		return
	}
	*lines = append(*lines, "```")
	report.Fixes = append(report.Fixes, Fix{
		Rule:        "fence-close",
		Line:        openLine + 1,
		Description: "unclosed ``` block → appended closing fence",
	})
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test -C . ./internal/generate/ -v`
Expected: all pass.

- [ ] **Step 5: Commit**

```bash
git add internal/generate/fixup.go internal/generate/fixup_test.go
git commit -m "feat(generate): fixup rule — close unclosed code fence"
```

---

## Task 9: Fixup rule — frontmatter missing terminator

**Files:**
- Modify: `internal/generate/fixup.go`
- Modify: `internal/generate/fixup_test.go`

- [ ] **Step 1: Write failing test**

Append to `internal/generate/fixup_test.go`:

```go
func TestApplyFrontmatterTerminator_InsertsClosing(t *testing.T) {
	md := "---\ntitle: Hello\ntheme: dark\n\n# Heading\n\nBody\n"
	fixed, report := Try(md, nil)

	require.Contains(t, fixed, "---\ntitle: Hello\ntheme: dark\n---\n")
	require.NotEmpty(t, report.Fixes)
	found := false
	for _, f := range report.Fixes {
		if f.Rule == "fm-terminator" {
			found = true
			require.Equal(t, 1, f.Line)
		}
	}
	require.True(t, found, "fm-terminator rule should fire")
}

func TestApplyFrontmatterTerminator_NoopWhenPresent(t *testing.T) {
	md := "---\ntitle: Hello\n---\n\nBody\n"
	_, report := Try(md, nil)
	for _, f := range report.Fixes {
		require.NotEqual(t, "fm-terminator", f.Rule)
	}
}

func TestApplyFrontmatterTerminator_NoopWithoutFrontmatter(t *testing.T) {
	md := "# Just a heading\n\nBody\n"
	_, report := Try(md, nil)
	require.Empty(t, report.Fixes)
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test -C . ./internal/generate/ -run TestApplyFrontmatterTerminator -v`
Expected: FAIL.

- [ ] **Step 3: Implement the rule**

Replace the `applyFrontmatterTerminator` stub:

```go
func applyFrontmatterTerminator(lines *[]string, report *FixReport) {
	if len(*lines) == 0 || strings.TrimSpace((*lines)[0]) != "---" {
		return
	}
	// Look for a closing --- anywhere in the first ~40 lines (frontmatter
	// is short in practice).
	limit := 40
	if limit > len(*lines) {
		limit = len(*lines)
	}
	for i := 1; i < limit; i++ {
		if strings.TrimSpace((*lines)[i]) == "---" {
			return // balanced
		}
	}
	// Insert a terminator after the last non-blank key:value line in the
	// first few lines (scan until first blank line).
	insertAt := 1
	for i := 1; i < len(*lines); i++ {
		if strings.TrimSpace((*lines)[i]) == "" {
			insertAt = i
			break
		}
		insertAt = i + 1
	}
	newLines := append([]string{}, (*lines)[:insertAt]...)
	newLines = append(newLines, "---")
	newLines = append(newLines, (*lines)[insertAt:]...)
	*lines = newLines
	report.Fixes = append(report.Fixes, Fix{
		Rule:        "fm-terminator",
		Line:        1,
		Description: "frontmatter missing terminator → inserted '---'",
	})
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test -C . ./internal/generate/ -v`
Expected: all pass.

- [ ] **Step 5: Commit**

```bash
git add internal/generate/fixup.go internal/generate/fixup_test.go
git commit -m "feat(generate): fixup rule — insert missing frontmatter terminator"
```

---

## Task 10: Fixup rule — YAML unquoted colon

**Files:**
- Modify: `internal/generate/fixup.go`
- Modify: `internal/generate/fixup_test.go`

- [ ] **Step 1: Write failing test**

Append to `internal/generate/fixup_test.go`:

```go
func TestApplyFrontmatterUnquotedColon_Quotes(t *testing.T) {
	md := "---\ntitle: Go: A Short Intro\ntheme: dark\n---\n\nBody\n"
	fixed, report := Try(md, nil)

	require.Contains(t, fixed, `title: "Go: A Short Intro"`)
	found := false
	for _, f := range report.Fixes {
		if f.Rule == "fm-unquoted-colon" {
			found = true
			require.Equal(t, 2, f.Line)
		}
	}
	require.True(t, found)
}

func TestApplyFrontmatterUnquotedColon_LeavesQuotedAlone(t *testing.T) {
	md := "---\ntitle: \"Go: already quoted\"\n---\n"
	_, report := Try(md, nil)
	for _, f := range report.Fixes {
		require.NotEqual(t, "fm-unquoted-colon", f.Rule)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test -C . ./internal/generate/ -run TestApplyFrontmatterUnquotedColon -v`
Expected: FAIL.

- [ ] **Step 3: Implement the rule**

Replace the `applyFrontmatterUnquotedColon` stub:

```go
func applyFrontmatterUnquotedColon(lines *[]string, report *FixReport) {
	// Must start with frontmatter
	if len(*lines) == 0 || strings.TrimSpace((*lines)[0]) != "---" {
		return
	}
	for i := 1; i < len(*lines); i++ {
		t := (*lines)[i]
		if strings.TrimSpace(t) == "---" {
			return // end of frontmatter
		}
		// "key: value" — split on first colon
		colon := strings.Index(t, ":")
		if colon < 0 {
			continue
		}
		key := t[:colon]
		value := strings.TrimSpace(t[colon+1:])
		if value == "" {
			continue
		}
		// Already quoted?
		if (strings.HasPrefix(value, `"`) && strings.HasSuffix(value, `"`)) ||
			(strings.HasPrefix(value, `'`) && strings.HasSuffix(value, `'`)) {
			continue
		}
		// Value itself contains a colon → needs quoting
		if !strings.Contains(value, ":") {
			continue
		}
		(*lines)[i] = key + `: "` + strings.ReplaceAll(value, `"`, `\"`) + `"`
		report.Fixes = append(report.Fixes, Fix{
			Rule:        "fm-unquoted-colon",
			Line:        i + 1,
			Description: "frontmatter value contained ':' → added quotes",
		})
	}
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test -C . ./internal/generate/ -v`
Expected: all pass.

- [ ] **Step 5: Commit**

```bash
git add internal/generate/fixup.go internal/generate/fixup_test.go
git commit -m "feat(generate): fixup rule — quote frontmatter values with colons"
```

---

## Task 11: Fixup rule — trailing newline

**Files:**
- Modify: `internal/generate/fixup.go`
- Modify: `internal/generate/fixup_test.go`

- [ ] **Step 1: Write failing test**

Append to `internal/generate/fixup_test.go`:

```go
func TestApplyTrailingNewline_Adds(t *testing.T) {
	md := "# Heading\n\nBody"
	fixed, report := Try(md, nil)
	require.True(t, strings.HasSuffix(fixed, "\n"))
	found := false
	for _, f := range report.Fixes {
		if f.Rule == "trailing-newline" {
			found = true
		}
	}
	require.True(t, found)
}

func TestApplyTrailingNewline_NoopWhenPresent(t *testing.T) {
	md := "# Heading\n\nBody\n"
	_, report := Try(md, nil)
	for _, f := range report.Fixes {
		require.NotEqual(t, "trailing-newline", f.Rule)
	}
}
```

Add `"strings"` import if not already present in the test file.

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test -C . ./internal/generate/ -run TestApplyTrailingNewline -v`
Expected: FAIL.

- [ ] **Step 3: Implement the rule**

Replace the `applyTrailingNewline` stub:

```go
func applyTrailingNewline(lines *[]string, report *FixReport) {
	if len(*lines) == 0 {
		return
	}
	last := (*lines)[len(*lines)-1]
	if last == "" {
		return // final "" from trailing \n
	}
	*lines = append(*lines, "")
	report.Fixes = append(report.Fixes, Fix{
		Rule:        "trailing-newline",
		Line:        len(*lines),
		Description: "file did not end with newline → appended",
	})
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test -C . ./internal/generate/ -v`
Expected: all pass.

- [ ] **Step 5: Commit**

```bash
git add internal/generate/fixup.go internal/generate/fixup_test.go
git commit -m "feat(generate): fixup rule — ensure trailing newline"
```

---

## Task 12: Orchestrator `Run` — happy path

**Files:**
- Create: `internal/generate/generate.go`
- Create: `internal/generate/generate_test.go`
- Create: `internal/generate/testdata/responses/good.md`

- [ ] **Step 1: Create fixture**

Create `internal/generate/testdata/responses/good.md`:

```markdown
---
title: Welcome
theme: dark
---

# Welcome

This is slide one.

---

## Slide Two

- Bullet A
- Bullet B
```

- [ ] **Step 2: Write failing test**

Create `internal/generate/generate_test.go`:

```go
package generate

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// fakeCompleter is an in-memory Completer for testing.
type fakeCompleter struct {
	content string
	usage   Usage
	err     error
}

func (f *fakeCompleter) Complete(_ context.Context, _ string, _ []Message) (string, Usage, error) {
	return f.content, f.usage, f.err
}

func TestRun_HappyPath_WritesOutput(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "out.md")

	good, err := os.ReadFile("testdata/responses/good.md")
	require.NoError(t, err)

	opts := Options{
		BaseURL: "unused",
		Model:   "gpt-4o",
		APIKey:  "k",
		Input:   Input{Topic: "Intro"},
		Output:  outPath,
	}
	err = runWith(context.Background(), opts, &fakeCompleter{content: string(good)}, os.Stderr)
	require.NoError(t, err)

	written, err := os.ReadFile(outPath)
	require.NoError(t, err)
	require.Equal(t, string(good), string(written))
}

func TestRun_RefusesOverwriteWithoutForce(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "out.md")
	require.NoError(t, os.WriteFile(outPath, []byte("existing\n"), 0644))

	opts := Options{
		Model:  "m",
		APIKey: "k",
		Input:  Input{Topic: "t"},
		Output: outPath,
		Force:  false,
	}
	err := runWith(context.Background(), opts, &fakeCompleter{content: "# ok\n"}, os.Stderr)
	require.Error(t, err)
	require.Contains(t, err.Error(), "exists")

	// existing content preserved
	got, _ := os.ReadFile(outPath)
	require.Equal(t, "existing\n", string(got))
}

func TestRun_ForceOverwrites(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "out.md")
	require.NoError(t, os.WriteFile(outPath, []byte("existing\n"), 0644))

	good, _ := os.ReadFile("testdata/responses/good.md")

	opts := Options{
		Model: "m", APIKey: "k", Input: Input{Topic: "t"}, Output: outPath, Force: true,
	}
	err := runWith(context.Background(), opts, &fakeCompleter{content: string(good)}, os.Stderr)
	require.NoError(t, err)

	got, _ := os.ReadFile(outPath)
	require.Equal(t, string(good), string(got))
}
```

- [ ] **Step 3: Run tests to verify they fail**

Run: `go test -C . ./internal/generate/ -run TestRun -v`
Expected: FAIL (undefined: runWith).

- [ ] **Step 4: Implement**

Create `internal/generate/generate.go`:

```go
package generate

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/GMfatcat/goslide/internal/parser"
)

// Run executes the full generate pipeline using a real HTTP Client.
func Run(ctx context.Context, opts Options) error {
	client := NewClient(opts.BaseURL, opts.APIKey, opts.Timeout)
	return runWith(ctx, opts, client, os.Stderr)
}

// runWith is Run with the Completer and progress-output sink injected for
// testing.
func runWith(ctx context.Context, opts Options, llm Completer, stderr io.Writer) error {
	if !opts.Force {
		if _, err := os.Stat(opts.Output); err == nil {
			return fmt.Errorf("file %s exists; use --force to overwrite", opts.Output)
		}
	}

	userMsg, err := BuildUserMessage(opts.Input)
	if err != nil {
		return err
	}
	msgs := []Message{
		{Role: "system", Content: SystemPrompt()},
		{Role: "user", Content: userMsg},
	}

	fmt.Fprintf(stderr, "Generating… (model=%s)\n", opts.Model)
	content, usage, err := llm.Complete(ctx, opts.Model, msgs)
	if err != nil {
		return err
	}
	if usage.TotalTokens > 0 {
		fmt.Fprintf(stderr, "Tokens: prompt=%d completion=%d total=%d\n",
			usage.PromptTokens, usage.CompletionTokens, usage.TotalTokens)
	}

	final, err := sanitizeAndFix(content, stderr)
	if err != nil {
		// Write raw + fixed artefacts next to the output path for diagnosis.
		writeArtefact(opts.Output+".raw.md", content)
		writeArtefact(opts.Output+".fixed.md", final)
		return err
	}

	if err := os.WriteFile(opts.Output, []byte(final), 0644); err != nil {
		return fmt.Errorf("write output: %w", err)
	}
	fmt.Fprintf(stderr, "✓ Written to %s\n", opts.Output)
	return nil
}

// sanitizeAndFix parses, applies heuristic fixup if parsing failed, and
// returns the final markdown. It reports applied fixes on stderr. If the
// fixed markdown still fails to parse, the second parse error is returned.
func sanitizeAndFix(md string, stderr io.Writer) (string, error) {
	if _, err := parser.Parse([]byte(md)); err == nil {
		return md, nil
	} else {
		fixed, report := Try(md, err)
		if _, err2 := parser.Parse([]byte(fixed)); err2 != nil {
			fmt.Fprintf(stderr, "✗ Generated markdown could not be auto-fixed.\n")
			fmt.Fprintf(stderr, "  Original parse error: %v\n", err)
			fmt.Fprintf(stderr, "  Post-fix parse error: %v\n", err2)
			return fixed, fmt.Errorf("generated markdown failed to parse; see .raw.md and .fixed.md")
		}
		if len(report.Fixes) > 0 {
			fmt.Fprintf(stderr, "⚠ Generated markdown had %d issue(s); auto-fix applied:\n", len(report.Fixes))
			for _, f := range report.Fixes {
				fmt.Fprintf(stderr, "  [%s] line %d: %s\n", f.Rule, f.Line, f.Description)
			}
			fmt.Fprintf(stderr, "  Original parse error: %v\n", err)
		}
		return fixed, nil
	}
}

func writeArtefact(path, content string) {
	_ = os.WriteFile(path, []byte(content), 0644)
}
```

Note: this task assumes `parser.Parse([]byte) (..., error)` exists. Verify by running the tests; if the parser package exposes a different signature, adapt the call — the test assertions do not depend on parser internals.

- [ ] **Step 5: Verify parser signature**

Run: `go doc -C . ./internal/parser`
Look for the public `Parse` function. If it takes a different argument shape (e.g., `Parse(string)` or `Parse(io.Reader)`), adjust the two call sites in `sanitizeAndFix` accordingly before proceeding.

- [ ] **Step 6: Run tests to verify they pass**

Run: `go test -C . ./internal/generate/ -run TestRun -v`
Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add internal/generate/generate.go internal/generate/generate_test.go internal/generate/testdata/responses/good.md
git commit -m "feat(generate): orchestrator Run (happy path + overwrite guard)"
```

---

## Task 13: Orchestrator — parse-failure & fixup paths

**Files:**
- Modify: `internal/generate/generate_test.go`
- Create: `internal/generate/testdata/responses/unclosed_fence.md`
- Create: `internal/generate/testdata/responses/broken_unfixable.md`

- [ ] **Step 1: Create fixtures**

Create `internal/generate/testdata/responses/unclosed_fence.md`:

```markdown
---
title: Demo
theme: dark
---

# Demo

```go
func main() {
    fmt.Println("hi")
}
```

`internal/generate/testdata/responses/broken_unfixable.md`:

```markdown
this is not valid frontmatter or markdown syntax at all
```yaml
{ unbalanced: [braces
```

(Adjust this fixture after running Step 3 if `parser.Parse` happens to accept
it — the goal is any content that neither parses cleanly nor is fixable by
the five heuristic rules.)

- [ ] **Step 2: Write failing tests**

Append to `internal/generate/generate_test.go`:

```go
func TestRun_FixupRecovers(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "out.md")

	broken, _ := os.ReadFile("testdata/responses/unclosed_fence.md")

	opts := Options{Model: "m", APIKey: "k", Input: Input{Topic: "t"}, Output: outPath}
	var stderr strings.Builder
	err := runWith(context.Background(), opts, &fakeCompleter{content: string(broken)}, &stderr)
	require.NoError(t, err)

	_, statErr := os.Stat(outPath)
	require.NoError(t, statErr)
	require.Contains(t, stderr.String(), "fence-close")
}

func TestRun_UnrecoverableWritesArtefacts(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "out.md")

	garbage, _ := os.ReadFile("testdata/responses/broken_unfixable.md")

	opts := Options{Model: "m", APIKey: "k", Input: Input{Topic: "t"}, Output: outPath}
	var stderr strings.Builder
	err := runWith(context.Background(), opts, &fakeCompleter{content: string(garbage)}, &stderr)
	require.Error(t, err)

	_, rawErr := os.Stat(outPath + ".raw.md")
	require.NoError(t, rawErr, ".raw.md should have been written")
	_, fixedErr := os.Stat(outPath + ".fixed.md")
	require.NoError(t, fixedErr, ".fixed.md should have been written")

	// primary output should NOT exist
	_, outErr := os.Stat(outPath)
	require.True(t, os.IsNotExist(outErr))
}
```

Add `"strings"` import.

- [ ] **Step 3: Run tests**

Run: `go test -C . ./internal/generate/ -v`

If `TestRun_FixupRecovers` fails because the happy-path fixture actually parses, the unclosed-fence fixture is correct; proceed.

If `TestRun_UnrecoverableWritesArtefacts` fails because `parser.Parse` accepts the garbage content, adjust the fixture to something `parser` definitely rejects (e.g., a fenced block with mismatched backtick count that fixup can't repair, or binary bytes). Rerun until the test fails meaningfully, then passes.

- [ ] **Step 4: Confirm all pass**

Run: `go test -C . ./internal/generate/ -v`
Expected: all tests pass.

- [ ] **Step 5: Commit**

```bash
git add internal/generate/generate_test.go internal/generate/testdata/responses/unclosed_fence.md internal/generate/testdata/responses/broken_unfixable.md
git commit -m "test(generate): cover fixup recovery + unrecoverable paths"
```

---

## Task 14: CLI command — `goslide generate`

**Files:**
- Create: `internal/cli/generate.go`
- Create: `internal/cli/generate_test.go`

- [ ] **Step 1: Write failing tests**

Create `internal/cli/generate_test.go`:

```go
package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenerate_DumpPrompt(t *testing.T) {
	cmd := newGenerateCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--dump-prompt"})

	err := cmd.Execute()
	require.NoError(t, err)
	require.Contains(t, out.String(), "layout:")
	require.Contains(t, out.String(), "theme:")
}

func TestGenerate_NoArgsNoDumpFails(t *testing.T) {
	cmd := newGenerateCmd()
	cmd.SetArgs([]string{})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	err := cmd.Execute()
	require.Error(t, err)
	require.Contains(t, strings.ToLower(err.Error()), "argument")
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test -C . ./internal/cli/ -run TestGenerate -v`
Expected: FAIL (undefined: newGenerateCmd).

- [ ] **Step 3: Implement**

Create `internal/cli/generate.go`:

```go
package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/GMfatcat/goslide/internal/config"
	"github.com/GMfatcat/goslide/internal/generate"
	"github.com/spf13/cobra"
)

func newGenerateCmd() *cobra.Command {
	var (
		output      string
		force       bool
		model       string
		baseURL     string
		apiKeyEnv   string
		dumpPrompt  bool
	)

	cmd := &cobra.Command{
		Use:   "generate [topic | prompt.md]",
		Short: "Generate a GoSlide presentation via an LLM",
		Long: "Call an OpenAI-compatible LLM endpoint to produce a GoSlide Markdown file.\n" +
			"Pass a topic string for simple mode, or a path to a prompt.md file for advanced mode.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if dumpPrompt {
				fmt.Fprint(cmd.OutOrStdout(), generate.SystemPrompt())
				return nil
			}
			if len(args) != 1 {
				return fmt.Errorf("exactly one argument required (topic or prompt.md path)")
			}

			in, err := inputFromArg(args[0])
			if err != nil {
				return err
			}

			cfg, err := config.Load(".")
			if err != nil {
				return err
			}

			opts, err := generate.Resolve(cfg, generate.Flags{
				BaseURL:   baseURL,
				Model:     model,
				APIKeyEnv: apiKeyEnv,
				Output:    output,
				Force:     force,
			}, in)
			if err != nil {
				return err
			}

			return generate.Run(context.Background(), opts)
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "", "output path (default talk.md)")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "overwrite existing output file")
	cmd.Flags().StringVar(&model, "model", "", "override generate.model from goslide.yaml")
	cmd.Flags().StringVar(&baseURL, "base-url", "", "override generate.base_url from goslide.yaml")
	cmd.Flags().StringVar(&apiKeyEnv, "api-key-env", "", "env var name holding the API key (default OPENAI_API_KEY)")
	cmd.Flags().BoolVar(&dumpPrompt, "dump-prompt", false, "print the built-in system prompt and exit")

	return cmd
}

// inputFromArg decides simple vs advanced mode: if arg resolves to an
// existing file, parse it; otherwise treat it as a topic string.
func inputFromArg(arg string) (generate.Input, error) {
	if fi, err := os.Stat(arg); err == nil && !fi.IsDir() {
		return generate.ParsePromptFile(arg)
	}
	return generate.Input{Topic: arg}, nil
}

func init() {
	rootCmd.AddCommand(newGenerateCmd())
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test -C . ./internal/cli/ -v`
Expected: all pass.

- [ ] **Step 5: Build the binary end-to-end**

Run: `go build -C . ./cmd/goslide`
Expected: success, produces `goslide` (or `goslide.exe`) in cwd.

Run: `./goslide generate --dump-prompt | head -20`
Expected: prints the embedded system prompt.

- [ ] **Step 6: Commit**

```bash
git add internal/cli/generate.go internal/cli/generate_test.go
git commit -m "feat(cli): add goslide generate command"
```

---

## Task 15: Update `PRD.md` and `README.md`

**Files:**
- Modify: `PRD.md`
- Modify: `README.md`

- [ ] **Step 1: Update PRD Phase 5 checklist**

Open `PRD.md`. Find the line `- [ ] \`goslide generate\` command (LLM-powered slide generation)` and change `[ ]` to `[x]`.

- [ ] **Step 2: Add a README section**

Insert a new subsection in `README.md` (place it after the section that documents `goslide init`; if unclear, append before the "License" or "Contributing" section):

````markdown
### AI slide generation

Generate a full presentation from a topic using any OpenAI-compatible LLM
endpoint (OpenAI, OpenRouter, Ollama, vllm, sglang, etc.).

Add a `generate:` section to `goslide.yaml`:

```yaml
generate:
  base_url: https://api.openai.com/v1
  model: gpt-4o
  api_key_env: OPENAI_API_KEY
  timeout: 120s
```

Export the API key, then run:

```bash
export OPENAI_API_KEY=sk-...
goslide generate "Introduction to Kubernetes"            # simple mode
goslide generate my-prompt.md -o talk.md                 # advanced mode
goslide generate --dump-prompt > system.txt              # inspect prompt
```

Advanced mode reads a `prompt.md` file:

```markdown
---
topic: Kubernetes Architecture
audience: Backend engineers
slides: 15
theme: dark
language: en
---
Emphasize Pod/Service/Ingress. End with a Q&A slide.
```

The command refuses to overwrite an existing output file unless `--force` is
passed. Generated Markdown is sanity-checked against the parser; common
issues (unclosed code fences, missing frontmatter terminator) are auto-fixed
with a transparent report.
````

- [ ] **Step 3: Commit**

```bash
git add PRD.md README.md
git commit -m "docs: document goslide generate in README and mark PRD checkbox"
```

---

## Task 16: Final verification

**Files:** none (verification only)

- [ ] **Step 1: Run the full test suite**

Run: `go test -C . ./...`
Expected: PASS across all packages.

- [ ] **Step 2: Build the binary**

Run: `go build -C . ./cmd/goslide`
Expected: success.

- [ ] **Step 3: Smoke test `--dump-prompt`**

Run: `./goslide generate --dump-prompt`
Expected: prints the system prompt to stdout.

- [ ] **Step 4: Smoke test missing config**

In an empty temp dir with no `goslide.yaml`:
```bash
cd $(mktemp -d)
/path/to/goslide generate "hello"
```
Expected: exits non-zero with a clear message about missing `base_url` or `model`.

- [ ] **Step 5: Smoke test missing API key**

Create a `goslide.yaml` with a `generate:` section using `api_key_env: NOPE_12345`, unset that env var, run `goslide generate "hi"`.
Expected: exits non-zero with message naming `NOPE_12345`.

- [ ] **Step 6: Manual quality check (optional, not in CI)**

Set `OPENAI_API_KEY` to a real key (or point `base_url` at a local Ollama with a model pulled). Run:
```bash
goslide generate "Introduction to CAP theorem" -o /tmp/cap.md
goslide serve /tmp/cap.md
```
Open the browser and verify the deck looks reasonable. This is human review — not automated.

- [ ] **Step 7: No commit needed**

If all previous steps passed, implementation is complete. The commits from Tasks 1-15 are the deliverable.

---

## Success Criteria (from spec §10)

- ✅ `goslide generate "topic"` writes a valid `talk.md` (Task 12 + 16.6 manual).
- ✅ `goslide generate prompt.md` honours all frontmatter fields (Task 4).
- ✅ `--dump-prompt` output is usable standalone (Task 2 + 14 + 16.3).
- ✅ Five fixup rules unit-tested (Tasks 8-11).
- ✅ All unit + integration tests pass; no real LLM in CI (Task 16.1).
