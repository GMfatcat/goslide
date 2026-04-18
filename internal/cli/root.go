package cli

import (
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
}

func Execute(version string) {
	rootCmd.Version = version
	rootCmd.Execute()
}
