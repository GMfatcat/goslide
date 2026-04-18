package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var verbose bool

var rootCmd = &cobra.Command{
	Use:   "goslide",
	Short: "Markdown-driven interactive presentations",
	Long:  "GoSlide renders Markdown files as Reveal.js presentations with live reload.",
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "verbose logging")

	stubs := []struct {
		use, short string
	}{
		{"build <file.md>", "Export presentation as static HTML"},
	}
	for _, s := range stubs {
		s := s
		rootCmd.AddCommand(&cobra.Command{
			Use:   s.use,
			Short: s.short,
			RunE: func(cmd *cobra.Command, args []string) error {
				return fmt.Errorf("not implemented: available in a future release")
			},
		})
	}
}

func Execute(version string) {
	rootCmd.Version = version
	rootCmd.Execute()
}
