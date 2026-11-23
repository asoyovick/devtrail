package user

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"devtrail/backend/internal/platform/httpserver"
)

type contextKey string

const (
	userContextKey    contextKey = "user"
	sessionContextKey contextKey = "auth_session"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/auth/register", h.Register)
	mux.HandleFunc("POST /api/v1/auth/login", h.Login)
	mux.HandleFunc("POST /api/v1/auth/logout", h.RequireAuth(h.Logout))
	mux.HandleFunc("GET /api/v1/me", h.RequireAuth(h.Me))
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpserver.Error(w, http.StatusBadRequest, "invalid_json", "invalid request body", nil)
		return
	}

	resp, err := h.service.Register(r.Context(), req, r.UserAgent())
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	httpserver.JSON(w, http.StatusCreated, resp)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpserver.Error(w, http.StatusBadRequest, "invalid_json", "invalid request body", nil)
		return
	}

	resp, err := h.service.Login(r.Context(), req, r.UserAgent())
	if err != nil {
		if err.Error() == "invalid credentials" {
			httpserver.Error(w, http.StatusUnauthorized, "invalid_credentials", "invalid email or password", nil)
			return
		}
		h.writeServiceError(w, err)
		return
	}
	httpserver.JSON(w, http.StatusOK, resp)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	sess, ok := SessionFromContext(r.Context())
	if !ok {
		httpserver.Error(w, http.StatusUnauthorized, "unauthorized", "authentication required", nil)
		return
	}
	if err := h.service.Logout(r.Context(), sess.ID); err != nil && !errors.Is(err, ErrNotFound) {
		httpserver.Error(w, http.StatusInternalServerError, "internal_error", "failed to logout", nil)
		return
	}
	httpserver.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	u, ok := UserFromContext(r.Context())
	if !ok {
		httpserver.Error(w, http.StatusUnauthorized, "unauthorized", "authentication required", nil)
		return
	}
	httpserver.JSON(w, http.StatusOK, u)
}

func (h *Handler) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := bearerToken(r.Header.Get("Authorization"))
		if token == "" {
			httpserver.Error(w, http.StatusUnauthorized, "unauthorized", "authentication required", nil)
			return
		}
		u, sess, err := h.service.Authenticate(r.Context(), token)
		if err != nil {
			httpserver.Error(w, http.StatusUnauthorized, "unauthorized", "invalid or expired session", nil)
			return
		}
		ctx := context.WithValue(r.Context(), userContextKey, u)
		ctx = context.WithValue(ctx, sessionContextKey, sess)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func UserFromContext(ctx context.Context) (User, bool) {
	u, ok := ctx.Value(userContextKey).(User)
	return u, ok
}

func SessionFromContext(ctx context.Context) (AuthSession, bool) {
	s, ok := ctx.Value(sessionContextKey).(AuthSession)
	return s, ok
}

func bearerToken(header string) string {
	if header == "" {
		return ""
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func (h *Handler) writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrEmailExists):
		httpserver.Error(w, http.StatusConflict, "email_exists", "email already registered", nil)
	case err.Error() == "invalid email", err.Error() == "display_name is required", err.Error() == "password must be at least 8 characters":
		httpserver.Error(w, http.StatusBadRequest, "validation_error", err.Error(), nil)
	case err.Error() == "invalid credentials":
		httpserver.Error(w, http.StatusUnauthorized, "invalid_credentials", "invalid email or password", nil)
	default:
		httpserver.Error(w, http.StatusInternalServerError, "internal_error", "internal server error", nil)
	}
}
