package pdfexport

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolvePaperSize_KnownPresets(t *testing.T) {
	cases := []struct {
		name   string
		wantW  float64
		wantH  float64
		wantOK bool
	}{
		{"slide-16x9", 20.0, 11.25, true},     // 1920 / 96 = 20 in, 1080 / 96 = 11.25 in
		{"slide-4x3", 16.667, 12.5, true},     // 1600/96=16.666, 1200/96=12.5
		{"a4-landscape", 11.693, 8.268, true}, // 297mm/25.4 = 11.693 in, 210mm/25.4 = 8.267
		{"letter-landscape", 11.0, 8.5, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			w, h, err := ResolvePaperSize(c.name)
			require.NoError(t, err)
			require.InDelta(t, c.wantW, w, 0.01)
			require.InDelta(t, c.wantH, h, 0.01)
		})
	}
}

func TestResolvePaperSize_UnknownError(t *testing.T) {
	_, _, err := ResolvePaperSize("foobar")
	require.Error(t, err)
	require.Contains(t, err.Error(), "unknown paper size")
	require.Contains(t, err.Error(), "slide-16x9")
}

func TestKnownPaperSizes(t *testing.T) {
	names := KnownPaperSizes()
	require.ElementsMatch(t, []string{"slide-16x9", "slide-4x3", "a4-landscape", "letter-landscape"}, names)
}
