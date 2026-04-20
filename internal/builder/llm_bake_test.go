package builder

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/GMfatcat/goslide/internal/ir"
	"github.com/GMfatcat/goslide/internal/llm"
	"github.com/stretchr/testify/require"
)

type fakeCompleter struct {
	content string
	err     error
	calls   int
}

func (f *fakeCompleter) Complete(_ context.Context, _ string, _ []llm.Message) (string, llm.Usage, error) {
	f.calls++
	return f.content, llm.Usage{}, f.err
}

func TestBakeLLM_CacheHitProducesBakes(t *testing.T) {
	dir := t.TempDir()
	cache := llm.NewDiskCache(filepath.Join(dir, ".goslide-cache"))
	req := llm.Request{Model: "m", Prompt: "hi {{data}}", Data: []byte(`{"x":1}`)}
	require.NoError(t, cache.Put(cache.Key(req), llm.CacheEntry{Version: 1, Content: "result"}))

	fixture := filepath.Join(dir, "data.json")
	require.NoError(t, os.WriteFile(fixture, []byte(`{"x":1}`), 0644))

	pres := &ir.Presentation{
		Slides: []ir.Slide{{
			Index: 0,
			Components: []ir.Component{{
				Index: 0,
				Type:  "api",
				Params: map[string]any{
					"endpoint": "/api/x",
					"fixture":  "data.json",
					"render": []any{
						map[string]any{"type": "chart:bar"},
						map[string]any{"type": "llm", "prompt": "hi {{data}}", "model": "m"},
					},
				},
			}},
		}},
	}

	fc := &fakeCompleter{}
	err := BakeLLM(pres, cache, fc, "m", dir, false)
	require.NoError(t, err)
	require.Equal(t, 0, fc.calls, "cache hit must not call completer")

	bakes := pres.Slides[0].Components[0].Params["_llm_bakes"].(map[string]string)
	require.Equal(t, "result", bakes["1"])
}

func TestBakeLLM_CacheMissWithoutRefreshReturnsError(t *testing.T) {
	dir := t.TempDir()
	cache := llm.NewDiskCache(filepath.Join(dir, ".goslide-cache"))
	fixture := filepath.Join(dir, "data.json")
	require.NoError(t, os.WriteFile(fixture, []byte(`{"x":1}`), 0644))

	pres := &ir.Presentation{
		Slides: []ir.Slide{{
			Index: 0,
			Components: []ir.Component{{
				Index: 0,
				Type:  "api",
				Params: map[string]any{
					"fixture": "data.json",
					"render": []any{
						map[string]any{"type": "llm", "prompt": "hi", "model": "m"},
					},
				},
			}},
		}},
	}

	err := BakeLLM(pres, cache, &fakeCompleter{}, "m", dir, false)
	require.Error(t, err)
	require.Contains(t, err.Error(), "cache miss")
	require.Contains(t, err.Error(), "slide 0")
}

func TestBakeLLM_CacheMissWithRefreshCallsCompleter(t *testing.T) {
	dir := t.TempDir()
	cache := llm.NewDiskCache(filepath.Join(dir, ".goslide-cache"))
	fixture := filepath.Join(dir, "data.json")
	require.NoError(t, os.WriteFile(fixture, []byte(`{"x":1}`), 0644))

	pres := &ir.Presentation{
		Slides: []ir.Slide{{
			Index: 0,
			Components: []ir.Component{{
				Index: 0,
				Type:  "api",
				Params: map[string]any{
					"fixture": "data.json",
					"render": []any{
						map[string]any{"type": "llm", "prompt": "hi {{data}}", "model": "m"},
					},
				},
			}},
		}},
	}

	fc := &fakeCompleter{content: "fresh"}
	err := BakeLLM(pres, cache, fc, "m", dir, true)
	require.NoError(t, err)
	require.Equal(t, 1, fc.calls)

	bakes := pres.Slides[0].Components[0].Params["_llm_bakes"].(map[string]string)
	require.Equal(t, "fresh", bakes["0"])
}

func TestBakeLLM_SkipsNonAPIComponents(t *testing.T) {
	dir := t.TempDir()
	cache := llm.NewDiskCache(filepath.Join(dir, ".goslide-cache"))

	pres := &ir.Presentation{
		Slides: []ir.Slide{{
			Index:      0,
			Components: []ir.Component{{Index: 0, Type: "chart", Params: map[string]any{}}},
		}},
	}

	err := BakeLLM(pres, cache, &fakeCompleter{}, "m", dir, false)
	require.NoError(t, err)
}

func TestBakeLLM_NoFixtureAndNoCacheReturnsError(t *testing.T) {
	dir := t.TempDir()
	cache := llm.NewDiskCache(filepath.Join(dir, ".goslide-cache"))

	pres := &ir.Presentation{
		Slides: []ir.Slide{{
			Index: 0,
			Components: []ir.Component{{
				Index: 0,
				Type:  "api",
				Params: map[string]any{
					"render": []any{
						map[string]any{"type": "llm", "prompt": "hi"},
					},
				},
			}},
		}},
	}

	err := BakeLLM(pres, cache, &fakeCompleter{}, "m", dir, false)
	require.Error(t, err)
}
