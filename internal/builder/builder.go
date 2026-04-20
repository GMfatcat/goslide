package builder

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/GMfatcat/goslide/internal/config"
	"github.com/GMfatcat/goslide/internal/generate"
	"github.com/GMfatcat/goslide/internal/ir"
	"github.com/GMfatcat/goslide/internal/llm"
	"github.com/GMfatcat/goslide/internal/parser"
	"github.com/GMfatcat/goslide/internal/renderer"
	"github.com/GMfatcat/goslide/web"
)

type Options struct {
	File       string
	Output     string
	Theme      string
	Accent     string
	LLMRefresh bool
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

	cfg, _ := config.Load(filepath.Dir(opts.File))

	// BakeLLM — skip when there's no generate config. Note BakeLLM walks
	// the full presentation and no-ops on slides/components without llm
	// render items, so it's safe to always call when generate is set.
	if cfg != nil && cfg.Generate.BaseURL != "" && cfg.Generate.Model != "" {
		mdDir := filepath.Dir(opts.File)
		cache := llm.NewDiskCache(filepath.Join(mdDir, ".goslide-cache"))
		completer := buildLLMCompleter(cfg)
		if err := BakeLLM(pres, cache, completer, cfg.Generate.Model, mdDir, opts.LLMRefresh); err != nil {
			return err
		}
	}

	html, err := renderer.Render(pres)
	if err != nil {
		return fmt.Errorf("render %s: %w", opts.File, err)
	}

	if cfg != nil && len(cfg.Theme.Overrides) > 0 {
		html = injectThemeOverrides(html, cfg.Theme.Overrides)
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

func injectThemeOverrides(html string, overrides map[string]string) string {
	var sb strings.Builder
	sb.WriteString("<style>:root {")
	for k, v := range overrides {
		sb.WriteString(" --")
		sb.WriteString(k)
		sb.WriteString(": ")
		sb.WriteString(v)
		sb.WriteString(";")
	}
	sb.WriteString(" }</style>")
	return strings.Replace(html, "</head>", sb.String()+"</head>", 1)
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
		{`href="/themes/transitions.css"`, web.ThemeFS, "themes/transitions.css"},
	}

	themeNames := []string{"default", "dark", "corporate", "minimal", "hacker", "dracula", "midnight", "gruvbox", "solarized", "catppuccin-mocha", "ink-wash", "instagram", "western", "pixel", "nord-light", "paper", "catppuccin-latte", "chalk", "synthwave", "forest", "rose", "amoled"}
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

func buildLLMCompleter(cfg *config.Config) llm.Completer {
	apiKey := os.Getenv(firstNonEmpty(cfg.Generate.APIKeyEnv, "OPENAI_API_KEY"))
	timeout := cfg.Generate.Timeout
	if timeout == 0 {
		timeout = 120 * time.Second
	}
	client := generate.NewClient(cfg.Generate.BaseURL, apiKey, timeout)
	return completerAdapter{client}
}

type completerAdapter struct{ c *generate.Client }

func (a completerAdapter) Complete(ctx context.Context, model string, msgs []llm.Message) (string, llm.Usage, error) {
	gmsgs := make([]generate.Message, len(msgs))
	for i, m := range msgs {
		gmsgs[i] = generate.Message{Role: m.Role, Content: m.Content}
	}
	content, usage, err := a.c.Complete(ctx, model, gmsgs)
	return content, llm.Usage(usage), err
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

func inlineFonts(css string) string {
	fonts := []struct {
		url  string
		path string
	}{
		{"url('/fonts/NotoSansTC-Regular.woff2')", "static/fonts/NotoSansTC-Regular.woff2"},
		{"url('/fonts/NotoSansTC-Bold.woff2')", "static/fonts/NotoSansTC-Bold.woff2"},
		{"url('/fonts/JetBrainsMono-Regular.woff2')", "static/fonts/JetBrainsMono-Regular.woff2"},
		{"url('/fonts/PressStart2P-Regular.woff2')", "static/fonts/PressStart2P-Regular.woff2"},
		{"url('/fonts/Rye-Regular.woff2')", "static/fonts/Rye-Regular.woff2"},
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
