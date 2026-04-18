package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/GMfatcat/goslide/internal/parser"
)

var listCmd = &cobra.Command{
	Use:   "list [directory]",
	Short: "List presentations in a directory",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	dir := "."
	if len(args) > 0 {
		dir = args[0]
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read directory %s: %w", dir, err)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "FILE\tTITLE\tTHEME\tSLIDES")
	fmt.Fprintln(w, "----\t-----\t-----\t------")

	found := false
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		pres, err := parser.Parse(data, path)
		if err != nil {
			fmt.Fprintf(w, "%s\t(parse error)\t-\t-\n", entry.Name())
			continue
		}

		title := pres.Meta.Title
		if title == "" {
			title = "(untitled)"
		}
		theme := pres.Meta.Theme
		if theme == "" {
			theme = "default"
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%d\n", entry.Name(), title, theme, len(pres.Slides))
		found = true
	}

	if !found {
		fmt.Println("No .md files found in", dir)
		return nil
	}

	w.Flush()
	return nil
}
