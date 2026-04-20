package llm

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestDiskCache_PutGet(t *testing.T) {
	dir := t.TempDir()
	c := NewDiskCache(dir)

	req := Request{Model: "gpt-4o", Prompt: "hi", Data: []byte(`{"x":1}`)}
	key := c.Key(req)
	require.Len(t, key, 64) // hex sha256

	_, ok, err := c.Get(key)
	require.NoError(t, err)
	require.False(t, ok, "expected miss")

	entry := CacheEntry{
		Version: 1,
		Created: time.Now().UTC(),
		Model:   "gpt-4o",
		Prompt:  "hi",
		Content: "hello!",
		Usage:   Usage{TotalTokens: 10},
	}
	require.NoError(t, c.Put(key, entry))

	got, ok, err := c.Get(key)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, "hello!", got.Content)
	require.Equal(t, 10, got.Usage.TotalTokens)
}

func TestDiskCache_KeyStableForSameInputs(t *testing.T) {
	c := NewDiskCache(t.TempDir())
	req := Request{Model: "m", Prompt: "p", Data: []byte(`{"a":1,"b":2}`)}
	a := c.Key(req)
	b := c.Key(req)
	require.Equal(t, a, b)
}

func TestDiskCache_KeyCanonicalizesDataJSON(t *testing.T) {
	c := NewDiskCache(t.TempDir())
	a := c.Key(Request{Model: "m", Prompt: "p", Data: []byte(`{"a":1,"b":2}`)})
	b := c.Key(Request{Model: "m", Prompt: "p", Data: []byte(`{"b":2,"a":1}`)})
	require.Equal(t, a, b, "logically equal JSON should hash the same")
}

func TestDiskCache_KeyDiffersOnModel(t *testing.T) {
	c := NewDiskCache(t.TempDir())
	a := c.Key(Request{Model: "m1", Prompt: "p", Data: []byte(`{}`)})
	b := c.Key(Request{Model: "m2", Prompt: "p", Data: []byte(`{}`)})
	require.NotEqual(t, a, b)
}

func TestDiskCache_CorruptFileTreatedAsMiss(t *testing.T) {
	dir := t.TempDir()
	c := NewDiskCache(dir)
	require.NoError(t, os.MkdirAll(dir, 0755))
	bad := filepath.Join(dir, "deadbeef.json")
	require.NoError(t, os.WriteFile(bad, []byte("not json"), 0644))

	_, ok, err := c.Get("deadbeef")
	require.NoError(t, err)
	require.False(t, ok)
}

func TestDiskCache_AutoCreatesDir(t *testing.T) {
	parent := t.TempDir()
	dir := filepath.Join(parent, "nested", ".goslide-cache")
	c := NewDiskCache(dir)

	entry := CacheEntry{Version: 1, Content: "x"}
	require.NoError(t, c.Put("key1", entry))

	_, err := os.Stat(dir)
	require.NoError(t, err)
}
