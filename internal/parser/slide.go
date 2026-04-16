package parser

import (
	"bytes"
	"html/template"
	"regexp"
	"strings"

	"github.com/user/goslide/internal/ir"
	"github.com/yuin/goldmark"
)

var (
	metadataRe = regexp.MustCompile(`^<!--\s*([\w][\w-]*)\s*:\s*(.+?)\s*-->$`)
	regionRe   = regexp.MustCompile(`^<!--\s*(\w+)\s*-->$`)
)

var layoutRegions = map[string][]string{
	"two-column":   {"left", "right"},
	"code-preview": {"code", "preview"},
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
		Index:    index,
		Meta:     meta,
		RawBody:  bodyText,
		BodyHTML: bodyHTML,
		Regions:  regions,
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

	return meta
}

func renderMarkdown(src string) template.HTML {
	md := goldmark.New()
	var buf bytes.Buffer
	if err := md.Convert([]byte(src), &buf); err != nil {
		return template.HTML("<p>Markdown render error: " + err.Error() + "</p>")
	}
	return template.HTML(buf.String())
}
