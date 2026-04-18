package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/user/goslide/templates"
)

var initTemplate string

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Scaffold a new presentation",
	Args:  cobra.NoArgs,
	RunE:  runInit,
}

func init() {
	initCmd.Flags().StringVarP(&initTemplate, "template", "t", "basic", "template to use: basic, demo, corporate")
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	filename := initTemplate + ".md"
	data, err := templates.FS.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("unknown template %q (available: basic, demo, corporate)", initTemplate)
	}

	outPath := filepath.Join(".", "talk.md")
	if _, err := os.Stat(outPath); err == nil {
		return fmt.Errorf("talk.md already exists in current directory")
	}

	if err := os.WriteFile(outPath, data, 0644); err != nil {
		return err
	}

	fmt.Printf("Created %s from %q template\n", outPath, initTemplate)
	return nil
}
