package pdfexport

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/GMfatcat/goslide/internal/builder"
)

// Options configures one Export call.
type Options struct {
	File       string   // input .md path (required)
	Output     string   // output .pdf path (defaults to <name>.pdf beside File)
	PaperSize  string   // preset name (required; see papersize.go)
	ShowNotes  bool     // append speaker notes under each slide
	Theme      string   // optional theme override (forwarded to builder.Build)
	Accent     string   // optional accent override
	ChromePath string   // explicit Chrome binary (required)
	Launcher   Launcher // chromedp-backed in production; fake in tests
}

// LaunchRequest is what Export hands to a Launcher after resolving the
// Options. It decouples the orchestrator from chromedp's concrete API.
type LaunchRequest struct {
	ChromePath    string
	URL           string // file:// URL of the built HTML, with ?print-pdf[&showNotes=true]
	PaperWidthIn  float64
	PaperHeightIn float64
	ShowNotes     bool
}

// Launcher produces PDF bytes from a LaunchRequest.
type Launcher interface {
	Launch(ctx context.Context, req LaunchRequest) ([]byte, error)
}

// Export runs the full build → launch → write pipeline.
func Export(opts Options) error {
	if opts.File == "" {
		return errors.New("pdfexport: File is required")
	}
	if opts.Launcher == nil {
		return errors.New("pdfexport: Launcher is required")
	}
	if opts.ChromePath == "" {
		return errors.New("pdfexport: ChromePath is required")
	}

	widthIn, heightIn, err := ResolvePaperSize(opts.PaperSize)
	if err != nil {
		return err
	}

	output := opts.Output
	if output == "" {
		base := strings.TrimSuffix(filepath.Base(opts.File), filepath.Ext(opts.File))
		output = filepath.Join(filepath.Dir(opts.File), base+".pdf")
	}

	// Stage the static HTML in a temp dir so we don't pollute the project.
	tmpDir, err := os.MkdirTemp("", "goslide-pdf-")
	if err != nil {
		return fmt.Errorf("pdfexport: make temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	htmlPath := filepath.Join(tmpDir, "deck.html")
	if err := builder.Build(builder.Options{
		File:   opts.File,
		Output: htmlPath,
		Theme:  opts.Theme,
		Accent: opts.Accent,
	}); err != nil {
		return fmt.Errorf("pdfexport: build: %w", err)
	}

	url := fileURL(htmlPath) + "?print-pdf"
	if opts.ShowNotes {
		url += "&showNotes=true"
	}

	pdfBytes, err := opts.Launcher.Launch(context.Background(), LaunchRequest{
		ChromePath:    opts.ChromePath,
		URL:           url,
		PaperWidthIn:  widthIn,
		PaperHeightIn: heightIn,
		ShowNotes:     opts.ShowNotes,
	})
	if err != nil {
		return fmt.Errorf("pdfexport: launch: %w", err)
	}
	if len(pdfBytes) == 0 {
		return errors.New("pdfexport: Chrome produced empty PDF")
	}

	if err := os.WriteFile(output, pdfBytes, 0644); err != nil {
		return fmt.Errorf("pdfexport: write output: %w", err)
	}
	return nil
}

// fileURL converts an absolute local path to a file:// URL. Windows
// paths use forward slashes and a leading "/" for the drive.
func fileURL(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}
	abs = strings.ReplaceAll(abs, "\\", "/")
	if len(abs) > 0 && abs[0] != '/' {
		abs = "/" + abs
	}
	return "file://" + abs
}
