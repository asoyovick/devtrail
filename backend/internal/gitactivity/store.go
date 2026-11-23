package gitactivity

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("not found")
var ErrProjectNotFound = errors.New("project not found")

type Store interface {
	Create(ctx context.Context, userID string, req CreateRequest) (Commit, error)
	ListByUser(ctx context.Context, userID, projectID string) ([]Commit, error)
}

type PGStore struct{ pool *pgxpool.Pool }

func NewPGStore(pool *pgxpool.Pool) *PGStore { return &PGStore{pool: pool} }

func (s *PGStore) Create(ctx context.Context, userID string, req CreateRequest) (Commit, error) {
	var c Commit
	err := s.pool.QueryRow(ctx, `
		INSERT INTO git_commits (user_id,project_id,session_id,commit_hash,branch,message_summary,author_name,author_email,committed_at,files_changed,insertions,deletions,event_id)
		SELECT $1,p.id,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12 FROM projects p WHERE p.id=$13 AND p.user_id=$1
		ON CONFLICT (project_id, commit_hash) DO UPDATE SET branch=EXCLUDED.branch, message_summary=EXCLUDED.message_summary, author_name=EXCLUDED.author_name, author_email=EXCLUDED.author_email, committed_at=EXCLUDED.committed_at, files_changed=EXCLUDED.files_changed, insertions=EXCLUDED.insertions, deletions=EXCLUDED.deletions, event_id=COALESCE(EXCLUDED.event_id, git_commits.event_id)
		RETURNING id,user_id,project_id,session_id,commit_hash,COALESCE(branch,''),COALESCE(message_summary,''),COALESCE(author_name,''),COALESCE(author_email,''),committed_at,files_changed,insertions,deletions,event_id,created_at
	`, userID, req.SessionID, req.CommitHash, nullIfEmpty(req.Branch), nullIfEmpty(req.MessageSummary), nullIfEmpty(req.AuthorName), nullIfEmpty(req.AuthorEmail), req.CommittedAt, req.FilesChanged, req.Insertions, req.Deletions, req.EventID, req.ProjectID).Scan(&c.ID, &c.UserID, &c.ProjectID, &c.SessionID, &c.CommitHash, &c.Branch, &c.MessageSummary, &c.AuthorName, &c.AuthorEmail, &c.CommittedAt, &c.FilesChanged, &c.Insertions, &c.Deletions, &c.EventID, &c.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Commit{}, ErrProjectNotFound
	}
	if err != nil {
		return Commit{}, fmt.Errorf("create git commit: %w", err)
	}
	return c, nil
}
func (s *PGStore) ListByUser(ctx context.Context, userID, projectID string) ([]Commit, error) {
	q := `SELECT id,user_id,project_id,session_id,commit_hash,COALESCE(branch,''),COALESCE(message_summary,''),COALESCE(author_name,''),COALESCE(author_email,''),committed_at,files_changed,insertions,deletions,event_id,created_at FROM git_commits WHERE user_id=$1`
	args := []any{userID}
	if strings.TrimSpace(projectID) != "" {
		q += " AND project_id=$2"
		args = append(args, projectID)
	}
	q += " ORDER BY committed_at DESC"
	rows, err := s.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Commit{}
	for rows.Next() {
		var c Commit
		if err := rows.Scan(&c.ID, &c.UserID, &c.ProjectID, &c.SessionID, &c.CommitHash, &c.Branch, &c.MessageSummary, &c.AuthorName, &c.AuthorEmail, &c.CommittedAt, &c.FilesChanged, &c.Insertions, &c.Deletions, &c.EventID, &c.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}
func nullIfEmpty(v string) any {
	if strings.TrimSpace(v) == "" {
		return nil
	}
	return v
}
