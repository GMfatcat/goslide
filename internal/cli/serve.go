package cli

import (
	"github.com/spf13/cobra"
	"github.com/GMfatcat/goslide/internal/server"
)

var (
	port    int
	theme   string
	accent  string
	noOpen  bool
	noWatch bool
)

var serveCmd = &cobra.Command{
	Use:   "serve <file.md>",
	Short: "Serve a markdown file as a presentation",
	Args:  cobra.ExactArgs(1),
	RunE:  runServe,
}

func init() {
	serveCmd.Flags().IntVarP(&port, "port", "p", 3000, "port number")
	serveCmd.Flags().StringVarP(&theme, "theme", "t", "", "override theme")
	serveCmd.Flags().StringVarP(&accent, "accent", "a", "", "override accent color")
	serveCmd.Flags().BoolVar(&noOpen, "no-open", false, "don't auto-open browser")
	serveCmd.Flags().BoolVar(&noWatch, "no-watch", false, "disable live reload")
	rootCmd.AddCommand(serveCmd)
}

func runServe(cmd *cobra.Command, args []string) error {
	return server.Run(server.Options{
		File:    args[0],
		Port:    port,
		Theme:   theme,
		Accent:  accent,
		NoOpen:  noOpen,
		NoWatch: noWatch,
		Verbose: verbose,
	})
}
