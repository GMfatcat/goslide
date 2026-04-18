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

func TestResolveTheme_Corporate(t *testing.T) {
	require.Equal(t, "corporate", ResolveTheme("corporate"))
}

func TestResolveTheme_Minimal(t *testing.T) {
	require.Equal(t, "minimal", ResolveTheme("minimal"))
}

func TestResolveTheme_Hacker(t *testing.T) {
	require.Equal(t, "hacker", ResolveTheme("hacker"))
}

func TestResolveAccent_Default(t *testing.T) {
	name := ResolveAccent("", "")
	require.Equal(t, "blue", name)
}

func TestResolveAccent_AllValid(t *testing.T) {
	accents := []string{"blue", "teal", "purple", "coral", "amber", "green", "red", "pink"}
	for _, a := range accents {
		require.Equal(t, a, ResolveAccent(a, "default"))
	}
}

func TestResolveAccent_DefaultForHacker(t *testing.T) {
	require.Equal(t, "green", ResolveAccent("", "hacker"))
}

func TestResolveAccent_DefaultForCorporate(t *testing.T) {
	require.Equal(t, "blue", ResolveAccent("", "corporate"))
}

func TestResolveAccent_ExplicitOverridesThemeDefault(t *testing.T) {
	require.Equal(t, "coral", ResolveAccent("coral", "hacker"))
}

func TestResolveAccent_EmptyTheme(t *testing.T) {
	require.Equal(t, "blue", ResolveAccent("", ""))
}

func TestThemeCSSPath(t *testing.T) {
	require.Equal(t, "themes/dark.css", ThemeCSSPath("dark"))
	require.Equal(t, "themes/default.css", ThemeCSSPath("default"))
	require.Equal(t, "themes/hacker.css", ThemeCSSPath("hacker"))
}

func TestResolveTheme_Dracula(t *testing.T) {
	require.Equal(t, "dracula", ResolveTheme("dracula"))
}

func TestResolveTheme_AllNewThemes(t *testing.T) {
	for _, name := range []string{"midnight", "gruvbox", "solarized", "catppuccin-mocha"} {
		require.Equal(t, name, ResolveTheme(name))
	}
}

func TestResolveAccent_DraculaDefault(t *testing.T) {
	require.Equal(t, "pink", ResolveAccent("", "dracula"))
}

func TestResolveAccent_GruvboxDefault(t *testing.T) {
	require.Equal(t, "amber", ResolveAccent("", "gruvbox"))
}

func TestResolveAccent_SolarizedDefault(t *testing.T) {
	require.Equal(t, "teal", ResolveAccent("", "solarized"))
}
