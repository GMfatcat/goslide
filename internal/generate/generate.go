package generate

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/GMfatcat/goslide/internal/parser"
)

// Run executes the full generate pipeline using a real HTTP Client.
func Run(ctx context.Context, opts Options) error {
	client := NewClient(opts.BaseURL, opts.APIKey, opts.Timeout)
	return runWith(ctx, opts, client, os.Stderr)
}

// runWith is Run with the Completer and progress-output sink injected for
// testing.
func runWith(ctx context.Context, opts Options, llm Completer, stderr io.Writer) error {
	if !opts.Force {
		if _, err := os.Stat(opts.Output); err == nil {
			return fmt.Errorf("file %s exists; use --force to overwrite", opts.Output)
		}
	}

	userMsg, err := BuildUserMessage(opts.Input)
	if err != nil {
		return err
	}
	msgs := []Message{
		{Role: "system", Content: SystemPrompt()},
		{Role: "user", Content: userMsg},
	}

	fmt.Fprintf(stderr, "Generating… (model=%s)\n", opts.Model)
	content, usage, err := llm.Complete(ctx, opts.Model, msgs)
	if err != nil {
		return err
	}
	if usage.TotalTokens > 0 {
		fmt.Fprintf(stderr, "Tokens: prompt=%d completion=%d total=%d\n",
			usage.PromptTokens, usage.CompletionTokens, usage.TotalTokens)
	}

	final, err := sanitizeAndFix(content, stderr)
	if err != nil {
		// Write raw + fixed artefacts next to the output path for diagnosis.
		writeArtefact(opts.Output+".raw.md", content)
		writeArtefact(opts.Output+".fixed.md", final)
		return err
	}

	if err := os.WriteFile(opts.Output, []byte(final), 0644); err != nil {
		return fmt.Errorf("write output: %w", err)
	}
	fmt.Fprintf(stderr, "✓ Written to %s\n", opts.Output)
	return nil
}

// sanitizeAndFix parses, applies heuristic fixup if parsing failed, and
// returns the final markdown. It reports applied fixes on stderr. If the
// fixed markdown still fails to parse, the second parse error is returned.
func sanitizeAndFix(md string, stderr io.Writer) (string, error) {
	if _, err := parser.Parse([]byte(md), "generated.md"); err == nil {
		return md, nil
	} else {
		fixed, report := Try(md, err)
		if _, err2 := parser.Parse([]byte(fixed), "generated.md"); err2 != nil {
			fmt.Fprintf(stderr, "✗ Generated markdown could not be auto-fixed.\n")
			fmt.Fprintf(stderr, "  Original parse error: %v\n", err)
			fmt.Fprintf(stderr, "  Post-fix parse error: %v\n", err2)
			return fixed, fmt.Errorf("generated markdown failed to parse; see .raw.md and .fixed.md")
		}
		if len(report.Fixes) > 0 {
			fmt.Fprintf(stderr, "⚠ Generated markdown had %d issue(s); auto-fix applied:\n", len(report.Fixes))
			for _, f := range report.Fixes {
				fmt.Fprintf(stderr, "  [%s] line %d: %s\n", f.Rule, f.Line, f.Description)
			}
			fmt.Fprintf(stderr, "  Original parse error: %v\n", err)
		}
		return fixed, nil
	}
}

func writeArtefact(path, content string) {
	_ = os.WriteFile(path, []byte(content), 0644)
}
