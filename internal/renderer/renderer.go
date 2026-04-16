package renderer

import (
	"bytes"
	"html/template"
	"io/fs"

	"github.com/user/goslide/internal/ir"
	"github.com/user/goslide/internal/theme"
	"github.com/user/goslide/web"
)

type templateData struct {
	Title       string
	Theme       string
	Accent      string
	Transition  string
	SlideNumber string
	Slides      []ir.Slide
}

func Render(pres *ir.Presentation) (string, error) {
	tmplFS, err := fs.Sub(web.TemplateFS, "templates")
	if err != nil {
		return "", err
	}

	tmpl, err := template.New("slide.html").ParseFS(tmplFS, "slide.html")
	if err != nil {
		return "", err
	}

	data := templateData{
		Title:       pres.Meta.Title,
		Theme:       theme.ResolveTheme(pres.Meta.Theme),
		Accent:      theme.ResolveAccent(pres.Meta.Accent),
		Transition:  resolveTransition(pres.Meta.Transition),
		SlideNumber: pres.Meta.SlideNumber,
		Slides:      pres.Slides,
	}

	if data.Title == "" {
		data.Title = "GoSlide"
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func resolveTransition(t string) string {
	if t == "" {
		return "slide"
	}
	return t
}
