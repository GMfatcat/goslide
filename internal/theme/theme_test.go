package theme

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolveTheme_Default(t *testing.T) {
	name := ResolveTheme("")
	require.Equal(t, "default", name)
}

func TestResolveTheme_Known(t *testing.T) {
	require.Equal(t, "dark", ResolveTheme("dark"))
	require.Equal(t, "default", ResolveTheme("default"))
}

func TestResolveTheme_Unknown(t *testing.T) {
	require.Equal(t, "default", ResolveTheme("matrix"))
}

func TestResolveAccent_Default(t *testing.T) {
	name := ResolveAccent("")
	require.Equal(t, "blue", name)
}

func TestResolveAccent_AllValid(t *testing.T) {
	accents := []string{"blue", "teal", "purple", "coral", "amber", "green", "red", "pink"}
	for _, a := range accents {
		require.Equal(t, a, ResolveAccent(a))
	}
}

func TestThemeCSSPath(t *testing.T) {
	require.Equal(t, "themes/dark.css", ThemeCSSPath("dark"))
	require.Equal(t, "themes/default.css", ThemeCSSPath("default"))
}
