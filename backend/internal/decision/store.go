package decision

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
	Create(ctx context.Context, userID string, req CreateRequest) (Decision, error)
	ListByUser(ctx context.Context, userID, projectID string) ([]Decision, error)
	GetByID(ctx context.Context, userID, id string) (Decision, error)
	Update(ctx context.Context, userID, id string, req UpdateRequest) (Decision, error)
}

type PGStore struct{ pool *pgxpool.Pool }

func NewPGStore(pool *pgxpool.Pool) *PGStore { return &PGStore{pool: pool} }

func (s *PGStore) Create(ctx context.Context, userID string, req CreateRequest) (Decision, error) {
	status := strings.TrimSpace(req.Status)
	if status == "" {
		status = "accepted"
	}
	var d Decision
	err := s.pool.QueryRow(ctx, `
		INSERT INTO decisions (user_id, project_id, title, context, problem, alternatives, selected_approach, reasoning, consequences, status)
		SELECT $1,p.id,$2,$3,$4,$5,$6,$7,$8,$9 FROM projects p WHERE p.id=$10 AND p.user_id=$1
		RETURNING id,user_id,project_id,title,context,problem,COALESCE(alternatives,''),selected_approach,reasoning,COALESCE(consequences,''),status,decided_at,created_at,updated_at
	`, userID, req.Title, req.Context, req.Problem, nullIfEmpty(req.Alternatives), req.SelectedApproach, req.Reasoning, nullIfEmpty(req.Consequences), status, req.ProjectID).Scan(&d.ID, &d.UserID, &d.ProjectID, &d.Title, &d.Context, &d.Problem, &d.Alternatives, &d.SelectedApproach, &d.Reasoning, &d.Consequences, &d.Status, &d.DecidedAt, &d.CreatedAt, &d.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Decision{}, ErrProjectNotFound
	}
	if err != nil {
		return Decision{}, fmt.Errorf("create decision: %w", err)
	}
	return d, nil
}

func (s *PGStore) ListByUser(ctx context.Context, userID, projectID string) ([]Decision, error) {
	q := `SELECT id,user_id,project_id,title,context,problem,COALESCE(alternatives,''),selected_approach,reasoning,COALESCE(consequences,''),status,decided_at,created_at,updated_at FROM decisions WHERE user_id=$1`
	args := []any{userID}
	if strings.TrimSpace(projectID) != "" {
		q += " AND project_id=$2"
		args = append(args, projectID)
	}
	q += " ORDER BY created_at DESC"
	rows, err := s.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("list decisions: %w", err)
	}
	defer rows.Close()
	out := []Decision{}
	for rows.Next() {
		var d Decision
		if err := rows.Scan(&d.ID, &d.UserID, &d.ProjectID, &d.Title, &d.Context, &d.Problem, &d.Alternatives, &d.SelectedApproach, &d.Reasoning, &d.Consequences, &d.Status, &d.DecidedAt, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func (s *PGStore) GetByID(ctx context.Context, userID, id string) (Decision, error) {
	var d Decision
	err := s.pool.QueryRow(ctx, `SELECT id,user_id,project_id,title,context,problem,COALESCE(alternatives,''),selected_approach,reasoning,COALESCE(consequences,''),status,decided_at,created_at,updated_at FROM decisions WHERE id=$1 AND user_id=$2`, id, userID).Scan(&d.ID, &d.UserID, &d.ProjectID, &d.Title, &d.Context, &d.Problem, &d.Alternatives, &d.SelectedApproach, &d.Reasoning, &d.Consequences, &d.Status, &d.DecidedAt, &d.CreatedAt, &d.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Decision{}, ErrNotFound
	}
	if err != nil {
		return Decision{}, fmt.Errorf("get decision: %w", err)
	}
	return d, nil
}

func (s *PGStore) Update(ctx context.Context, userID, id string, req UpdateRequest) (Decision, error) {
	cur, err := s.GetByID(ctx, userID, id)
	if err != nil {
		return Decision{}, err
	}
	apply := func(dst, src *string) {
		if src != nil {
			*dst = strings.TrimSpace(*src)
		}
	}
	apply(&cur.Title, req.Title)
	apply(&cur.Context, req.Context)
	apply(&cur.Problem, req.Problem)
	apply(&cur.Alternatives, req.Alternatives)
	apply(&cur.SelectedApproach, req.SelectedApproach)
	apply(&cur.Reasoning, req.Reasoning)
	apply(&cur.Consequences, req.Consequences)
	if req.Status != nil {
		st := strings.TrimSpace(*req.Status)
		if st != "proposed" && st != "accepted" && st != "superseded" && st != "deprecated" {
			return Decision{}, fmt.Errorf("invalid status")
		}
		cur.Status = st
	}
	if cur.Title == "" || cur.Context == "" || cur.Problem == "" || cur.SelectedApproach == "" || cur.Reasoning == "" {
		return Decision{}, fmt.Errorf("missing required fields")
	}
	var d Decision
	err = s.pool.QueryRow(ctx, `UPDATE decisions SET title=$1,context=$2,problem=$3,alternatives=$4,selected_approach=$5,reasoning=$6,consequences=$7,status=$8,updated_at=NOW() WHERE id=$9 AND user_id=$10 RETURNING id,user_id,project_id,title,context,problem,COALESCE(alternatives,''),selected_approach,reasoning,COALESCE(consequences,''),status,decided_at,created_at,updated_at`, cur.Title, cur.Context, cur.Problem, nullIfEmpty(cur.Alternatives), cur.SelectedApproach, cur.Reasoning, nullIfEmpty(cur.Consequences), cur.Status, id, userID).Scan(&d.ID, &d.UserID, &d.ProjectID, &d.Title, &d.Context, &d.Problem, &d.Alternatives, &d.SelectedApproach, &d.Reasoning, &d.Consequences, &d.Status, &d.DecidedAt, &d.CreatedAt, &d.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Decision{}, ErrNotFound
	}
	if err != nil {
		return Decision{}, fmt.Errorf("update decision: %w", err)
	}
	return d, nil
}

func nullIfEmpty(v string) any {
	if strings.TrimSpace(v) == "" {
		return nil
	}
	return v
}
