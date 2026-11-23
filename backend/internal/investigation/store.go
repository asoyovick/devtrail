package investigation

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
var ErrProjectNotFound = errors.New("project not found")

type Store interface {
	Create(ctx context.Context, userID string, req CreateRequest) (Investigation, error)
	ListByUser(ctx context.Context, userID, projectID string) ([]Investigation, error)
	GetByID(ctx context.Context, userID, id string) (Investigation, error)
	Update(ctx context.Context, userID, id string, req UpdateRequest) (Investigation, error)
}

type PGStore struct{ pool *pgxpool.Pool }

func NewPGStore(pool *pgxpool.Pool) *PGStore { return &PGStore{pool: pool} }

func (s *PGStore) Create(ctx context.Context, userID string, req CreateRequest) (Investigation, error) {
	var inv Investigation
	err := s.pool.QueryRow(ctx, `
		INSERT INTO investigations (user_id, project_id, session_id, title, problem_description, observed_behavior, investigation_steps, attempted_approaches, root_cause, resolution, lessons_learned, prevention_strategies, severity)
		SELECT $1,p.id,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12
		FROM projects p WHERE p.id=$13 AND p.user_id=$1
		RETURNING id,user_id,project_id,session_id,title,problem_description,COALESCE(observed_behavior,''),COALESCE(investigation_steps,''),COALESCE(attempted_approaches,''),COALESCE(root_cause,''),COALESCE(resolution,''),COALESCE(lessons_learned,''),COALESCE(prevention_strategies,''),status,COALESCE(severity,''),created_at,updated_at,resolved_at
	`, userID, req.SessionID, req.Title, req.ProblemDescription, nullIfEmpty(req.ObservedBehavior), nullIfEmpty(req.InvestigationSteps), nullIfEmpty(req.AttemptedApproaches), nullIfEmpty(req.RootCause), nullIfEmpty(req.Resolution), nullIfEmpty(req.LessonsLearned), nullIfEmpty(req.PreventionStrategies), nullIfEmpty(req.Severity), req.ProjectID).Scan(
		&inv.ID, &inv.UserID, &inv.ProjectID, &inv.SessionID, &inv.Title, &inv.ProblemDescription, &inv.ObservedBehavior, &inv.InvestigationSteps, &inv.AttemptedApproaches, &inv.RootCause, &inv.Resolution, &inv.LessonsLearned, &inv.PreventionStrategies, &inv.Status, &inv.Severity, &inv.CreatedAt, &inv.UpdatedAt, &inv.ResolvedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Investigation{}, ErrProjectNotFound
	}
	if err != nil {
		return Investigation{}, fmt.Errorf("create investigation: %w", err)
	}
	return inv, nil
}

func (s *PGStore) ListByUser(ctx context.Context, userID, projectID string) ([]Investigation, error) {
	q := `SELECT id,user_id,project_id,session_id,title,problem_description,COALESCE(observed_behavior,''),COALESCE(investigation_steps,''),COALESCE(attempted_approaches,''),COALESCE(root_cause,''),COALESCE(resolution,''),COALESCE(lessons_learned,''),COALESCE(prevention_strategies,''),status,COALESCE(severity,''),created_at,updated_at,resolved_at FROM investigations WHERE user_id=$1`
	args := []any{userID}
	if strings.TrimSpace(projectID) != "" {
		q += " AND project_id=$2"
		args = append(args, projectID)
	}
	q += " ORDER BY created_at DESC"
	rows, err := s.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("list investigations: %w", err)
	}
	defer rows.Close()
	out := []Investigation{}
	for rows.Next() {
		var inv Investigation
		if err := rows.Scan(&inv.ID, &inv.UserID, &inv.ProjectID, &inv.SessionID, &inv.Title, &inv.ProblemDescription, &inv.ObservedBehavior, &inv.InvestigationSteps, &inv.AttemptedApproaches, &inv.RootCause, &inv.Resolution, &inv.LessonsLearned, &inv.PreventionStrategies, &inv.Status, &inv.Severity, &inv.CreatedAt, &inv.UpdatedAt, &inv.ResolvedAt); err != nil {
			return nil, fmt.Errorf("scan investigation: %w", err)
		}
		out = append(out, inv)
	}
	return out, rows.Err()
}

