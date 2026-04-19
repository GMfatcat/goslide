package generate

import (
	"fmt"
	"os"
	"time"

	"github.com/GMfatcat/goslide/internal/config"
)

// Flags holds values that come from CLI flags (empty means "not provided").
type Flags struct {
	BaseURL   string
	Model     string
	APIKeyEnv string
	Output    string
	Force     bool
}

// Options is the fully-resolved configuration used by Run.
type Options struct {
	BaseURL string
	Model   string
	APIKey  string
	Timeout time.Duration

	Input  Input
	Output string
	Force  bool
}

const defaultTimeout = 120 * time.Second

// Resolve merges YAML config, CLI flags, and environment variables.
// Precedence: flags > yaml > defaults. Returns an error if required fields
// are missing or the API key env var is unset.
func Resolve(cfg *config.Config, flags Flags, in Input) (Options, error) {
	g := cfg.Generate

	baseURL := firstNonEmpty(flags.BaseURL, g.BaseURL)
	model := firstNonEmpty(flags.Model, g.Model)
	apiKeyEnv := firstNonEmpty(flags.APIKeyEnv, g.APIKeyEnv, "OPENAI_API_KEY")
	output := firstNonEmpty(flags.Output, "talk.md")
	timeout := g.Timeout
	if timeout <= 0 {
		timeout = defaultTimeout
	}

	if baseURL == "" {
		return Options{}, fmt.Errorf("generate.base_url is not set (use --base-url or add it to goslide.yaml)")
	}
	if model == "" {
		return Options{}, fmt.Errorf("generate.model is not set (use --model or add it to goslide.yaml)")
	}

	apiKey := os.Getenv(apiKeyEnv)
	if apiKey == "" {
		return Options{}, fmt.Errorf("environment variable %s is not set", apiKeyEnv)
	}

	return Options{
		BaseURL: baseURL,
		Model:   model,
		APIKey:  apiKey,
		Timeout: timeout,
		Input:   in,
		Output:  output,
		Force:   flags.Force,
	}, nil
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
