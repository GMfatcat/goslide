package llm

import "strings"

// Render substitutes the literal string "{{data}}" in template with the
// raw bytes of data. No other template syntax is supported in the MVP.
func Render(template string, data []byte) string {
	return strings.ReplaceAll(template, "{{data}}", string(data))
}
