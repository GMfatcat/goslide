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

func applyFenceClose(lines *[]string, report *FixReport)               {}
func applyFrontmatterTerminator(lines *[]string, report *FixReport)    {}
func applyFrontmatterUnquotedColon(lines *[]string, report *FixReport) {}
func applyTrailingNewline(lines *[]string, report *FixReport)          {}
