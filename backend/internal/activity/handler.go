package activity

import (
	"encoding/json"
	"net/http"
	"strconv"

	"devtrail/backend/internal/platform/httpserver"
	"devtrail/backend/internal/user"
)

type Handler struct {
	service     *Service
	requireAuth func(http.HandlerFunc) http.HandlerFunc
}

func NewHandler(service *Service, requireAuth func(http.HandlerFunc) http.HandlerFunc) *Handler {
	return &Handler{service: service, requireAuth: requireAuth}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/sync/events", h.requireAuth(h.Ingest))
	mux.HandleFunc("GET /api/v1/activity", h.requireAuth(h.List))
}

func (h *Handler) Ingest(w http.ResponseWriter, r *http.Request) {
	u, ok := user.UserFromContext(r.Context())
	if !ok {
		httpserver.Error(w, http.StatusUnauthorized, "unauthorized", "authentication required", nil)
		return
	}
	var req IngestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpserver.Error(w, http.StatusBadRequest, "invalid_json", "invalid request body", nil)
		return
	}
	resp, err := h.service.IngestBatch(r.Context(), u.ID, req)
	if err != nil {
		httpserver.Error(w, http.StatusInternalServerError, "internal_error", "failed to ingest events", nil)
		return
	}
	httpserver.JSON(w, http.StatusOK, resp)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	u, ok := user.UserFromContext(r.Context())
	if !ok {
		httpserver.Error(w, http.StatusUnauthorized, "unauthorized", "authentication required", nil)
		return
	}
	limit := 50
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			limit = n
		}
	}
	events, err := h.service.List(r.Context(), u.ID, r.URL.Query().Get("project_id"), limit)
	if err != nil {
		httpserver.Error(w, http.StatusInternalServerError, "internal_error", "failed to list events", nil)
		return
	}
	httpserver.JSON(w, http.StatusOK, map[string]any{"events": events})
}
