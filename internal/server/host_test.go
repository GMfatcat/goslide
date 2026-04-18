package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func writeMD(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	require.NoError(t, os.WriteFile(path, []byte(content), 0644))
	return path
}

func TestHostApp_ScanDirectory(t *testing.T) {
	dir := t.TempDir()
	writeMD(t, dir, "talk-a.md", "---\ntitle: Talk A\nauthor: Alice\n---\n\n# Slide 1\n")
	writeMD(t, dir, "talk-b.md", "---\ntitle: Talk B\ndate: 2026-01-01\ntags: [go, slides]\n---\n\n# Slide 1\n")
	writeMD(t, dir, "not-md.txt", "ignore me")

	h, err := newHostApp(HostOptions{Dir: dir, Port: 0})
	require.NoError(t, err)
	require.Len(t, h.pages, 2)
	require.Contains(t, h.pages, "talk-a")
	require.Contains(t, h.pages, "talk-b")
	require.Equal(t, "Talk A", h.pages["talk-a"].Title)
	require.Equal(t, "Alice", h.pages["talk-a"].Author)
}

func TestHostApp_IndexPage(t *testing.T) {
	dir := t.TempDir()
	writeMD(t, dir, "demo.md", "---\ntitle: Demo Talk\n---\n\n# Hello\n")

	h, err := newHostApp(HostOptions{Dir: dir, Port: 0})
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	h.mux.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), "Demo Talk")
	require.Contains(t, rec.Body.String(), "/talks/demo")
}

func TestHostApp_ServeTalk(t *testing.T) {
	dir := t.TempDir()
	writeMD(t, dir, "my-talk.md", "---\ntitle: My Talk\n---\n\n# Slide 1\n\nContent.\n")

	h, err := newHostApp(HostOptions{Dir: dir, Port: 0})
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/talks/my-talk", nil)
	rec := httptest.NewRecorder()
	h.mux.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), "Slide 1")
	require.Contains(t, rec.Body.String(), "reveal")
}

func TestHostApp_TalkNotFound(t *testing.T) {
	dir := t.TempDir()
	writeMD(t, dir, "exists.md", "# Slide\n")

	h, err := newHostApp(HostOptions{Dir: dir, Port: 0})
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/talks/nonexist", nil)
	rec := httptest.NewRecorder()
	h.mux.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
}

func TestHostApp_EmptyDirectory(t *testing.T) {
	dir := t.TempDir()

	h, err := newHostApp(HostOptions{Dir: dir, Port: 0})
	require.NoError(t, err)
	require.Len(t, h.pages, 0)

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	h.mux.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), "No presentations found")
}

func TestHostApp_StaticAssets(t *testing.T) {
	dir := t.TempDir()

	h, err := newHostApp(HostOptions{Dir: dir, Port: 0})
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/themes/tokens.css", nil)
	rec := httptest.NewRecorder()
	h.mux.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), "--accent-blue")
}

func TestHostApp_FileCreate(t *testing.T) {
	dir := t.TempDir()
	writeMD(t, dir, "initial.md", "---\ntitle: Initial\n---\n\n# Slide\n")

	h, err := newHostApp(HostOptions{Dir: dir, Port: 0})
	require.NoError(t, err)
	require.Len(t, h.pages, 1)

	writeMD(t, dir, "new-talk.md", "---\ntitle: New Talk\n---\n\n# New\n")
	h.handleFileEvent("create", filepath.Join(dir, "new-talk.md"))

	h.mu.RLock()
	count := len(h.pages)
	h.mu.RUnlock()
	require.Equal(t, 2, count)
}

func TestHostApp_FileDelete(t *testing.T) {
	dir := t.TempDir()
	writeMD(t, dir, "to-delete.md", "---\ntitle: Delete Me\n---\n\n# Slide\n")

	h, err := newHostApp(HostOptions{Dir: dir, Port: 0})
	require.NoError(t, err)
	require.Len(t, h.pages, 1)

	os.Remove(filepath.Join(dir, "to-delete.md"))
	h.handleFileEvent("remove", filepath.Join(dir, "to-delete.md"))

	h.mu.RLock()
	count := len(h.pages)
	h.mu.RUnlock()
	require.Equal(t, 0, count)
}

func TestHostApp_FileModify(t *testing.T) {
	dir := t.TempDir()
	writeMD(t, dir, "modify.md", "---\ntitle: Original\n---\n\n# Old\n")

	h, err := newHostApp(HostOptions{Dir: dir, Port: 0})
	require.NoError(t, err)
	require.Equal(t, "Original", h.pages["modify"].Title)

	writeMD(t, dir, "modify.md", "---\ntitle: Updated\n---\n\n# New\n")
	h.handleFileEvent("write", filepath.Join(dir, "modify.md"))

	h.mu.RLock()
	title := h.pages["modify"].Title
	h.mu.RUnlock()
	require.Equal(t, "Updated", title)
}
