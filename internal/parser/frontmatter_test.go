package parser

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseFrontmatter_Complete(t *testing.T) {
	raw := "title: My Talk\nauthor: Alice\ndate: 2026-01-01\ntags: [go, slides]\ntheme: dark\naccent: teal\ntransition: fade\nfragments: true\nfragment-style: highlight-current\n"
	fm, err := parseFrontmatter(raw)
	require.NoError(t, err)
	require.Equal(t, "My Talk", fm.Title)
	require.Equal(t, "Alice", fm.Author)
	require.Equal(t, "dark", fm.Theme)
	require.Equal(t, "teal", fm.Accent)
	require.Equal(t, "fade", fm.Transition)
	require.True(t, fm.Fragments)
	require.Equal(t, "highlight-current", fm.FragmentStyle)
	require.Equal(t, []string{"go", "slides"}, fm.Tags)
}

func TestParseFrontmatter_MissingFields(t *testing.T) {
	raw := "title: Minimal\n"
	fm, err := parseFrontmatter(raw)
	require.NoError(t, err)
	require.Equal(t, "Minimal", fm.Title)
	require.Empty(t, fm.Theme)
	require.Empty(t, fm.Accent)
	require.False(t, fm.Fragments)
}

func TestParseFrontmatter_UnknownFieldsIgnored(t *testing.T) {
	raw := "title: T\ncustom_field: whatever\ntheme: dark\n"
	fm, err := parseFrontmatter(raw)
	require.NoError(t, err)
	require.Equal(t, "T", fm.Title)
	require.Equal(t, "dark", fm.Theme)
}

func TestParseFrontmatter_YAMLSyntaxError(t *testing.T) {
	raw := "title: T\n  bad indent:\n"
	_, err := parseFrontmatter(raw)
	require.Error(t, err)
}

func TestParseFrontmatter_CaseNormalization(t *testing.T) {
	raw := "theme: DARK\naccent: Teal\ntransition: Fade\nfragment-style: Highlight-Current\n"
	fm, err := parseFrontmatter(raw)
	require.NoError(t, err)
	require.Equal(t, "dark", fm.Theme)
	require.Equal(t, "teal", fm.Accent)
	require.Equal(t, "fade", fm.Transition)
	require.Equal(t, "highlight-current", fm.FragmentStyle)
}

func TestParseFrontmatter_Empty(t *testing.T) {
	fm, err := parseFrontmatter("")
	require.NoError(t, err)
	require.Empty(t, fm.Title)
}
