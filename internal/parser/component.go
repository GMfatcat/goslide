package parser

import (
	"fmt"
	"strings"

	"github.com/GMfatcat/goslide/internal/ir"
	"gopkg.in/yaml.v3"
)

var knownComponentPrefixes = map[string]bool{
	"chart": true, "mermaid": true, "table": true,
	"tabs": true, "panel": true, "slider": true,
	"toggle": true, "api": true, "embed": true, "card": true,
	"placeholder": true,
}

func isComponentFence(lang string) bool {
	if knownComponentPrefixes[lang] {
		return true
	}
	prefix := lang
	if idx := strings.Index(lang, ":"); idx != -1 {
		prefix = lang[:idx]
	}
	return knownComponentPrefixes[prefix]
}

func extractComponents(body string) (string, []ir.Component) {
	lines := strings.Split(body, "\n")
	var result []string
	var components []ir.Component

	i := 0
	for i < len(lines) {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "~~~") && len(trimmed) > 3 {
			lang := strings.TrimSpace(trimmed[3:])
			if isComponentFence(lang) {
				var contentLines []string
				i++
				for i < len(lines) {
					if strings.TrimSpace(lines[i]) == "~~~" {
						break
					}
					contentLines = append(contentLines, lines[i])
					i++
				}
				raw := strings.Join(contentLines, "\n")

				var params map[string]any
				if err := yaml.Unmarshal([]byte(raw), &params); err != nil {
					params = nil
				}

				comp := ir.Component{
					Index:  len(components),
					Type:   lang,
					Raw:    raw,
					Params: params,
				}
				components = append(components, comp)
				result = append(result, fmt.Sprintf("<!--goslide:component:%d-->", comp.Index))
				i++
				continue
			}
		}

		result = append(result, line)
		i++
	}

	cleaned := strings.Join(result, "\n")
	return cleaned, components
}
