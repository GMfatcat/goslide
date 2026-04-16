package parser

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParse_BasicPresentation(t *testing.T) {
	raw := []byte("---\ntitle: Hello\ntheme: dark\n---\n\n# Slide 1\n\nContent.\n\n---\n\n# Slide 2\n\nMore.\n")
	pres, err := Parse(raw, "test.md")
	require.NoError(t, err)
	require.Equal(t, "test.md", pres.Source)
	require.Equal(t, "Hello", pres.Meta.Title)
	require.Equal(t, "dark", pres.Meta.Theme)
	require.Len(t, pres.Slides, 2)
	require.Equal(t, 1, pres.Slides[0].Index)
	require.Equal(t, 2, pres.Slides[1].Index)
	require.Contains(t, string(pres.Slides[0].BodyHTML), "Slide 1")
}

func TestParse_NoFrontmatter(t *testing.T) {
	raw := []byte("# Only slide\n")
	pres, err := Parse(raw, "test.md")
	require.NoError(t, err)
	require.Empty(t, pres.Meta.Title)
	require.Len(t, pres.Slides, 1)
}

func TestParse_FrontmatterYAMLError(t *testing.T) {
	raw := []byte("---\ntitle: T\n  bad indent:\n---\n# Slide\n")
	_, err := Parse(raw, "test.md")
	require.Error(t, err)
}

func TestParse_SlideInheritsDefaults(t *testing.T) {
	raw := []byte("---\ntransition: fade\nfragments: true\n---\n\n# Slide\n\n- A\n- B\n")
	pres, err := Parse(raw, "test.md")
	require.NoError(t, err)
	require.Equal(t, "fade", pres.Slides[0].Meta.Transition)
	require.True(t, pres.Slides[0].Meta.Fragments)
}

func TestParse_SlideOverridesDefault(t *testing.T) {
	raw := []byte("---\ntransition: slide\n---\n\n<!-- transition: zoom -->\n\n# Overridden\n")
	pres, err := Parse(raw, "test.md")
	require.NoError(t, err)
	require.Equal(t, "zoom", pres.Slides[0].Meta.Transition)
}
