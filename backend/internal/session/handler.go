package session

import (
	"context"
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
	mux.HandleFunc("POST /api/v1/sessions", h.requireAuth(h.Start))
	mux.HandleFunc("GET /api/v1/sessions", h.requireAuth(h.List))
	mux.HandleFunc("GET /api/v1/sessions/{sessionId}", h.requireAuth(h.Get))
	mux.HandleFunc("PATCH /api/v1/sessions/{sessionId}/pause", h.requireAuth(h.Pause))
	mux.HandleFunc("PATCH /api/v1/sessions/{sessionId}/resume", h.requireAuth(h.Resume))
	mux.HandleFunc("PATCH /api/v1/sessions/{sessionId}/end", h.requireAuth(h.End))
}

func (h *Handler) Start(w http.ResponseWriter, r *http.Request) {
	u, ok := user.UserFromContext(r.Context())
	if !ok {
		httpserver.Error(w, http.StatusUnauthorized, "unauthorized", "authentication required", nil)
		return
	}
	var req StartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpserver.Error(w, http.StatusBadRequest, "invalid_json", "invalid request body", nil)
		return
	}
	sess, err := h.service.Start(r.Context(), u.ID, req)
	if err != nil {
		h.writeError(w, err)
		return
	}
	httpserver.JSON(w, http.StatusCreated, sess)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	u, ok := user.UserFromContext(r.Context())
	if !ok {
		httpserver.Error(w, http.StatusUnauthorized, "unauthorized", "authentication required", nil)
		return
	}
	sessions, err := h.service.List(r.Context(), u.ID, r.URL.Query().Get("project_id"))
	if err != nil {
		httpserver.Error(w, http.StatusInternalServerError, "internal_error", "failed to list sessions", nil)
		return
	}
	httpserver.JSON(w, http.StatusOK, map[string]any{"sessions": sessions})
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	u, ok := user.UserFromContext(r.Context())
	if !ok {
		httpserver.Error(w, http.StatusUnauthorized, "unauthorized", "authentication required", nil)
		return
	}
	sess, err := h.service.Get(r.Context(), u.ID, r.PathValue("sessionId"))
	if err != nil {
		h.writeError(w, err)
		return
	}
	httpserver.JSON(w, http.StatusOK, sess)
}

func (h *Handler) Pause(w http.ResponseWriter, r *http.Request) { h.transition(w, r, h.service.Pause) }
func (h *Handler) Resume(w http.ResponseWriter, r *http.Request) {
	h.transition(w, r, h.service.Resume)
}
func (h *Handler) End(w http.ResponseWriter, r *http.Request) { h.transition(w, r, h.service.End) }

func (h *Handler) transition(w http.ResponseWriter, r *http.Request, fn func(context.Context, string, string) (Session, error)) {
	u, ok := user.UserFromContext(r.Context())
	if !ok {
		httpserver.Error(w, http.StatusUnauthorized, "unauthorized", "authentication required", nil)
		return
	}
	sess, err := fn(r.Context(), u.ID, r.PathValue("sessionId"))
	if err != nil {
		h.writeError(w, err)
		return
	}
	httpserver.JSON(w, http.StatusOK, sess)
}

func (h *Handler) writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		httpserver.Error(w, http.StatusNotFound, "not_found", "session not found", nil)
	case errors.Is(err, ErrProjectNotFound):
		httpserver.Error(w, http.StatusNotFound, "project_not_found", "project not found", nil)
	case errors.Is(err, ErrInvalidTransition):
		httpserver.Error(w, http.StatusConflict, "invalid_transition", "invalid session state transition", nil)
	case err.Error() == "project_id is required", err.Error() == "invalid source":
		httpserver.Error(w, http.StatusBadRequest, "validation_error", err.Error(), nil)
	default:
		httpserver.Error(w, http.StatusInternalServerError, "internal_error", "internal server error", nil)
	}
}
