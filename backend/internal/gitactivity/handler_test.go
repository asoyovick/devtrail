package gitactivity

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"devtrail/backend/internal/user"
)

type fakeStore struct{ items []Commit }

func (s *fakeStore) Create(_ context.Context, userID string, req CreateRequest) (Commit, error) {
	c := Commit{ID: "c1", UserID: userID, ProjectID: req.ProjectID, CommitHash: req.CommitHash, CommittedAt: req.CommittedAt, CreatedAt: time.Now()}
	s.items = append(s.items, c)
	return c, nil
}
func (s *fakeStore) ListByUser(_ context.Context, userID, _ string) ([]Commit, error) {
	out := []Commit{}
	for _, c := range s.items {
		if c.UserID == userID {
			out = append(out, c)
		}
	}
	return out, nil
}
func auth(userID string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r.WithContext(user.WithUser(r.Context(), user.User{ID: userID})))
		}
	}
}

func TestCreateCommit(t *testing.T) {
	h := NewHandler(NewService(&fakeStore{}), auth("u1"))
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	body := bytes.NewBufferString(`{"project_id":"p1","commit_hash":"abc123","committed_at":"2026-11-23T10:00:00Z"}`)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/api/v1/git/commits", body))
	if rr.Code != 201 {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
}
