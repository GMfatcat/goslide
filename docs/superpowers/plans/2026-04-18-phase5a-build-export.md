# Phase 5a: Static HTML Export Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add `goslide build <file.md>` that exports a self-contained single HTML file with all CSS, JS, and fonts inlined, runnable offline without a server.

**Architecture:** New `internal/builder` package reads the .md, runs the parse→validate→render pipeline, then replaces all `<link href>` and `<script src>` references with inline content from `go:embed`. Font URLs in CSS become base64 data URIs. A `data-mode="static"` flag on `<body>` tells runtime.js and components.js to skip WebSocket and API features.

**Tech Stack:** Go 1.21.6, existing parser/renderer, `encoding/base64`, `io/fs`. No new dependencies.

**Shell rules (Windows):** Never chain commands with `&&`. Use SEPARATE Bash calls for `git add`, `git commit`, `go test`, `go build`. Use `GOTOOLCHAIN=local` prefix for ALL go commands. Use `-C` flag for go/git to specify directory.

---

## Task 1: Static Mode JS Changes

**Files:**
- Modify: `web/static/runtime.js`
- Modify: `web/static/components.js`

- [ ] **Step 1: Update runtime.js — add isStatic guard**

Read `web/static/runtime.js`. Add `var isStatic = document.body.dataset.mode === 'static';` after `'use strict';`.

Find the WebSocket section starting with:
```javascript
  var toast = document.getElementById('goslide-toast');
  var proto = location.protocol === 'https:' ? 'wss:' : 'ws:';
  var ws = new WebSocket(proto + '//' + location.host + '/ws');
```

Wrap the WS creation and ALL WS event listeners (message, close) plus the presenter tracking section in `if (!isStatic) {`. Close the brace before the page number section. The toast div declaration stays outside.

The structure becomes:
```javascript
  var toast = document.getElementById('goslide-toast');

  if (!isStatic) {
    var proto = ...;
    var ws = new WebSocket(...);
    ws.addEventListener('message', ...);
    ws.addEventListener('close', ...);

    // Presenter slide tracking
    var isPresenter = ...;
    if (isPresenter) { ... }

    // Viewer: show presenter indicator
    var presenterIndicator = null;
    ws.addEventListener('message', ...);
  }

  // Page number indicator (always active)
  var pageNumEl = ...;
```

- [ ] **Step 2: Update components.js — add isStatic guard**

Read `web/static/components.js`. Add `var isStatic = document.body.dataset.mode === 'static';` after `'use strict';`.

Find the api component branch in `initAllComponents`:
```javascript
      else if (type === 'api') initApiComponent(el);
```

Replace with:
```javascript
      else if (type === 'api') {
        if (isStatic) {
          el.innerHTML = '<div style="color:var(--slide-muted);font-size:0.75em;text-align:center;padding:1rem;">API data requires goslide serve</div>';
        } else {
          initApiComponent(el);
        }
        initialized[id] = true;
        return;
      }
```

Find the polling slidechanged handler at the bottom (the `Reveal.on('slidechanged'` that manages api polling). Wrap it in `if (!isStatic) { ... }`.

- [ ] **Step 3: Verify build**

Run: `GOTOOLCHAIN=local go build -C D:/CLAUDE-CODE-GOSLIDE ./...`

- [ ] **Step 4: Verify serve mode still works**

Run: `GOTOOLCHAIN=local go test -C D:/CLAUDE-CODE-GOSLIDE ./internal/... -count=1`

- [ ] **Step 5: Commit**

```
git -C D:/CLAUDE-CODE-GOSLIDE add web/static/runtime.js web/static/components.js
```
```
git -C D:/CLAUDE-CODE-GOSLIDE commit -m "feat: add static mode guards in runtime.js and components.js for build export"
```

---

## Task 2: Builder Package

**Files:**
- Create: `internal/builder/builder.go`
- Create: `internal/builder/builder_test.go`

- [ ] **Step 1: Write failing tests**

Create `internal/builder/builder_test.go`:

```go
package builder

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuild_BasicOutput(t *testing.T) {
	dir := t.TempDir()
	mdPath := filepath.Join(dir, "test.md")
	os.WriteFile(mdPath, []byte("---\ntitle: Test\ntheme: default\n---\n\n# Slide 1\n\nHello world.\n"), 0644)

	outPath := filepath.Join(dir, "test.html")
	err := Build(Options{File: mdPath, Output: outPath})
	require.NoError(t, err)

	data, err := os.ReadFile(outPath)
	require.NoError(t, err)
	html := string(data)

	require.Contains(t, html, "Slide 1")
	require.Contains(t, html, "Hello world")
	require.Contains(t, html, `data-mode="static"`)
}

func TestBuild_NoExternalRefs(t *testing.T) {
	dir := t.TempDir()
	mdPath := filepath.Join(dir, "test.md")
	os.WriteFile(mdPath, []byte("---\ntitle: T\n---\n\n# S\n"), 0644)

	outPath := filepath.Join(dir, "out.html")
	err := Build(Options{File: mdPath, Output: outPath})
	require.NoError(t, err)

	data, _ := os.ReadFile(outPath)
	html := string(data)

	require.NotContains(t, html, `href="/themes/`)
	require.NotContains(t, html, `href="/static/`)
	require.NotContains(t, html, `src="/static/`)
}

