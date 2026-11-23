package testrun

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"devtrail/backend/internal/user"
)

type fakeStore struct{ items map[string]TestRun }

func newFakeStore() *fakeStore { return &fakeStore{items: map[string]TestRun{}} }
func (s *fakeStore) Create(_ context.Context, userID string, req CreateRequest) (TestRun, error) {
	t := TestRun{ID: "t1", UserID: userID, ProjectID: req.ProjectID, Command: req.Command, Status: req.Status, StartedAt: req.StartedAt, CreatedAt: time.Now()}
	s.items[t.ID] = t
	return t, nil
}
func (s *fakeStore) ListByUser(_ context.Context, userID, _ string) ([]TestRun, error) {
	out := []TestRun{}
	for _, t := range s.items {
		if t.UserID == userID {
			out = append(out, t)
		}
	}
	return out, nil
}
func (s *fakeStore) GetByID(_ context.Context, userID, id string) (TestRun, error) {
	t, ok := s.items[id]
	if !ok || t.UserID != userID {
		return TestRun{}, ErrNotFound
	}
	return t, nil
}
func auth(userID string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r.WithContext(user.WithUser(r.Context(), user.User{ID: userID})))
		}
	}
}

func TestCreateTestRun(t *testing.T) {
	h := NewHandler(NewService(newFakeStore()), auth("u1"))
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	body := bytes.NewBufferString(`{"project_id":"p1","command":"go test ./...","status":"passed","started_at":"2026-11-23T10:00:00Z"}`)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/api/v1/test-runs", body))
	if rr.Code != 201 {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
}
