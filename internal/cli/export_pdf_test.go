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
