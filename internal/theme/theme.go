package theme

var validThemes = map[string]bool{"default": true, "dark": true}
var defaultAccent = "blue"

func ResolveTheme(name string) string {
	if name == "" || !validThemes[name] {
		return "default"
	}
	return name
}

func ResolveAccent(name string) string {
	if name == "" {
		return defaultAccent
	}
	return name
}

func ThemeCSSPath(name string) string {
	return "themes/" + name + ".css"
}
