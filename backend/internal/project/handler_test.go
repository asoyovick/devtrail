package project

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

type fakeStore struct {
	projects map[string]Project
}

func newFakeStore() *fakeStore { return &fakeStore{projects: map[string]Project{}} }

func (s *fakeStore) Create(_ context.Context, userID string, req CreateRequest) (Project, error) {
	for _, p := range s.projects {
		if p.UserID == userID && p.Name == req.Name {
			return Project{}, ErrNameExists
		}
	}
	p := Project{ID: "proj-1", UserID: userID, Name: req.Name, Description: req.Description, RepoRootHash: req.RepoRootHash, DefaultBranch: req.DefaultBranch, Status: "active", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	s.projects[p.ID] = p
	return p, nil
}

func (s *fakeStore) ListByUser(_ context.Context, userID string) ([]Project, error) {
	out := []Project{}
	for _, p := range s.projects {
		if p.UserID == userID {
			out = append(out, p)
		}
	}
	return out, nil
}

func (s *fakeStore) GetByID(_ context.Context, userID, projectID string) (Project, error) {
	p, ok := s.projects[projectID]
	if !ok || p.UserID != userID {
		return Project{}, ErrNotFound
	}
	return p, nil
}

func (s *fakeStore) Update(_ context.Context, userID, projectID string, req UpdateRequest) (Project, error) {
	p, ok := s.projects[projectID]
	if !ok || p.UserID != userID {
		return Project{}, ErrNotFound
	}
	if req.Name != nil {
		p.Name = *req.Name
	}
	if req.Description != nil {
		p.Description = *req.Description
	}
	if req.DefaultBranch != nil {
		p.DefaultBranch = *req.DefaultBranch
	}
	if req.Status != nil {
		p.Status = *req.Status
	}
	p.UpdatedAt = time.Now()
	s.projects[projectID] = p
	return p, nil
}

func fakeRequireAuth(userID string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx := user.WithUser(r.Context(), user.User{ID: userID, Email: userID + "@example.com"})
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	}
}

func setupProjectTest(t *testing.T, userID string) (*http.ServeMux, *fakeStore) {
	t.Helper()
	store := newFakeStore()
	h := NewHandler(NewService(store), fakeRequireAuth(userID))
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	return mux, store
}

func TestCreateAndListProjects(t *testing.T) {
	mux, _ := setupProjectTest(t, "user-1")
	body := bytes.NewBufferString(`{"name":"DevTrail","description":"test"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects", body)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("create status=%d body=%s", rr.Code, rr.Body.String())
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/projects", nil)
	listRR := httptest.NewRecorder()
	mux.ServeHTTP(listRR, listReq)
	if listRR.Code != http.StatusOK {
		t.Fatalf("list status=%d body=%s", listRR.Code, listRR.Body.String())
	}
	var resp struct {
		Projects []Project `json:"projects"`
	}
	if err := json.Unmarshal(listRR.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(resp.Projects) != 1 || resp.Projects[0].Name != "DevTrail" {
		t.Fatalf("unexpected projects: %+v", resp.Projects)
	}
}

func TestGetProjectNotFoundForOtherUser(t *testing.T) {
	mux, store := setupProjectTest(t, "user-2")
	store.projects["proj-1"] = Project{ID: "proj-1", UserID: "user-1", Name: "Secret", Status: "active"}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/proj-1", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
}

func TestArchiveProject(t *testing.T) {
	mux, store := setupProjectTest(t, "user-1")
	store.projects["proj-1"] = Project{ID: "proj-1", UserID: "user-1", Name: "DevTrail", Status: "active"}
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/projects/proj-1", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	if store.projects["proj-1"].Status != "archived" {
		t.Fatalf("expected archived, got %s", store.projects["proj-1"].Status)
	}
}
