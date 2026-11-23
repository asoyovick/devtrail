package user

import (
	"context"
	"fmt"
	"strings"
	"time"

	"devtrail/backend/internal/platform/auth"
)

type Service struct {
	store    Store
	tokenTTL time.Duration
}

func NewService(store Store, tokenTTL time.Duration) *Service {
	if tokenTTL <= 0 {
		tokenTTL = 72 * time.Hour
	}
	return &Service{store: store, tokenTTL: tokenTTL}
}

func (s *Service) Register(ctx context.Context, req RegisterRequest, userAgent string) (AuthResponse, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))
	displayName := strings.TrimSpace(req.DisplayName)
	password := req.Password

	if email == "" || !strings.Contains(email, "@") {
		return AuthResponse{}, fmt.Errorf("invalid email")
	}
	if len(password) < 8 {
		return AuthResponse{}, fmt.Errorf("password must be at least 8 characters")
	}
	if displayName == "" {
		return AuthResponse{}, fmt.Errorf("display_name is required")
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		return AuthResponse{}, err
	}

	u, err := s.store.CreateUser(ctx, email, hash, displayName)
	if err != nil {
		return AuthResponse{}, err
	}

	token, err := s.createSession(ctx, u.ID, userAgent)
	if err != nil {
		return AuthResponse{}, err
	}

	return AuthResponse{Token: token, User: sanitizeUser(u)}, nil
}

func (s *Service) Login(ctx context.Context, req LoginRequest, userAgent string) (AuthResponse, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))
	if email == "" || req.Password == "" {
		return AuthResponse{}, fmt.Errorf("invalid credentials")
	}

	u, err := s.store.GetUserByEmail(ctx, email)
	if err != nil {
		return AuthResponse{}, fmt.Errorf("invalid credentials")
	}
	if err := auth.CheckPassword(u.PasswordHash, req.Password); err != nil {
		return AuthResponse{}, fmt.Errorf("invalid credentials")
	}

	token, err := s.createSession(ctx, u.ID, userAgent)
	if err != nil {
		return AuthResponse{}, err
	}
	return AuthResponse{Token: token, User: sanitizeUser(u)}, nil
}

func (s *Service) Authenticate(ctx context.Context, token string) (User, AuthSession, error) {
	if strings.TrimSpace(token) == "" {
		return User{}, AuthSession{}, ErrNotFound
	}
	hash := auth.HashToken(token)
	sess, err := s.store.GetSessionByTokenHash(ctx, hash)
	if err != nil {
		return User{}, AuthSession{}, err
	}
	if sess.RevokedAt != nil || time.Now().After(sess.ExpiresAt) {
		return User{}, AuthSession{}, ErrNotFound
	}
	u, err := s.store.GetUserByID(ctx, sess.UserID)
	if err != nil {
		return User{}, AuthSession{}, err
	}
	return sanitizeUser(u), sess, nil
}

func (s *Service) Logout(ctx context.Context, sessionID string) error {
	return s.store.RevokeSession(ctx, sessionID)
}

func (s *Service) createSession(ctx context.Context, userID, userAgent string) (string, error) {
	token, err := auth.NewToken()
	if err != nil {
		return "", err
	}
	hash := auth.HashToken(token)
	expiresAt := time.Now().Add(s.tokenTTL)
	if _, err := s.store.CreateSession(ctx, userID, hash, userAgent, expiresAt); err != nil {
		return "", err
	}
	return token, nil
}

func sanitizeUser(u User) User {
	u.PasswordHash = ""
	return u
}
