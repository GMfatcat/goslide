package builder

import (
	"os"
	"path/filepath"
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
