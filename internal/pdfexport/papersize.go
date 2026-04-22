package pdfexport

import (
	"fmt"
	"sort"
	"strings"
)

// paperSizes maps a preset name to its width and height in inches.
// chromedp's PrintToPDFParams takes inches, so we normalise here.
var paperSizes = map[string]struct {
	widthIn, heightIn float64
}{
	"slide-16x9":       {1920.0 / 96.0, 1080.0 / 96.0},
	"slide-4x3":        {1600.0 / 96.0, 1200.0 / 96.0},
	"a4-landscape":     {297.0 / 25.4, 210.0 / 25.4},
	"letter-landscape": {11.0, 8.5},
}

// ResolvePaperSize returns width and height in inches for a known preset.
func ResolvePaperSize(name string) (float64, float64, error) {
	p, ok := paperSizes[name]
	if !ok {
		return 0, 0, fmt.Errorf("unknown paper size %q (valid: %s)", name, strings.Join(KnownPaperSizes(), ", "))
	}
	return p.widthIn, p.heightIn, nil
}

// KnownPaperSizes returns the list of valid preset names, sorted.
func KnownPaperSizes() []string {
	out := make([]string, 0, len(paperSizes))
	for k := range paperSizes {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
