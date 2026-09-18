package server

import (
	"io/fs"
	"net/http"

	"pivotal/internal/job"
	"pivotal/internal/link"
	"pivotal/internal/middleware"
)

type Server struct {
	handler http.Handler
}

func NewServer(linkSvc link.Service, webFS fs.FS, jobMgr *job.Manager) *Server {
	mux := http.NewServeMux()

	linkHandler := link.NewHandler(linkSvc)
	linkHandler.RegisterRoutes(mux)

	mux.HandleFunc("GET /_health", handleHealth)

	mux.Handle("/", WebHandler(webFS, linkSvc))

	s := &Server{
		handler: middleware.Chain(mux, middleware.Recoverer, middleware.Logger),
	}
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.handler.ServeHTTP(w, r)
}
