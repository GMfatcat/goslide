package llm

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
)

// Transform applies the request: build the final prompt via substitution,
// consult the cache, call the completer on miss, persist the result, and
// return Result{FromCache: true|false}. Errors never write to cache.
func Transform(ctx context.Context, req Request, completer Completer, cache *DiskCache) (Result, error) {
	if completer == nil {
		return Result{}, errors.New("llm: completer is nil")
	}
	if cache == nil {
		return Result{}, errors.New("llm: cache is nil")
	}

	key := cache.Key(req)
	if entry, ok, err := cache.Get(key); err == nil && ok {
		return Result{Content: entry.Content, Usage: entry.Usage, FromCache: true}, nil
	}

	finalPrompt := Render(req.Prompt, req.Data)
	msgs := []Message{{Role: "user", Content: finalPrompt}}

	content, usage, err := completer.Complete(ctx, req.Model, msgs)
	if err != nil {
		return Result{}, fmt.Errorf("llm: completer: %w", err)
	}
	if content == "" {
		return Result{}, errors.New("llm: completer returned empty content")
	}

	entry := CacheEntry{
		Version:  1,
		Created:  time.Now().UTC(),
		Model:    req.Model,
		Prompt:   req.Prompt,
		DataHash: dataHash(req.Data),
		Content:  content,
		Usage:    usage,
	}
	if putErr := cache.Put(key, entry); putErr != nil {
		// Cache write failure is non-fatal; return the result anyway.
		_ = putErr
	}

	return Result{Content: content, Usage: usage, FromCache: false}, nil
}

func dataHash(data []byte) string {
	h := sha256.Sum256(canonicalJSON(data))
	return "sha256:" + hex.EncodeToString(h[:])
}
