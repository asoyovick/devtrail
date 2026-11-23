package insight

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct{ pool *pgxpool.Pool }

func NewService(pool *pgxpool.Pool) *Service { return &Service{pool: pool} }

func (s *Service) List(ctx context.Context, userID string) ([]Insight, error) {
	recurring, err := s.RecurringProblems(ctx, userID)
	if err != nil {
		return nil, err
	}
	improve, err := s.ImprovementAreas(ctx, userID)
	if err != nil {
		return nil, err
	}
	return append(recurring, improve...), nil
}

func (s *Service) RecurringProblems(ctx context.Context, userID string) ([]Insight, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT project_id, command, COUNT(*) AS failures
		FROM test_runs
		WHERE user_id=$1 AND status IN ('failed','errored')
		GROUP BY project_id, command
		HAVING COUNT(*) >= 2
		ORDER BY failures DESC LIMIT 10
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Insight{}
	for rows.Next() {
		var projectID, command string
		var failures int
		if err := rows.Scan(&projectID, &command, &failures); err != nil {
			return nil, err
		}
		out = append(out, Insight{
			Type: "recurring_test_failure", Scope: "project", Confidence: "medium",
			Summary:        "Repeated test failures detected",
			Observation:    "The same test command has failed multiple times in the recorded history.",
			Evidence:       map[string]any{"project_id": projectID, "command": command, "failure_count": failures},
			Recommendation: "Review the failing command, inspect related investigations, and add prevention steps after resolving the root cause.",
		})
	}
	return out, rows.Err()
}

func (s *Service) ImprovementAreas(ctx context.Context, userID string) ([]Insight, error) {
	var openInvestigations int
	if err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM investigations WHERE user_id=$1 AND status='open'`, userID).Scan(&openInvestigations); err != nil {
		return nil, err
	}
	var projectsWithoutTests int
	if err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM projects p WHERE p.user_id=$1 AND NOT EXISTS (SELECT 1 FROM test_runs t WHERE t.project_id=p.id)`, userID).Scan(&projectsWithoutTests); err != nil {
		return nil, err
	}
	out := []Insight{}
	if openInvestigations > 0 {
		out = append(out, Insight{Type: "open_investigations", Scope: "global", Confidence: "high", Summary: "Open investigations need attention", Observation: "There are unresolved debugging investigations.", Evidence: map[string]any{"open_investigations": openInvestigations}, Recommendation: "Review open investigations and document root causes/resolutions to preserve engineering knowledge."})
	}
	if projectsWithoutTests > 0 {
		out = append(out, Insight{Type: "missing_test_history", Scope: "global", Confidence: "medium", Summary: "Projects without recorded test history", Observation: "Some registered projects have no recorded test runs.", Evidence: map[string]any{"projects_without_test_runs": projectsWithoutTests}, Recommendation: "Record test runs through the CLI or API to improve reliability visibility."})
	}
	return out, nil
}
