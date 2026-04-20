package llm

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type fakeCompleter struct {
	content string
	usage   Usage
	err     error
	calls   int
}

func (f *fakeCompleter) Complete(_ context.Context, _ string, _ []Message) (string, Usage, error) {
	f.calls++
	return f.content, f.usage, f.err
}

func TestTransform_CacheMissCallsCompleterAndStores(t *testing.T) {
	cache := NewDiskCache(t.TempDir())
	fc := &fakeCompleter{content: "out", usage: Usage{TotalTokens: 42}}
	req := Request{Model: "m", Prompt: "hi {{data}}", Data: []byte(`{"x":1}`), MaxTokens: 128}

	res, err := Transform(context.Background(), req, fc, cache)
	require.NoError(t, err)
	require.Equal(t, "out", res.Content)
	require.False(t, res.FromCache)
	require.Equal(t, 42, res.Usage.TotalTokens)
	require.Equal(t, 1, fc.calls)

	res2, err := Transform(context.Background(), req, fc, cache)
	require.NoError(t, err)
	require.True(t, res2.FromCache)
	require.Equal(t, "out", res2.Content)
	require.Equal(t, 1, fc.calls, "completer must not be called again on hit")
}

func TestTransform_SubstitutesPromptBeforeCalling(t *testing.T) {
	cache := NewDiskCache(t.TempDir())
	var observed string
	fc := &stubCompleter{fn: func(msgs []Message) (string, Usage, error) {
		observed = msgs[0].Content
		return "ok", Usage{}, nil
	}}
	req := Request{Model: "m", Prompt: "Summarise {{data}}", Data: []byte(`{"x":1}`)}

	_, err := Transform(context.Background(), req, fc, cache)
	require.NoError(t, err)
	require.Contains(t, observed, `{"x":1}`)
	require.NotContains(t, observed, "{{data}}")
}

func TestTransform_CompleterErrorSkipsCacheWrite(t *testing.T) {
	cache := NewDiskCache(t.TempDir())
	fc := &fakeCompleter{err: errors.New("boom")}
	req := Request{Model: "m", Prompt: "hi", Data: []byte(`{}`)}

	_, err := Transform(context.Background(), req, fc, cache)
	require.Error(t, err)

	_, hit, _ := cache.Get(cache.Key(req))
	require.False(t, hit, "error responses must not be cached")
}

func TestTransform_EmptyContentIsError(t *testing.T) {
	cache := NewDiskCache(t.TempDir())
	fc := &fakeCompleter{content: "", usage: Usage{}}
	req := Request{Model: "m", Prompt: "hi", Data: []byte(`{}`)}

	_, err := Transform(context.Background(), req, fc, cache)
	require.Error(t, err)
	require.Contains(t, err.Error(), "empty")
}

type stubCompleter struct {
	fn func(msgs []Message) (string, Usage, error)
}

func (s *stubCompleter) Complete(_ context.Context, _ string, msgs []Message) (string, Usage, error) {
	return s.fn(msgs)
}

var _ = time.Now // keep import even if unused inline
