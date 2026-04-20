package llm

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

// DiskCache stores CacheEntry values as JSON files named <hex-key>.json
// inside a directory (typically .goslide-cache/ in the project root).
type DiskCache struct {
	dir string
}

func NewDiskCache(dir string) *DiskCache {
	return &DiskCache{dir: dir}
}

// Key returns the stable hex-sha256 of (model || 0 || prompt || 0 ||
// canonical(data)) for the request.
func (c *DiskCache) Key(req Request) string {
	h := sha256.New()
	h.Write([]byte(req.Model))
	h.Write([]byte{0})
	h.Write([]byte(req.Prompt))
	h.Write([]byte{0})
	h.Write(canonicalJSON(req.Data))
	return hex.EncodeToString(h.Sum(nil))
}

func (c *DiskCache) Get(key string) (CacheEntry, bool, error) {
	path := filepath.Join(c.dir, key+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return CacheEntry{}, false, nil
		}
		return CacheEntry{}, false, err
	}
	var entry CacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return CacheEntry{}, false, nil
	}
	return entry, true, nil
}

func (c *DiskCache) Put(key string, entry CacheEntry) error {
	if err := os.MkdirAll(c.dir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return err
	}
	path := filepath.Join(c.dir, key+".json")
	return os.WriteFile(path, data, 0644)
}

// canonicalJSON round-trips the input through json.Unmarshal +
// json.Marshal so that two logically-equal JSON inputs produce the
// same canonical byte sequence (key sort, whitespace stripped).
// Non-JSON input falls through unchanged.
func canonicalJSON(data []byte) []byte {
	if len(data) == 0 {
		return data
	}
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		return data
	}
	out, err := json.Marshal(v)
	if err != nil {
		return data
	}
	return out
}
