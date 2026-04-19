package parser

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExtractComponents_BasicChart(t *testing.T) {
	body := "# Title\n\n~~~chart:bar\ntitle: Yield\nlabels: [\"A\", \"B\"]\ndata: [96, 93]\n~~~\n\nMore text.\n"
	cleaned, comps := extractComponents(body)
	require.Len(t, comps, 1)
	require.Equal(t, "chart:bar", comps[0].Type)
	require.Equal(t, 0, comps[0].Index)
	require.Contains(t, comps[0].Raw, "title: Yield")
	require.Equal(t, "Yield", comps[0].Params["title"])
	require.Contains(t, cleaned, "<!--goslide:component:0-->")
	require.Contains(t, cleaned, "# Title")
	require.Contains(t, cleaned, "More text.")
	require.NotContains(t, cleaned, "~~~chart:bar")
}

func TestExtractComponents_Mermaid(t *testing.T) {
	body := "~~~mermaid\ngraph TD\n    A --> B\n~~~\n"
	cleaned, comps := extractComponents(body)
	require.Len(t, comps, 1)
	require.Equal(t, "mermaid", comps[0].Type)
	require.Contains(t, comps[0].Raw, "graph TD")
	require.Contains(t, cleaned, "<!--goslide:component:0-->")
}

func TestExtractComponents_MermaidParamsNil(t *testing.T) {
	body := "~~~mermaid\ngraph TD\n    A --> B\n~~~\n"
	_, comps := extractComponents(body)
	require.Nil(t, comps[0].Params)
}

func TestExtractComponents_MultipleComponents(t *testing.T) {
	body := "~~~chart:bar\ntitle: A\n~~~\n\nMiddle text.\n\n~~~chart:line\ntitle: B\n~~~\n"
	cleaned, comps := extractComponents(body)
	require.Len(t, comps, 2)
	require.Equal(t, "chart:bar", comps[0].Type)
	require.Equal(t, 0, comps[0].Index)
	require.Equal(t, "chart:line", comps[1].Type)
	require.Equal(t, 1, comps[1].Index)
	require.Contains(t, cleaned, "<!--goslide:component:0-->")
	require.Contains(t, cleaned, "<!--goslide:component:1-->")
	require.Contains(t, cleaned, "Middle text.")
}

func TestExtractComponents_NonComponentFencePassthrough(t *testing.T) {
	body := "~~~go\nfmt.Println(\"hello\")\n~~~\n"
	cleaned, comps := extractComponents(body)
	require.Len(t, comps, 0)
	require.Contains(t, cleaned, "~~~go")
	require.Contains(t, cleaned, "fmt.Println")
}

func TestExtractComponents_UnknownComponentPassthrough(t *testing.T) {
	body := "~~~unknown-widget\nfoo: bar\n~~~\n"
	cleaned, comps := extractComponents(body)
	require.Len(t, comps, 0)
	require.Contains(t, cleaned, "~~~unknown-widget")
}

func TestExtractComponents_NoComponents(t *testing.T) {
	body := "# Just markdown\n\nParagraph.\n"
	cleaned, comps := extractComponents(body)
	require.Len(t, comps, 0)
	require.Equal(t, body, cleaned)
}

func TestExtractComponents_ComponentWithColonType(t *testing.T) {
	body := "~~~embed:html\n<div>custom</div>\n~~~\n"
	cleaned, comps := extractComponents(body)
	require.Len(t, comps, 1)
	require.Equal(t, "embed:html", comps[0].Type)
	require.Contains(t, cleaned, "<!--goslide:component:0-->")
}

func TestExtractComponents_Table(t *testing.T) {
	body := "~~~table\ncolumns: [Name, Role]\nrows:\n  - [\"Alice\", \"Engineer\"]\nsortable: true\n~~~\n"
	_, comps := extractComponents(body)
	require.Len(t, comps, 1)
	require.Equal(t, "table", comps[0].Type)
	require.Equal(t, true, comps[0].Params["sortable"])
}

func TestExtractComponents_YAMLParseError(t *testing.T) {
	body := "~~~chart:bar\nkey: [unclosed bracket\n~~~\n"
	cleaned, comps := extractComponents(body)
	require.Len(t, comps, 1)
	require.Equal(t, "chart:bar", comps[0].Type)
	require.Nil(t, comps[0].Params)
	require.Contains(t, comps[0].Raw, "unclosed")
	require.Contains(t, cleaned, "<!--goslide:component:0-->")
}

func TestExtractComponents_Placeholder(t *testing.T) {
	body := "~~~placeholder\nhint: K8s architecture\nicon: 🗺️\naspect: 16:9\n~~~\n"
	_, comps := extractComponents(body)
	require.Len(t, comps, 1)
	require.Equal(t, "placeholder", comps[0].Type)
	require.Equal(t, "K8s architecture", comps[0].Params["hint"])
	require.Equal(t, "🗺️", comps[0].Params["icon"])
	require.Equal(t, "16:9", comps[0].Params["aspect"])
}

func TestExtractComponents_PlaceholderWithBody(t *testing.T) {
	// Actual splitting happens in slide.parseSlide, not extractComponents.
	// Here we just verify the raw body is captured; Task 5 verifies the split.
	body := "~~~placeholder\nhint: Cluster\n---\nsubtitle text\n~~~\n"
	_, comps := extractComponents(body)
	require.Len(t, comps, 1)
	require.Contains(t, comps[0].Raw, "subtitle text")
}
