package session

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
var ErrInvalidTransition = errors.New("invalid session transition")
var ErrProjectNotFound = errors.New("project not found")

type Store interface {
	Create(ctx context.Context, userID string, req StartRequest) (Session, error)
	GetByID(ctx context.Context, userID, sessionID string) (Session, error)
	ListByUser(ctx context.Context, userID, projectID string) ([]Session, error)
	SetStatus(ctx context.Context, userID, sessionID, status string) (Session, error)
}

type PGStore struct{ pool *pgxpool.Pool }

func NewPGStore(pool *pgxpool.Pool) *PGStore { return &PGStore{pool: pool} }

func (s *PGStore) Create(ctx context.Context, userID string, req StartRequest) (Session, error) {
	source := strings.TrimSpace(req.Source)
	if source == "" {
		source = "web"
	}
	var sess Session
	err := s.pool.QueryRow(ctx, `
		INSERT INTO development_sessions (user_id, project_id, title, status, source)
		SELECT $1, p.id, $2, 'active', $3
		FROM projects p WHERE p.id = $4 AND p.user_id = $1
		RETURNING id, user_id, project_id, COALESCE(title,''), status, started_at, ended_at, duration_seconds, source, created_at, updated_at
	`, userID, nullIfEmpty(req.Title), source, req.ProjectID).Scan(
		&sess.ID, &sess.UserID, &sess.ProjectID, &sess.Title, &sess.Status, &sess.StartedAt, &sess.EndedAt, &sess.DurationSeconds, &sess.Source, &sess.CreatedAt, &sess.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Session{}, ErrProjectNotFound
	}
	if err != nil {
		return Session{}, fmt.Errorf("create session: %w", err)
	}
	return sess, nil
}

func (s *PGStore) GetByID(ctx context.Context, userID, sessionID string) (Session, error) {
	var sess Session
	err := s.pool.QueryRow(ctx, `
		SELECT id, user_id, project_id, COALESCE(title,''), status, started_at, ended_at, duration_seconds, source, created_at, updated_at
		FROM development_sessions WHERE id = $1 AND user_id = $2
	`, sessionID, userID).Scan(&sess.ID, &sess.UserID, &sess.ProjectID, &sess.Title, &sess.Status, &sess.StartedAt, &sess.EndedAt, &sess.DurationSeconds, &sess.Source, &sess.CreatedAt, &sess.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Session{}, ErrNotFound
	}
	if err != nil {
		return Session{}, fmt.Errorf("get session: %w", err)
	}
	return sess, nil
}

func (s *PGStore) ListByUser(ctx context.Context, userID, projectID string) ([]Session, error) {
	query := `
		SELECT id, user_id, project_id, COALESCE(title,''), status, started_at, ended_at, duration_seconds, source, created_at, updated_at
		FROM development_sessions WHERE user_id = $1`
	args := []any{userID}
	if strings.TrimSpace(projectID) != "" {
		query += " AND project_id = $2"
		args = append(args, projectID)
	}
	query += " ORDER BY started_at DESC"
	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list sessions: %w", err)
	}
	defer rows.Close()
	out := []Session{}
	for rows.Next() {
		var sess Session
		if err := rows.Scan(&sess.ID, &sess.UserID, &sess.ProjectID, &sess.Title, &sess.Status, &sess.StartedAt, &sess.EndedAt, &sess.DurationSeconds, &sess.Source, &sess.CreatedAt, &sess.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan session: %w", err)
		}
		out = append(out, sess)
	}
	return out, rows.Err()
}

func (s *PGStore) SetStatus(ctx context.Context, userID, sessionID, status string) (Session, error) {
	current, err := s.GetByID(ctx, userID, sessionID)
	if err != nil {
		return Session{}, err
	}

	if !validTransition(current.Status, status) {
		return Session{}, ErrInvalidTransition
	}

	var endedAt *time.Time
	var duration *int
	now := time.Now()
	if status == "ended" {
		endedAt = &now
		d := int(now.Sub(current.StartedAt).Seconds())
		if d < 0 {
			d = 0
		}
		duration = &d
	} else {
		endedAt = current.EndedAt
		duration = current.DurationSeconds
	}

	var sess Session
	err = s.pool.QueryRow(ctx, `
		UPDATE development_sessions
		SET status=$1, ended_at=$2, duration_seconds=$3, updated_at=NOW()
		WHERE id=$4 AND user_id=$5
		RETURNING id, user_id, project_id, COALESCE(title,''), status, started_at, ended_at, duration_seconds, source, created_at, updated_at
	`, status, endedAt, duration, sessionID, userID).Scan(&sess.ID, &sess.UserID, &sess.ProjectID, &sess.Title, &sess.Status, &sess.StartedAt, &sess.EndedAt, &sess.DurationSeconds, &sess.Source, &sess.CreatedAt, &sess.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Session{}, ErrNotFound
	}
	if err != nil {
		return Session{}, fmt.Errorf("update session: %w", err)
	}
	return sess, nil
}

func validTransition(from, to string) bool {
	switch from {
	case "active":
		return to == "paused" || to == "ended"
	case "paused":
		return to == "active" || to == "ended"
	case "ended":
		return false
	default:
		return false
	}
}

func nullIfEmpty(v string) any {
	if strings.TrimSpace(v) == "" {
		return nil
	}
	return v
}
