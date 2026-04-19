package generate

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Message is a single chat-completions message.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Usage is the token accounting returned by the API (may be zero if absent).
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Completer is the abstraction Run depends on. Client implements it; tests
// can substitute a fake.
type Completer interface {
	Complete(ctx context.Context, model string, msgs []Message) (string, Usage, error)
}

// Client is an OpenAI-compatible chat-completions client.
type Client struct {
	baseURL string
	apiKey  string
	http    *http.Client
}

// NewClient builds a Client. baseURL should be the API root that exposes
// /chat/completions (e.g. https://api.openai.com/v1).
func NewClient(baseURL, apiKey string, timeout time.Duration) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		apiKey:  apiKey,
		http:    &http.Client{Timeout: timeout},
	}
}

type chatReq struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

type chatResp struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
	Usage Usage `json:"usage"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// Complete posts to /chat/completions and returns the content of the first
// choice plus token usage.
func (c *Client) Complete(ctx context.Context, model string, msgs []Message) (string, Usage, error) {
	body, err := json.Marshal(chatReq{Model: model, Messages: msgs, Stream: false})
	if err != nil {
		return "", Usage{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", Usage{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return "", Usage{}, fmt.Errorf("llm request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		msg := strings.TrimSpace(string(respBody))
		var parsed chatResp
		if json.Unmarshal(respBody, &parsed) == nil && parsed.Error != nil && parsed.Error.Message != "" {
			msg = parsed.Error.Message
		}
		return "", Usage{}, fmt.Errorf("llm http %d: %s", resp.StatusCode, msg)
	}

	var parsed chatResp
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", Usage{}, fmt.Errorf("decode llm response: %w", err)
	}
	if len(parsed.Choices) == 0 {
		return "", Usage{}, fmt.Errorf("llm returned empty choices")
	}
	return parsed.Choices[0].Message.Content, parsed.Usage, nil
}
