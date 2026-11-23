package investigation

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"devtrail/backend/internal/user"
)

type fakeStore struct{ items map[string]Investigation }

func newFakeStore() *fakeStore { return &fakeStore{items: map[string]Investigation{}} }
func (s *fakeStore) Create(_ context.Context, userID string, req CreateRequest) (Investigation, error) {
	if req.ProjectID == "missing" {
		return Investigation{}, ErrProjectNotFound
	}
	inv := Investigation{ID: "inv-1", UserID: userID, ProjectID: req.ProjectID, Title: req.Title, ProblemDescription: req.ProblemDescription, Status: "open", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	s.items[inv.ID] = inv
	return inv, nil
}
func (s *fakeStore) ListByUser(_ context.Context, userID, projectID string) ([]Investigation, error) {
	out := []Investigation{}
	for _, i := range s.items {
		if i.UserID == userID && (projectID == "" || i.ProjectID == projectID) {
			out = append(out, i)
		}
	}
	return out, nil
}
func (s *fakeStore) GetByID(_ context.Context, userID, id string) (Investigation, error) {
	i, ok := s.items[id]
	if !ok || i.UserID != userID {
		return Investigation{}, ErrNotFound
	}
	return i, nil
}
func (s *fakeStore) Update(_ context.Context, userID, id string, req UpdateRequest) (Investigation, error) {
	i, err := s.GetByID(context.Background(), userID, id)
	if err != nil {
		return Investigation{}, err
	}
	if req.Title != nil {
		i.Title = *req.Title
	}
	if req.Status != nil {
		i.Status = *req.Status
	}
	s.items[id] = i
	return i, nil
}
func auth(userID string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r.WithContext(user.WithUser(r.Context(), user.User{ID: userID})))
		}
	}
}

func TestInvestigationCRUD(t *testing.T) {
	h := NewHandler(NewService(newFakeStore()), auth("u1"))
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	body := bytes.NewBufferString(`{"project_id":"p1","title":"Bug","problem_description":"fails"}`)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/api/v1/investigations", body))
	if rr.Code != 201 {
		t.Fatalf("create=%d %s", rr.Code, rr.Body.String())
	}
	rr2 := httptest.NewRecorder()
	mux.ServeHTTP(rr2, httptest.NewRequest(http.MethodGet, "/api/v1/investigations", nil))
	if rr2.Code != 200 {
		t.Fatalf("list=%d", rr2.Code)
	}
	var list struct {
		Investigations []Investigation `json:"investigations"`
	}
	_ = json.Unmarshal(rr2.Body.Bytes(), &list)
	if len(list.Investigations) != 1 {
		t.Fatalf("expected 1")
	}
}
