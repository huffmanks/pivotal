package link

import (
	"errors"
	"net/http"
	"strconv"

	"url-shortener/internal/platform/utils"
)

type Handler struct {
	service Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{service: svc}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/shorten", h.handleShorten)
	mux.HandleFunc("GET /api/links", h.handleListLinks)
	mux.HandleFunc("GET /api/links/{id}", h.handleGetLink)
	mux.HandleFunc("GET /api/links/{id}/clicks", h.handleListClicks)
	mux.HandleFunc("GET /api/links/{id}/clicks/{clickID}", h.handleGetClick)
}

func (h *Handler) handleShorten(w http.ResponseWriter, r *http.Request) {
	var req CreateLinkRequest
	if err := utils.ParseJSON(w, r, &req); err != nil || req.DestinationURL == "" {
		http.Error(w, `{"error":"Invalid payload"}`, http.StatusBadRequest)
		return
	}

	created, err := h.service.Create(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidURL):
			http.Error(w, `{"error":"Invalid destination URL"}`, http.StatusBadRequest)
		case errors.Is(err, ErrReservedSlug):
			http.Error(w, `{"error":"Slug is a reserved path"}`, http.StatusBadRequest)
		case errors.Is(err, ErrSlugExists):
			http.Error(w, `{"error":"Slug already exists"}`, http.StatusConflict)
		default:
			http.Error(w, `{"error":"Database error"}`, http.StatusInternalServerError)
		}
		return
	}

	utils.WriteJSON(w, http.StatusCreated, created)
}

func (h *Handler) handleListLinks(w http.ResponseWriter, r *http.Request) {
	links, err := h.service.List(r.Context())
	if err != nil {
		http.Error(w, `{"error":"Database error"}`, http.StatusInternalServerError)
		return
	}
	utils.WriteJSON(w, http.StatusOK, links)
}

func (h *Handler) handleGetLink(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, `{"error":"Invalid link ID"}`, http.StatusBadRequest)
		return
	}

	l, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			http.Error(w, `{"error":"Link not found"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error":"Database error"}`, http.StatusInternalServerError)
		return
	}

	utils.WriteJSON(w, http.StatusOK, l)
}

func (h *Handler) handleListClicks(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, `{"error":"Invalid link ID"}`, http.StatusBadRequest)
		return
	}

	clicks, err := h.service.ListClicks(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"Database error"}`, http.StatusInternalServerError)
		return
	}

	utils.WriteJSON(w, http.StatusOK, clicks)
}

func (h *Handler) handleGetClick(w http.ResponseWriter, r *http.Request) {
	clickID, err := strconv.ParseInt(r.PathValue("clickID"), 10, 64)
	if err != nil {
		http.Error(w, `{"error":"Invalid click ID"}`, http.StatusBadRequest)
		return
	}

	click, err := h.service.GetClickByID(r.Context(), clickID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			http.Error(w, `{"error":"Click not found"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error":"Database error"}`, http.StatusInternalServerError)
		return
	}

	utils.WriteJSON(w, http.StatusOK, click)
}
