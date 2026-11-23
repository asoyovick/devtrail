package decision

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"devtrail/backend/internal/user"
)

type fakeStore struct{ items map[string]Decision }

func newFakeStore() *fakeStore { return &fakeStore{items: map[string]Decision{}} }
func (s *fakeStore) Create(_ context.Context, userID string, req CreateRequest) (Decision, error) {
	if req.ProjectID == "missing" {
		return Decision{}, ErrProjectNotFound
	}
	d := Decision{ID: "d1", UserID: userID, ProjectID: req.ProjectID, Title: req.Title, Context: req.Context, Problem: req.Problem, SelectedApproach: req.SelectedApproach, Reasoning: req.Reasoning, Status: "accepted", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	s.items[d.ID] = d
	return d, nil
}
func (s *fakeStore) ListByUser(_ context.Context, userID, projectID string) ([]Decision, error) {
	out := []Decision{}
	for _, d := range s.items {
		if d.UserID == userID {
			out = append(out, d)
		}
	}
	return out, nil
}
func (s *fakeStore) GetByID(_ context.Context, userID, id string) (Decision, error) {
	d, ok := s.items[id]
	if !ok || d.UserID != userID {
		return Decision{}, ErrNotFound
	}
	return d, nil
}
func (s *fakeStore) Update(_ context.Context, userID, id string, req UpdateRequest) (Decision, error) {
	d, err := s.GetByID(context.Background(), userID, id)
	if err != nil {
		return Decision{}, err
	}
	if req.Title != nil {
		d.Title = *req.Title
	}
	s.items[id] = d
	return d, nil
}
func auth(userID string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r.WithContext(user.WithUser(r.Context(), user.User{ID: userID})))
		}
	}
}

func TestDecisionCreateList(t *testing.T) {
	h := NewHandler(NewService(newFakeStore()), auth("u1"))
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	body := bytes.NewBufferString(`{"project_id":"p1","title":"T","context":"C","problem":"P","selected_approach":"A","reasoning":"R"}`)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/api/v1/decisions", body))
	if rr.Code != 201 {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
}
