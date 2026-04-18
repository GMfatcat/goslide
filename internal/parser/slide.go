package parser

import (
	"bytes"
	"html/template"
	"regexp"
	"strconv"
	"strings"

	"github.com/user/goslide/internal/ir"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
	yaml "gopkg.in/yaml.v3"
)

var (
	metadataRe = regexp.MustCompile(`^<!--\s*([\w][\w-]*)\s*:\s*(.+?)\s*-->$`)
	regionRe   = regexp.MustCompile(`^<!--\s*(\w+)\s*-->$`)
)

var layoutRegions = map[string][]string{
	"two-column":    {"left", "right"},
	"code-preview":  {"code", "preview"},
	"three-column":  {"col1", "col2", "col3"},
	"image-left":    {"image", "text"},
	"image-right":   {"text", "image"},
	"split-heading": {"heading", "body"},
	"top-bottom":    {"top", "bottom"},
}

func parseSlide(index int, raw string, defaults ir.Frontmatter) ir.Slide {
	lines := strings.Split(raw, "\n")

	metaMap := map[string]string{}
	bodyStart := 0
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			bodyStart = i + 1
			continue
		}
		if m := metadataRe.FindStringSubmatch(trimmed); m != nil {
			metaMap[normalizeEnum(m[1])] = m[2]
			bodyStart = i + 1
		} else {
			break
		}
	}

	meta := buildSlideMeta(metaMap, defaults)

	bodyLines := lines[bodyStart:]
	bodyText := strings.Join(bodyLines, "\n")

	var notesHTML template.HTML
	if idx := strings.Index(bodyText, "\n<!-- notes -->\n"); idx != -1 {
		notesRaw := bodyText[idx+len("\n<!-- notes -->\n"):]
		bodyText = bodyText[:idx]
		notesHTML = renderMarkdown(notesRaw)
	} else if strings.HasPrefix(bodyText, "<!-- notes -->\n") {
		notesRaw := bodyText[len("<!-- notes -->\n"):]
		bodyText = ""
		notesHTML = renderMarkdown(notesRaw)
	}

	cleanedBody, components := extractComponents(bodyText)
	for i := range components {
		if strings.HasPrefix(components[i].Type, "panel:") {
			components[i].ContentHTML = string(renderMarkdown(components[i].Raw))
		} else if components[i].Type == "card" {
			parts := strings.SplitN(components[i].Raw, "\n---\n", 2)
			if len(parts) == 2 {
				var summaryParams map[string]any
				if err := yaml.Unmarshal([]byte(parts[0]), &summaryParams); err == nil {
					components[i].Params = summaryParams
				}
				components[i].ContentHTML = string(renderMarkdown(parts[1]))
			}
		} else if components[i].Type == "embed:html" {
			components[i].ContentHTML = components[i].Raw
		}
	}
	bodyLines = strings.Split(cleanedBody, "\n")

	validRegions := layoutRegions[meta.Layout]
	validSet := make(map[string]bool, len(validRegions))
	for _, r := range validRegions {
		validSet[r] = true
	}

	var regions []ir.Region
	var mainLines []string
	currentRegion := ""
	regionContent := map[string][]string{}
	var regionOrder []string

	for _, line := range bodyLines {
		trimmed := strings.TrimSpace(line)
		if m := regionRe.FindStringSubmatch(trimmed); m != nil {
			name := strings.ToLower(m[1])
			if validSet[name] {
				currentRegion = name
				if _, exists := regionContent[name]; !exists {
					regionOrder = append(regionOrder, name)
					regionContent[name] = []string{}
				}
				continue
			}
		}
		if currentRegion == "" {
			mainLines = append(mainLines, line)
		} else {
			regionContent[currentRegion] = append(regionContent[currentRegion], line)
		}
	}

	mainBody := strings.Join(mainLines, "\n")
	bodyHTML := renderMarkdown(mainBody)

	for _, name := range regionOrder {
		regionText := strings.Join(regionContent[name], "\n")
		regions = append(regions, ir.Region{
			Name: name,
			HTML: renderMarkdown(regionText),
		})
	}

	return ir.Slide{
		Index:      index,
		Meta:       meta,
		RawBody:    bodyText,
		BodyHTML:   bodyHTML,
		Regions:    regions,
		Components: components,
		Notes:      notesHTML,
	}
}

func buildSlideMeta(metaMap map[string]string, defaults ir.Frontmatter) ir.SlideMeta {
	meta := ir.SlideMeta{
		Layout:        "default",
		Transition:    defaults.Transition,
		Fragments:     defaults.Fragments,
		FragmentStyle: defaults.FragmentStyle,
	}

	if v, ok := metaMap["layout"]; ok {
		meta.Layout = normalizeEnum(v)
	}
	if v, ok := metaMap["transition"]; ok {
		meta.Transition = normalizeEnum(v)
	}
	if v, ok := metaMap["fragments"]; ok {
		meta.Fragments = strings.ToLower(strings.TrimSpace(v)) == "true"
	}
	if v, ok := metaMap["fragment-style"]; ok {
		meta.FragmentStyle = normalizeEnum(v)
	}
	if v, ok := metaMap["slide-number"]; ok {
		meta.SlideNumberHidden = strings.ToLower(strings.TrimSpace(v)) == "false"
	}
	if v, ok := metaMap["columns"]; ok {
		n, err := strconv.Atoi(strings.TrimSpace(v))
		if err == nil {
			meta.Columns = n
		}
	}

	return meta
}

func renderMarkdown(src string) template.HTML {
	md := goldmark.New(
		goldmark.WithExtensions(extension.Table),
		goldmark.WithRendererOptions(html.WithUnsafe()),
	)
	var buf bytes.Buffer
	if err := md.Convert([]byte(src), &buf); err != nil {
		return template.HTML("<p>Markdown render error: " + err.Error() + "</p>")
	}
	return template.HTML(buf.String())
}
