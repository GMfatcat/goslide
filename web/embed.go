package web

import "embed"

//go:embed all:static
var StaticFS embed.FS

//go:embed all:themes
var ThemeFS embed.FS

//go:embed all:templates
var TemplateFS embed.FS