func TestBuild_FontsInlined(t *testing.T) {
	dir := t.TempDir()
	mdPath := filepath.Join(dir, "test.md")
	os.WriteFile(mdPath, []byte("# S\n"), 0644)

	outPath := filepath.Join(dir, "out.html")
	err := Build(Options{File: mdPath, Output: outPath})
	require.NoError(t, err)

	data, _ := os.ReadFile(outPath)
	html := string(data)

	require.Contains(t, html, "data:font/woff2;base64,")
}

func TestBuild_RevealPresent(t *testing.T) {
	dir := t.TempDir()
	mdPath := filepath.Join(dir, "test.md")
	os.WriteFile(mdPath, []byte("# S\n"), 0644)

	outPath := filepath.Join(dir, "out.html")
	Build(Options{File: mdPath, Output: outPath})

	data, _ := os.ReadFile(outPath)
	require.Contains(t, string(data), "Reveal")
}

func TestBuild_DefaultOutputName(t *testing.T) {
	dir := t.TempDir()
	mdPath := filepath.Join(dir, "my-talk.md")
	os.WriteFile(mdPath, []byte("# S\n"), 0644)

	err := Build(Options{File: mdPath})
	require.NoError(t, err)

	outPath := filepath.Join(".", "my-talk.html")
	defer os.Remove(outPath)

	_, err = os.Stat(outPath)
	require.NoError(t, err)
}
```

- [ ] **Step 2: Implement builder.go**

Create `internal/builder/builder.go`:

```go
package builder

import (
	"encoding/base64"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/user/goslide/internal/ir"
	"github.com/user/goslide/internal/parser"
	"github.com/user/goslide/internal/renderer"
	"github.com/user/goslide/web"
)

type Options struct {
	File   string
	Output string
}

func Build(opts Options) error {
	data, err := os.ReadFile(opts.File)
	if err != nil {
		return fmt.Errorf("read %s: %w", opts.File, err)
	}

	pres, err := parser.Parse(data, opts.File)
	if err != nil {
		return fmt.Errorf("parse %s: %w", opts.File, err)
	}

	valErrs := pres.Validate()
	if len(valErrs) > 0 {
		fmt.Fprint(os.Stderr, ir.FormatErrors(opts.File, valErrs))
	}
	if ir.HasErrors(valErrs) {
		return fmt.Errorf("validation failed for %s", opts.File)
	}

	html, err := renderer.Render(pres)
	if err != nil {
		return fmt.Errorf("render %s: %w", opts.File, err)
	}

	html = inlineAssets(html)
	html = addStaticMode(html)

	output := opts.Output
	if output == "" {
		base := strings.TrimSuffix(filepath.Base(opts.File), ".md")
		output = base + ".html"
	}

	if err := os.WriteFile(output, []byte(html), 0644); err != nil {
		return fmt.Errorf("write %s: %w", output, err)
	}

	fmt.Printf("Built %s → %s\n", opts.File, output)
	return nil
}

func addStaticMode(html string) string {
	return strings.Replace(html, "<body ", "<body data-mode=\"static\" ", 1)
}

func inlineAssets(html string) string {
	cssFiles := []struct {
		href string
		fsys fs.FS
		path string
	}{
		{`href="/themes/tokens.css"`, web.ThemeFS, "themes/tokens.css"},
		{`href="/themes/layouts.css"`, web.ThemeFS, "themes/layouts.css"},
	}

	themeNames := []string{"default", "dark", "corporate", "minimal", "hacker"}
	for _, t := range themeNames {
		cssFiles = append(cssFiles, struct {
			href string
			fsys fs.FS
			path string
		}{
			fmt.Sprintf(`href="/themes/%s.css"`, t),
			web.ThemeFS,
			fmt.Sprintf("themes/%s.css", t),
		})
	}

	for _, cf := range cssFiles {
		content, err := fs.ReadFile(cf.fsys, cf.path)
		if err != nil {
			continue
		}
		css := inlineFonts(string(content))
		old := fmt.Sprintf(`<link rel="stylesheet" %s>`, cf.href)
		html = strings.Replace(html, old, "<style>"+css+"</style>", 1)
	}

	jsFiles := []struct {
		src  string
		fsys fs.FS
		path string
	}{
		{`src="/static/chartjs/chart.min.js"`, web.StaticFS, "static/chartjs/chart.min.js"},
		{`src="/static/mermaid/mermaid.min.js"`, web.StaticFS, "static/mermaid/mermaid.min.js"},
		{`src="/static/reveal/reveal.js"`, web.StaticFS, "static/reveal/reveal.js"},
		{`src="/static/runtime.js"`, web.StaticFS, "static/runtime.js"},
		{`src="/static/reactive.js"`, web.StaticFS, "static/reactive.js"},
		{`src="/static/components.js"`, web.StaticFS, "static/components.js"},
	}

	for _, jf := range jsFiles {
		content, err := fs.ReadFile(jf.fsys, jf.path)
		if err != nil {
			continue
		}
		old := fmt.Sprintf(`<script %s></script>`, jf.src)
		html = strings.Replace(html, old, "<script>"+string(content)+"</script>", 1)
	}

	return html
}

