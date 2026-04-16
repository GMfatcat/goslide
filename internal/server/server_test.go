package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestHandleIndex_ServesHTML(t *testing.T) {
	dir := t.TempDir()
	mdPath := filepath.Join(dir, "test.md")
	os.WriteFile(mdPath, []byte("---\ntitle: Hello\n---\n\n# Slide 1\n"), 0644)

	a, err := newApp(Options{File: mdPath, Port: 0, NoWatch: true, NoOpen: true})
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	a.mux.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), "<title>Hello</title>")
	require.Contains(t, rec.Body.String(), "Slide 1")
}

func TestHandleIndex_LastGoodOnBadReload(t *testing.T) {
	dir := t.TempDir()
	mdPath := filepath.Join(dir, "test.md")
	os.WriteFile(mdPath, []byte("---\ntitle: Good\n---\n\n# OK\n"), 0644)

	a, err := newApp(Options{File: mdPath, Port: 0, NoWatch: true, NoOpen: true})
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	a.mux.ServeHTTP(rec, req)
	require.Contains(t, rec.Body.String(), "Good")

	os.WriteFile(mdPath, []byte("---\ntitle: Good\n  bad yaml:\n---\n"), 0644)
	a.reload()

	req2 := httptest.NewRequest("GET", "/", nil)
	rec2 := httptest.NewRecorder()
	a.mux.ServeHTTP(rec2, req2)
	require.Equal(t, http.StatusOK, rec2.Code)
	require.Contains(t, rec2.Body.String(), "Good")
}

func TestHandleThemes_ServesCSS(t *testing.T) {
	dir := t.TempDir()
	mdPath := filepath.Join(dir, "test.md")
	os.WriteFile(mdPath, []byte("# Slide\n"), 0644)

	a, err := newApp(Options{File: mdPath, Port: 0, NoWatch: true, NoOpen: true})
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/themes/tokens.css", nil)
	rec := httptest.NewRecorder()
	a.mux.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), "--accent-blue")
}

func TestDebounce(t *testing.T) {
	var count int32
	fn := func() { atomic.AddInt32(&count, 1) }
	d := newDebouncer(50*time.Millisecond, fn)

	d.trigger()
	d.trigger()
	d.trigger()

	time.Sleep(100 * time.Millisecond)
	require.Equal(t, int32(1), atomic.LoadInt32(&count))
}
