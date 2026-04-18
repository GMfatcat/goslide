package server

import (
	"io/fs"
	"net/http"
	"strings"

	"github.com/user/goslide/web"
)

func setupStaticRoutes(mux *http.ServeMux) {
	themeSub, _ := fs.Sub(web.ThemeFS, "themes")
	mux.Handle("/themes/", http.StripPrefix("/themes/", http.FileServer(http.FS(themeSub))))

	staticSub, _ := fs.Sub(web.StaticFS, "static")
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticSub))))

	fontsSub, _ := fs.Sub(web.StaticFS, "static/fonts")
	mux.Handle("/fonts/", http.StripPrefix("/fonts/", addCacheHeader(http.FileServer(http.FS(fontsSub)))))
}

func (a *app) setupRoutes() {
	a.mux.HandleFunc("/", a.handleIndex)
	a.mux.HandleFunc("/ws", a.broadcast.handleWS)
	setupStaticRoutes(a.mux)
}

func (a *app) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	a.mu.RLock()
	html := a.cachedHTML
	a.mu.RUnlock()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

func addCacheHeader(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, ".woff2") {
			w.Header().Set("Content-Type", "font/woff2")
		}
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		h.ServeHTTP(w, r)
	})
}
