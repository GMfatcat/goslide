package ir

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func findError(errs []Error, code string) *Error {
	for i := range errs {
		if errs[i].Code == code {
			return &errs[i]
		}
	}
	return nil
}

func TestValidate_ValidPresentation(t *testing.T) {
	p := &Presentation{
		Meta: Frontmatter{Theme: "dark", Accent: "teal", Transition: "fade"},
		Slides: []Slide{
			{Index: 1, Meta: SlideMeta{Layout: "default"}},
		},
	}
	errs := p.Validate()
	for _, e := range errs {
		require.NotEqual(t, "error", e.Severity, "unexpected error: %s", e.Message)
	}
}

func TestValidate_UnknownTheme(t *testing.T) {
	p := &Presentation{
		Meta:   Frontmatter{Theme: "matrix"},
		Slides: []Slide{{Index: 1, Meta: SlideMeta{Layout: "default"}}},
	}
	errs := p.Validate()
	e := findError(errs, "unknown-theme")
	require.NotNil(t, e)
	require.Equal(t, "error", e.Severity)
}

func TestValidate_UnknownAccent(t *testing.T) {
	p := &Presentation{
		Meta:   Frontmatter{Accent: "neon"},
		Slides: []Slide{{Index: 1, Meta: SlideMeta{Layout: "default"}}},
	}
	errs := p.Validate()
	e := findError(errs, "unknown-accent")
	require.NotNil(t, e)
	require.Equal(t, "error", e.Severity)
}

func TestValidate_UnknownTransition(t *testing.T) {
	p := &Presentation{
		Meta:   Frontmatter{Transition: "swirl"},
		Slides: []Slide{{Index: 1, Meta: SlideMeta{Layout: "default"}}},
	}
	errs := p.Validate()
	e := findError(errs, "unknown-transition")
	require.NotNil(t, e)
	require.Equal(t, "warning", e.Severity)
}

func TestValidate_TypoLayout_Distance1(t *testing.T) {
	p := &Presentation{
		Slides: []Slide{{Index: 3, Meta: SlideMeta{Layout: "two-colum"}}},
	}
	errs := p.Validate()
	e := findError(errs, "typo-suggestion")
	require.NotNil(t, e)
	require.Equal(t, "error", e.Severity)
	require.Equal(t, 3, e.Slide)
	require.Contains(t, e.Hint, "two-column")
}

func TestValidate_TypoLayout_Distance2(t *testing.T) {
	p := &Presentation{
		Slides: []Slide{{Index: 1, Meta: SlideMeta{Layout: "two-colmn"}}},
	}
	errs := p.Validate()
	e := findError(errs, "typo-suggestion")
	require.NotNil(t, e)
	require.Equal(t, "error", e.Severity)
}

func TestValidate_TypoLayout_Distance3_NotTypo(t *testing.T) {
	p := &Presentation{
		Slides: []Slide{{Index: 1, Meta: SlideMeta{Layout: "two-xyz"}}},
	}
	errs := p.Validate()
	e := findError(errs, "typo-suggestion")
	require.Nil(t, e)
	e = findError(errs, "unknown-layout")
	require.NotNil(t, e)
	require.Equal(t, "warning", e.Severity)
}

func TestValidate_FutureLayout(t *testing.T) {
	// grid-cards is now a known layout; verify no future-layout warning is emitted
	p := &Presentation{
		Slides: []Slide{{Index: 4, Meta: SlideMeta{Layout: "grid-cards"}}},
	}
	errs := p.Validate()
	e := findError(errs, "future-layout")
	require.Nil(t, e)
}

func TestValidate_FutureComponent(t *testing.T) {
	p := &Presentation{
		Slides: []Slide{{Index: 2, Meta: SlideMeta{Layout: "default"}, RawBody: "# Title\n\n~~~chart:bar\ntitle: Test\n~~~\n"}},
	}
	errs := p.Validate()
	e := findError(errs, "future-component")
	require.NotNil(t, e)
	require.Equal(t, "warning", e.Severity)
}

func TestValidate_FragmentsNoop(t *testing.T) {
	p := &Presentation{
		Slides: []Slide{{Index: 1, Meta: SlideMeta{Layout: "default", Fragments: true}, RawBody: "# Title\n\nJust a paragraph.\n"}},
	}
	errs := p.Validate()
	e := findError(errs, "fragments-noop")
	require.NotNil(t, e)
	require.Equal(t, "warning", e.Severity)
}

func TestValidate_FragmentsWithList_NoWarning(t *testing.T) {
	p := &Presentation{
		Slides: []Slide{{Index: 1, Meta: SlideMeta{Layout: "default", Fragments: true}, RawBody: "# Title\n\n- A\n- B\n"}},
	}
	errs := p.Validate()
	e := findError(errs, "fragments-noop")
	require.Nil(t, e)
}

func TestValidate_MissingRegion(t *testing.T) {
	p := &Presentation{
		Slides: []Slide{{Index: 5, Meta: SlideMeta{Layout: "two-column"}, Regions: []Region{{Name: "left", HTML: "content"}}}},
	}
	errs := p.Validate()
	e := findError(errs, "missing-region")
	require.NotNil(t, e)
	require.Equal(t, "error", e.Severity)
	require.Contains(t, e.Message, "right")
}

func TestValidate_TwoColumnComplete_NoError(t *testing.T) {
	p := &Presentation{
		Slides: []Slide{{
			Index: 1, Meta: SlideMeta{Layout: "two-column"},
			Regions: []Region{{Name: "left", HTML: "L"}, {Name: "right", HTML: "R"}},
		}},
	}
	errs := p.Validate()
	e := findError(errs, "missing-region")
	require.Nil(t, e)
}

