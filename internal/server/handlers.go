package server

import (
	"encoding/json"
	"net/http"
	"net/url"
	"path"
	"strings"

	"url-shortener/internal/link"
)

type ShortenRequest struct {
	URL        string `json:"url"`
	CustomSlug string `json:"custom_slug,omitempty"`
}

type ShortenResponse struct {
	Slug           string `json:"slug"`
	URL            string `json:"url"`
	DestinationURL string `json:"destination_url"`
}

func (s *Server) handleShorten(w http.ResponseWriter, r *http.Request) {
	var req ShortenRequest
	if err := parseJSON(w, r, &req); err != nil || req.URL == "" {
		http.Error(w, `{"error":"Invalid payload"}`, http.StatusBadRequest)
		return
	}

	parsedURL, err := url.ParseRequestURI(req.URL)
	if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
		http.Error(w, `{"error":"Invalid destination URL"}`, http.StatusBadRequest)
		return
	}

	slug := strings.Trim(strings.TrimSpace(req.CustomSlug), "/")
	isCustom := slug != ""

	if isCustom {
		if isReservedSlug(slug) {
			http.Error(w, `{"error":"Slug is a reserved path"}`, http.StatusBadRequest)
			return
		}
	} else {
		slug = link.GenerateBase62ID(6)
	}

	created, err := s.store.Create(r.Context(), slug, req.URL, isCustom)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			http.Error(w, `{"error":"Slug already exists"}`, http.StatusConflict)
			return
		}
		http.Error(w, `{"error":"Database error"}`, http.StatusInternalServerError)
		return
	}

	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}

	shortURL := scheme + "://" + r.Host + "/" + created.Slug

	writeJSON(w, http.StatusCreated, ShortenResponse{
		Slug:           created.Slug,
		URL:            shortURL,
		DestinationURL: created.DestinationURL,
	})
}

func (s *Server) handleRedirect(w http.ResponseWriter, r *http.Request) {
	requestedPath := strings.Trim(r.URL.Path, "/")

	if requestedPath == "" {
		http.NotFound(w, r)
		return
	}

	l, extraPath, found := s.store.Resolve(r.Context(), requestedPath)
	if !found {
		http.NotFound(w, r)
		return
	}

	handleResolvedRedirect(w, r, s.store, l, extraPath)
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

func isReservedSlug(slug string) bool {
	switch slug {
	case "api", "_health", "_app":
		return true
	}

	return strings.HasPrefix(slug, "api/")
}

func parseJSON(w http.ResponseWriter, r *http.Request, v any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1048576)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
