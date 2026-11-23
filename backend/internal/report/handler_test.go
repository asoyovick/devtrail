package report

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"devtrail/backend/internal/user"
)

type fakeGen struct{}

func (fakeGen) Generate(_ context.Context, _, _, reportType string, start, end time.Time) (Summary, error) {
	return Summary{ReportType: reportType, PeriodStart: start, PeriodEnd: end, Sessions: 2}, nil
}
func auth(userID string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r.WithContext(user.WithUser(r.Context(), user.User{ID: userID})))
		}
	}
}

func TestDailyReport(t *testing.T) {
	h := NewHandler(fakeGen{}, auth("u1"))
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/v1/reports/daily?date=2026-11-23&format=md", nil))
	if rr.Code != 200 {
		t.Fatalf("status=%d", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); ct == "" {
		t.Fatal("missing content type")
	}
}
