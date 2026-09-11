package server

import (
	"database/sql"
	"io/fs"
	"net/http"

	"url-shortener/internal/link"
)

type Server struct {
	store   *link.Store
	db      *sql.DB
	handler http.Handler
}

func NewServer(store *link.Store, db *sql.DB, web fs.FS) *Server {
	s := &Server{
		store: store,
		db:    db,
	}

	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/shorten", s.handleShorten)

	mux.HandleFunc("GET /api/links", s.handleListLinks)
	mux.HandleFunc("GET /api/links/{id}", s.handleGetLink)

	mux.HandleFunc("GET /api/links/{id}/clicks", s.handleListClicks)
	mux.HandleFunc("GET /api/links/{id}/clicks/{clickID}", s.handleGetClick)

	mux.HandleFunc("GET /_health", s.handleHealth)

	mux.Handle("/", WebHandler(web, store))

	s.handler = Chain(mux, Recoverer, Logger)
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.handler.ServeHTTP(w, r)
}
