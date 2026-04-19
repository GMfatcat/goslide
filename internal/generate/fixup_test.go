package generate

import (
	"strings"
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

func TestApplyFrontmatterTerminator_InsertsClosing(t *testing.T) {
	md := "---\ntitle: Hello\ntheme: dark\n\n# Heading\n\nBody\n"
	fixed, report := Try(md, nil)

	require.Contains(t, fixed, "---\ntitle: Hello\ntheme: dark\n---\n")
	require.NotEmpty(t, report.Fixes)
	found := false
	for _, f := range report.Fixes {
		if f.Rule == "fm-terminator" {
			found = true
			require.Equal(t, 1, f.Line)
		}
	}
	require.True(t, found, "fm-terminator rule should fire")
}

func TestApplyFrontmatterTerminator_NoopWhenPresent(t *testing.T) {
	md := "---\ntitle: Hello\n---\n\nBody\n"
	_, report := Try(md, nil)
	for _, f := range report.Fixes {
		require.NotEqual(t, "fm-terminator", f.Rule)
	}
}

func TestApplyFrontmatterTerminator_NoopWithoutFrontmatter(t *testing.T) {
	md := "# Just a heading\n\nBody\n"
	_, report := Try(md, nil)
	require.Empty(t, report.Fixes)
}

func TestApplyFrontmatterUnquotedColon_Quotes(t *testing.T) {
	md := "---\ntitle: Go: A Short Intro\ntheme: dark\n---\n\nBody\n"
	fixed, report := Try(md, nil)

	require.Contains(t, fixed, `title: "Go: A Short Intro"`)
	found := false
	for _, f := range report.Fixes {
		if f.Rule == "fm-unquoted-colon" {
			found = true
			require.Equal(t, 2, f.Line)
		}
	}
	require.True(t, found)
}

func TestApplyFrontmatterUnquotedColon_LeavesQuotedAlone(t *testing.T) {
	md := "---\ntitle: \"Go: already quoted\"\n---\n"
	_, report := Try(md, nil)
	for _, f := range report.Fixes {
		require.NotEqual(t, "fm-unquoted-colon", f.Rule)
	}
}

func TestApplyTrailingNewline_Adds(t *testing.T) {
	md := "# Heading\n\nBody"
	fixed, report := Try(md, nil)
	require.True(t, strings.HasSuffix(fixed, "\n"))
	found := false
	for _, f := range report.Fixes {
		if f.Rule == "trailing-newline" {
			found = true
		}
	}
	require.True(t, found)
}

func TestApplyTrailingNewline_NoopWhenPresent(t *testing.T) {
	md := "# Heading\n\nBody\n"
	_, report := Try(md, nil)
	for _, f := range report.Fixes {
		require.NotEqual(t, "trailing-newline", f.Rule)
	}
}

func TestTry_MultipleRulesFireTogether(t *testing.T) {
	// Unquoted colon + missing fm terminator + unclosed fence + no trailing newline
	md := "---\ntitle: Go: Intro\n\n```\ncode"
	_, report := Try(md, nil)

	rules := map[string]bool{}
	for _, f := range report.Fixes {
		rules[f.Rule] = true
	}
	require.True(t, rules["fence-close"], "fence-close should fire")
	require.True(t, rules["fm-terminator"], "fm-terminator should fire")
	require.True(t, rules["fm-unquoted-colon"], "fm-unquoted-colon should fire")
	require.True(t, rules["trailing-newline"], "trailing-newline should fire")
}
