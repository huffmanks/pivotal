package server

import (
	"net/http"
	"strconv"

	"url-shortener/internal/db"
	"url-shortener/internal/link"
)

func (s *Server) handleListClicks(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, `{"error":"Invalid link ID"}`, http.StatusBadRequest)
		return
	}

	clicks, err := db.List[link.ClickEvent](r.Context(), s.db, `
		SELECT
			link_id,
			referer,
			user_agent,
			clicked_at
		FROM link_clicks
		WHERE link_id = ?
		ORDER BY clicked_at DESC
	`, id)
	if err != nil {
		http.Error(w, `{"error":"Database error"}`, http.StatusInternalServerError)
		return
	}

	WriteJSON(w, http.StatusOK, clicks)
}

func (s *Server) handleGetClick(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("clickID"), 10, 64)
	if err != nil {
		http.Error(w, `{"error":"Invalid click ID"}`, http.StatusBadRequest)
		return
	}

	click, found, err := db.GetByID[link.ClickEvent](
		r.Context(),
		s.db,
		`
			SELECT
				link_id,
				referer,
				user_agent,
				clicked_at
			FROM link_clicks
			WHERE id = ?
		`,
		id,
	)
	if err != nil {
		http.Error(w, `{"error":"Database error"}`, http.StatusInternalServerError)
		return
	}

	if !found {
		http.Error(w, `{"error":"Click not found"}`, http.StatusNotFound)
		return
	}

	WriteJSON(w, http.StatusOK, click)
}
