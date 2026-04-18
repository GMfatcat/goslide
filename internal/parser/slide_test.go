package parser

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/user/goslide/internal/ir"
)

func TestParseSlide_MetadataComments(t *testing.T) {
	raw := "<!-- layout: two-column -->\n<!-- transition: fade -->\n\n# Title\n\nContent.\n"
	slide := parseSlide(1, raw, ir.Frontmatter{})
	require.Equal(t, "two-column", slide.Meta.Layout)
	require.Equal(t, "fade", slide.Meta.Transition)
	require.Contains(t, string(slide.BodyHTML), "<h1>Title</h1>")
}

func TestParseSlide_MetadataCaseNormalize(t *testing.T) {
	raw := "<!-- layout: Two-Column -->\n\n# Title\n"
	slide := parseSlide(1, raw, ir.Frontmatter{})
	require.Equal(t, "two-column", slide.Meta.Layout)
}

func TestParseSlide_FragmentsFromComment(t *testing.T) {
	raw := "<!-- fragments: true -->\n<!-- fragment-style: highlight-current -->\n\n# Title\n\n- A\n- B\n"
	slide := parseSlide(1, raw, ir.Frontmatter{})
	require.True(t, slide.Meta.Fragments)
	require.Equal(t, "highlight-current", slide.Meta.FragmentStyle)
}

func TestParseSlide_DefaultsFromFrontmatter(t *testing.T) {
	raw := "# Title\n\nContent.\n"
	defaults := ir.Frontmatter{Transition: "fade", Fragments: true}
	slide := parseSlide(1, raw, defaults)
	require.Equal(t, "fade", slide.Meta.Transition)
	require.True(t, slide.Meta.Fragments)
}

func TestParseSlide_CommentOverridesFrontmatter(t *testing.T) {
	raw := "<!-- transition: zoom -->\n\n# Title\n"
	defaults := ir.Frontmatter{Transition: "fade"}
	slide := parseSlide(1, raw, defaults)
	require.Equal(t, "zoom", slide.Meta.Transition)
}

func TestParseSlide_RegionSplitting_TwoColumn(t *testing.T) {
	raw := "<!-- layout: two-column -->\n\n# Title\n\n<!-- left -->\n\nLeft content\n\n<!-- right -->\n\nRight content\n"
	slide := parseSlide(1, raw, ir.Frontmatter{})
	require.Equal(t, "two-column", slide.Meta.Layout)
	require.Len(t, slide.Regions, 2)
	require.Equal(t, "left", slide.Regions[0].Name)
	require.Contains(t, string(slide.Regions[0].HTML), "Left content")
	require.Equal(t, "right", slide.Regions[1].Name)
	require.Contains(t, string(slide.Regions[1].HTML), "Right content")
}

func TestParseSlide_RegionSplitting_CodePreview(t *testing.T) {
	raw := "<!-- layout: code-preview -->\n\n# Demo\n\n<!-- code -->\n\n```go\nfmt.Println()\n```\n\n<!-- preview -->\n\nOutput here.\n"
	slide := parseSlide(1, raw, ir.Frontmatter{})
	require.Len(t, slide.Regions, 2)
	require.Equal(t, "code", slide.Regions[0].Name)
	require.Equal(t, "preview", slide.Regions[1].Name)
}

func TestParseSlide_NoLayout_DefaultUsed(t *testing.T) {
	raw := "# Title\n\nContent.\n"
	slide := parseSlide(1, raw, ir.Frontmatter{})
	require.Equal(t, "default", slide.Meta.Layout)
	require.Empty(t, slide.Regions)
}

func TestParseSlide_GoldmarkRendering(t *testing.T) {
	raw := "# Title\n\n**Bold** and *italic*.\n\n- Item 1\n- Item 2\n"
	slide := parseSlide(1, raw, ir.Frontmatter{})
	require.Contains(t, string(slide.BodyHTML), "<strong>Bold</strong>")
	require.Contains(t, string(slide.BodyHTML), "<em>italic</em>")
	require.Contains(t, string(slide.BodyHTML), "<li>Item 1</li>")
}

