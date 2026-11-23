package project

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("not found")
var ErrNameExists = errors.New("project name already exists")

type Store interface {
	Create(ctx context.Context, userID string, req CreateRequest) (Project, error)
	ListByUser(ctx context.Context, userID string) ([]Project, error)
	GetByID(ctx context.Context, userID, projectID string) (Project, error)
	Update(ctx context.Context, userID, projectID string, req UpdateRequest) (Project, error)
}

type PGStore struct {
	pool *pgxpool.Pool
}

func NewPGStore(pool *pgxpool.Pool) *PGStore { return &PGStore{pool: pool} }

func (s *PGStore) Create(ctx context.Context, userID string, req CreateRequest) (Project, error) {
	var p Project
	err := s.pool.QueryRow(ctx, `
		INSERT INTO projects (user_id, name, description, repo_root_hash, default_branch)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, user_id, name, COALESCE(description,''), COALESCE(repo_root_hash,''), COALESCE(default_branch,''), status, created_at, updated_at
	`, userID, req.Name, nullIfEmpty(req.Description), nullIfEmpty(req.RepoRootHash), nullIfEmpty(req.DefaultBranch)).Scan(
		&p.ID, &p.UserID, &p.Name, &p.Description, &p.RepoRootHash, &p.DefaultBranch, &p.Status, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") || strings.Contains(strings.ToLower(err.Error()), "unique") {
			return Project{}, ErrNameExists
		}
		return Project{}, fmt.Errorf("create project: %w", err)
	}
	return p, nil
}

func (s *PGStore) ListByUser(ctx context.Context, userID string) ([]Project, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, user_id, name, COALESCE(description,''), COALESCE(repo_root_hash,''), COALESCE(default_branch,''), status, created_at, updated_at
		FROM projects WHERE user_id = $1 ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}
	defer rows.Close()

	var projects []Project
	for rows.Next() {
		var p Project
		if err := rows.Scan(&p.ID, &p.UserID, &p.Name, &p.Description, &p.RepoRootHash, &p.DefaultBranch, &p.Status, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan project: %w", err)
		}
		projects = append(projects, p)
	}
	if projects == nil {
		projects = []Project{}
	}
	return projects, rows.Err()
}

func (s *PGStore) GetByID(ctx context.Context, userID, projectID string) (Project, error) {
	var p Project
	err := s.pool.QueryRow(ctx, `
		SELECT id, user_id, name, COALESCE(description,''), COALESCE(repo_root_hash,''), COALESCE(default_branch,''), status, created_at, updated_at
		FROM projects WHERE id = $1 AND user_id = $2
	`, projectID, userID).Scan(&p.ID, &p.UserID, &p.Name, &p.Description, &p.RepoRootHash, &p.DefaultBranch, &p.Status, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Project{}, ErrNotFound
	}
	if err != nil {
		return Project{}, fmt.Errorf("get project: %w", err)
	}
	return p, nil
}

func (s *PGStore) Update(ctx context.Context, userID, projectID string, req UpdateRequest) (Project, error) {
	current, err := s.GetByID(ctx, userID, projectID)
	if err != nil {
		return Project{}, err
	}
	if req.Name != nil {
		current.Name = strings.TrimSpace(*req.Name)
	}
	if req.Description != nil {
		current.Description = strings.TrimSpace(*req.Description)
	}
	if req.DefaultBranch != nil {
		current.DefaultBranch = strings.TrimSpace(*req.DefaultBranch)
	}
	if req.Status != nil {
		current.Status = strings.TrimSpace(*req.Status)
	}

	var p Project
	err = s.pool.QueryRow(ctx, `
		UPDATE projects
		SET name=$1, description=$2, default_branch=$3, status=$4, updated_at=NOW()
		WHERE id=$5 AND user_id=$6
		RETURNING id, user_id, name, COALESCE(description,''), COALESCE(repo_root_hash,''), COALESCE(default_branch,''), status, created_at, updated_at
	`, current.Name, nullIfEmpty(current.Description), nullIfEmpty(current.DefaultBranch), current.Status, projectID, userID).Scan(
		&p.ID, &p.UserID, &p.Name, &p.Description, &p.RepoRootHash, &p.DefaultBranch, &p.Status, &p.CreatedAt, &p.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Project{}, ErrNotFound
	}
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") || strings.Contains(strings.ToLower(err.Error()), "unique") {
			return Project{}, ErrNameExists
		}
		return Project{}, fmt.Errorf("update project: %w", err)
	}
	return p, nil
}

func nullIfEmpty(v string) any {
	if strings.TrimSpace(v) == "" {
		return nil
	}
	return v
}
