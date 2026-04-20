package server

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/GMfatcat/goslide/internal/config"
	"github.com/GMfatcat/goslide/internal/generate"
	"github.com/GMfatcat/goslide/internal/llm"
)

// llmTransformerAdapter satisfies the Transformer interface (from
// llm_handler.go) by wiring llm.Transform with a DiskCache and a
// generate.Client-backed Completer.
type llmTransformerAdapter struct {
	cfg   *config.Config
	cache *llm.DiskCache
}

func newLLMTransformerAdapter(cfg *config.Config, presentationFile string) *llmTransformerAdapter {
	cacheDir := filepath.Join(filepath.Dir(presentationFile), ".goslide-cache")
	return &llmTransformerAdapter{
		cfg:   cfg,
		cache: llm.NewDiskCache(cacheDir),
	}
}

func (a *llmTransformerAdapter) Transform(ctx context.Context, req llm.Request) (llm.Result, error) {
	if req.Model == "" {
		req.Model = a.cfg.Generate.Model
	}

	keyEnv := a.cfg.Generate.APIKeyEnv
	if keyEnv == "" {
		keyEnv = "OPENAI_API_KEY"
	}
	apiKey := os.Getenv(keyEnv)

	timeout := a.cfg.Generate.Timeout
	if timeout == 0 {
		timeout = 120 * time.Second
	}

	client := generate.NewClient(a.cfg.Generate.BaseURL, apiKey, timeout)
	completer := clientAsCompleter{c: client}
	return llm.Transform(ctx, req, completer, a.cache)
}

// clientAsCompleter adapts *generate.Client to llm.Completer. The Message
// and Usage types of the two packages are parallel, so the copy is
// trivial.
type clientAsCompleter struct {
	c *generate.Client
}

func (a clientAsCompleter) Complete(ctx context.Context, model string, msgs []llm.Message) (string, llm.Usage, error) {
	gmsgs := make([]generate.Message, len(msgs))
	for i, m := range msgs {
		gmsgs[i] = generate.Message{Role: m.Role, Content: m.Content}
	}
	content, usage, err := a.c.Complete(ctx, model, gmsgs)
	return content, llm.Usage(usage), err
}