func TestParseSlide_RawBodyStored(t *testing.T) {
	raw := "<!-- layout: default -->\n\n# Title\n\n- A\n- B\n"
	slide := parseSlide(1, raw, ir.Frontmatter{})
	require.Contains(t, slide.RawBody, "- A")
	require.Contains(t, slide.RawBody, "- B")
}

func TestParseSlide_Index(t *testing.T) {
	raw := "# Slide 7\n"
	slide := parseSlide(7, raw, ir.Frontmatter{})
	require.Equal(t, 7, slide.Index)
}

func TestParseSlide_SlideNumberHidden(t *testing.T) {
	raw := "<!-- slide-number: false -->\n\n# Title\n"
	slide := parseSlide(1, raw, ir.Frontmatter{})
	require.True(t, slide.Meta.SlideNumberHidden)
}

func TestParseSlide_SlideNumberNotHidden(t *testing.T) {
	raw := "# Title\n"
	slide := parseSlide(1, raw, ir.Frontmatter{})
	require.False(t, slide.Meta.SlideNumberHidden)
}

func TestParseSlide_WithComponent(t *testing.T) {
	raw := "# Title\n\n~~~chart:bar\ntitle: Yield\ndata: [96, 93]\n~~~\n\nAfter chart.\n"
	slide := parseSlide(1, raw, ir.Frontmatter{})
	require.Len(t, slide.Components, 1)
	require.Equal(t, "chart:bar", slide.Components[0].Type)
	require.Contains(t, string(slide.BodyHTML), "<!--goslide:component:0-->")
	require.Contains(t, string(slide.BodyHTML), "After chart.")
	require.NotContains(t, string(slide.BodyHTML), "~~~chart:bar")
}

func TestParseSlide_ComponentInRegion(t *testing.T) {
	raw := "<!-- layout: two-column -->\n\n<!-- left -->\n\nText left.\n\n<!-- right -->\n\n~~~chart:pie\ntitle: Share\n~~~\n"
	slide := parseSlide(1, raw, ir.Frontmatter{})
	require.Len(t, slide.Components, 1)
	require.Len(t, slide.Regions, 2)
	require.Contains(t, string(slide.Regions[1].HTML), "<!--goslide:component:0-->")
}

func TestParseSlide_NonComponentFenceUntouched(t *testing.T) {
	raw := "# Code\n\n~~~go\nfmt.Println()\n~~~\n"
	slide := parseSlide(1, raw, ir.Frontmatter{})
	require.Len(t, slide.Components, 0)
	require.Contains(t, string(slide.BodyHTML), "<code")
}

func TestParseSlide_CardWithDetail(t *testing.T) {
	raw := "<!-- layout: grid-cards -->\n\n~~~card\ntitle: Test Card\ndesc: A description\n---\n## Detail\n\nDetail content.\n~~~\n"
	slide := parseSlide(1, raw, ir.Frontmatter{})
	require.Len(t, slide.Components, 1)
	require.Equal(t, "card", slide.Components[0].Type)
	require.Equal(t, "Test Card", slide.Components[0].Params["title"])
	require.Contains(t, slide.Components[0].ContentHTML, "Detail content.")
	require.Contains(t, slide.Components[0].ContentHTML, "<h2>Detail</h2>")
}

func TestParseSlide_CardWithoutDetail(t *testing.T) {
	raw := "~~~card\ntitle: Simple\ndesc: No detail\n~~~\n"
	slide := parseSlide(1, raw, ir.Frontmatter{})
	require.Len(t, slide.Components, 1)
	require.Equal(t, "Simple", slide.Components[0].Params["title"])
	require.Empty(t, slide.Components[0].ContentHTML)
}

func TestParseSlide_Columns(t *testing.T) {
	raw := "<!-- layout: grid-cards -->\n<!-- columns: 3 -->\n\n# Title\n"
	slide := parseSlide(1, raw, ir.Frontmatter{})
	require.Equal(t, "grid-cards", slide.Meta.Layout)
	require.Equal(t, 3, slide.Meta.Columns)
}
