package report

import (
	"context"
	"net/http"
	"time"

	"devtrail/backend/internal/platform/httpserver"
	"devtrail/backend/internal/user"
)

type Generator interface {
	Generate(ctx context.Context, userID, projectID, reportType string, start, end time.Time) (Summary, error)
}

type Handler struct {
	service     Generator
	requireAuth func(http.HandlerFunc) http.HandlerFunc
}

func NewHandler(service Generator, requireAuth func(http.HandlerFunc) http.HandlerFunc) *Handler {
	return &Handler{service: service, requireAuth: requireAuth}
}
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/reports/daily", h.requireAuth(h.Daily))
	mux.HandleFunc("GET /api/v1/reports/weekly", h.requireAuth(h.Weekly))
	mux.HandleFunc("GET /api/v1/reports/projects/{projectId}", h.requireAuth(h.Project))
}
func (h *Handler) Daily(w http.ResponseWriter, r *http.Request) {
	u, ok := user.UserFromContext(r.Context())
	if !ok {
		httpserver.Error(w, 401, "unauthorized", "authentication required", nil)
		return
	}
	day := parseDay(r.URL.Query().Get("date"))
	h.respond(w, r, u.ID, "", "daily", day, day.Add(24*time.Hour))
}
func (h *Handler) Weekly(w http.ResponseWriter, r *http.Request) {
	u, ok := user.UserFromContext(r.Context())
	if !ok {
		httpserver.Error(w, 401, "unauthorized", "authentication required", nil)
		return
	}
	day := parseDay(r.URL.Query().Get("date"))
	weekday := int(day.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	start := day.AddDate(0, 0, -(weekday - 1))
	h.respond(w, r, u.ID, "", "weekly", start, start.AddDate(0, 0, 7))
}
func (h *Handler) Project(w http.ResponseWriter, r *http.Request) {
	u, ok := user.UserFromContext(r.Context())
	if !ok {
		httpserver.Error(w, 401, "unauthorized", "authentication required", nil)
		return
	}
	end := time.Now().UTC()
	h.respond(w, r, u.ID, r.PathValue("projectId"), "project", end.AddDate(0, 0, -30), end)
}
func (h *Handler) respond(w http.ResponseWriter, r *http.Request, userID, projectID, reportType string, start, end time.Time) {
	sum, err := h.service.Generate(r.Context(), userID, projectID, reportType, start, end)
	if err != nil {
		httpserver.Error(w, 500, "internal_error", "failed to generate report", nil)
		return
	}
	if r.URL.Query().Get("format") == "md" {
		w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(sum.Markdown()))
		return
	}
	httpserver.JSON(w, 200, sum)
}
func parseDay(v string) time.Time {
	if v == "" {
		now := time.Now().UTC()
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	}
	t, err := time.Parse("2006-01-02", v)
	if err != nil {
		now := time.Now().UTC()
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	}
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}
