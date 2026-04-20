package builder

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/GMfatcat/goslide/internal/ir"
	"github.com/GMfatcat/goslide/internal/llm"
)

// BakeLLM resolves every `llm` render item inside every `api` component
// of the presentation, storing cached results (or freshly-fetched ones
// when refreshOnMiss is true) under the Component param `_llm_bakes`
// as a map[string]string keyed by render-item index.
//
// mdDir is the directory containing the source .md file; it anchors
// relative paths in `fixture:` fields. defaultModel is used when a
// render item omits `model`.
func BakeLLM(pres *ir.Presentation, cache *llm.DiskCache, completer llm.Completer, defaultModel string, mdDir string, refreshOnMiss bool) error {
	var missErrors []string

	for si := range pres.Slides {
		slide := &pres.Slides[si]
		for ci := range slide.Components {
			comp := &slide.Components[ci]
			if comp.Type != "api" {
				continue
			}

			renderList, ok := comp.Params["render"].([]any)
			if !ok {
				continue
			}

			hasLLMItem := false
			for _, raw := range renderList {
				item, ok := raw.(map[string]any)
				if !ok {
					continue
				}
				if item["type"] == "llm" {
					hasLLMItem = true
					break
				}
			}
			if !hasLLMItem {
				continue
			}

			fixture, _ := comp.Params["fixture"].(string)
			if fixture == "" {
				missErrors = append(missErrors, fmt.Sprintf("slide %d, component %d: api component has llm render items but no 'fixture:' file", slide.Index, comp.Index))
				continue
			}
			data, readErr := os.ReadFile(filepath.Join(mdDir, fixture))
			if readErr != nil {
				missErrors = append(missErrors, fmt.Sprintf("slide %d, component %d: read fixture %s: %v", slide.Index, comp.Index, fixture, readErr))
				continue
			}

			bakes := map[string]string{}
			for itemIdx, raw := range renderList {
				item, ok := raw.(map[string]any)
				if !ok || item["type"] != "llm" {
					continue
				}
				prompt, _ := item["prompt"].(string)
				model, _ := item["model"].(string)
				if model == "" {
					model = defaultModel
				}
				maxTokens := 1024
				if mt, ok := item["max_tokens"].(int); ok {
					maxTokens = mt
				}

				req := llm.Request{Model: model, Prompt: prompt, Data: data, MaxTokens: maxTokens}
				key := cache.Key(req)

				entry, hit, _ := cache.Get(key)
				if hit {
					bakes[strconv.Itoa(itemIdx)] = entry.Content
					continue
				}
				if !refreshOnMiss {
					missErrors = append(missErrors, fmt.Sprintf("slide %d, component %d, item %d: cache miss (run 'goslide serve' to warm cache, or pass --llm-refresh)", slide.Index, comp.Index, itemIdx))
					continue
				}
				res, err := llm.Transform(context.Background(), req, completer, cache)
				if err != nil {
					missErrors = append(missErrors, fmt.Sprintf("slide %d, component %d, item %d: llm call failed: %v", slide.Index, comp.Index, itemIdx, err))
					continue
				}
				bakes[strconv.Itoa(itemIdx)] = res.Content
			}

			if len(bakes) > 0 {
				if comp.Params == nil {
					comp.Params = map[string]any{}
				}
				comp.Params["_llm_bakes"] = bakes
			}
		}
	}

	if len(missErrors) > 0 {
		return fmt.Errorf("BakeLLM: %d problem(s):\n  %s", len(missErrors), strings.Join(missErrors, "\n  "))
	}
	return nil
}
