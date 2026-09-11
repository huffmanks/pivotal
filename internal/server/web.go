package server

import (
	"io/fs"
	"net/http"
	"path"
	"strings"

	"pivotal/internal/link"
)

func WebHandler(files fs.FS, linkSvc link.Service) http.Handler {
	fileServer := http.FileServer(http.FS(files))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cleanPath := strings.TrimPrefix(path.Clean(r.URL.Path), "/")

		if cleanPath == "" || cleanPath == "." {
			serveFile(w, r, files, "200.html")
			return
		}

		if l, extraPath, err := linkSvc.Resolve(r.Context(), cleanPath); err == nil {
			dest := l.DestinationURL
			if extraPath != "" {
				dest = strings.TrimSuffix(dest, "/") + "/" + strings.TrimPrefix(extraPath, "/")
			}

			linkSvc.RecordClick(link.ClickEvent{
				LinkID:    l.ID,
				Referer:   r.Referer(),
				UserAgent: r.UserAgent(),
			})

			http.Redirect(w, r, dest, http.StatusFound)
			return
		}

		if info, err := fs.Stat(files, cleanPath); err == nil && !info.IsDir() {
			fileServer.ServeHTTP(w, r)
			return
		}

		if path.Ext(cleanPath) != "" {
			http.NotFound(w, r)
			return
		}

		serveFile(w, r, files, "200.html")
	})
}

func serveFile(w http.ResponseWriter, r *http.Request, files fs.FS, name string) {
	if f, err := files.Open(name); err == nil {
		_ = f.Close()
		http.ServeFileFS(w, r, files, name)
		return
	}
	http.NotFound(w, r)
}
