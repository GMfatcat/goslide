package theme

var validThemes = map[string]bool{
	"default": true, "dark": true,
	"corporate": true, "minimal": true, "hacker": true,
	"dracula": true, "midnight": true, "gruvbox": true, "solarized": true, "catppuccin-mocha": true,
	"ink-wash": true, "instagram": true, "western": true, "pixel": true,
	"nord-light": true, "paper": true, "catppuccin-latte": true, "chalk": true,
	"synthwave": true, "forest": true, "rose": true, "amoled": true,
}

var themeDefaultAccents = map[string]string{
	"default": "blue", "dark": "blue", "corporate": "blue",
	"minimal": "blue", "hacker": "green",
	"dracula": "pink", "midnight": "blue", "gruvbox": "amber", "solarized": "teal", "catppuccin-mocha": "pink",
	"ink-wash": "teal", "instagram": "pink", "western": "amber", "pixel": "green",
	"nord-light": "teal", "paper": "amber", "catppuccin-latte": "pink", "chalk": "purple",
	"synthwave": "pink", "forest": "green", "rose": "pink", "amoled": "blue",
}

func ResolveTheme(name string) string {
	if name == "" || !validThemes[name] {
		return "default"
	}
	return name
}

func ResolveAccent(accent, themeName string) string {
	if accent != "" {
		return accent
	}
	if def, ok := themeDefaultAccents[themeName]; ok {
		return def
	}
	return "blue"
}

func ThemeCSSPath(name string) string {
	return "themes/" + name + ".css"
}
