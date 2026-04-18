package cli

import (
	"github.com/spf13/cobra"
	"github.com/user/goslide/internal/builder"
)

var buildOutput string

var buildCmd = &cobra.Command{
	Use:   "build <file.md>",
	Short: "Export presentation as self-contained HTML",
	Args:  cobra.ExactArgs(1),
	RunE:  runBuild,
}

func init() {
	buildCmd.Flags().StringVarP(&buildOutput, "output", "o", "", "output file (default: {name}.html)")
	rootCmd.AddCommand(buildCmd)
}

func runBuild(cmd *cobra.Command, args []string) error {
	return builder.Build(builder.Options{
		File:   args[0],
		Output: buildOutput,
	})
}
