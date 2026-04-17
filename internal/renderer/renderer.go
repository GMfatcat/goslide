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
	Title             string
	Theme             string
	Accent            string
	Transition        string
	SlideNumber       string
	SlideNumberFormat string
	Slides            []ir.Slide
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

	resolvedTheme := theme.ResolveTheme(pres.Meta.Theme)
	data := templateData{
		Title:             pres.Meta.Title,
		Theme:             resolvedTheme,
		Accent:            theme.ResolveAccent(pres.Meta.Accent, resolvedTheme),
		Transition:        resolveTransition(pres.Meta.Transition),
		SlideNumber:       pres.Meta.SlideNumber,
		SlideNumberFormat: resolveSlideNumberFormat(pres.Meta.SlideNumberFormat),
		Slides:            pres.Slides,
	}

	if data.Title == "" {
		data.Title = "GoSlide"
	}

	for i := range data.Slides {
		s := &data.Slides[i]
		s.BodyHTML = template.HTML(renderComponents(string(s.BodyHTML), *s))
		for j := range s.Regions {
			s.Regions[j].HTML = template.HTML(renderComponents(string(s.Regions[j].HTML), *s))
		}
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func resolveSlideNumberFormat(f string) string {
	if f == "" {
		return "total"
	}
	return f
}

func resolveTransition(t string) string {
	if t == "" {
		return "slide"
	}
	return t
}
