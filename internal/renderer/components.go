package renderer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/user/goslide/internal/ir"
)

func renderComponents(html string, slide ir.Slide) string {
	if len(slide.Components) == 0 {
		return html
	}

	for _, comp := range slide.Components {
		placeholder := fmt.Sprintf("<!--goslide:component:%d-->", comp.Index)
		replacement := buildComponentDiv(slide.Index, comp)
		html = strings.Replace(html, placeholder, replacement, 1)
	}

	return html
}

func buildComponentDiv(slideIndex int, comp ir.Component) string {
	compID := fmt.Sprintf("s%d-c%d", slideIndex, comp.Index)

	var paramsAttr string
	var rawAttr string

	if comp.Type == "mermaid" {
		paramsAttr = "{}"
		rawAttr = escapeAttr(comp.Raw)
	} else {
		var buf bytes.Buffer
		enc := json.NewEncoder(&buf)
		enc.SetEscapeHTML(false)
		if err := enc.Encode(comp.Params); err != nil {
			buf.Reset()
			buf.WriteString("{}")
		}
		paramsAttr = escapeAttr(strings.TrimRight(buf.String(), "\n"))
		rawAttr = ""
	}

	return fmt.Sprintf(
		`<div class="goslide-component" data-type="%s" data-params="%s" data-raw="%s" data-comp-id="%s">%s</div>`,
		escapeAttr(comp.Type),
		paramsAttr,
		rawAttr,
		compID,
		comp.ContentHTML,
	)
}

func escapeAttr(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	s = strings.ReplaceAll(s, "'", "&#39;")
	return s
}
