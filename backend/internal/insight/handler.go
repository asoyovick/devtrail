package insight

import (
	"context"
	"net/http"

	"devtrail/backend/internal/platform/httpserver"
	"devtrail/backend/internal/user"
)

type Provider interface {
	List(ctx context.Context, userID string) ([]Insight, error)
	RecurringProblems(ctx context.Context, userID string) ([]Insight, error)
	ImprovementAreas(ctx context.Context, userID string) ([]Insight, error)
}

type Handler struct {
	service     Provider
	requireAuth func(http.HandlerFunc) http.HandlerFunc
}

func NewHandler(service Provider, requireAuth func(http.HandlerFunc) http.HandlerFunc) *Handler {
	return &Handler{service: service, requireAuth: requireAuth}
}
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/insights", h.requireAuth(h.List))
	mux.HandleFunc("GET /api/v1/insights/recurring-problems", h.requireAuth(h.Recurring))
	mux.HandleFunc("GET /api/v1/insights/improvement-areas", h.requireAuth(h.Improvement))
}
func (h *Handler) List(w http.ResponseWriter, r *http.Request) { h.respond(w, r, h.service.List) }
func (h *Handler) Recurring(w http.ResponseWriter, r *http.Request) {
	h.respond(w, r, h.service.RecurringProblems)
}
func (h *Handler) Improvement(w http.ResponseWriter, r *http.Request) {
	h.respond(w, r, h.service.ImprovementAreas)
}
func (h *Handler) respond(w http.ResponseWriter, r *http.Request, fn func(context.Context, string) ([]Insight, error)) {
	u, ok := user.UserFromContext(r.Context())
	if !ok {
		httpserver.Error(w, 401, "unauthorized", "authentication required", nil)
		return
	}
	items, err := fn(r.Context(), u.ID)
	if err != nil {
		httpserver.Error(w, 500, "internal_error", "failed to generate insights", nil)
		return
	}
	if items == nil {
		items = []Insight{}
	}
	httpserver.JSON(w, 200, map[string]any{"insights": items})
}
