package generate

import (
	"os"
	"path/filepath"
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
