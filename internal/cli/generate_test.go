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
