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
