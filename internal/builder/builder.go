package builder

import (
	"encoding/base64"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/user/goslide/internal/ir"
	"github.com/user/goslide/internal/parser"
	"github.com/user/goslide/internal/renderer"
	"github.com/user/goslide/web"
)

type Options struct {
	File   string
	Output string
	Theme  string
	Accent string
}

func Build(opts Options) error {
	data, err := os.ReadFile(opts.File)
	if err != nil {
		return fmt.Errorf("read %s: %w", opts.File, err)
	}

	pres, err := parser.Parse(data, opts.File)
	if err != nil {
		return fmt.Errorf("parse %s: %w", opts.File, err)
	}

	if opts.Theme != "" {
		pres.Meta.Theme = opts.Theme
	}
	if opts.Accent != "" {
		pres.Meta.Accent = opts.Accent
	}

	valErrs := pres.Validate()
	if len(valErrs) > 0 {
		fmt.Fprint(os.Stderr, ir.FormatErrors(opts.File, valErrs))
	}
	if ir.HasErrors(valErrs) {
		return fmt.Errorf("validation failed for %s", opts.File)
	}

	html, err := renderer.Render(pres)
	if err != nil {
		return fmt.Errorf("render %s: %w", opts.File, err)
	}

	html = inlineAssets(html)
	html = addStaticMode(html)

	output := opts.Output
	if output == "" {
		base := strings.TrimSuffix(filepath.Base(opts.File), ".md")
		output = base + ".html"
	}

	if err := os.WriteFile(output, []byte(html), 0644); err != nil {
		return fmt.Errorf("write %s: %w", output, err)
	}

	fmt.Printf("Built %s → %s\n", opts.File, output)
	return nil
}

func addStaticMode(html string) string {
	return strings.Replace(html, "<body ", "<body data-mode=\"static\" ", 1)
}

func inlineAssets(html string) string {
	cssReplacements := []struct {
		href string
		fsys fs.FS
		path string
	}{
		{`href="/static/reveal/reveal.css"`, web.StaticFS, "static/reveal/reveal.css"},
		{`href="/themes/tokens.css"`, web.ThemeFS, "themes/tokens.css"},
		{`href="/themes/layouts.css"`, web.ThemeFS, "themes/layouts.css"},
	}

	themeNames := []string{"default", "dark", "corporate", "minimal", "hacker"}
	for _, t := range themeNames {
		cssReplacements = append(cssReplacements, struct {
			href string
			fsys fs.FS
			path string
		}{
			fmt.Sprintf(`href="/themes/%s.css"`, t),
			web.ThemeFS,
			fmt.Sprintf("themes/%s.css", t),
		})
	}

	for _, cf := range cssReplacements {
		content, err := fs.ReadFile(cf.fsys, cf.path)
		if err != nil {
			continue
		}
		css := inlineFonts(string(content))
		old := fmt.Sprintf(`<link rel="stylesheet" %s>`, cf.href)
		html = strings.Replace(html, old, "<style>"+css+"</style>", 1)
	}

	jsReplacements := []struct {
		src  string
		fsys fs.FS
		path string
	}{
		{`src="/static/chartjs/chart.min.js"`, web.StaticFS, "static/chartjs/chart.min.js"},
		{`src="/static/mermaid/mermaid.min.js"`, web.StaticFS, "static/mermaid/mermaid.min.js"},
		{`src="/static/reveal/reveal.js"`, web.StaticFS, "static/reveal/reveal.js"},
		{`src="/static/reveal/plugin-notes.js"`, web.StaticFS, "static/reveal/plugin-notes.js"},
		{`src="/static/runtime.js"`, web.StaticFS, "static/runtime.js"},
		{`src="/static/reactive.js"`, web.StaticFS, "static/reactive.js"},
		{`src="/static/components.js"`, web.StaticFS, "static/components.js"},
	}

	for _, jf := range jsReplacements {
		content, err := fs.ReadFile(jf.fsys, jf.path)
		if err != nil {
			continue
		}
		old := fmt.Sprintf(`<script %s></script>`, jf.src)
		html = strings.Replace(html, old, "<script>"+string(content)+"</script>", 1)
	}

	return html
}

func inlineFonts(css string) string {
	fonts := []struct {
		url  string
		path string
	}{
		{"url('/fonts/NotoSansTC-Regular.woff2')", "static/fonts/NotoSansTC-Regular.woff2"},
		{"url('/fonts/NotoSansTC-Bold.woff2')", "static/fonts/NotoSansTC-Bold.woff2"},
		{"url('/fonts/JetBrainsMono-Regular.woff2')", "static/fonts/JetBrainsMono-Regular.woff2"},
	}

	for _, f := range fonts {
		data, err := fs.ReadFile(web.StaticFS, f.path)
		if err != nil {
			continue
		}
		b64 := base64.StdEncoding.EncodeToString(data)
		dataURI := fmt.Sprintf("url('data:font/woff2;base64,%s')", b64)
		css = strings.Replace(css, f.url, dataURI, 1)
	}

	return css
}
