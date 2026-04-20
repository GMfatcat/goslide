package llm

import (
	"context"
	"time"
)

// Request is the input to Transform. Data is the raw bytes of the API
// response that will be substituted into the prompt template.
type Request struct {
	Model     string
	Prompt    string
	Data      []byte
	MaxTokens int
}

// Usage mirrors the subset of OpenAI /chat/completions usage fields we
// surface to the caller.
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Result is the output of Transform. FromCache is true when the content
// was served from disk cache and no HTTP call happened.
type Result struct {
	Content   string
	Usage     Usage
	FromCache bool
}

// CacheEntry is what DiskCache stores on disk per key.
type CacheEntry struct {
	Version  int       `json:"version"`
	Created  time.Time `json:"created"`
	Model    string    `json:"model"`
	Prompt   string    `json:"prompt"`
	DataHash string    `json:"data_hash"`
	Content  string    `json:"content"`
	Usage    Usage     `json:"usage"`
}

// Completer is the abstraction Transform depends on for HTTP calls.
// internal/generate.Client already satisfies this shape; tests can
// substitute a fake.
type Completer interface {
	Complete(ctx context.Context, model string, msgs []Message) (string, Usage, error)
}

// Message is the chat-completions message shape. It mirrors
// internal/generate.Message so adapters stay trivial.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
