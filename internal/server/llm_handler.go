package server

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/GMfatcat/goslide/internal/llm"
)

// Transformer is the narrow interface llm_handler depends on. The
// production wiring (Task 7) will pass an adapter that calls
// llm.Transform with a shared DiskCache and Completer; tests pass a
// fake.
type Transformer interface {
	Transform(ctx context.Context, req llm.Request) (llm.Result, error)
}

type llmHandler struct {
	t Transformer
}

func newLLMHandler(t Transformer) *llmHandler {
	return &llmHandler{t: t}
}

type llmRequestBody struct {
	Model     string          `json:"model"`
	Prompt    string          `json:"prompt"`
	Data      json.RawMessage `json:"data"`
	MaxTokens int             `json:"max_tokens"`
}

type llmResponseBody struct {
	Content   string    `json:"content"`
	FromCache bool      `json:"from_cache"`
	Usage     llm.Usage `json:"usage"`
}

func (h *llmHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var body llmRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	req := llm.Request{
		Model:     body.Model,
		Prompt:    body.Prompt,
		Data:      []byte(body.Data),
		MaxTokens: body.MaxTokens,
	}
	res, err := h.t.Transform(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(llmResponseBody{
		Content:   res.Content,
		FromCache: res.FromCache,
		Usage:     res.Usage,
	})
}
