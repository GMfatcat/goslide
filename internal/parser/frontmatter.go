package parser

import (
	"github.com/user/goslide/internal/ir"
	"gopkg.in/yaml.v3"
)

func parseFrontmatter(raw string) (ir.Frontmatter, error) {
	var fm ir.Frontmatter
	if raw == "" {
		return fm, nil
	}
	if err := yaml.Unmarshal([]byte(raw), &fm); err != nil {
		return fm, err
	}
	fm.Theme = normalizeEnum(fm.Theme)
	fm.Accent = normalizeEnum(fm.Accent)
	fm.Transition = normalizeEnum(fm.Transition)
	fm.FragmentStyle = normalizeEnum(fm.FragmentStyle)
	fm.SlideNumber = normalizeEnum(fm.SlideNumber)
	return fm, nil
}