func TestValidate_BatchesAllErrors(t *testing.T) {
	p := &Presentation{
		Meta: Frontmatter{Theme: "invalid", Accent: "neon"},
		Slides: []Slide{
			{Index: 1, Meta: SlideMeta{Layout: "two-colum"}},
			{Index: 2, Meta: SlideMeta{Layout: "grid-cards"}},
		},
	}
	errs := p.Validate()
	require.GreaterOrEqual(t, len(errs), 3)
}

func TestValidate_EmptyThemeAndAccent_NoError(t *testing.T) {
	p := &Presentation{
		Meta:   Frontmatter{},
		Slides: []Slide{{Index: 1, Meta: SlideMeta{Layout: "default"}}},
	}
	errs := p.Validate()
	for _, e := range errs {
		require.NotEqual(t, "error", e.Severity, "unexpected error: %s %s", e.Code, e.Message)
	}
}

func TestValidate_UnknownSlideNumber(t *testing.T) {
	p := &Presentation{
		Meta:   Frontmatter{SlideNumber: "maybe"},
		Slides: []Slide{{Index: 1, Meta: SlideMeta{Layout: "default"}}},
	}
	errs := p.Validate()
	e := findError(errs, "unknown-slide-number")
	require.NotNil(t, e)
	require.Equal(t, "error", e.Severity)
}

func TestValidate_ValidSlideNumber(t *testing.T) {
	for _, v := range []string{"auto", "true", "false"} {
		p := &Presentation{
			Meta:   Frontmatter{SlideNumber: v},
			Slides: []Slide{{Index: 1, Meta: SlideMeta{Layout: "default"}}},
		}
		errs := p.Validate()
		e := findError(errs, "unknown-slide-number")
		require.Nil(t, e, "should accept slide-number=%q", v)
	}
}

func TestValidate_UnknownSlideNumberFormat(t *testing.T) {
	p := &Presentation{
		Meta:   Frontmatter{SlideNumberFormat: "fancy"},
		Slides: []Slide{{Index: 1, Meta: SlideMeta{Layout: "default"}}},
	}
	errs := p.Validate()
	e := findError(errs, "unknown-slide-number-format")
	require.NotNil(t, e)
	require.Equal(t, "error", e.Severity)
}

func TestValidate_ValidSlideNumberFormat(t *testing.T) {
	for _, v := range []string{"total", "current"} {
		p := &Presentation{
			Meta:   Frontmatter{SlideNumberFormat: v},
			Slides: []Slide{{Index: 1, Meta: SlideMeta{Layout: "default"}}},
		}
		errs := p.Validate()
		e := findError(errs, "unknown-slide-number-format")
		require.Nil(t, e, "should accept slide-number-format=%q", v)
	}
}

func TestValidate_ThreeColumnLayout_Valid(t *testing.T) {
	p := &Presentation{
		Slides: []Slide{{Index: 1, Meta: SlideMeta{Layout: "three-column"},
			Regions: []Region{{Name: "col1", HTML: "a"}, {Name: "col2", HTML: "b"}, {Name: "col3", HTML: "c"}}}},
	}
	errs := p.Validate()
	e := findError(errs, "future-layout")
	require.Nil(t, e)
	e = findError(errs, "missing-region")
	require.Nil(t, e)
}

func TestValidate_ThreeColumnMissingRegion(t *testing.T) {
	p := &Presentation{
		Slides: []Slide{{Index: 1, Meta: SlideMeta{Layout: "three-column"},
			Regions: []Region{{Name: "col1", HTML: "a"}}}},
	}
	errs := p.Validate()
	e := findError(errs, "missing-region")
	require.NotNil(t, e)
}

func TestValidate_QuoteLayout_Valid(t *testing.T) {
	p := &Presentation{
		Slides: []Slide{{Index: 1, Meta: SlideMeta{Layout: "quote"}}},
	}
	errs := p.Validate()
	e := findError(errs, "future-layout")
	require.Nil(t, e)
}

func TestValidate_BlankLayout_Valid(t *testing.T) {
	p := &Presentation{
		Slides: []Slide{{Index: 1, Meta: SlideMeta{Layout: "blank"}}},
	}
	errs := p.Validate()
	e := findError(errs, "future-layout")
	require.Nil(t, e)
}

func TestValidate_GridCardsIsKnown(t *testing.T) {
	// grid-cards has been promoted to a known layout; no future-layout warning expected
	p := &Presentation{
		Slides: []Slide{{Index: 1, Meta: SlideMeta{Layout: "grid-cards"}}},
	}
	errs := p.Validate()
	e := findError(errs, "future-layout")
	require.Nil(t, e)
}

func TestValidate_ImageLeftRequiredRegions(t *testing.T) {
	p := &Presentation{
		Slides: []Slide{{Index: 1, Meta: SlideMeta{Layout: "image-left"},
			Regions: []Region{{Name: "image", HTML: "img"}}}},
	}
	errs := p.Validate()
	e := findError(errs, "missing-region")
	require.NotNil(t, e)
	require.Contains(t, e.Message, "text")
}

func TestValidate_TopBottomComplete(t *testing.T) {
	p := &Presentation{
		Slides: []Slide{{Index: 1, Meta: SlideMeta{Layout: "top-bottom"},
			Regions: []Region{{Name: "top", HTML: "t"}, {Name: "bottom", HTML: "b"}}}},
	}
	errs := p.Validate()
	e := findError(errs, "missing-region")
	require.Nil(t, e)
}
