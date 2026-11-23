package gitactivity

import (
	"encoding/json"
	"errors"
	"net/http"

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
	mux.HandleFunc("POST /api/v1/git/commits", h.requireAuth(h.Create))
	mux.HandleFunc("GET /api/v1/git/commits", h.requireAuth(h.List))
}
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	u, ok := user.UserFromContext(r.Context())
	if !ok {
		httpserver.Error(w, 401, "unauthorized", "authentication required", nil)
		return
	}
	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpserver.Error(w, 400, "invalid_json", "invalid request body", nil)
		return
	}
	c, err := h.service.Create(r.Context(), u.ID, req)
	if err != nil {
		h.writeError(w, err)
		return
	}
	httpserver.JSON(w, 201, c)
}
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	u, ok := user.UserFromContext(r.Context())
	if !ok {
		httpserver.Error(w, 401, "unauthorized", "authentication required", nil)
		return
	}
	items, err := h.service.List(r.Context(), u.ID, r.URL.Query().Get("project_id"))
	if err != nil {
		httpserver.Error(w, 500, "internal_error", "failed to list git commits", nil)
		return
	}
	httpserver.JSON(w, 200, map[string]any{"commits": items})
}
func (h *Handler) writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrProjectNotFound):
		httpserver.Error(w, 404, "project_not_found", "project not found", nil)
	case err.Error() == "project_id, commit_hash, and committed_at are required":
		httpserver.Error(w, 400, "validation_error", err.Error(), nil)
	default:
		httpserver.Error(w, 500, "internal_error", "internal server error", nil)
	}
}
