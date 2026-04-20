package llm

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRender_ReplacesDataPlaceholder(t *testing.T) {
	got := Render("Summarise {{data}} please", []byte(`{"x":1}`))
	require.Equal(t, `Summarise {"x":1} please`, got)
}

func TestRender_NoPlaceholder(t *testing.T) {
	got := Render("No variables here", []byte(`{"x":1}`))
	require.Equal(t, "No variables here", got)
}

func TestRender_EmptyData(t *testing.T) {
	got := Render("Summarise {{data}}", nil)
	require.Equal(t, "Summarise ", got)
}

func TestRender_MultiplePlaceholders(t *testing.T) {
	got := Render("{{data}} and again {{data}}", []byte(`X`))
	require.Equal(t, "X and again X", got)
}
