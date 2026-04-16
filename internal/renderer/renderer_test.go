package renderer

import (
	"html/template"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/user/goslide/internal/ir"
)

func TestRender_BasicSlide(t *testing.T) {
	pres := &ir.Presentation{
		Meta: ir.Frontmatter{Title: "Test", Theme: "default", Accent: "blue", Transition: "slide"},
		Slides: []ir.Slide{
			{Index: 1, Meta: ir.SlideMeta{Layout: "default"}, BodyHTML: "<h1>Hello</h1>"},
		},
	}
	html, err := Render(pres)
	require.NoError(t, err)
	require.Contains(t, html, "<title>Test</title>")
	require.Contains(t, html, `data-accent="blue"`)
	require.Contains(t, html, `href="/themes/default.css"`)
	require.Contains(t, html, "<h1>Hello</h1>")
}

func TestRender_DarkTheme(t *testing.T) {
	pres := &ir.Presentation{
		Meta:   ir.Frontmatter{Title: "Dark", Theme: "dark", Accent: "teal"},
		Slides: []ir.Slide{{Index: 1, Meta: ir.SlideMeta{Layout: "default"}, BodyHTML: "<p>content</p>"}},
	}
	html, err := Render(pres)
	require.NoError(t, err)
	require.Contains(t, html, `href="/themes/dark.css"`)
	require.Contains(t, html, `data-accent="teal"`)
}

func TestRender_SlideTransitionOverride(t *testing.T) {
	pres := &ir.Presentation{
		Meta: ir.Frontmatter{Title: "T", Theme: "default", Transition: "slide"},
		Slides: []ir.Slide{
			{Index: 1, Meta: ir.SlideMeta{Layout: "default", Transition: "fade"}, BodyHTML: "<p>x</p>"},
		},
	}
	html, err := Render(pres)
	require.NoError(t, err)
	require.Contains(t, html, `data-transition="fade"`)
}

func TestRender_FragmentsAttributes(t *testing.T) {
	pres := &ir.Presentation{
		Meta: ir.Frontmatter{Title: "T", Theme: "default"},
		Slides: []ir.Slide{
			{Index: 1, Meta: ir.SlideMeta{Layout: "default", Fragments: true, FragmentStyle: "highlight-current"}, BodyHTML: "<ul><li>A</li><li>B</li></ul>"},
		},
	}
	html, err := Render(pres)
	require.NoError(t, err)
	require.Contains(t, html, `data-fragments="true"`)
	require.Contains(t, html, `data-fragment-style="highlight-current"`)
}

func TestRender_TwoColumnRegions(t *testing.T) {
	pres := &ir.Presentation{
		Meta: ir.Frontmatter{Title: "T", Theme: "default"},
		Slides: []ir.Slide{
			{
				Index: 1,
				Meta:  ir.SlideMeta{Layout: "two-column"},
				Regions: []ir.Region{
					{Name: "left", HTML: template.HTML("<p>Left</p>")},
					{Name: "right", HTML: template.HTML("<p>Right</p>")},
				},
			},
		},
	}
	html, err := Render(pres)
	require.NoError(t, err)
	require.Contains(t, html, `data-layout="two-column"`)
	require.Contains(t, html, `class="region-left"`)
	require.Contains(t, html, `class="region-right"`)
	leftIdx := strings.Index(html, "region-left")
	rightIdx := strings.Index(html, "region-right")
	require.Less(t, leftIdx, rightIdx)
}

func TestRender_DefaultValues(t *testing.T) {
	pres := &ir.Presentation{
		Meta:   ir.Frontmatter{},
		Slides: []ir.Slide{{Index: 1, Meta: ir.SlideMeta{Layout: "default"}, BodyHTML: "<p>x</p>"}},
	}
	html, err := Render(pres)
	require.NoError(t, err)
	require.Contains(t, html, `href="/themes/default.css"`)
	require.Contains(t, html, `data-accent="blue"`)
	require.Contains(t, html, `transition: 'slide'`)
}
