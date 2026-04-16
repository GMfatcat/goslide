package ir

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/agnivade/levenshtein"
)

var (
	validThemes      = map[string]bool{"default": true, "dark": true, "corporate": true, "minimal": true, "hacker": true}
	validAccents     = map[string]bool{"blue": true, "teal": true, "purple": true, "coral": true, "amber": true, "green": true, "red": true, "pink": true}
	validTransitions = map[string]bool{"slide": true, "fade": true, "convex": true, "concave": true, "zoom": true, "none": true}

	phase1Layouts = map[string]bool{"default": true, "title": true, "section": true, "two-column": true, "code-preview": true}
	futureLayouts = map[string]bool{"three-column": true, "image-left": true, "image-right": true, "quote": true, "split-heading": true, "top-bottom": true, "grid-cards": true, "blank": true}

	requiredRegions = map[string][]string{
		"two-column":   {"left", "right"},
		"code-preview": {"code", "preview"},
	}

	futureComponentRe = regexp.MustCompile(`(?m)^~~~(chart(?::\w+)?|mermaid|table|tabs|panel(?::\S+)?|slider|toggle|api|embed(?::\w+)?|card)\s*$`)
	listLineRe        = regexp.MustCompile(`(?m)^[ \t]*[-*+][ \t]|^[ \t]*\d+\.[ \t]`)
)

func (p *Presentation) Validate() []Error {
	var errs []Error

	if p.Meta.Theme != "" && !validThemes[p.Meta.Theme] {
		errs = append(errs, Error{
			Slide: 0, Severity: "error", Code: "unknown-theme",
			Message: fmt.Sprintf("frontmatter: theme %q not recognized", p.Meta.Theme),
		})
	}

	if p.Meta.Accent != "" && !validAccents[p.Meta.Accent] {
		errs = append(errs, Error{
			Slide: 0, Severity: "error", Code: "unknown-accent",
			Message: fmt.Sprintf("frontmatter: accent %q not recognized", p.Meta.Accent),
		})
	}

	if p.Meta.Transition != "" && !validTransitions[p.Meta.Transition] {
		errs = append(errs, Error{
			Slide: 0, Severity: "warning", Code: "unknown-transition",
			Message: fmt.Sprintf("frontmatter: transition %q not recognized (using default)", p.Meta.Transition),
		})
	}

	validSlideNumbers := map[string]bool{"auto": true, "true": true, "false": true}
	if p.Meta.SlideNumber != "" && !validSlideNumbers[p.Meta.SlideNumber] {
		errs = append(errs, Error{
			Slide: 0, Severity: "error", Code: "unknown-slide-number",
			Message: fmt.Sprintf("frontmatter: slide-number %q not recognized (use auto, true, or false)", p.Meta.SlideNumber),
		})
	}

	validSlideNumberFormats := map[string]bool{"total": true, "current": true}
	if p.Meta.SlideNumberFormat != "" && !validSlideNumberFormats[p.Meta.SlideNumberFormat] {
		errs = append(errs, Error{
			Slide: 0, Severity: "error", Code: "unknown-slide-number-format",
			Message: fmt.Sprintf("frontmatter: slide-number-format %q not recognized (use total or current)", p.Meta.SlideNumberFormat),
		})
	}

	for _, slide := range p.Slides {
		errs = append(errs, validateSlide(slide)...)
	}

	p.Warnings = errs
	return errs
}

func validateSlide(s Slide) []Error {
	var errs []Error
	layout := s.Meta.Layout

	if layout != "" && layout != "default" && !phase1Layouts[layout] {
		if futureLayouts[layout] {
			errs = append(errs, Error{
				Slide: s.Index, Severity: "warning", Code: "future-layout",
				Message: fmt.Sprintf("slide %d: layout %q not implemented in Phase 1 (using default)", s.Index, layout),
			})
		} else if best, dist := closestLayout(layout); dist <= 2 {
			errs = append(errs, Error{
				Slide: s.Index, Severity: "error", Code: "typo-suggestion",
				Message: fmt.Sprintf("slide %d: layout %q not recognized", s.Index, layout),
				Hint:    fmt.Sprintf("did you mean %q?", best),
			})
		} else {
			errs = append(errs, Error{
				Slide: s.Index, Severity: "warning", Code: "unknown-layout",
				Message: fmt.Sprintf("slide %d: layout %q not recognized (using default)", s.Index, layout),
			})
		}
	}

	if required, ok := requiredRegions[layout]; ok && phase1Layouts[layout] {
		regionSet := make(map[string]bool, len(s.Regions))
		for _, r := range s.Regions {
			regionSet[r.Name] = true
		}
		for _, name := range required {
			if !regionSet[name] {
				errs = append(errs, Error{
					Slide: s.Index, Severity: "error", Code: "missing-region",
					Message: fmt.Sprintf("slide %d: layout %q but no <!-- %s --> region found", s.Index, layout, name),
				})
			}
		}
	}

	if s.Meta.Fragments && !listLineRe.MatchString(s.RawBody) {
		errs = append(errs, Error{
			Slide: s.Index, Severity: "warning", Code: "fragments-noop",
			Message: fmt.Sprintf("slide %d: fragments enabled but slide has no list", s.Index),
		})
	}

	if matches := futureComponentRe.FindAllString(s.RawBody, -1); len(matches) > 0 {
		for _, m := range matches {
			comp := strings.TrimPrefix(m, "~~~")
			errs = append(errs, Error{
				Slide: s.Index, Severity: "warning", Code: "future-component",
				Message: fmt.Sprintf("slide %d: component %q not implemented in Phase 1 (rendered as code block)", s.Index, comp),
			})
		}
	}

	return errs
}

func closestLayout(name string) (string, int) {
	best := ""
	bestDist := 999
	for layout := range phase1Layouts {
		d := levenshtein.ComputeDistance(name, layout)
		if d < bestDist {
			bestDist = d
			best = layout
		}
	}
	for layout := range futureLayouts {
		d := levenshtein.ComputeDistance(name, layout)
		if d < bestDist {
			bestDist = d
			best = layout
		}
	}
	return best, bestDist
}

func HasErrors(errs []Error) bool {
	for _, e := range errs {
		if e.IsError() {
			return true
		}
	}
	return false
}

func FormatErrors(source string, errs []Error) string {
	if len(errs) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString(source + ":\n")
	errorCount, warnCount := 0, 0
	for _, e := range errs {
		if e.IsError() {
			errorCount++
		} else {
			warnCount++
		}
		fmt.Fprintf(&b, "  %-7s [%s] %s", e.Severity, e.Code, e.Message)
		if e.Hint != "" {
			fmt.Fprintf(&b, " -- %s", e.Hint)
		}
		b.WriteString("\n")
	}
	fmt.Fprintf(&b, "\n%d errors, %d warnings", errorCount, warnCount)
	if errorCount > 0 {
		b.WriteString(" -- refusing to serve")
	}
	b.WriteString("\n")
	return b.String()
}
