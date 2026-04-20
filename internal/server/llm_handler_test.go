package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GMfatcat/goslide/internal/llm"
	"github.com/stretchr/testify/require"
)

type fakeTransformer struct {
	result llm.Result
	err    error
	got    llm.Request
}

func (f *fakeTransformer) Transform(_ context.Context, req llm.Request) (llm.Result, error) {
	f.got = req
	return f.result, f.err
}

func TestLLMHandler_ValidRequest(t *testing.T) {
	ft := &fakeTransformer{result: llm.Result{Content: "hi", Usage: llm.Usage{TotalTokens: 9}, FromCache: false}}
	h := newLLMHandler(ft)

	body := []byte(`{"model":"m","prompt":"hi {{data}}","data":{"x":1},"max_tokens":256}`)
	r := httptest.NewRequest(http.MethodPost, "/api/llm", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.ServeHTTP(w, r)

	require.Equal(t, http.StatusOK, w.Code)
	var resp struct {
		Content   string    `json:"content"`
		FromCache bool      `json:"from_cache"`
		Usage     llm.Usage `json:"usage"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Equal(t, "hi", resp.Content)
	require.False(t, resp.FromCache)
	require.Equal(t, 9, resp.Usage.TotalTokens)

	require.JSONEq(t, `{"x":1}`, string(ft.got.Data))
	require.Equal(t, "m", ft.got.Model)
	require.Equal(t, 256, ft.got.MaxTokens)
}

func TestLLMHandler_MethodNotAllowed(t *testing.T) {
	h := newLLMHandler(&fakeTransformer{})
	r := httptest.NewRequest(http.MethodGet, "/api/llm", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	require.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestLLMHandler_MalformedJSON(t *testing.T) {
	h := newLLMHandler(&fakeTransformer{})
	r := httptest.NewRequest(http.MethodPost, "/api/llm", bytes.NewReader([]byte("not json")))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLLMHandler_TransformError(t *testing.T) {
	h := newLLMHandler(&fakeTransformer{err: errors.New("upstream failed")})
	r := httptest.NewRequest(http.MethodPost, "/api/llm", bytes.NewReader([]byte(`{"prompt":"x"}`)))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	require.Equal(t, http.StatusBadGateway, w.Code)
}
