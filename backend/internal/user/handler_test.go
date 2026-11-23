package user

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type fakeStore struct {
	users    map[string]User
	sessions map[string]AuthSession
	byEmail  map[string]string
}

func newFakeStore() *fakeStore {
	return &fakeStore{users: map[string]User{}, sessions: map[string]AuthSession{}, byEmail: map[string]string{}}
}

func (s *fakeStore) CreateUser(_ context.Context, email, passwordHash, displayName string) (User, error) {
	if _, ok := s.byEmail[email]; ok {
		return User{}, ErrEmailExists
	}
	u := User{ID: "user-1", Email: email, DisplayName: displayName, PasswordHash: passwordHash, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	s.users[u.ID] = u
	s.byEmail[email] = u.ID
	return u, nil
}

func (s *fakeStore) GetUserByEmail(_ context.Context, email string) (User, error) {
	id, ok := s.byEmail[email]
	if !ok {
		return User{}, ErrNotFound
	}
	return s.users[id], nil
}

func (s *fakeStore) GetUserByID(_ context.Context, id string) (User, error) {
	u, ok := s.users[id]
	if !ok {
		return User{}, ErrNotFound
	}
	return u, nil
}

func (s *fakeStore) CreateSession(_ context.Context, userID, tokenHash, userAgent string, expiresAt time.Time) (AuthSession, error) {
	sess := AuthSession{ID: "sess-1", UserID: userID, TokenHash: tokenHash, ExpiresAt: expiresAt, UserAgent: userAgent, CreatedAt: time.Now()}
	s.sessions[tokenHash] = sess
	return sess, nil
}

func (s *fakeStore) GetSessionByTokenHash(_ context.Context, tokenHash string) (AuthSession, error) {
	sess, ok := s.sessions[tokenHash]
	if !ok {
		return AuthSession{}, ErrNotFound
	}
	return sess, nil
}

func (s *fakeStore) RevokeSession(_ context.Context, sessionID string) error {
	for k, sess := range s.sessions {
		if sess.ID == sessionID {
			now := time.Now()
			sess.RevokedAt = &now
			s.sessions[k] = sess
			return nil
		}
	}
	return ErrNotFound
}

func setupAuthTest(t *testing.T) (*Handler, *fakeStore) {
	t.Helper()
	store := newFakeStore()
	svc := NewService(store, time.Hour)
	return NewHandler(svc), store
}

func TestRegisterAndMe(t *testing.T) {
	h, _ := setupAuthTest(t)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	body := bytes.NewBufferString(`{"email":"Test@example.com","password":"password123","display_name":"Test User"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", body)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("register status=%d body=%s", rr.Code, rr.Body.String())
	}
	var authResp AuthResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &authResp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if authResp.Token == "" || authResp.User.Email != "test@example.com" {
		t.Fatalf("unexpected response: %+v", authResp)
	}

	meReq := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	meReq.Header.Set("Authorization", "Bearer "+authResp.Token)
	meRR := httptest.NewRecorder()
	mux.ServeHTTP(meRR, meReq)
	if meRR.Code != http.StatusOK {
		t.Fatalf("me status=%d body=%s", meRR.Code, meRR.Body.String())
	}
}

func TestLoginInvalidCredentials(t *testing.T) {
	h, _ := setupAuthTest(t)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	body := bytes.NewBufferString(`{"email":"nobody@example.com","password":"password123"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", body)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
}

func TestMeUnauthorized(t *testing.T) {
	h, _ := setupAuthTest(t)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
}
