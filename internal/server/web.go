package server

import (
	"io/fs"
	"net"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

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
			if l.Status != link.LinkStatusActive {
				if l.FallbackURL != "" {
					recordClick(linkSvc, l, r)
					http.Redirect(w, r, l.FallbackURL, getRedirectCode(l.RedirectType))
					return
				}
				if l.Status == link.LinkStatusExpired {
					http.Error(w, `{"error":"Link expired"}`, http.StatusGone)
					return
				}
				http.Error(w, `{"error":"Link disabled"}`, http.StatusNotFound)
				return
			}

			dest := l.DestinationURL
			if extraPath != "" {
				dest = strings.TrimSuffix(dest, "/") + "/" + strings.TrimPrefix(extraPath, "/")
			}

			recordClick(linkSvc, l, r)

			http.Redirect(w, r, dest, getRedirectCode(l.RedirectType))
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

func recordClick(linkSvc link.Service, l link.Link, r *http.Request) {
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	clickEvent := link.ClickEvent{
		LinkID:    l.ID,
		Referer:   r.Referer(),
		UserAgent: r.UserAgent(),
		ClickedAt: time.Now(),
		IP:        ip,
	}
	if qrID := r.URL.Query().Get("qr"); qrID != "" {
		if id, err := strconv.ParseInt(qrID, 10, 64); err == nil {
			clickEvent.QRCodeID = &id
		}
	}
	linkSvc.RecordClick(clickEvent)
}

func getRedirectCode(redirectType string) int {
	if redirectType == "301" {
		return http.StatusMovedPermanently
	}
	return http.StatusFound
}

func serveFile(w http.ResponseWriter, r *http.Request, files fs.FS, name string) {
	if f, err := files.Open(name); err == nil {
		_ = f.Close()
		http.ServeFileFS(w, r, files, name)
		return
	}
	http.NotFound(w, r)
}
