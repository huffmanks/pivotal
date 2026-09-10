package server

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"url-shortener/internal/link"
)

type ShortenRequest struct {
	Slug           string `json:"slug,omitempty"`
	DestinationURL string `json:"destination_url"`
}

type ShortenResponse struct {
	Slug           string `json:"slug"`
	DestinationURL string `json:"destination_url"`
}

func (s *Server) handleShorten(w http.ResponseWriter, r *http.Request) {
	var req ShortenRequest
	if err := ParseJSON(w, r, &req); err != nil || req.DestinationURL == "" {
		http.Error(w, `{"error":"Invalid payload"}`, http.StatusBadRequest)
		return
	}

	parsedURL, err := url.ParseRequestURI(req.DestinationURL)
	if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
		http.Error(w, `{"error":"Invalid destination URL"}`, http.StatusBadRequest)
		return
	}

	slug := strings.Trim(strings.TrimSpace(req.Slug), "/")
	isCustom := slug != ""

	if isCustom {
		if isReservedSlug(slug) {
			http.Error(w, `{"error":"Slug is a reserved path"}`, http.StatusBadRequest)
			return
		}
	} else {
		slug = link.GenerateBase62ID(6)
	}

	created, err := s.store.Create(r.Context(), slug, req.DestinationURL, isCustom)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			http.Error(w, `{"error":"Slug already exists"}`, http.StatusConflict)
			return
		}
		http.Error(w, `{"error":"Database error"}`, http.StatusInternalServerError)
		return
	}

	WriteJSON(w, http.StatusCreated, ShortenResponse{
		Slug:           created.Slug,
		DestinationURL: created.DestinationURL,
	})
}

func (s *Server) handleListLinks(w http.ResponseWriter, r *http.Request) {
	links, err := s.store.List(r.Context())
	if err != nil {
		http.Error(w, `{"error":"Database error"}`, http.StatusInternalServerError)
		return
	}

	WriteJSON(w, http.StatusOK, links)
}

func (s *Server) handleGetLink(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, `{"error":"Invalid link ID"}`, http.StatusBadRequest)
		return
	}

	l, found, err := s.store.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"Database error"}`, http.StatusInternalServerError)
		return
	}

	if !found {
		http.Error(w, `{"error":"Link not found"}`, http.StatusNotFound)
		return
	}

	WriteJSON(w, http.StatusOK, l)
}

func isReservedSlug(slug string) bool {
	switch slug {
	case "api", "_health", "_app":
		return true
	}

	return strings.HasPrefix(slug, "api/")
}
