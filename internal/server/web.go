package server

import (
	"io/fs"
	"net/http"
	"path"
	"strings"

	"url-shortener/internal/link"
)

func WebHandler(files fs.FS, store *link.Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestedPath := strings.Trim(r.URL.Path, "/")

		if requestedPath != "" {
			if l, extraPath, found := store.Resolve(r.Context(), requestedPath); found {
				handleResolvedRedirect(w, r, store, l, extraPath)
				return
			}
		}

		serveWeb(w, r, files)
	})
}

func serveWeb(w http.ResponseWriter, r *http.Request, files fs.FS) {
	name := strings.TrimPrefix(path.Clean(r.URL.Path), "/")

	if name == "." || name == "" {
		http.ServeFileFS(w, r, files, "200.html")
		return
	}

	if info, err := fs.Stat(files, name); err == nil {
		if info.IsDir() {
			name = strings.TrimSuffix(name, "/") + "/index.html"

			if _, err := fs.Stat(files, name); err != nil {
				http.NotFound(w, r)
				return
			}
		}

		http.ServeFileFS(w, r, files, name)
		return
	}

	if path.Ext(name) != "" {
		http.NotFound(w, r)
		return
	}

	http.ServeFileFS(w, r, files, "200.html")
}
