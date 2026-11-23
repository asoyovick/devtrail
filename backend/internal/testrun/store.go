package testrun

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
	Create(ctx context.Context, userID string, req CreateRequest) (TestRun, error)
	ListByUser(ctx context.Context, userID, projectID string) ([]TestRun, error)
	GetByID(ctx context.Context, userID, id string) (TestRun, error)
}

type PGStore struct{ pool *pgxpool.Pool }

func NewPGStore(pool *pgxpool.Pool) *PGStore { return &PGStore{pool: pool} }

func (s *PGStore) Create(ctx context.Context, userID string, req CreateRequest) (TestRun, error) {
	var t TestRun
	err := s.pool.QueryRow(ctx, `
		INSERT INTO test_runs (user_id,project_id,session_id,command,status,started_at,completed_at,duration_ms,total_tests,passed_tests,failed_tests,failure_summary,event_id)
		SELECT $1,p.id,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12 FROM projects p WHERE p.id=$13 AND p.user_id=$1
		RETURNING id,user_id,project_id,session_id,command,status,started_at,completed_at,duration_ms,total_tests,passed_tests,failed_tests,COALESCE(failure_summary,''),event_id,created_at
	`, userID, req.SessionID, req.Command, req.Status, req.StartedAt, req.CompletedAt, req.DurationMs, req.TotalTests, req.PassedTests, req.FailedTests, nullIfEmpty(req.FailureSummary), req.EventID, req.ProjectID).Scan(&t.ID, &t.UserID, &t.ProjectID, &t.SessionID, &t.Command, &t.Status, &t.StartedAt, &t.CompletedAt, &t.DurationMs, &t.TotalTests, &t.PassedTests, &t.FailedTests, &t.FailureSummary, &t.EventID, &t.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return TestRun{}, ErrProjectNotFound
	}
	if err != nil {
		return TestRun{}, fmt.Errorf("create test run: %w", err)
	}
	return t, nil
}
func (s *PGStore) ListByUser(ctx context.Context, userID, projectID string) ([]TestRun, error) {
	q := `SELECT id,user_id,project_id,session_id,command,status,started_at,completed_at,duration_ms,total_tests,passed_tests,failed_tests,COALESCE(failure_summary,''),event_id,created_at FROM test_runs WHERE user_id=$1`
	args := []any{userID}
	if strings.TrimSpace(projectID) != "" {
		q += " AND project_id=$2"
		args = append(args, projectID)
	}
	q += " ORDER BY started_at DESC"
	rows, err := s.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []TestRun{}
	for rows.Next() {
		var t TestRun
		if err := rows.Scan(&t.ID, &t.UserID, &t.ProjectID, &t.SessionID, &t.Command, &t.Status, &t.StartedAt, &t.CompletedAt, &t.DurationMs, &t.TotalTests, &t.PassedTests, &t.FailedTests, &t.FailureSummary, &t.EventID, &t.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}
func (s *PGStore) GetByID(ctx context.Context, userID, id string) (TestRun, error) {
	var t TestRun
	err := s.pool.QueryRow(ctx, `SELECT id,user_id,project_id,session_id,command,status,started_at,completed_at,duration_ms,total_tests,passed_tests,failed_tests,COALESCE(failure_summary,''),event_id,created_at FROM test_runs WHERE id=$1 AND user_id=$2`, id, userID).Scan(&t.ID, &t.UserID, &t.ProjectID, &t.SessionID, &t.Command, &t.Status, &t.StartedAt, &t.CompletedAt, &t.DurationMs, &t.TotalTests, &t.PassedTests, &t.FailedTests, &t.FailureSummary, &t.EventID, &t.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return TestRun{}, ErrNotFound
	}
	if err != nil {
		return TestRun{}, fmt.Errorf("get test run: %w", err)
	}
	return t, nil
}
func nullIfEmpty(v string) any {
	if strings.TrimSpace(v) == "" {
		return nil
	}
	return v
}
