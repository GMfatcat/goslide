package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/user/goslide/internal/config"
)

func TestProxy_BasicForward(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("upstream:" + r.URL.Path))
	}))
	defer upstream.Close()

	mux := http.NewServeMux()
	setupProxy(mux, map[string]config.ProxyTarget{
		"/api/test": {Target: upstream.URL},
	})

	req := httptest.NewRequest("GET", "/api/test/status", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "upstream:/status", rec.Body.String())
}

func TestProxy_PrefixStripping(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("path:" + r.URL.Path))
	}))
	defer upstream.Close()

	mux := http.NewServeMux()
	setupProxy(mux, map[string]config.ProxyTarget{
		"/api/v1": {Target: upstream.URL},
	})

	req := httptest.NewRequest("GET", "/api/v1/users/123", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	require.Equal(t, "path:/users/123", rec.Body.String())
}

func TestProxy_HeaderInjection(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("auth:" + r.Header.Get("Authorization")))
	}))
	defer upstream.Close()

	mux := http.NewServeMux()
	setupProxy(mux, map[string]config.ProxyTarget{
		"/api/secure": {
			Target:  upstream.URL,
			Headers: map[string]string{"Authorization": "Bearer token123"},
		},
	})

	req := httptest.NewRequest("GET", "/api/secure/data", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	require.Equal(t, "auth:Bearer token123", rec.Body.String())
}

func TestProxy_MultipleRoutes(t *testing.T) {
	upstream1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("svc1"))
	}))
	defer upstream1.Close()

	upstream2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("svc2"))
	}))
	defer upstream2.Close()

	mux := http.NewServeMux()
	setupProxy(mux, map[string]config.ProxyTarget{
		"/api/one": {Target: upstream1.URL},
		"/api/two": {Target: upstream2.URL},
	})

	req1 := httptest.NewRequest("GET", "/api/one/x", nil)
	rec1 := httptest.NewRecorder()
	mux.ServeHTTP(rec1, req1)
	require.Equal(t, "svc1", rec1.Body.String())

	req2 := httptest.NewRequest("GET", "/api/two/y", nil)
	rec2 := httptest.NewRecorder()
	mux.ServeHTTP(rec2, req2)
	require.Equal(t, "svc2", rec2.Body.String())
}

func TestProxy_UpstreamDown(t *testing.T) {
	mux := http.NewServeMux()
	setupProxy(mux, map[string]config.ProxyTarget{
		"/api/down": {Target: "http://127.0.0.1:59999"},
	})

	req := httptest.NewRequest("GET", "/api/down/test", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadGateway, rec.Code)
}

func TestProxy_RootPath(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("root:" + r.URL.Path))
	}))
	defer upstream.Close()

	mux := http.NewServeMux()
	setupProxy(mux, map[string]config.ProxyTarget{
		"/api/test": {Target: upstream.URL},
	})

	req := httptest.NewRequest("GET", "/api/test/", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	body, _ := io.ReadAll(rec.Result().Body)
	require.Equal(t, "root:/", string(body))
}
