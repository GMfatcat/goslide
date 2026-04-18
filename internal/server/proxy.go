package server

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/user/goslide/internal/config"
)

func setupProxy(mux *http.ServeMux, proxies map[string]config.ProxyTarget) {
	for prefix, target := range proxies {
		handler := makeProxyHandler(prefix, target)
		pattern := prefix
		if !strings.HasSuffix(pattern, "/") {
			pattern += "/"
		}
		mux.Handle(pattern, handler)
	}
}

func makeProxyHandler(prefix string, target config.ProxyTarget) http.Handler {
	targetURL, _ := url.Parse(target.Target)
	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = targetURL.Scheme
			req.URL.Host = targetURL.Host
			req.URL.Path = strings.TrimPrefix(req.URL.Path, prefix)
			if req.URL.Path == "" || req.URL.Path[0] != '/' {
				req.URL.Path = "/" + req.URL.Path
			}
			req.Host = targetURL.Host
			for k, v := range target.Headers {
				req.Header.Set(k, v)
			}
		},
	}
	return proxy
}
