package generate

import (
	"errors"
	"fmt"
	"strings"
)

// Input describes what the user wants generated. Topic is required; every
// other field is optional and used in advanced mode.
type Input struct {
	Topic    string
	Audience string
	Slides   int
	Theme    string
	Language string
	Notes    string // free-text body from prompt.md
}

// BuildUserMessage composes the user-role message sent to the LLM.
func BuildUserMessage(in Input) (string, error) {
	if strings.TrimSpace(in.Topic) == "" {
		return "", errors.New("topic is required")
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Topic: %s\n", in.Topic)

	if in.Audience != "" {
		fmt.Fprintf(&b, "Audience: %s\n", in.Audience)
	}
	if in.Slides > 0 {
		fmt.Fprintf(&b, "Slides: %d\n", in.Slides)
	} else {
		b.WriteString("Slides: 10-15\n")
	}
	if in.Theme != "" {
		fmt.Fprintf(&b, "Theme: %s\n", in.Theme)
	}
	if in.Language != "" {
		fmt.Fprintf(&b, "Language: %s\n", in.Language)
	} else {
		b.WriteString("Language: en\n")
	}
	if strings.TrimSpace(in.Notes) != "" {
		b.WriteString("\nAdditional notes:\n")
		b.WriteString(strings.TrimSpace(in.Notes))
		b.WriteString("\n")
	}

	b.WriteString("\nGenerate the full GoSlide Markdown now.\n")
	return b.String(), nil
}
