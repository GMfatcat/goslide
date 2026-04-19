package generate

import _ "embed"

//go:embed system_prompt.md
var systemPromptRaw string

// SystemPrompt returns the embedded GoSlide system prompt.
func SystemPrompt() string {
	return systemPromptRaw
}
