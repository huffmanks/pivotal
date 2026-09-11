package server

import "net/http"

func (s *Server) handleListClicks(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		http.Error(w, `{"error":"Invalid link ID"}`, http.StatusBadRequest)
		return
	}

	clicks, err := s.store.ListClicksByLinkID(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"Database error"}`, http.StatusInternalServerError)
		return
	}

	WriteJSON(w, http.StatusOK, clicks)
}

func (s *Server) handleGetClick(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "clickID")
	if err != nil {
		http.Error(w, `{"error":"Invalid click ID"}`, http.StatusBadRequest)
		return
	}

	click, found, err := s.store.GetClickByID(r.Context(), id)
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
