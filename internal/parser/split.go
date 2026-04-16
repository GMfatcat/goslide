package parser

import (
	"strings"
)

func splitRaw(raw []byte) (frontmatter string, slides []string) {
	content := strings.ReplaceAll(string(raw), "\r\n", "\n")
	// strings.Split on a newline-terminated string produces a trailing empty
	// element; remove it so joins reconstruct the original content exactly.
	rawLines := strings.Split(content, "\n")
	lines := rawLines
	if len(rawLines) > 0 && rawLines[len(rawLines)-1] == "" {
		lines = rawLines[:len(rawLines)-1]
	}

	idx := 0

	if idx < len(lines) && strings.TrimSpace(lines[idx]) == "---" {
		idx++
		fmStart := idx
		for idx < len(lines) && strings.TrimSpace(lines[idx]) != "---" {
			idx++
		}
		if idx < len(lines) {
			frontmatter = strings.Join(lines[fmStart:idx], "\n") + "\n"
			idx++
		} else {
			idx = 0
		}
	}

	var current []string
	var fence fenceState

	for ; idx < len(lines); idx++ {
		line := lines[idx]
		trimmed := strings.TrimSpace(line)

		if fence.update(trimmed) {
			current = append(current, line)
			continue
		}

		if !fence.inside && trimmed == "---" && line == strings.TrimRight(line, " \t") && !strings.HasPrefix(line, " ") {
			if current != nil {
				slides = append(slides, strings.Join(current, "\n")+"\n")
			}
			current = []string{}
			continue
		}

		current = append(current, line)
	}

	if current != nil {
		slides = append(slides, strings.Join(current, "\n")+"\n")
	}

	if len(slides) == 0 {
		slides = []string{"\n"}
	}

	return
}

type fenceState struct {
	inside bool
	char   byte
	count  int
}

func (f *fenceState) update(trimmed string) bool {
	if len(trimmed) < 3 {
		return f.inside
	}

	ch := trimmed[0]
	if ch != '`' && ch != '~' {
		return f.inside
	}

	n := 0
	for n < len(trimmed) && trimmed[n] == ch {
		n++
	}
	if n < 3 {
		return f.inside
	}

	if !f.inside {
		f.inside = true
		f.char = ch
		f.count = n
		return true
	}

	if ch == f.char && n >= f.count {
		rest := strings.TrimSpace(trimmed[n:])
		if rest == "" {
			f.inside = false
			return true
		}
	}

	return f.inside
}
