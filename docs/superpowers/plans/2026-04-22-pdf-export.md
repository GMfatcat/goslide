# PDF Export Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a `goslide export-pdf <file.md>` command that builds the deck to static HTML, launches a locally-installed Chrome/Edge/Chromium in headless mode, and uses Chrome DevTools Protocol `Page.printToPDF` to produce a PDF. Chrome-or-fail: no bundled chromium, no auto-download.

**Architecture:** New `internal/pdfexport` package with three units — `papersize` (preset table), `chrome` (Chrome binary discovery with env-var / PATH / platform-path fallback), and `export` (orchestrator that calls `builder.Build` for the HTML, then drives chromedp). A small front-end marker (`window.__goslideReady`) lets the headless browser know when async component rendering (Mermaid, charts) has settled. New `goslide export-pdf` CLI command.

**Tech Stack:** Go 1.21.6; new dependency `github.com/chromedp/chromedp` (CDP client); no front-end additions beyond the ready marker.

**Spec:** `docs/superpowers/specs/2026-04-22-pdf-export-design.md`

---

## File Structure

**Create:**
- `internal/pdfexport/papersize.go` — preset table + resolution
- `internal/pdfexport/papersize_test.go`
- `internal/pdfexport/chrome.go` — `FindChrome()` with DI hooks for testing
- `internal/pdfexport/chrome_test.go`
- `internal/pdfexport/export.go` — `Export(Options) error` orchestrator; `Launcher` interface
- `internal/pdfexport/export_test.go` — unit (fake launcher) + integration (real Chrome, skipped when absent)
- `internal/pdfexport/testdata/fixture.md` — small deck for integration tests
- `internal/cli/export_pdf.go` — Cobra command

**Modify:**
- `go.mod` / `go.sum` — add `github.com/chromedp/chromedp`
- `web/static/components.js` — set `window.__goslideReady = true` when init loop + mermaid promises settle
- `README.md`, `README_zh-TW.md` — new "PDF export" subsection
- `PRD.md` — tick checkbox

**No changes to:**
- Existing `internal/builder/` — we call `Build()` as-is
- Other `internal/server` routes — export-pdf has no runtime component

---

## Task 1: Add chromedp dependency

**Files:**
- Modify: `go.mod`, `go.sum`

- [ ] **Step 1: Add chromedp**

Run: `GOTOOLCHAIN=local go get -C D:/CLAUDE-CODE-GOSLIDE github.com/chromedp/chromedp@latest`

Expected: `go.mod` picks up a new `require github.com/chromedp/chromedp vX.Y.Z` line; `go.sum` updates.

- [ ] **Step 2: Verify tidy**

Run: `GOTOOLCHAIN=local go mod tidy -C D:/CLAUDE-CODE-GOSLIDE`
Run: `GOTOOLCHAIN=local go build -C D:/CLAUDE-CODE-GOSLIDE ./...`
Expected: success.

- [ ] **Step 3: Commit**

```bash
git -C D:/CLAUDE-CODE-GOSLIDE add go.mod go.sum
```

```bash
git -C D:/CLAUDE-CODE-GOSLIDE commit -m "chore: add github.com/chromedp/chromedp dependency"
```

---

## Task 2: papersize presets

**Files:**
- Create: `internal/pdfexport/papersize.go`
- Create: `internal/pdfexport/papersize_test.go`

- [ ] **Step 1: Write failing test**

Create `internal/pdfexport/papersize_test.go`:

```go
package pdfexport

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolvePaperSize_KnownPresets(t *testing.T) {
	cases := []struct {
		name   string
		wantW  float64
		wantH  float64
		wantOK bool
	}{
		{"slide-16x9", 20.0, 11.25, true},         // 1920 / 96 = 20 in, 1080 / 96 = 11.25 in
		{"slide-4x3", 16.667, 12.5, true},         // 1600/96=16.666, 1200/96=12.5
		{"a4-landscape", 11.693, 8.268, true},     // 297mm/25.4 = 11.693 in, 210mm/25.4 = 8.267
		{"letter-landscape", 11.0, 8.5, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			w, h, err := ResolvePaperSize(c.name)
			require.NoError(t, err)
			require.InDelta(t, c.wantW, w, 0.01)
			require.InDelta(t, c.wantH, h, 0.01)
		})
	}
}

func TestResolvePaperSize_UnknownError(t *testing.T) {
	_, _, err := ResolvePaperSize("foobar")
	require.Error(t, err)
	require.Contains(t, err.Error(), "unknown paper size")
	require.Contains(t, err.Error(), "slide-16x9")
}

func TestKnownPaperSizes(t *testing.T) {
	names := KnownPaperSizes()
	require.ElementsMatch(t, []string{"slide-16x9", "slide-4x3", "a4-landscape", "letter-landscape"}, names)
}
```

- [ ] **Step 2: Run — verify fail**

