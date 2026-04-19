package generate

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestClient_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/chat/completions", r.URL.Path)
		require.Equal(t, "Bearer sk-test", r.Header.Get("Authorization"))

		body, _ := io.ReadAll(r.Body)
		var req struct {
			Model    string `json:"model"`
			Messages []struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"messages"`
		}
		require.NoError(t, json.Unmarshal(body, &req))
		require.Equal(t, "gpt-4o", req.Model)
		require.Len(t, req.Messages, 2)
		require.Equal(t, "system", req.Messages[0].Role)
		require.Equal(t, "user", req.Messages[1].Role)

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
            "choices": [{"message": {"content": "# slide\n"}}],
            "usage": {"prompt_tokens": 10, "completion_tokens": 20, "total_tokens": 30}
        }`))
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "sk-test", 5*time.Second)
	content, usage, err := c.Complete(context.Background(), "gpt-4o", []Message{
		{Role: "system", Content: "sys"},
		{Role: "user", Content: "hi"},
	})
	require.NoError(t, err)
	require.Equal(t, "# slide\n", content)
	require.Equal(t, 30, usage.TotalTokens)
}

func TestClient_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
		w.Write([]byte(`{"error": {"message": "invalid API key"}}`))
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "bad", time.Second)
	_, _, err := c.Complete(context.Background(), "gpt-4o", []Message{{Role: "user", Content: "x"}})
	require.Error(t, err)
	require.Contains(t, err.Error(), "401")
	require.Contains(t, err.Error(), "invalid API key")
}

func TestClient_EmptyChoices(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"choices": []}`))
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "k", time.Second)
	_, _, err := c.Complete(context.Background(), "gpt-4o", []Message{{Role: "user", Content: "x"}})
	require.Error(t, err)
	require.Contains(t, strings.ToLower(err.Error()), "empty")
}

func TestClient_Timeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "k", 50*time.Millisecond)
	_, _, err := c.Complete(context.Background(), "gpt-4o", []Message{{Role: "user", Content: "x"}})
	require.Error(t, err)
}

func TestClient_TrimsTrailingSlashInBaseURL(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Write([]byte(`{"choices":[{"message":{"content":"ok"}}]}`))
	}))
	defer srv.Close()

	c := NewClient(srv.URL+"/", "k", time.Second)
	_, _, err := c.Complete(context.Background(), "m", []Message{{Role: "user", Content: "x"}})
	require.NoError(t, err)
	require.Equal(t, "/chat/completions", gotPath)
}
