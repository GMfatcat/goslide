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