Run: `GOTOOLCHAIN=local go test -C D:/CLAUDE-CODE-GOSLIDE ./internal/pdfexport/ -v`
Expected: compile error (package doesn't exist).

- [ ] **Step 3: Implement**

Create `internal/pdfexport/papersize.go`:

```go
package pdfexport

import (
	"fmt"
	"sort"
	"strings"
)

// paperSizes maps a preset name to its width and height in inches.
// chromedp's PrintToPDFParams takes inches, so we normalise here.
var paperSizes = map[string]struct {
	widthIn, heightIn float64
}{
	"slide-16x9":       {1920.0 / 96.0, 1080.0 / 96.0},
	"slide-4x3":        {1600.0 / 96.0, 1200.0 / 96.0},
	"a4-landscape":     {297.0 / 25.4, 210.0 / 25.4},
	"letter-landscape": {11.0, 8.5},
}

// ResolvePaperSize returns width and height in inches for a known preset.
func ResolvePaperSize(name string) (float64, float64, error) {
	p, ok := paperSizes[name]
	if !ok {
		return 0, 0, fmt.Errorf("unknown paper size %q (valid: %s)", name, strings.Join(KnownPaperSizes(), ", "))
	}
	return p.widthIn, p.heightIn, nil
}

// KnownPaperSizes returns the list of valid preset names, sorted.
func KnownPaperSizes() []string {
	out := make([]string, 0, len(paperSizes))
	for k := range paperSizes {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
```

- [ ] **Step 4: Verify pass + gofmt**

Run: `GOTOOLCHAIN=local go test -C D:/CLAUDE-CODE-GOSLIDE ./internal/pdfexport/ -v`
Expected: 3 tests pass.

Run: `GOTOOLCHAIN=local gofmt -w D:/CLAUDE-CODE-GOSLIDE/internal/pdfexport/papersize.go`
Run: `GOTOOLCHAIN=local gofmt -w D:/CLAUDE-CODE-GOSLIDE/internal/pdfexport/papersize_test.go`
Run: `GOTOOLCHAIN=local gofmt -l D:/CLAUDE-CODE-GOSLIDE/internal/pdfexport/`
Expected: empty.

- [ ] **Step 5: Commit**

```bash
git -C D:/CLAUDE-CODE-GOSLIDE add internal/pdfexport/papersize.go internal/pdfexport/papersize_test.go
```

```bash
git -C D:/CLAUDE-CODE-GOSLIDE commit -m "feat(pdfexport): paper size presets"
```

---

## Task 3: Chrome discovery

**Files:**
- Create: `internal/pdfexport/chrome.go`
- Create: `internal/pdfexport/chrome_test.go`

- [ ] **Step 1: Write failing tests**

Create `internal/pdfexport/chrome_test.go`:

```go
package pdfexport

import (
	"errors"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFindChrome_EnvOverrideWins(t *testing.T) {
	t.Setenv("GOSLIDE_CHROME_PATH", "/fake/chrome")
	finder := &chromeFinder{
		fileExists: func(path string) bool { return path == "/fake/chrome" },
		lookPath:   func(string) (string, error) { return "", errors.New("not found") },
	}
	got, err := finder.find()
	require.NoError(t, err)
	require.Equal(t, "/fake/chrome", got)
}

func TestFindChrome_EnvOverrideMissingErrors(t *testing.T) {
	t.Setenv("GOSLIDE_CHROME_PATH", "/nonexistent/chrome")
	finder := &chromeFinder{
		fileExists: func(string) bool { return false },
		lookPath:   func(string) (string, error) { return "", errors.New("not found") },
	}
	_, err := finder.find()
	require.Error(t, err)
	require.Contains(t, err.Error(), "GOSLIDE_CHROME_PATH")
}

func TestFindChrome_PATHSearch(t *testing.T) {
	t.Setenv("GOSLIDE_CHROME_PATH", "")
	finder := &chromeFinder{
		fileExists: func(string) bool { return false },
		lookPath: func(name string) (string, error) {
			if name == "chromium" {
				return "/usr/bin/chromium", nil
			}
			return "", errors.New("not found")
		},
	}
	got, err := finder.find()
	require.NoError(t, err)
	require.Equal(t, "/usr/bin/chromium", got)
}

func TestFindChrome_PlatformPaths(t *testing.T) {
	t.Setenv("GOSLIDE_CHROME_PATH", "")
	var probed []string
	finder := &chromeFinder{
		fileExists: func(path string) bool {
			probed = append(probed, path)
			// simulate Chrome installed at the first platform path only
			return false
		},
		lookPath: func(string) (string, error) { return "", errors.New("not found") },
	}
	_, err := finder.find()
	require.Error(t, err)
	require.Contains(t, err.Error(), "Chrome/Edge/Chromium")
	// confirm at least one platform path was probed
	require.NotEmpty(t, probed)
	// all probed paths were OS-specific — at minimum, none empty
	for _, p := range probed {
		require.NotEmpty(t, p)
	}
}

func TestFindChrome_PlatformPathMatch(t *testing.T) {
	t.Setenv("GOSLIDE_CHROME_PATH", "")
	// Simulate a platform path that exists. Pick based on runtime.GOOS.
	var target string
	switch runtime.GOOS {
	case "windows":
		target = `C:\Program Files\Google\Chrome\Application\chrome.exe`
	case "darwin":
		target = "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
	default:
		t.Skip("platform paths test only meaningful on windows/darwin; linux uses PATH")
	}
	finder := &chromeFinder{
		fileExists: func(path string) bool { return path == target },
		lookPath:   func(string) (string, error) { return "", errors.New("not found") },
	}
	got, err := finder.find()
	require.NoError(t, err)
	require.Equal(t, target, got)
}
```

- [ ] **Step 2: Run — verify fail**

Run: `GOTOOLCHAIN=local go test -C D:/CLAUDE-CODE-GOSLIDE ./internal/pdfexport/ -run TestFindChrome -v`
Expected: FAIL (undefined: chromeFinder).

- [ ] **Step 3: Implement**

Create `internal/pdfexport/chrome.go`:

```go
package pdfexport

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// FindChrome locates a Chrome/Edge/Chromium binary. Search order:
//   1. GOSLIDE_CHROME_PATH env var (user override)
//   2. PATH (via exec.LookPath) for common names
//   3. Platform-specific known install locations
func FindChrome() (string, error) {
	return defaultFinder.find()
}

var defaultFinder = &chromeFinder{
	fileExists: func(path string) bool {
		_, err := os.Stat(path)
		return err == nil
	},
	lookPath: exec.LookPath,
}

type chromeFinder struct {
	fileExists func(path string) bool
	lookPath   func(name string) (string, error)
}

func (f *chromeFinder) find() (string, error) {
	// 1. Env override
	if env := os.Getenv("GOSLIDE_CHROME_PATH"); env != "" {
		if f.fileExists(env) {
			return env, nil
		}
		return "", fmt.Errorf("GOSLIDE_CHROME_PATH=%s but the file does not exist", env)
	}

	// 2. PATH search
	for _, name := range []string{"chrome", "chromium", "chromium-browser", "google-chrome", "microsoft-edge"} {
		if path, err := f.lookPath(name); err == nil {
			return path, nil
		}
	}

	// 3. Platform paths
	candidates := platformPaths()
	for _, p := range candidates {
		if f.fileExists(p) {
			return p, nil
		}
	}

	return "", notFoundError(candidates)
}

func platformPaths() []string {
	switch runtime.GOOS {
	case "windows":
		local := os.Getenv("LOCALAPPDATA")
		paths := []string{
			`C:\Program Files\Google\Chrome\Application\chrome.exe`,
			`C:\Program Files (x86)\Google\Chrome\Application\chrome.exe`,
			`C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe`,
			`C:\Program Files\Microsoft\Edge\Application\msedge.exe`,
		}
		if local != "" {
			paths = append(paths, local+`\Google\Chrome\Application\chrome.exe`)
		}
		return paths
	case "darwin":
		return []string{
			"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
			"/Applications/Chromium.app/Contents/MacOS/Chromium",
			"/Applications/Microsoft Edge.app/Contents/MacOS/Microsoft Edge",
		}
	default:
		// Linux + other: rely entirely on PATH. Nothing to add here.
		return nil
	}
}

func notFoundError(checked []string) error {
	var sb strings.Builder
	sb.WriteString("Chrome/Edge/Chromium not found.\n\n")
	sb.WriteString("Checked PATH for: chrome, chromium, chromium-browser, google-chrome, microsoft-edge\n")
	if len(checked) > 0 {
		sb.WriteString("Checked known install paths:\n")
		for _, p := range checked {
			sb.WriteString("  - ")
			sb.WriteString(p)
			sb.WriteString("\n")
		}
	}
	sb.WriteString("\nInstall Chrome/Edge/Chromium, or set GOSLIDE_CHROME_PATH to an explicit binary path.")
	return errors.New(sb.String())
}
```

- [ ] **Step 4: Verify pass + gofmt**

Run: `GOTOOLCHAIN=local go test -C D:/CLAUDE-CODE-GOSLIDE ./internal/pdfexport/ -v`
Expected: all pass.

Run: `GOTOOLCHAIN=local gofmt -w D:/CLAUDE-CODE-GOSLIDE/internal/pdfexport/chrome.go`
Run: `GOTOOLCHAIN=local gofmt -w D:/CLAUDE-CODE-GOSLIDE/internal/pdfexport/chrome_test.go`
Run: `GOTOOLCHAIN=local gofmt -l D:/CLAUDE-CODE-GOSLIDE/internal/pdfexport/`
Expected: empty.

- [ ] **Step 5: Commit**

```bash
git -C D:/CLAUDE-CODE-GOSLIDE add internal/pdfexport/chrome.go internal/pdfexport/chrome_test.go
```

```bash
git -C D:/CLAUDE-CODE-GOSLIDE commit -m "feat(pdfexport): Chrome binary discovery (env / PATH / platform paths)"
```

---

## Task 4: Front-end ready marker

**Files:**
- Modify: `web/static/components.js`

The headless browser needs a deterministic signal that async rendering
(Mermaid's promise-based render in particular) has completed. Set
`window.__goslideReady = true` after `initAllComponents()` and the
mermaid render promises resolve.

- [ ] **Step 1: Locate the Reveal ready handler**

Open `web/static/components.js`. Around line 554 there's:

```javascript
Reveal.on('ready', function () {
    initAllMermaid();
    initAllComponents();
});
```

- [ ] **Step 2: Instrument it**

Replace that block with a version that awaits mermaid render. Mermaid's
existing `initAllMermaid` invokes `mermaid.render(...).then(...)` per
diagram. Track promise count and flip the ready flag when all settle.

Replace the handler with:

```javascript
Reveal.on('ready', function () {
    // Count mermaid diagrams before init so we know when all have rendered.
    var mermaidCount = document.querySelectorAll('.goslide-component[data-type="mermaid"]').length;
    var mermaidDone = 0;
    window.__goslideReady = false;

    function markReadyIfSettled() {
        if (mermaidDone >= mermaidCount) {
            // One animation frame after last paint to be safe.
            requestAnimationFrame(function () {
                window.__goslideReady = true;
            });
        }
    }

    // Hook mermaid promise resolution: patch initAllMermaid's render
    // callback. Simplest approach — wrap mermaid.render to increment
    // mermaidDone on every settle.
    if (typeof mermaid !== 'undefined' && mermaid.render) {
        var origRender = mermaid.render.bind(mermaid);
        mermaid.render = function (id, text) {
            return origRender(id, text).then(function (r) {
                mermaidDone++;
                markReadyIfSettled();
                return r;
            }, function (err) {
                mermaidDone++;
                markReadyIfSettled();
                throw err;
            });
        };
    }

    initAllMermaid();
    initAllComponents();

    // If there were no mermaid diagrams, signal ready immediately.
    if (mermaidCount === 0) {
        markReadyIfSettled();
    }
});
```

- [ ] **Step 3: Build + existing tests**

Run: `GOTOOLCHAIN=local go build -C D:/CLAUDE-CODE-GOSLIDE ./cmd/goslide`
Expected: success.

Run: `GOTOOLCHAIN=local go test -C D:/CLAUDE-CODE-GOSLIDE ./...`
Expected: all existing tests pass (JS is not unit-tested in this repo; we rely on the integration test in Task 6 to exercise the marker).

- [ ] **Step 4: Commit**

```bash
git -C D:/CLAUDE-CODE-GOSLIDE add web/static/components.js
```

```bash
git -C D:/CLAUDE-CODE-GOSLIDE commit -m "feat(web): window.__goslideReady marker after components + mermaid settle"
```

---

## Task 5: Export orchestrator — Options + Launcher interface + unit test

**Files:**
- Create: `internal/pdfexport/export.go`
- Create: `internal/pdfexport/export_test.go`

This task sets up the orchestrator with a `Launcher` interface that
tests can fake. Task 6 adds the chromedp-backed real launcher. Task 7
adds an integration test that exercises the chromedp path end-to-end
but skips when Chrome is absent.

- [ ] **Step 1: Write failing tests**

Create `internal/pdfexport/export_test.go`:

```go
package pdfexport

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

type fakeLauncher struct {
	pdf []byte
	err error
	got LaunchRequest
}

func (f *fakeLauncher) Launch(ctx context.Context, req LaunchRequest) ([]byte, error) {
	f.got = req
	return f.pdf, f.err
}

func TestExport_HappyPath(t *testing.T) {
	dir := t.TempDir()
	mdPath := filepath.Join(dir, "deck.md")
	require.NoError(t, os.WriteFile(mdPath, []byte("---\ntitle: Test\n---\n\n# Hello\n"), 0644))
	outPath := filepath.Join(dir, "deck.pdf")

	fake := &fakeLauncher{pdf: []byte("%PDF-1.4\n...fake...\n%%EOF\n")}
	err := Export(Options{
		File:       mdPath,
		Output:     outPath,
		PaperSize:  "slide-16x9",
		ChromePath: "/fake/chrome",
		Launcher:   fake,
	})
	require.NoError(t, err)

	got, err := os.ReadFile(outPath)
	require.NoError(t, err)
	require.Equal(t, "%PDF-1.4\n...fake...\n%%EOF\n", string(got))

	// Confirm launcher got the right dimensions (slide-16x9 = 20 x 11.25 in)
	require.InDelta(t, 20.0, fake.got.PaperWidthIn, 0.01)
	require.InDelta(t, 11.25, fake.got.PaperHeightIn, 0.01)
	require.False(t, fake.got.ShowNotes)
	require.Equal(t, "/fake/chrome", fake.got.ChromePath)
	require.Contains(t, fake.got.URL, "file://")
	require.Contains(t, fake.got.URL, "?print-pdf")
}

func TestExport_UnknownPaperSize(t *testing.T) {
	dir := t.TempDir()
	mdPath := filepath.Join(dir, "deck.md")
	require.NoError(t, os.WriteFile(mdPath, []byte("# Hello\n"), 0644))
	err := Export(Options{
		File:       mdPath,
		PaperSize:  "garbage",
		ChromePath: "/fake/chrome",
		Launcher:   &fakeLauncher{pdf: []byte("...")},
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "unknown paper size")
}

func TestExport_ShowNotesPropagated(t *testing.T) {
	dir := t.TempDir()
	mdPath := filepath.Join(dir, "deck.md")
	require.NoError(t, os.WriteFile(mdPath, []byte("# Hello\n"), 0644))
	fake := &fakeLauncher{pdf: []byte("%PDF-1.4\n%%EOF\n")}
	err := Export(Options{
		File:       mdPath,
		Output:     filepath.Join(dir, "out.pdf"),
		PaperSize:  "slide-16x9",
		ChromePath: "/fake/chrome",
		ShowNotes:  true,
		Launcher:   fake,
	})
	require.NoError(t, err)
	require.True(t, fake.got.ShowNotes)
	require.Contains(t, fake.got.URL, "showNotes=true")
}

func TestExport_DefaultOutputPath(t *testing.T) {
	dir := t.TempDir()
	mdPath := filepath.Join(dir, "talk.md")
	require.NoError(t, os.WriteFile(mdPath, []byte("# Hello\n"), 0644))
	fake := &fakeLauncher{pdf: []byte("%PDF-1.4\n%%EOF\n")}
	err := Export(Options{
		File:       mdPath,
		PaperSize:  "slide-16x9",
		ChromePath: "/fake/chrome",
		Launcher:   fake,
		// Output not set — should default to <name>.pdf next to input
	})
	require.NoError(t, err)
	expected := filepath.Join(dir, "talk.pdf")
	_, statErr := os.Stat(expected)
	require.NoError(t, statErr)
}

func TestExport_LauncherError(t *testing.T) {
	dir := t.TempDir()
	mdPath := filepath.Join(dir, "deck.md")
	require.NoError(t, os.WriteFile(mdPath, []byte("# Hello\n"), 0644))
	fake := &fakeLauncher{err: errors.New("chrome crashed")}
	err := Export(Options{
		File:       mdPath,
		Output:     filepath.Join(dir, "out.pdf"),
		PaperSize:  "slide-16x9",
		ChromePath: "/fake/chrome",
		Launcher:   fake,
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "chrome crashed")
}

func TestExport_EmptyPDFIsError(t *testing.T) {
	dir := t.TempDir()
	mdPath := filepath.Join(dir, "deck.md")
	require.NoError(t, os.WriteFile(mdPath, []byte("# Hello\n"), 0644))
	fake := &fakeLauncher{pdf: []byte{}}
	err := Export(Options{
		File:       mdPath,
		Output:     filepath.Join(dir, "out.pdf"),
		PaperSize:  "slide-16x9",
		ChromePath: "/fake/chrome",
		Launcher:   fake,
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "empty PDF")
}
```

- [ ] **Step 2: Run — verify fail**

Run: `GOTOOLCHAIN=local go test -C D:/CLAUDE-CODE-GOSLIDE ./internal/pdfexport/ -run TestExport -v`
Expected: FAIL (undefined: Export, Options, LaunchRequest, Launcher).

- [ ] **Step 3: Implement**

Create `internal/pdfexport/export.go`:

```go
package pdfexport

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/GMfatcat/goslide/internal/builder"
)

// Options configures one Export call.
type Options struct {
	File       string    // input .md path (required)
	Output     string    // output .pdf path (defaults to <name>.pdf beside File)
	PaperSize  string    // preset name (required; see papersize.go)
	ShowNotes  bool      // append speaker notes under each slide
	Theme      string    // optional theme override (forwarded to builder.Build)
	Accent     string    // optional accent override
	ChromePath string    // explicit Chrome binary (required)
	Launcher   Launcher  // chromedp-backed in production; fake in tests
}

// LaunchRequest is what Export hands to a Launcher after resolving the
// Options. It decouples the orchestrator from chromedp's concrete API.
type LaunchRequest struct {
	ChromePath    string
	URL           string // file:// URL of the built HTML, with ?print-pdf[&showNotes=true]
	PaperWidthIn  float64
	PaperHeightIn float64
	ShowNotes     bool
}

// Launcher produces PDF bytes from a LaunchRequest.
type Launcher interface {
	Launch(ctx context.Context, req LaunchRequest) ([]byte, error)
}

// Export runs the full build → launch → write pipeline.
func Export(opts Options) error {
	if opts.File == "" {
		return errors.New("pdfexport: File is required")
	}
	if opts.Launcher == nil {
		return errors.New("pdfexport: Launcher is required")
	}
	if opts.ChromePath == "" {
		return errors.New("pdfexport: ChromePath is required")
	}

	widthIn, heightIn, err := ResolvePaperSize(opts.PaperSize)
	if err != nil {
		return err
	}

	output := opts.Output
	if output == "" {
		base := strings.TrimSuffix(filepath.Base(opts.File), filepath.Ext(opts.File))
		output = filepath.Join(filepath.Dir(opts.File), base+".pdf")
	}

	// Stage the static HTML in a temp dir so we don't pollute the project.
	tmpDir, err := os.MkdirTemp("", "goslide-pdf-")
	if err != nil {
		return fmt.Errorf("pdfexport: make temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	htmlPath := filepath.Join(tmpDir, "deck.html")
	if err := builder.Build(builder.Options{
		File:   opts.File,
		Output: htmlPath,
		Theme:  opts.Theme,
		Accent: opts.Accent,
	}); err != nil {
		return fmt.Errorf("pdfexport: build: %w", err)
	}

	url := fileURL(htmlPath) + "?print-pdf"
	if opts.ShowNotes {
		url += "&showNotes=true"
	}

	pdfBytes, err := opts.Launcher.Launch(context.Background(), LaunchRequest{
		ChromePath:    opts.ChromePath,
		URL:           url,
		PaperWidthIn:  widthIn,
		PaperHeightIn: heightIn,
		ShowNotes:     opts.ShowNotes,
	})
	if err != nil {
		return fmt.Errorf("pdfexport: launch: %w", err)
	}
	if len(pdfBytes) == 0 {
		return errors.New("pdfexport: Chrome produced empty PDF")
	}

	if err := os.WriteFile(output, pdfBytes, 0644); err != nil {
		return fmt.Errorf("pdfexport: write output: %w", err)
	}
	return nil
}

// fileURL converts an absolute local path to a file:// URL. Windows
// paths use forward slashes and a leading "/" for the drive.
func fileURL(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}
	abs = strings.ReplaceAll(abs, "\\", "/")
	if len(abs) > 0 && abs[0] != '/' {
		abs = "/" + abs
	}
	return "file://" + abs
}
```

- [ ] **Step 4: Verify pass + gofmt**

Run: `GOTOOLCHAIN=local go test -C D:/CLAUDE-CODE-GOSLIDE ./internal/pdfexport/ -v`
Expected: all pass.

Run: `GOTOOLCHAIN=local gofmt -w D:/CLAUDE-CODE-GOSLIDE/internal/pdfexport/export.go`
Run: `GOTOOLCHAIN=local gofmt -w D:/CLAUDE-CODE-GOSLIDE/internal/pdfexport/export_test.go`
Run: `GOTOOLCHAIN=local gofmt -l D:/CLAUDE-CODE-GOSLIDE/internal/pdfexport/`
Expected: empty.

- [ ] **Step 5: Commit**

```bash
git -C D:/CLAUDE-CODE-GOSLIDE add internal/pdfexport/export.go internal/pdfexport/export_test.go
```

```bash
git -C D:/CLAUDE-CODE-GOSLIDE commit -m "feat(pdfexport): Export orchestrator + Launcher interface (fake-driven tests)"
```

---

## Task 6: chromedp-backed Launcher

**Files:**
- Create: `internal/pdfexport/launcher_chromedp.go`

- [ ] **Step 1: Implement**

Create `internal/pdfexport/launcher_chromedp.go`:

```go
package pdfexport

import (
	"context"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

// ChromedpLauncher drives a real Chrome instance via chromedp.
type ChromedpLauncher struct{}

// NewChromedpLauncher returns a Launcher that uses chromedp.
func NewChromedpLauncher() *ChromedpLauncher {
	return &ChromedpLauncher{}
}

func (l *ChromedpLauncher) Launch(ctx context.Context, req LaunchRequest) ([]byte, error) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ExecPath(req.ChromePath),
		chromedp.Flag("headless", "new"),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
	)

	allocCtx, cancelAlloc := chromedp.NewExecAllocator(ctx, opts...)
	defer cancelAlloc()

	browserCtx, cancelBrowser := chromedp.NewContext(allocCtx)
	defer cancelBrowser()

	// 60-second overall timeout; Chrome launch + page load is typically < 5s.
	tCtx, cancelT := context.WithTimeout(browserCtx, 60*time.Second)
	defer cancelT()

	var pdfData []byte
	err := chromedp.Run(tCtx,
		chromedp.Navigate(req.URL),
		chromedp.WaitReady(".reveal", chromedp.ByQuery),
		chromedp.Poll("window.__goslideReady === true", nil, chromedp.WithPollingInterval(100*time.Millisecond), chromedp.WithPollingTimeout(30*time.Second)),
		chromedp.ActionFunc(func(ctx context.Context) error {
			params := page.PrintToPDFParams{
				PaperWidth:              req.PaperWidthIn,
				PaperHeight:             req.PaperHeightIn,
				PrintBackground:         true,
				PreferCSSPageSize:       true,
				MarginTop:               0,
				MarginBottom:            0,
				MarginLeft:              0,
				MarginRight:             0,
			}
			b, _, err := params.Do(ctx)
			if err != nil {
				return err
			}
			pdfData = b
			return nil
		}),
	)
	if err != nil {
		return nil, err
	}
	return pdfData, nil
}
```

- [ ] **Step 2: Build**

Run: `GOTOOLCHAIN=local go build -C D:/CLAUDE-CODE-GOSLIDE ./...`
Expected: success. (No new tests — this file is exercised by Task 7's integration test which skips on machines without Chrome.)

- [ ] **Step 3: gofmt**

Run: `GOTOOLCHAIN=local gofmt -w D:/CLAUDE-CODE-GOSLIDE/internal/pdfexport/launcher_chromedp.go`
Run: `GOTOOLCHAIN=local gofmt -l D:/CLAUDE-CODE-GOSLIDE/internal/pdfexport/`
Expected: empty.

- [ ] **Step 4: Commit**

```bash
git -C D:/CLAUDE-CODE-GOSLIDE add internal/pdfexport/launcher_chromedp.go
```

```bash
git -C D:/CLAUDE-CODE-GOSLIDE commit -m "feat(pdfexport): chromedp-backed Launcher (Page.printToPDF)"
```

---

## Task 7: Integration test — real Chrome, skip when absent

**Files:**
- Create: `internal/pdfexport/testdata/fixture.md`
- Modify: `internal/pdfexport/export_test.go` (append integration test)

- [ ] **Step 1: Create a tiny fixture deck**

Create `internal/pdfexport/testdata/fixture.md`:

```markdown
---
title: PDF Export Fixture
theme: dark
---

# Slide One

Intro bullet.

---

# Slide Two

- A
- B

---

# Slide Three

The end.
```

- [ ] **Step 2: Append integration test**

Append to `internal/pdfexport/export_test.go`:

```go
func TestExport_Integration_RealChrome(t *testing.T) {
	chromePath, err := FindChrome()
	if err != nil {
		t.Skipf("Chrome not available: %v", err)
	}

	dir := t.TempDir()
	outPath := filepath.Join(dir, "fixture.pdf")

	err = Export(Options{
		File:       "testdata/fixture.md",
		Output:     outPath,
		PaperSize:  "slide-16x9",
		ChromePath: chromePath,
		Launcher:   NewChromedpLauncher(),
	})
	require.NoError(t, err)

	data, err := os.ReadFile(outPath)
	require.NoError(t, err)
	require.True(t, len(data) > 10*1024, "expected PDF > 10KB, got %d bytes", len(data))
	require.True(t, len(data) < 50*1024*1024, "expected PDF < 50MB, got %d bytes", len(data))
	require.Equal(t, "%PDF-", string(data[:5]), "output is not a PDF")
}
```

- [ ] **Step 3: Run — works if Chrome available, skips otherwise**

Run: `GOTOOLCHAIN=local go test -C D:/CLAUDE-CODE-GOSLIDE ./internal/pdfexport/ -v`
Expected:
- On a machine with Chrome: 7 tests pass (unit + integration).
- Without Chrome: integration test is marked SKIP; other tests pass.

- [ ] **Step 4: gofmt**

Run: `GOTOOLCHAIN=local gofmt -w D:/CLAUDE-CODE-GOSLIDE/internal/pdfexport/export_test.go`
Run: `GOTOOLCHAIN=local gofmt -l D:/CLAUDE-CODE-GOSLIDE/internal/pdfexport/`
Expected: empty.

- [ ] **Step 5: Commit**

```bash
git -C D:/CLAUDE-CODE-GOSLIDE add internal/pdfexport/export_test.go internal/pdfexport/testdata/fixture.md
```

```bash
git -C D:/CLAUDE-CODE-GOSLIDE commit -m "test(pdfexport): integration test with real Chrome (skipped when absent)"
```

---

## Task 8: CLI command

**Files:**
- Create: `internal/cli/export_pdf.go`
- Create: `internal/cli/export_pdf_test.go`

- [ ] **Step 1: Write failing test**

Create `internal/cli/export_pdf_test.go`:

```go
package cli

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExportPDF_HelpFlag(t *testing.T) {
	cmd := newExportPDFCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--help"})
	err := cmd.Execute()
	require.NoError(t, err)
	require.Contains(t, out.String(), "export-pdf")
	require.Contains(t, out.String(), "--paper-size")
	require.Contains(t, out.String(), "--notes")
}

func TestExportPDF_RequiresFileArg(t *testing.T) {
	cmd := newExportPDFCmd()
	cmd.SetArgs([]string{})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	err := cmd.Execute()
	require.Error(t, err)
}
```

- [ ] **Step 2: Verify fail**

Run: `GOTOOLCHAIN=local go test -C D:/CLAUDE-CODE-GOSLIDE ./internal/cli/ -run TestExportPDF -v`
Expected: FAIL (undefined: newExportPDFCmd).

- [ ] **Step 3: Implement**

Create `internal/cli/export_pdf.go`:

```go
package cli

import (
	"fmt"

	"github.com/GMfatcat/goslide/internal/pdfexport"
	"github.com/spf13/cobra"
)

var (
	exportPDFOutput    string
	exportPDFPaperSize string
	exportPDFNotes     bool
	exportPDFTheme     string
	exportPDFAccent    string
)

func newExportPDFCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export-pdf <file.md>",
		Short: "Export the presentation as a PDF (requires Chrome/Edge/Chromium)",
		Args:  cobra.ExactArgs(1),
		RunE:  runExportPDF,
	}
	cmd.Flags().StringVarP(&exportPDFOutput, "output", "o", "", "output PDF path (default: {name}.pdf)")
	cmd.Flags().StringVar(&exportPDFPaperSize, "paper-size", "slide-16x9", "paper size preset (slide-16x9, slide-4x3, a4-landscape, letter-landscape)")
	cmd.Flags().BoolVar(&exportPDFNotes, "notes", false, "include speaker notes beneath each slide")
	cmd.Flags().StringVarP(&exportPDFTheme, "theme", "t", "", "override theme")
	cmd.Flags().StringVarP(&exportPDFAccent, "accent", "a", "", "override accent color")
	return cmd
}

func init() {
	rootCmd.AddCommand(newExportPDFCmd())
}

func runExportPDF(cmd *cobra.Command, args []string) error {
	chromePath, err := pdfexport.FindChrome()
	if err != nil {
		return err
	}
	if err := pdfexport.Export(pdfexport.Options{
		File:       args[0],
		Output:     exportPDFOutput,
		PaperSize:  exportPDFPaperSize,
		ShowNotes:  exportPDFNotes,
		Theme:      exportPDFTheme,
		Accent:     exportPDFAccent,
		ChromePath: chromePath,
		Launcher:   pdfexport.NewChromedpLauncher(),
	}); err != nil {
		return err
	}
	out := exportPDFOutput
	if out == "" {
		out = "(default <name>.pdf)"
	}
	fmt.Printf("Exported %s → %s\n", args[0], out)
	return nil
}
```

- [ ] **Step 4: Verify pass + gofmt**

Run: `GOTOOLCHAIN=local go test -C D:/CLAUDE-CODE-GOSLIDE ./internal/cli/ -v`
Expected: all pass.

Run: `GOTOOLCHAIN=local gofmt -w D:/CLAUDE-CODE-GOSLIDE/internal/cli/export_pdf.go`
Run: `GOTOOLCHAIN=local gofmt -w D:/CLAUDE-CODE-GOSLIDE/internal/cli/export_pdf_test.go`
Run: `GOTOOLCHAIN=local gofmt -l D:/CLAUDE-CODE-GOSLIDE/internal/cli/`
Expected: empty.

- [ ] **Step 5: Commit**

```bash
git -C D:/CLAUDE-CODE-GOSLIDE add internal/cli/export_pdf.go internal/cli/export_pdf_test.go
```

```bash
git -C D:/CLAUDE-CODE-GOSLIDE commit -m "feat(cli): goslide export-pdf command"
```

---

## Task 9: Docs — README + PRD

**Files:**
- Modify: `README.md`
- Modify: `README_zh-TW.md`
- Modify: `PRD.md`

- [ ] **Step 1: README English**

Find the "LLM transformer inside api components (experimental)" subsection
(added in v1.4.0). Immediately after it, before the `## ⚙️ Configuration`
major section, insert:

````markdown
### PDF export

Export your deck as a PDF via headless Chrome:

```bash
goslide export-pdf talk.md
goslide export-pdf talk.md -o handout.pdf
goslide export-pdf talk.md --notes              # include speaker notes
goslide export-pdf talk.md --paper-size a4-landscape
```

Paper sizes: `slide-16x9` (default), `slide-4x3`, `a4-landscape`,
`letter-landscape`.

Requires a locally-installed Chrome / Edge / Chromium (discovered via
PATH and standard install locations). Set `GOSLIDE_CHROME_PATH` to
point at a specific binary. No bundled Chromium — GoSlide stays a
single ~8MB binary.

Under the hood, `export-pdf` runs `goslide build` to produce a static
HTML, then drives Chrome's `Page.printToPDF`. Charts, Mermaid, themes,
and LLM-baked API results all render exactly as you see them in the
browser.
````

- [ ] **Step 2: README 繁體中文**

In `README_zh-TW.md`, insert the parallel subsection after "api component 的 LLM 轉換器（experimental）":

````markdown
### PDF 匯出

用 headless Chrome 把簡報匯出為 PDF：

```bash
goslide export-pdf talk.md
goslide export-pdf talk.md -o handout.pdf
goslide export-pdf talk.md --notes              # 包含講者筆記
goslide export-pdf talk.md --paper-size a4-landscape
```

Paper size：`slide-16x9`（預設）、`slide-4x3`、`a4-landscape`、`letter-landscape`。

需要本機已安裝 Chrome / Edge / Chromium（會從 PATH 與標準安裝位置自動尋找）。可用 `GOSLIDE_CHROME_PATH` 環境變數指定特定 binary。不會內建 Chromium，GoSlide 仍是單一 ~8MB binary。

實作上 `export-pdf` 先跑 `goslide build` 產出靜態 HTML，再透過 Chrome 的 `Page.printToPDF` 匯出。圖表、Mermaid、主題、甚至烘焙好的 LLM 結果都會跟瀏覽器看到的一致。
````

- [ ] **Step 3: PRD checkbox**

Open `PRD.md`. Find the deferred-list area. Add (or tick an existing similar line):

```markdown
- [x] PDF export via headless Chrome (Phase 7b, v1.5.0)
```

Place it alongside the existing `- [x] LLM transformer render type ...` line if present.

- [ ] **Step 4: Commit**

```bash
git -C D:/CLAUDE-CODE-GOSLIDE add README.md README_zh-TW.md PRD.md
```

```bash
git -C D:/CLAUDE-CODE-GOSLIDE commit -m "docs: README + PRD for PDF export"
```

---

## Task 10: Final verification

**Files:** none (verification only)

- [ ] **Step 1: gofmt the whole repo**

Run: `GOTOOLCHAIN=local gofmt -l D:/CLAUDE-CODE-GOSLIDE`
Expected: empty.

If any file is listed, run `gofmt -w` on it and add a tidy-up commit.

- [ ] **Step 2: Full test suite**

Run: `GOTOOLCHAIN=local go test -C D:/CLAUDE-CODE-GOSLIDE ./...`
Expected: all pass. The pdfexport integration test runs on machines with Chrome, skips otherwise.

- [ ] **Step 3: go vet**

Run: `GOTOOLCHAIN=local go vet -C D:/CLAUDE-CODE-GOSLIDE ./...`
Expected: clean.

- [ ] **Step 4: Build**

Run: `GOTOOLCHAIN=local go build -C D:/CLAUDE-CODE-GOSLIDE -o D:/CLAUDE-CODE-GOSLIDE/goslide.exe ./cmd/goslide`
Expected: success.

- [ ] **Step 5: Backward compat — existing commands still work**

Run: `D:/CLAUDE-CODE-GOSLIDE/goslide.exe build D:/CLAUDE-CODE-GOSLIDE/examples/demo.md -o /tmp/compat.html`
Expected: success.

Run: `D:/CLAUDE-CODE-GOSLIDE/goslide.exe --help`
Expected: output lists `export-pdf` alongside existing commands.

- [ ] **Step 6: Manual smoke — real PDF export**

With Chrome installed:

Run: `D:/CLAUDE-CODE-GOSLIDE/goslide.exe export-pdf D:/CLAUDE-CODE-GOSLIDE/examples/demo.md -o /tmp/demo.pdf`
Expected: success; `/tmp/demo.pdf` exists and opens in a PDF viewer. Visually confirm:
- One page per slide
- Theme colours / fonts / charts match `goslide serve`
- No blank pages
- Fragment animations collapsed

Try `--notes`:

Run: `D:/CLAUDE-CODE-GOSLIDE/goslide.exe export-pdf D:/CLAUDE-CODE-GOSLIDE/examples/demo.md -o /tmp/demo-notes.pdf --notes`
Expected: Speaker notes appear beneath slides that have `<!-- notes: -->`.

Try an unknown paper size:

Run: `D:/CLAUDE-CODE-GOSLIDE/goslide.exe export-pdf D:/CLAUDE-CODE-GOSLIDE/examples/demo.md --paper-size nonsense`
Expected: exits non-zero with "unknown paper size" listing valid options.

Try with `GOSLIDE_CHROME_PATH=/no/such/path`:

Expected: exits non-zero with the env-override error.

- [ ] **Step 7: No commit needed**

If all steps above passed, implementation is complete.

---

## Success Criteria (from spec §7)

- ✅ `goslide export-pdf talk.md` produces a valid PDF (Task 7 + Task 10.6).
- ✅ One page per slide; fragments collapsed (Task 7 fixture + Task 10.6 manual).
- ✅ `--notes` appends speaker notes (Task 5 unit test + Task 10.6 manual).
- ✅ Missing Chrome → actionable error, non-zero exit (Task 3 + Task 10.6).
- ✅ No regressions (Task 10.2 full suite + 10.5 compat).
- ✅ Only new Go dep is `github.com/chromedp/chromedp` (Task 1 + Task 10.2).
