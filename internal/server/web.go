package server

import (
	"io/fs"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

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

func handleResolvedRedirect(
	w http.ResponseWriter,
	r *http.Request,
	store *link.Store,
	l link.Link,
	extraPath string,
) {
	targetURL, err := mergeQueryParamsAndPath(
		l.DestinationURL,
		r.URL.Query(),
		extraPath,
	)
	if err != nil {
		http.Error(
			w,
			"Failed to resolve redirect URL",
			http.StatusInternalServerError,
		)
		return
	}

	store.RecordClick(link.ClickEvent{
		LinkID:    l.ID,
		Referer:   r.Referer(),
		UserAgent: r.UserAgent(),
		ClickedAt: time.Now(),
	})

	http.Redirect(w, r, targetURL, http.StatusFound)
}

func mergeQueryParamsAndPath(rawDest string, incomingQuery url.Values, extraPath string) (string, error) {
	destURL, err := url.Parse(rawDest)
	if err != nil {
		return "", err
	}

	if extraPath != "" {
		destURL.Path = path.Join(destURL.Path, extraPath)
	}

	destQuery := destURL.Query()
	for key, values := range incomingQuery {
		destQuery.Del(key)
		for _, v := range values {
			destQuery.Add(key, v)
		}
	}

	destURL.RawQuery = destQuery.Encode()
	return destURL.String(), nil
}
