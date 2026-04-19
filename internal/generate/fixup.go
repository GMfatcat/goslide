package generate

import "strings"

// FixReport describes which heuristic fixes were applied.
type FixReport struct {
	Fixes []Fix
}

// Fix is one applied heuristic.
type Fix struct {
	Rule        string // e.g. "fence-close"
	Line        int    // 1-based line number where the issue was detected
	Description string // short human-readable summary of what was done
}

// Try applies heuristic fixes in a fixed order and returns the possibly
// modified markdown plus a report. parseErr may be nil (the caller will
// decide whether to re-parse); the current implementation ignores it.
func Try(md string, parseErr error) (string, FixReport) {
	report := FixReport{}
	lines := strings.Split(md, "\n")

	// Rules are applied in order; each mutates `lines` and appends to report.
	applyFenceClose(&lines, &report)
	applyFrontmatterTerminator(&lines, &report)
	applyFrontmatterUnquotedColon(&lines, &report)
	applyTrailingNewline(&lines, &report)

	return strings.Join(lines, "\n"), report
}

// --- rule stubs (filled in by later tasks) ---

func applyFenceClose(lines *[]string, report *FixReport) {
	openLine := -1
	for i, ln := range *lines {
		t := strings.TrimSpace(ln)
		if strings.HasPrefix(t, "```") {
			if openLine == -1 {
				openLine = i
			} else {
				openLine = -1
			}
		}
	}
	if openLine == -1 {
		return
	}
	// If the last element is an empty string (trailing newline), insert before it.
	ls := *lines
	if len(ls) > 0 && ls[len(ls)-1] == "" {
		*lines = append(ls[:len(ls)-1], "```", "")
	} else {
		*lines = append(ls, "```")
	}
	report.Fixes = append(report.Fixes, Fix{
		Rule:        "fence-close",
		Line:        openLine + 1,
		Description: "unclosed ``` block → appended closing fence",
	})
}
func applyFrontmatterTerminator(lines *[]string, report *FixReport) {
	if len(*lines) == 0 || strings.TrimSpace((*lines)[0]) != "---" {
		return
	}
	// Look for a closing --- anywhere in the first ~40 lines (frontmatter
	// is short in practice).
	limit := 40
	if limit > len(*lines) {
		limit = len(*lines)
	}
	for i := 1; i < limit; i++ {
		if strings.TrimSpace((*lines)[i]) == "---" {
			return // balanced
		}
	}
	// Insert a terminator after the last non-blank key:value line in the
	// first few lines (scan until first blank line).
	insertAt := 1
	for i := 1; i < len(*lines); i++ {
		if strings.TrimSpace((*lines)[i]) == "" {
			insertAt = i
			break
		}
		insertAt = i + 1
	}
	newLines := append([]string{}, (*lines)[:insertAt]...)
	newLines = append(newLines, "---")
	newLines = append(newLines, (*lines)[insertAt:]...)
	*lines = newLines
	report.Fixes = append(report.Fixes, Fix{
		Rule:        "fm-terminator",
		Line:        1,
		Description: "frontmatter missing terminator → inserted '---'",
	})
}
func applyFrontmatterUnquotedColon(lines *[]string, report *FixReport) {
	// Must start with frontmatter
	if len(*lines) == 0 || strings.TrimSpace((*lines)[0]) != "---" {
		return
	}
	for i := 1; i < len(*lines); i++ {
		t := (*lines)[i]
		if strings.TrimSpace(t) == "---" {
			return // end of frontmatter
		}
		// "key: value" — split on first colon
		colon := strings.Index(t, ":")
		if colon < 0 {
			continue
		}
		key := t[:colon]
		value := strings.TrimSpace(t[colon+1:])
		if value == "" {
			continue
		}
		// Already quoted?
		if (strings.HasPrefix(value, `"`) && strings.HasSuffix(value, `"`)) ||
			(strings.HasPrefix(value, `'`) && strings.HasSuffix(value, `'`)) {
			continue
		}
		// Value itself contains a colon → needs quoting
		if !strings.Contains(value, ":") {
			continue
		}
		(*lines)[i] = key + `: "` + strings.ReplaceAll(value, `"`, `\"`) + `"`
		report.Fixes = append(report.Fixes, Fix{
			Rule:        "fm-unquoted-colon",
			Line:        i + 1,
			Description: "frontmatter value contained ':' → added quotes",
		})
	}
}
func applyTrailingNewline(lines *[]string, report *FixReport) {
	if len(*lines) == 0 {
		return
	}
	last := (*lines)[len(*lines)-1]
	if last == "" {
		return // final "" from trailing \n
	}
	*lines = append(*lines, "")
	report.Fixes = append(report.Fixes, Fix{
		Rule:        "trailing-newline",
		Line:        len(*lines),
		Description: "file did not end with newline → appended",
	})
}
