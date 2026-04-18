package parser

import (
	"github.com/GMfatcat/goslide/internal/ir"
)

func Parse(raw []byte, source string) (*ir.Presentation, error) {
	fmRaw, slideRaws := splitRaw(raw)

	fm, err := parseFrontmatter(fmRaw)
	if err != nil {
		return nil, err
	}

	slides := make([]ir.Slide, len(slideRaws))
	for i, s := range slideRaws {
		slides[i] = parseSlide(i+1, s, fm)
	}

	return &ir.Presentation{
		Source: source,
		Meta:   fm,
		Slides: slides,
	}, nil
}
