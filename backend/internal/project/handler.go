package project

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
	mux.HandleFunc("POST /api/v1/projects", h.requireAuth(h.Create))
	mux.HandleFunc("GET /api/v1/projects", h.requireAuth(h.List))
	mux.HandleFunc("GET /api/v1/projects/{projectId}", h.requireAuth(h.Get))
	mux.HandleFunc("PATCH /api/v1/projects/{projectId}", h.requireAuth(h.Update))
	mux.HandleFunc("DELETE /api/v1/projects/{projectId}", h.requireAuth(h.Archive))
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	u, ok := user.UserFromContext(r.Context())
	if !ok {
		httpserver.Error(w, http.StatusUnauthorized, "unauthorized", "authentication required", nil)
		return
	}
	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpserver.Error(w, http.StatusBadRequest, "invalid_json", "invalid request body", nil)
		return
	}
	p, err := h.service.Create(r.Context(), u.ID, req)
	if err != nil {
		h.writeError(w, err)
		return
	}
	httpserver.JSON(w, http.StatusCreated, p)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	u, ok := user.UserFromContext(r.Context())
	if !ok {
		httpserver.Error(w, http.StatusUnauthorized, "unauthorized", "authentication required", nil)
		return
	}
	projects, err := h.service.List(r.Context(), u.ID)
	if err != nil {
		httpserver.Error(w, http.StatusInternalServerError, "internal_error", "failed to list projects", nil)
		return
	}
	httpserver.JSON(w, http.StatusOK, map[string]any{"projects": projects})
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	u, ok := user.UserFromContext(r.Context())
	if !ok {
		httpserver.Error(w, http.StatusUnauthorized, "unauthorized", "authentication required", nil)
		return
	}
	p, err := h.service.Get(r.Context(), u.ID, r.PathValue("projectId"))
	if err != nil {
		h.writeError(w, err)
		return
	}
	httpserver.JSON(w, http.StatusOK, p)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	u, ok := user.UserFromContext(r.Context())
	if !ok {
		httpserver.Error(w, http.StatusUnauthorized, "unauthorized", "authentication required", nil)
		return
	}
	var req UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpserver.Error(w, http.StatusBadRequest, "invalid_json", "invalid request body", nil)
		return
	}
	p, err := h.service.Update(r.Context(), u.ID, r.PathValue("projectId"), req)
	if err != nil {
		h.writeError(w, err)
		return
	}
	httpserver.JSON(w, http.StatusOK, p)
}

func (h *Handler) Archive(w http.ResponseWriter, r *http.Request) {
	u, ok := user.UserFromContext(r.Context())
	if !ok {
		httpserver.Error(w, http.StatusUnauthorized, "unauthorized", "authentication required", nil)
		return
	}
	status := "archived"
	p, err := h.service.Update(r.Context(), u.ID, r.PathValue("projectId"), UpdateRequest{Status: &status})
	if err != nil {
		h.writeError(w, err)
		return
	}
	httpserver.JSON(w, http.StatusOK, p)
}

func (h *Handler) writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		httpserver.Error(w, http.StatusNotFound, "not_found", "project not found", nil)
	case errors.Is(err, ErrNameExists):
		httpserver.Error(w, http.StatusConflict, "project_name_exists", "project name already exists", nil)
	case err.Error() == "name is required", err.Error() == "invalid status":
		httpserver.Error(w, http.StatusBadRequest, "validation_error", err.Error(), nil)
	default:
		httpserver.Error(w, http.StatusInternalServerError, "internal_error", "internal server error", nil)
	}
}
