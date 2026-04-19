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
