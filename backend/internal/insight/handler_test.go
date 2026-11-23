package insight

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"devtrail/backend/internal/user"
)

type fakeSvc struct{}

func (fakeSvc) List(context.Context, string) ([]Insight, error) {
	return []Insight{{Type: "t", Scope: "global", Summary: "s", Observation: "o", Evidence: map[string]any{"n": 1}, Confidence: "high"}}, nil
}
func (fakeSvc) RecurringProblems(context.Context, string) ([]Insight, error) { return []Insight{}, nil }
func (fakeSvc) ImprovementAreas(context.Context, string) ([]Insight, error)  { return []Insight{}, nil }
func auth(userID string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r.WithContext(user.WithUser(r.Context(), user.User{ID: userID})))
		}
	}
}

func TestListInsights(t *testing.T) {
	h := NewHandler(fakeSvc{}, auth("u1"))
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/v1/insights", nil))
	if rr.Code != 200 {
		t.Fatalf("status=%d", rr.Code)
	}
}