func inlineFonts(css string) string {
	fonts := []struct {
		url  string
		path string
	}{
		{"url('/fonts/NotoSansTC-Regular.woff2')", "static/fonts/NotoSansTC-Regular.woff2"},
		{"url('/fonts/NotoSansTC-Bold.woff2')", "static/fonts/NotoSansTC-Bold.woff2"},
		{"url('/fonts/JetBrainsMono-Regular.woff2')", "static/fonts/JetBrainsMono-Regular.woff2"},
	}

	for _, f := range fonts {
		data, err := fs.ReadFile(web.StaticFS, f.path)
		if err != nil {
			continue
		}
		b64 := base64.StdEncoding.EncodeToString(data)
		dataURI := fmt.Sprintf("url('data:font/woff2;base64,%s')", b64)
		css = strings.Replace(css, f.url, dataURI, 1)
	}

	return css
}
```

- [ ] **Step 3: Run tests**

Run: `GOTOOLCHAIN=local go test -C D:/CLAUDE-CODE-GOSLIDE/internal/builder -v`

Run: `GOTOOLCHAIN=local go test -C D:/CLAUDE-CODE-GOSLIDE ./internal/... -count=1`

- [ ] **Step 4: Commit**

```
git -C D:/CLAUDE-CODE-GOSLIDE add internal/builder/builder.go internal/builder/builder_test.go
```
```
git -C D:/CLAUDE-CODE-GOSLIDE commit -m "feat(builder): add static HTML export with inline assets and base64 fonts"
```

---

## Task 3: CLI Build Command + Checklist

**Files:**
- Create: `internal/cli/build.go`
- Modify: `internal/cli/root.go`
- Modify: `MANUAL_CHECKLIST.md`

- [ ] **Step 1: Create build.go**

Create `internal/cli/build.go`:

```go
package cli

import (
	"github.com/spf13/cobra"
	"github.com/user/goslide/internal/builder"
)

var buildOutput string

var buildCmd = &cobra.Command{
	Use:   "build <file.md>",
	Short: "Export presentation as self-contained HTML",
	Args:  cobra.ExactArgs(1),
	RunE:  runBuild,
}

func init() {
	buildCmd.Flags().StringVarP(&buildOutput, "output", "o", "", "output file (default: {name}.html)")
	rootCmd.AddCommand(buildCmd)
}

func runBuild(cmd *cobra.Command, args []string) error {
	return builder.Build(builder.Options{
		File:   args[0],
		Output: buildOutput,
	})
}
```

- [ ] **Step 2: Remove build stub from root.go**

Read `internal/cli/root.go`. The stubs slice should now only have `build`. Remove the entire stubs block (the `stubs` variable declaration + the loop that adds them). If there are no stubs left, remove the whole block.

- [ ] **Step 3: Update MANUAL_CHECKLIST.md**

Read `MANUAL_CHECKLIST.md`. Append:

```markdown

## Static Export (Phase 5a)
- [ ] `goslide build examples/demo.md` produces demo.html
- [ ] demo.html opens in browser offline (no server needed)
- [ ] All slides render correctly
- [ ] Charts display with correct data
- [ ] Mermaid diagrams render as SVG
- [ ] Tables are sortable
- [ ] Tabs/slider/toggle work
- [ ] Card overlay works
- [ ] Fragments work
- [ ] Page numbers display
- [ ] No WS connection attempt (no console errors)
- [ ] API slides show "requires goslide serve" message
- [ ] Keyboard navigation works
```

- [ ] **Step 4: Verify build + test**

Run: `GOTOOLCHAIN=local go build -C D:/CLAUDE-CODE-GOSLIDE -o D:/CLAUDE-CODE-GOSLIDE/goslide.exe ./cmd/goslide`

Run: `D:/CLAUDE-CODE-GOSLIDE/goslide.exe build --help`

Run: `GOTOOLCHAIN=local go test -C D:/CLAUDE-CODE-GOSLIDE ./... -count=1 -race`

- [ ] **Step 5: Commit**

```
git -C D:/CLAUDE-CODE-GOSLIDE add internal/cli/build.go internal/cli/root.go MANUAL_CHECKLIST.md
```
```
git -C D:/CLAUDE-CODE-GOSLIDE commit -m "feat(cli): add build command for static HTML export"
```
