package theme

var validThemes = map[string]bool{
	"default": true, "dark": true,
	"corporate": true, "minimal": true, "hacker": true,
}

var themeDefaultAccents = map[string]string{
	"default": "blue", "dark": "blue", "corporate": "blue",
	"minimal": "blue", "hacker": "green",
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
