package cli

import (
	"fmt"

	"github.com/GMfatcat/goslide/internal/pdfexport"
	"github.com/spf13/cobra"
)

var (
	exportPDFOutput    string
	exportPDFPaperSize string
	exportPDFNotes     bool
	exportPDFTheme     string
	exportPDFAccent    string
)

func newExportPDFCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export-pdf <file.md>",
		Short: "Export the presentation as a PDF (requires Chrome/Edge/Chromium)",
		Args:  cobra.ExactArgs(1),
		RunE:  runExportPDF,
	}
	cmd.Flags().StringVarP(&exportPDFOutput, "output", "o", "", "output PDF path (default: {name}.pdf)")
	cmd.Flags().StringVar(&exportPDFPaperSize, "paper-size", "slide-16x9", "paper size preset (slide-16x9, slide-4x3, a4-landscape, letter-landscape)")
	cmd.Flags().BoolVar(&exportPDFNotes, "notes", false, "include speaker notes beneath each slide")
	cmd.Flags().StringVarP(&exportPDFTheme, "theme", "t", "", "override theme")
	cmd.Flags().StringVarP(&exportPDFAccent, "accent", "a", "", "override accent color")
	return cmd
}

func init() {
	rootCmd.AddCommand(newExportPDFCmd())
}

func runExportPDF(cmd *cobra.Command, args []string) error {
	chromePath, err := pdfexport.FindChrome()
	if err != nil {
		return err
	}
	if err := pdfexport.Export(pdfexport.Options{
		File:       args[0],
		Output:     exportPDFOutput,
		PaperSize:  exportPDFPaperSize,
		ShowNotes:  exportPDFNotes,
		Theme:      exportPDFTheme,
		Accent:     exportPDFAccent,
		ChromePath: chromePath,
		Launcher:   pdfexport.NewChromedpLauncher(),
	}); err != nil {
		return err
	}
	out := exportPDFOutput
	if out == "" {
		out = "(default <name>.pdf)"
	}
	fmt.Printf("Exported %s → %s\n", args[0], out)
	return nil
}
