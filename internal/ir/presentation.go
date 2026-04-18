package ir

import "html/template"

type Presentation struct {
	Source   string
	Meta     Frontmatter
	Slides   []Slide
	Warnings []Error
}

type Slide struct {
	Index      int
	Meta       SlideMeta
	RawBody    string
	BodyHTML   template.HTML
	Regions    []Region
	Components []Component
}

type Region struct {
	Name string
	HTML template.HTML
}

type Component struct {
	Index       int
	Type        string
	Raw         string
	Params      map[string]any
	ContentHTML string
}

type SlideMeta struct {
	Title         string
	Layout        string
	Transition    string
	Fragments         bool
	FragmentStyle     string
	SlideNumberHidden bool
}

type Frontmatter struct {
	Title         string   `yaml:"title"`
	Author        string   `yaml:"author"`
	Date          string   `yaml:"date"`
	Tags          []string `yaml:"tags"`
	Theme         string   `yaml:"theme"`
	Accent        string   `yaml:"accent"`
	Transition    string   `yaml:"transition"`
	Fragments     bool     `yaml:"fragments"`
	FragmentStyle string   `yaml:"fragment-style"`
	SlideNumber       string `yaml:"slide-number"`
	SlideNumberFormat string `yaml:"slide-number-format"`
}

type Error struct {
	Slide    int
	Severity string
	Code     string
	Message  string
	Hint     string
}

func (e Error) IsError() bool {
	return e.Severity == "error"
}