func (s *PGStore) GetByID(ctx context.Context, userID, id string) (Investigation, error) {
	var inv Investigation
	err := s.pool.QueryRow(ctx, `SELECT id,user_id,project_id,session_id,title,problem_description,COALESCE(observed_behavior,''),COALESCE(investigation_steps,''),COALESCE(attempted_approaches,''),COALESCE(root_cause,''),COALESCE(resolution,''),COALESCE(lessons_learned,''),COALESCE(prevention_strategies,''),status,COALESCE(severity,''),created_at,updated_at,resolved_at FROM investigations WHERE id=$1 AND user_id=$2`, id, userID).Scan(&inv.ID, &inv.UserID, &inv.ProjectID, &inv.SessionID, &inv.Title, &inv.ProblemDescription, &inv.ObservedBehavior, &inv.InvestigationSteps, &inv.AttemptedApproaches, &inv.RootCause, &inv.Resolution, &inv.LessonsLearned, &inv.PreventionStrategies, &inv.Status, &inv.Severity, &inv.CreatedAt, &inv.UpdatedAt, &inv.ResolvedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Investigation{}, ErrNotFound
	}
	if err != nil {
		return Investigation{}, fmt.Errorf("get investigation: %w", err)
	}
	return inv, nil
}

func (s *PGStore) Update(ctx context.Context, userID, id string, req UpdateRequest) (Investigation, error) {
	cur, err := s.GetByID(ctx, userID, id)
	if err != nil {
		return Investigation{}, err
	}
	apply := func(dst *string, src *string) {
		if src != nil {
			*dst = strings.TrimSpace(*src)
		}
	}
	apply(&cur.Title, req.Title)
	apply(&cur.ProblemDescription, req.ProblemDescription)
	apply(&cur.ObservedBehavior, req.ObservedBehavior)
	apply(&cur.InvestigationSteps, req.InvestigationSteps)
	apply(&cur.AttemptedApproaches, req.AttemptedApproaches)
	apply(&cur.RootCause, req.RootCause)
	apply(&cur.Resolution, req.Resolution)
	apply(&cur.LessonsLearned, req.LessonsLearned)
	apply(&cur.PreventionStrategies, req.PreventionStrategies)
	apply(&cur.Severity, req.Severity)
	if req.Status != nil {
		s := strings.TrimSpace(*req.Status)
		if s != "open" && s != "resolved" && s != "abandoned" {
			return Investigation{}, fmt.Errorf("invalid status")
		}
		cur.Status = s
		if s == "resolved" && cur.ResolvedAt == nil {
			now := time.Now()
			cur.ResolvedAt = &now
		}
		if s != "resolved" {
			cur.ResolvedAt = nil
		}
	}
	if cur.Title == "" || cur.ProblemDescription == "" {
		return Investigation{}, fmt.Errorf("title and problem_description are required")
	}

	var inv Investigation
	err = s.pool.QueryRow(ctx, `
		UPDATE investigations SET title=$1,problem_description=$2,observed_behavior=$3,investigation_steps=$4,attempted_approaches=$5,root_cause=$6,resolution=$7,lessons_learned=$8,prevention_strategies=$9,status=$10,severity=$11,resolved_at=$12,updated_at=NOW()
		WHERE id=$13 AND user_id=$14
		RETURNING id,user_id,project_id,session_id,title,problem_description,COALESCE(observed_behavior,''),COALESCE(investigation_steps,''),COALESCE(attempted_approaches,''),COALESCE(root_cause,''),COALESCE(resolution,''),COALESCE(lessons_learned,''),COALESCE(prevention_strategies,''),status,COALESCE(severity,''),created_at,updated_at,resolved_at
	`, cur.Title, cur.ProblemDescription, nullIfEmpty(cur.ObservedBehavior), nullIfEmpty(cur.InvestigationSteps), nullIfEmpty(cur.AttemptedApproaches), nullIfEmpty(cur.RootCause), nullIfEmpty(cur.Resolution), nullIfEmpty(cur.LessonsLearned), nullIfEmpty(cur.PreventionStrategies), cur.Status, nullIfEmpty(cur.Severity), cur.ResolvedAt, id, userID).Scan(&inv.ID, &inv.UserID, &inv.ProjectID, &inv.SessionID, &inv.Title, &inv.ProblemDescription, &inv.ObservedBehavior, &inv.InvestigationSteps, &inv.AttemptedApproaches, &inv.RootCause, &inv.Resolution, &inv.LessonsLearned, &inv.PreventionStrategies, &inv.Status, &inv.Severity, &inv.CreatedAt, &inv.UpdatedAt, &inv.ResolvedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Investigation{}, ErrNotFound
	}
	if err != nil {
		return Investigation{}, fmt.Errorf("update investigation: %w", err)
	}
	return inv, nil
}

func nullIfEmpty(v string) any {
	if strings.TrimSpace(v) == "" {
		return nil
	}
	return v
}
