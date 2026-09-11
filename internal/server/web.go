package server

import (
	"io/fs"
	"net/http"
	"path"
	"strings"

	"url-shortener/internal/link"
)

func WebHandler(files fs.FS, linkSvc link.Service) http.Handler {
	fileServer := http.FileServer(http.FS(files))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := strings.TrimPrefix(path.Clean(r.URL.Path), "/")

		if p == "" {
			fileServer.ServeHTTP(w, r)
			return
		}

		l, extraPath, err := linkSvc.Resolve(r.Context(), p)
		if err == nil {
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

		fileServer.ServeHTTP(w, r)
	})
}
