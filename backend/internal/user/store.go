package user

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("not found")
var ErrEmailExists = errors.New("email already exists")

type Store interface {
	CreateUser(ctx context.Context, email, passwordHash, displayName string) (User, error)
	GetUserByEmail(ctx context.Context, email string) (User, error)
	GetUserByID(ctx context.Context, id string) (User, error)
	CreateSession(ctx context.Context, userID, tokenHash, userAgent string, expiresAt time.Time) (AuthSession, error)
	GetSessionByTokenHash(ctx context.Context, tokenHash string) (AuthSession, error)
	RevokeSession(ctx context.Context, sessionID string) error
}

type PGStore struct {
	pool *pgxpool.Pool
}

func NewPGStore(pool *pgxpool.Pool) *PGStore {
	return &PGStore{pool: pool}
}

func (s *PGStore) CreateUser(ctx context.Context, email, passwordHash, displayName string) (User, error) {
	email = normalizeEmail(email)
	var u User
	err := s.pool.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, display_name)
		VALUES ($1, $2, $3)
		RETURNING id, email, display_name, password_hash, created_at, updated_at
	`, email, passwordHash, displayName).Scan(&u.ID, &u.Email, &u.DisplayName, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique") {
			return User{}, ErrEmailExists
		}
		return User{}, fmt.Errorf("create user: %w", err)
	}
	return u, nil
}

func (s *PGStore) GetUserByEmail(ctx context.Context, email string) (User, error) {
	email = normalizeEmail(email)
	var u User
	err := s.pool.QueryRow(ctx, `
		SELECT id, email, display_name, password_hash, created_at, updated_at
		FROM users WHERE email = $1
	`, email).Scan(&u.ID, &u.Email, &u.DisplayName, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrNotFound
	}
	if err != nil {
		return User{}, fmt.Errorf("get user by email: %w", err)
	}
	return u, nil
}

func (s *PGStore) GetUserByID(ctx context.Context, id string) (User, error) {
	var u User
	err := s.pool.QueryRow(ctx, `
		SELECT id, email, display_name, password_hash, created_at, updated_at
		FROM users WHERE id = $1
	`, id).Scan(&u.ID, &u.Email, &u.DisplayName, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrNotFound
	}
	if err != nil {
		return User{}, fmt.Errorf("get user by id: %w", err)
	}
	return u, nil
}

func (s *PGStore) CreateSession(ctx context.Context, userID, tokenHash, userAgent string, expiresAt time.Time) (AuthSession, error) {
	var sess AuthSession
	err := s.pool.QueryRow(ctx, `
		INSERT INTO auth_sessions (user_id, token_hash, expires_at, user_agent)
		VALUES ($1, $2, $3, $4)
		RETURNING id, user_id, token_hash, expires_at, revoked_at, user_agent, created_at
	`, userID, tokenHash, expiresAt, userAgent).Scan(&sess.ID, &sess.UserID, &sess.TokenHash, &sess.ExpiresAt, &sess.RevokedAt, &sess.UserAgent, &sess.CreatedAt)
	if err != nil {
		return AuthSession{}, fmt.Errorf("create session: %w", err)
	}
	return sess, nil
}

func (s *PGStore) GetSessionByTokenHash(ctx context.Context, tokenHash string) (AuthSession, error) {
	var sess AuthSession
	err := s.pool.QueryRow(ctx, `
		SELECT id, user_id, token_hash, expires_at, revoked_at, user_agent, created_at
		FROM auth_sessions WHERE token_hash = $1
	`, tokenHash).Scan(&sess.ID, &sess.UserID, &sess.TokenHash, &sess.ExpiresAt, &sess.RevokedAt, &sess.UserAgent, &sess.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return AuthSession{}, ErrNotFound
	}
	if err != nil {
		return AuthSession{}, fmt.Errorf("get session: %w", err)
	}
	return sess, nil
}

func (s *PGStore) RevokeSession(ctx context.Context, sessionID string) error {
	tag, err := s.pool.Exec(ctx, `UPDATE auth_sessions SET revoked_at = NOW() WHERE id = $1 AND revoked_at IS NULL`, sessionID)
	if err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
