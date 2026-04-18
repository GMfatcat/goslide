package renderer

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/GMfatcat/goslide/internal/ir"
	"github.com/GMfatcat/goslide/internal/parser"
	"github.com/stretchr/testify/require"
)

var update = flag.Bool("update", false, "update golden files")

func TestGolden(t *testing.T) {
	entries, err := filepath.Glob("testdata/golden/*.md")
	require.NoError(t, err)
	require.NotEmpty(t, entries, "no golden test inputs found")

	for _, mdPath := range entries {
		name := strings.TrimSuffix(filepath.Base(mdPath), ".md")
		t.Run(name, func(t *testing.T) {
			input, err := os.ReadFile(mdPath)
			require.NoError(t, err)

			pres, err := parser.Parse(input, mdPath)
			require.NoError(t, err)

			valErrs := pres.Validate()
			stderrOut := ir.FormatErrors(mdPath, valErrs)

			stderrPath := "testdata/golden/" + name + ".stderr"
			if *update {
				os.WriteFile(stderrPath, []byte(stderrOut), 0644)
			} else if _, statErr := os.Stat(stderrPath); statErr == nil {
				want, _ := os.ReadFile(stderrPath)
				require.Equal(t, string(want), stderrOut, "stderr mismatch for %s", name)
			}

			if ir.HasErrors(valErrs) {
				return
			}

			html, err := Render(pres)
			require.NoError(t, err)

			htmlPath := "testdata/golden/" + name + ".html"
			if *update {
				os.WriteFile(htmlPath, []byte(html), 0644)
			} else {
				want, err := os.ReadFile(htmlPath)
				require.NoError(t, err, "golden file %s not found; run with -update", htmlPath)
				require.Equal(t, string(want), html, "HTML mismatch for %s", name)
			}
		})
	}
}
