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
