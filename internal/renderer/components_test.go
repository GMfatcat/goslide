package renderer

import (
	"testing"

	"github.com/GMfatcat/goslide/internal/ir"
	"github.com/stretchr/testify/require"
)

func TestRenderComponents_BasicChart(t *testing.T) {
	slide := ir.Slide{
		Index: 1,
		Components: []ir.Component{
			{Index: 0, Type: "chart:bar", Raw: "title: Yield", Params: map[string]any{"title": "Yield", "data": []any{96.0, 93.0}}},
		},
	}
	html := "before<!--goslide:component:0-->after"
	result := renderComponents(html, slide)
	require.Contains(t, result, `data-type="chart:bar"`)
	require.Contains(t, result, `data-comp-id="s1-c0"`)
	require.Contains(t, result, `data-params=`)
	require.Contains(t, result, "before")
	require.Contains(t, result, "after")
	require.NotContains(t, result, "<!--goslide:component:0-->")
}

func TestRenderComponents_Mermaid(t *testing.T) {
	slide := ir.Slide{
		Index: 2,
		Components: []ir.Component{
			{Index: 0, Type: "mermaid", Raw: "graph TD\n    A --> B"},
		},
	}
	html := "<!--goslide:component:0-->"
	result := renderComponents(html, slide)
	require.Contains(t, result, `data-type="mermaid"`)
	require.Contains(t, result, `data-raw=`)
	require.Contains(t, result, "graph TD")
	require.Contains(t, result, `data-comp-id="s2-c0"`)
}

func TestRenderComponents_MultipleComponents(t *testing.T) {
	slide := ir.Slide{
		Index: 1,
		Components: []ir.Component{
			{Index: 0, Type: "chart:bar", Params: map[string]any{"title": "A"}},
			{Index: 1, Type: "chart:line", Params: map[string]any{"title": "B"}},
		},
	}
	html := "<!--goslide:component:0-->middle<!--goslide:component:1-->"
	result := renderComponents(html, slide)
	require.Contains(t, result, `data-comp-id="s1-c0"`)
	require.Contains(t, result, `data-comp-id="s1-c1"`)
	require.Contains(t, result, "middle")
}

func TestRenderComponents_NoComponents(t *testing.T) {
	slide := ir.Slide{Index: 1}
	html := "<p>no components here</p>"
	result := renderComponents(html, slide)
	require.Equal(t, html, result)
}

func TestRenderComponents_HTMLEscape(t *testing.T) {
	slide := ir.Slide{
		Index: 1,
		Components: []ir.Component{
			{Index: 0, Type: "chart:bar", Params: map[string]any{"title": "A < B & C's"}},
		},
	}
	html := "<!--goslide:component:0-->"
	result := renderComponents(html, slide)
	require.NotContains(t, result, `A < B`)
	require.Contains(t, result, `A &lt; B`)
}

func TestRenderComponents_Table(t *testing.T) {
	slide := ir.Slide{
		Index: 1,
		Components: []ir.Component{
			{Index: 0, Type: "table", Params: map[string]any{
				"columns":  []any{"Name", "Role"},
				"rows":     []any{[]any{"Alice", "Engineer"}},
				"sortable": true,
			}},
		},
	}
	html := "<!--goslide:component:0-->"
	result := renderComponents(html, slide)
	require.Contains(t, result, `data-type="table"`)
}
