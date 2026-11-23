package note

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("not found")

type Store interface {
	Create(ctx context.Context, userID string, req CreateRequest) (Note, error)
	ListByUser(ctx context.Context, userID, projectID string) ([]Note, error)
	GetByID(ctx context.Context, userID, id string) (Note, error)
	Update(ctx context.Context, userID, id string, req UpdateRequest) (Note, error)
	Delete(ctx context.Context, userID, id string) error
}

type PGStore struct{ pool *pgxpool.Pool }

func NewPGStore(pool *pgxpool.Pool) *PGStore { return &PGStore{pool: pool} }

func (s *PGStore) Create(ctx context.Context, userID string, req CreateRequest) (Note, error) {
	tags := req.Tags
	if tags == nil {
		tags = []string{}
	}
	var n Note
	err := s.pool.QueryRow(ctx, `
		INSERT INTO notes (user_id, project_id, session_id, title, body, tags)
		VALUES ($1,$2,$3,$4,$5,$6)
		RETURNING id,user_id,project_id,session_id,title,body,tags,created_at,updated_at
	`, userID, req.ProjectID, req.SessionID, req.Title, req.Body, tags).Scan(&n.ID, &n.UserID, &n.ProjectID, &n.SessionID, &n.Title, &n.Body, &n.Tags, &n.CreatedAt, &n.UpdatedAt)
	if err != nil {
		return Note{}, fmt.Errorf("create note: %w", err)
	}
	if n.Tags == nil {
		n.Tags = []string{}
	}
	return n, nil
}
func (s *PGStore) ListByUser(ctx context.Context, userID, projectID string) ([]Note, error) {
	q := `SELECT id,user_id,project_id,session_id,title,body,tags,created_at,updated_at FROM notes WHERE user_id=$1`
	args := []any{userID}
	if strings.TrimSpace(projectID) != "" {
		q += " AND project_id=$2"
		args = append(args, projectID)
	}
	q += " ORDER BY created_at DESC"
	rows, err := s.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("list notes: %w", err)
	}
	defer rows.Close()
	out := []Note{}
	for rows.Next() {
		var n Note
		if err := rows.Scan(&n.ID, &n.UserID, &n.ProjectID, &n.SessionID, &n.Title, &n.Body, &n.Tags, &n.CreatedAt, &n.UpdatedAt); err != nil {
			return nil, err
		}
		if n.Tags == nil {
			n.Tags = []string{}
		}
		out = append(out, n)
	}
	return out, rows.Err()
}
func (s *PGStore) GetByID(ctx context.Context, userID, id string) (Note, error) {
	var n Note
	err := s.pool.QueryRow(ctx, `SELECT id,user_id,project_id,session_id,title,body,tags,created_at,updated_at FROM notes WHERE id=$1 AND user_id=$2`, id, userID).Scan(&n.ID, &n.UserID, &n.ProjectID, &n.SessionID, &n.Title, &n.Body, &n.Tags, &n.CreatedAt, &n.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Note{}, ErrNotFound
	}
	if err != nil {
		return Note{}, fmt.Errorf("get note: %w", err)
	}
	if n.Tags == nil {
		n.Tags = []string{}
	}
	return n, nil
}
func (s *PGStore) Update(ctx context.Context, userID, id string, req UpdateRequest) (Note, error) {
	cur, err := s.GetByID(ctx, userID, id)
	if err != nil {
		return Note{}, err
	}
	if req.Title != nil {
		cur.Title = strings.TrimSpace(*req.Title)
	}
	if req.Body != nil {
		cur.Body = strings.TrimSpace(*req.Body)
	}
	if req.Tags != nil {
		cur.Tags = *req.Tags
	}
	if cur.Title == "" || cur.Body == "" {
		return Note{}, fmt.Errorf("title and body are required")
	}
	var n Note
	err = s.pool.QueryRow(ctx, `UPDATE notes SET title=$1,body=$2,tags=$3,updated_at=NOW() WHERE id=$4 AND user_id=$5 RETURNING id,user_id,project_id,session_id,title,body,tags,created_at,updated_at`, cur.Title, cur.Body, cur.Tags, id, userID).Scan(&n.ID, &n.UserID, &n.ProjectID, &n.SessionID, &n.Title, &n.Body, &n.Tags, &n.CreatedAt, &n.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Note{}, ErrNotFound
	}
	if err != nil {
		return Note{}, fmt.Errorf("update note: %w", err)
	}
	if n.Tags == nil {
		n.Tags = []string{}
	}
	return n, nil
}
func (s *PGStore) Delete(ctx context.Context, userID, id string) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM notes WHERE id=$1 AND user_id=$2`, id, userID)
	if err != nil {
		return fmt.Errorf("delete note: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
