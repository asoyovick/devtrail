package investigation

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
	mux.HandleFunc("POST /api/v1/investigations", h.requireAuth(h.Create))
	mux.HandleFunc("GET /api/v1/investigations", h.requireAuth(h.List))
	mux.HandleFunc("GET /api/v1/investigations/{id}", h.requireAuth(h.Get))
	mux.HandleFunc("PATCH /api/v1/investigations/{id}", h.requireAuth(h.Update))
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
	inv, err := h.service.Create(r.Context(), u.ID, req)
	if err != nil {
		h.writeError(w, err)
		return
	}
	httpserver.JSON(w, 201, inv)
}
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	u, ok := user.UserFromContext(r.Context())
	if !ok {
		httpserver.Error(w, 401, "unauthorized", "authentication required", nil)
		return
	}
	items, err := h.service.List(r.Context(), u.ID, r.URL.Query().Get("project_id"))
	if err != nil {
		httpserver.Error(w, 500, "internal_error", "failed to list investigations", nil)
		return
	}
	httpserver.JSON(w, 200, map[string]any{"investigations": items})
}
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	u, ok := user.UserFromContext(r.Context())
	if !ok {
		httpserver.Error(w, 401, "unauthorized", "authentication required", nil)
		return
	}
	inv, err := h.service.Get(r.Context(), u.ID, r.PathValue("id"))
	if err != nil {
		h.writeError(w, err)
		return
	}
	httpserver.JSON(w, 200, inv)
}
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	u, ok := user.UserFromContext(r.Context())
	if !ok {
		httpserver.Error(w, 401, "unauthorized", "authentication required", nil)
		return
	}
	var req UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpserver.Error(w, 400, "invalid_json", "invalid request body", nil)
		return
	}
	inv, err := h.service.Update(r.Context(), u.ID, r.PathValue("id"), req)
	if err != nil {
		h.writeError(w, err)
		return
	}
	httpserver.JSON(w, 200, inv)
}
func (h *Handler) writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		httpserver.Error(w, 404, "not_found", "investigation not found", nil)
	case errors.Is(err, ErrProjectNotFound):
		httpserver.Error(w, 404, "project_not_found", "project not found", nil)
	case err.Error() == "project_id, title, and problem_description are required" || err.Error() == "title and problem_description are required" || err.Error() == "invalid status":
		httpserver.Error(w, 400, "validation_error", err.Error(), nil)
	default:
		httpserver.Error(w, 500, "internal_error", "internal server error", nil)
	}
}
