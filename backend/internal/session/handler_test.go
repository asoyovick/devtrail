package session

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

type fakeStore struct{ sessions map[string]Session }

func newFakeStore() *fakeStore { return &fakeStore{sessions: map[string]Session{}} }

func (s *fakeStore) Create(_ context.Context, userID string, req StartRequest) (Session, error) {
	if req.ProjectID == "missing" {
		return Session{}, ErrProjectNotFound
	}
	sess := Session{ID: "sess-1", UserID: userID, ProjectID: req.ProjectID, Title: req.Title, Status: "active", StartedAt: time.Now(), Source: req.Source, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	if sess.Source == "" {
		sess.Source = "web"
	}
	s.sessions[sess.ID] = sess
	return sess, nil
}
func (s *fakeStore) GetByID(_ context.Context, userID, sessionID string) (Session, error) {
	sess, ok := s.sessions[sessionID]
	if !ok || sess.UserID != userID {
		return Session{}, ErrNotFound
	}
	return sess, nil
}
func (s *fakeStore) ListByUser(_ context.Context, userID, projectID string) ([]Session, error) {
	out := []Session{}
	for _, sess := range s.sessions {
		if sess.UserID == userID && (projectID == "" || sess.ProjectID == projectID) {
			out = append(out, sess)
		}
	}
	return out, nil
}
func (s *fakeStore) SetStatus(_ context.Context, userID, sessionID, status string) (Session, error) {
	sess, err := s.GetByID(context.Background(), userID, sessionID)
	if err != nil {
		return Session{}, err
	}
	if !validTransition(sess.Status, status) {
		return Session{}, ErrInvalidTransition
	}
	sess.Status = status
	if status == "ended" {
		now := time.Now()
		sess.EndedAt = &now
		d := 1
		sess.DurationSeconds = &d
	}
	s.sessions[sessionID] = sess
	return sess, nil
}

func fakeRequireAuth(userID string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx := user.WithUser(r.Context(), user.User{ID: userID})
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	}
}

func TestSessionLifecycle(t *testing.T) {
	store := newFakeStore()
	h := NewHandler(NewService(store), fakeRequireAuth("user-1"))
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	body := bytes.NewBufferString(`{"project_id":"proj-1","title":"Work"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/sessions", body)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("start status=%d body=%s", rr.Code, rr.Body.String())
	}

	for _, step := range []struct {
		method, path string
		want         int
	}{
		{http.MethodPatch, "/api/v1/sessions/sess-1/pause", http.StatusOK},
		{http.MethodPatch, "/api/v1/sessions/sess-1/resume", http.StatusOK},
		{http.MethodPatch, "/api/v1/sessions/sess-1/end", http.StatusOK},
	} {
		r := httptest.NewRequest(step.method, step.path, nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)
		if w.Code != step.want {
			t.Fatalf("%s %s status=%d body=%s", step.method, step.path, w.Code, w.Body.String())
		}
	}

	var ended Session
	if err := json.Unmarshal(rr.Body.Bytes(), &ended); err == nil {
		// ignore; we checked statuses via calls
	}
	_ = ended
}

func TestSessionInvalidTransition(t *testing.T) {
	store := newFakeStore()
	store.sessions["sess-1"] = Session{ID: "sess-1", UserID: "user-1", ProjectID: "proj-1", Status: "ended", StartedAt: time.Now()}
	h := NewHandler(NewService(store), fakeRequireAuth("user-1"))
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/sessions/sess-1/pause", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusConflict {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
}
