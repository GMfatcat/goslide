package renderer

import (
	"html/template"
	"strings"
	"testing"

	"github.com/GMfatcat/goslide/internal/ir"
	"github.com/stretchr/testify/require"
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

func TestRender_WithChartComponent(t *testing.T) {
	pres := &ir.Presentation{
		Meta: ir.Frontmatter{Title: "Chart", Theme: "default"},
		Slides: []ir.Slide{
			{
				Index:    1,
				Meta:     ir.SlideMeta{Layout: "default"},
				BodyHTML: "<h1>Dashboard</h1>\n<!--goslide:component:0-->\n",
				Components: []ir.Component{
					{Index: 0, Type: "chart:bar", Params: map[string]any{"title": "Yield"}},
				},
			},
		},
	}
	html, err := Render(pres)
	require.NoError(t, err)
	require.Contains(t, html, `data-type="chart:bar"`)
	require.Contains(t, html, `data-comp-id="s1-c0"`)
	require.NotContains(t, html, "<!--goslide:component:0-->")
}

func TestRender_WithMermaidComponent(t *testing.T) {
	pres := &ir.Presentation{
		Meta: ir.Frontmatter{Title: "Mermaid", Theme: "default"},
		Slides: []ir.Slide{
			{
				Index:    1,
				Meta:     ir.SlideMeta{Layout: "default"},
				BodyHTML: "<!--goslide:component:0-->",
				Components: []ir.Component{
					{Index: 0, Type: "mermaid", Raw: "graph TD\n    A --> B"},
				},
			},
		},
	}
	html, err := Render(pres)
	require.NoError(t, err)
	require.Contains(t, html, `data-type="mermaid"`)
	require.Contains(t, html, "graph TD")
}

func TestRender_ComponentInRegion(t *testing.T) {
	pres := &ir.Presentation{
		Meta: ir.Frontmatter{Title: "Region", Theme: "default"},
		Slides: []ir.Slide{
			{
				Index: 1,
				Meta:  ir.SlideMeta{Layout: "two-column"},
				Regions: []ir.Region{
					{Name: "left", HTML: "<p>text</p>"},
					{Name: "right", HTML: "<!--goslide:component:0-->"},
				},
				Components: []ir.Component{
					{Index: 0, Type: "chart:pie", Params: map[string]any{"title": "Share"}},
				},
			},
		},
	}
	html, err := Render(pres)
	require.NoError(t, err)
	require.Contains(t, html, `data-type="chart:pie"`)
	require.NotContains(t, html, "<!--goslide:component:0-->")
}

func TestRender_Placeholder(t *testing.T) {
	pres := &ir.Presentation{
		Meta: ir.Frontmatter{Theme: "dark"},
		Slides: []ir.Slide{{
			Index: 0,
			Meta:  ir.SlideMeta{Layout: "default"},
			Components: []ir.Component{{
				Index:  0,
				Type:   "placeholder",
				Params: map[string]any{"hint": "K8s architecture", "icon": "🗺️", "aspect": "16:9"},
			}},
			BodyHTML: template.HTML(`<!--goslide:component:0-->`),
		}},
	}
	html, err := Render(pres)
	require.NoError(t, err)
	require.Contains(t, html, `data-type="placeholder"`)
	require.Contains(t, html, `K8s architecture`)
}

func TestRender_ImageGrid(t *testing.T) {
	pres := &ir.Presentation{
		Meta: ir.Frontmatter{Theme: "dark"},
		Slides: []ir.Slide{{
			Index: 0,
			Meta:  ir.SlideMeta{Layout: "image-grid", Columns: 2},
			Regions: []ir.Region{
				{Name: "cell", HTML: template.HTML("<p>A</p>")},
				{Name: "cell", HTML: template.HTML("<p>B</p>")},
			},
		}},
	}
	html, err := Render(pres)
	require.NoError(t, err)
	require.Contains(t, html, `data-layout="image-grid"`)
	require.Contains(t, html, `data-columns="2"`)
	require.Equal(t, 2, strings.Count(html, `class="region-cell"`))
}

func TestRender_EmitsLLMBakes(t *testing.T) {
	pres := &ir.Presentation{
		Meta: ir.Frontmatter{Theme: "dark"},
		Slides: []ir.Slide{{
			Index: 0,
			Meta:  ir.SlideMeta{Layout: "default"},
			Components: []ir.Component{{
				Index: 0,
				Type:  "api",
				Params: map[string]any{
					"endpoint":   "/api/x",
					"_llm_bakes": map[string]any{"1": "## Insights"},
				},
			}},
			BodyHTML: `<!--goslide:component:0-->`,
		}},
	}
	html, err := Render(pres)
	require.NoError(t, err)
	require.Contains(t, html, `data-llm-bakes=`)
	require.Contains(t, html, `Insights`)
}
