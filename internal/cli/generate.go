package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/GMfatcat/goslide/internal/config"
	"github.com/GMfatcat/goslide/internal/generate"
	"github.com/spf13/cobra"
)

func newGenerateCmd() *cobra.Command {
	var (
		output     string
		force      bool
		model      string
		baseURL    string
		apiKeyEnv  string
		dumpPrompt bool
	)

	cmd := &cobra.Command{
		Use:   "generate [topic | prompt.md]",
		Short: "Generate a GoSlide presentation via an LLM",
		Long: "Call an OpenAI-compatible LLM endpoint to produce a GoSlide Markdown file.\n" +
			"Pass a topic string for simple mode, or a path to a prompt.md file for advanced mode.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if dumpPrompt {
				fmt.Fprint(cmd.OutOrStdout(), generate.SystemPrompt())
				return nil
			}
			if len(args) != 1 {
				return fmt.Errorf("exactly one argument required (topic or prompt.md path)")
			}

			in, err := inputFromArg(args[0])
			if err != nil {
				return err
			}

			cfg, err := config.Load(".")
			if err != nil {
				return err
			}

			opts, err := generate.Resolve(cfg, generate.Flags{
				BaseURL:   baseURL,
				Model:     model,
				APIKeyEnv: apiKeyEnv,
				Output:    output,
				Force:     force,
			}, in)
			if err != nil {
				return err
			}

			return generate.Run(context.Background(), opts)
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "", "output path (default talk.md)")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "overwrite existing output file")
	cmd.Flags().StringVar(&model, "model", "", "override generate.model from goslide.yaml")
	cmd.Flags().StringVar(&baseURL, "base-url", "", "override generate.base_url from goslide.yaml")
	cmd.Flags().StringVar(&apiKeyEnv, "api-key-env", "", "env var name holding the API key (default OPENAI_API_KEY)")
	cmd.Flags().BoolVar(&dumpPrompt, "dump-prompt", false, "print the built-in system prompt and exit")

	return cmd
}

// inputFromArg decides simple vs advanced mode: if arg resolves to an
// existing file, parse it; otherwise treat it as a topic string.
func inputFromArg(arg string) (generate.Input, error) {
	if fi, err := os.Stat(arg); err == nil && !fi.IsDir() {
		return generate.ParsePromptFile(arg)
	}
	return generate.Input{Topic: arg}, nil
}

func init() {
	rootCmd.AddCommand(newGenerateCmd())
}
