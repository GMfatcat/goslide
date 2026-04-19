package generate

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
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

// ParsePromptFile reads a prompt.md file (YAML frontmatter + Markdown body)
// into an Input. The `topic` frontmatter field is required.
func ParsePromptFile(path string) (Input, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Input{}, fmt.Errorf("read prompt file: %w", err)
	}

	fmBytes, body, err := splitFrontmatter(data)
	if err != nil {
		return Input{}, err
	}

	var raw struct {
		Topic    string `yaml:"topic"`
		Audience string `yaml:"audience"`
		Slides   int    `yaml:"slides"`
		Theme    string `yaml:"theme"`
		Language string `yaml:"language"`
	}
	if err := yaml.Unmarshal(fmBytes, &raw); err != nil {
		return Input{}, fmt.Errorf("parse frontmatter: %w", err)
	}
	if strings.TrimSpace(raw.Topic) == "" {
		return Input{}, errors.New("prompt.md frontmatter: `topic` is required")
	}

	return Input{
		Topic:    raw.Topic,
		Audience: raw.Audience,
		Slides:   raw.Slides,
		Theme:    raw.Theme,
		Language: raw.Language,
		Notes:    string(body),
	}, nil
}

// splitFrontmatter returns (yaml, body, error). The file MUST start with a
// line that is exactly "---"; the frontmatter ends at the next line that is
// exactly "---". Everything after is the body.
func splitFrontmatter(data []byte) (fm []byte, body []byte, err error) {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)

	var fmBuf, bodyBuf bytes.Buffer
	state := 0 // 0=expect opening ---, 1=inside fm, 2=body
	line := 0
	for scanner.Scan() {
		line++
		t := scanner.Text()
		switch state {
		case 0:
			if strings.TrimSpace(t) != "---" {
				return nil, nil, fmt.Errorf("prompt.md must start with '---' (line 1)")
			}
			state = 1
		case 1:
			if strings.TrimSpace(t) == "---" {
				state = 2
				continue
			}
			fmBuf.WriteString(t)
			fmBuf.WriteByte('\n')
		case 2:
			bodyBuf.WriteString(t)
			bodyBuf.WriteByte('\n')
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, nil, err
	}
	if state != 2 {
		return nil, nil, errors.New("prompt.md frontmatter missing closing '---'")
	}
	return fmBuf.Bytes(), bodyBuf.Bytes(), nil
}
