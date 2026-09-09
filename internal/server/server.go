package server

import (
	"io/fs"
	"net/http"

	"url-shortener/internal/link"
)

type Server struct {
	store   *link.Store
	handler http.Handler
}

func New(store *link.Store, web fs.FS) *Server {
	mux := http.NewServeMux()

	s := &Server{
		store: store,
	}

	mux.HandleFunc("POST /api/shorten", s.handleShorten)
	mux.HandleFunc("GET /_health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})
	mux.Handle("/", WebHandler(web, store))

	s.handler = Chain(mux, Recoverer, Logger)
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.handler.ServeHTTP(w, r)
}
