package parser

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSplitRaw_PlainTextNoSeparator(t *testing.T) {
	raw := []byte("# Hello\n\nSome content.\n")
	fm, slides := splitRaw(raw)
	require.Empty(t, fm)
	require.Len(t, slides, 1)
	require.Equal(t, "# Hello\n\nSome content.\n", slides[0])
}

func TestSplitRaw_WithFrontmatter(t *testing.T) {
	raw := []byte("---\ntitle: Test\ntheme: dark\n---\n\n# Slide 1\n")
	fm, slides := splitRaw(raw)
	require.Equal(t, "title: Test\ntheme: dark\n", fm)
	require.Len(t, slides, 1)
	require.Equal(t, "\n# Slide 1\n", slides[0])
}

func TestSplitRaw_MultipleSlides(t *testing.T) {
	raw := []byte("---\ntitle: T\n---\n\n# A\n\n---\n\n# B\n\n---\n\n# C\n")
	fm, slides := splitRaw(raw)
	require.Equal(t, "title: T\n", fm)
	require.Len(t, slides, 3)
	require.Contains(t, slides[0], "# A")
	require.Contains(t, slides[1], "# B")
	require.Contains(t, slides[2], "# C")
}

func TestSplitRaw_DashesInsideCodeFence(t *testing.T) {
	raw := []byte("# A\n\n```\nsome code\n---\nmore code\n```\n\n---\n\n# B\n")
	_, slides := splitRaw(raw)
	require.Len(t, slides, 2)
	require.Contains(t, slides[0], "---\nmore code")
}

func TestSplitRaw_TildeFence(t *testing.T) {
	raw := []byte("# A\n\n~~~\n---\n~~~\n\n---\n\n# B\n")
	_, slides := splitRaw(raw)
	require.Len(t, slides, 2)
	require.Contains(t, slides[0], "~~~\n---\n~~~")
}

func TestSplitRaw_WindowsCRLF(t *testing.T) {
	raw := []byte("---\r\ntitle: T\r\n---\r\n\r\n# A\r\n\r\n---\r\n\r\n# B\r\n")
	fm, slides := splitRaw(raw)
	require.Contains(t, fm, "title: T")
	require.Len(t, slides, 2)
}

func TestSplitRaw_FrontmatterOnly(t *testing.T) {
	raw := []byte("---\ntitle: T\n---\n")
	fm, slides := splitRaw(raw)
	require.Equal(t, "title: T\n", fm)
	require.Len(t, slides, 1)
}

func TestSplitRaw_NestedFenceDifferentChar(t *testing.T) {
	raw := []byte("# A\n\n````\n```\n---\n```\n````\n\n---\n\n# B\n")
	_, slides := splitRaw(raw)
	require.Len(t, slides, 2)
}

func TestSplitRaw_DashesInIndentedLine(t *testing.T) {
	raw := []byte("# A\n\n    ---\n\n---\n\n# B\n")
	_, slides := splitRaw(raw)
	require.Len(t, slides, 2)
	require.Contains(t, slides[0], "    ---")
}
