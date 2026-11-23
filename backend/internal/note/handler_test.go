package note

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"devtrail/backend/internal/user"
)

type fakeStore struct{ items map[string]Note }

func newFakeStore() *fakeStore { return &fakeStore{items: map[string]Note{}} }
func (s *fakeStore) Create(_ context.Context, userID string, req CreateRequest) (Note, error) {
	n := Note{ID: "n1", UserID: userID, ProjectID: req.ProjectID, Title: req.Title, Body: req.Body, Tags: req.Tags, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	if n.Tags == nil {
		n.Tags = []string{}
	}
	s.items[n.ID] = n
	return n, nil
}
func (s *fakeStore) ListByUser(_ context.Context, userID, _ string) ([]Note, error) {
	out := []Note{}
	for _, n := range s.items {
		if n.UserID == userID {
			out = append(out, n)
		}
	}
	return out, nil
}
func (s *fakeStore) GetByID(_ context.Context, userID, id string) (Note, error) {
	n, ok := s.items[id]
	if !ok || n.UserID != userID {
		return Note{}, ErrNotFound
	}
	return n, nil
}
func (s *fakeStore) Update(_ context.Context, userID, id string, req UpdateRequest) (Note, error) {
	n, err := s.GetByID(context.Background(), userID, id)
	if err != nil {
		return Note{}, err
	}
	if req.Title != nil {
		n.Title = *req.Title
	}
	if req.Body != nil {
		n.Body = *req.Body
	}
	s.items[id] = n
	return n, nil
}
func (s *fakeStore) Delete(_ context.Context, userID, id string) error {
	n, ok := s.items[id]
	if !ok || n.UserID != userID {
		return ErrNotFound
	}
	delete(s.items, id)
	return nil
}
func auth(userID string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r.WithContext(user.WithUser(r.Context(), user.User{ID: userID})))
		}
	}
}

func TestNoteCreate(t *testing.T) {
	h := NewHandler(NewService(newFakeStore()), auth("u1"))
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/api/v1/notes", bytes.NewBufferString(`{"title":"T","body":"B"}`)))
	if rr.Code != 201 {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
}
