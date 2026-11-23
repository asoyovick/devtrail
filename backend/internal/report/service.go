package report

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct{ pool *pgxpool.Pool }

func NewService(pool *pgxpool.Pool) *Service { return &Service{pool: pool} }

func (s *Service) Generate(ctx context.Context, userID, projectID, reportType string, start, end time.Time) (Summary, error) {
	sum := Summary{ReportType: reportType, PeriodStart: start, PeriodEnd: end, ProjectID: projectID}
	count := func(q string, args ...any) (int, error) {
		var n int
		if err := s.pool.QueryRow(ctx, q, args...).Scan(&n); err != nil {
			return 0, err
		}
		return n, nil
	}
	var err error
	if projectID != "" {
		if sum.Sessions, err = count(`SELECT COUNT(*) FROM development_sessions WHERE user_id=$1 AND project_id=$2 AND started_at >= $3 AND started_at < $4`, userID, projectID, start, end); err != nil {
			return Summary{}, err
		}
		if sum.Events, err = count(`SELECT COUNT(*) FROM activity_events WHERE user_id=$1 AND project_id=$2 AND occurred_at >= $3 AND occurred_at < $4`, userID, projectID, start, end); err != nil {
			return Summary{}, err
		}
		if sum.Commits, err = count(`SELECT COUNT(*) FROM git_commits WHERE user_id=$1 AND project_id=$2 AND committed_at >= $3 AND committed_at < $4`, userID, projectID, start, end); err != nil {
			return Summary{}, err
		}
		if sum.TestRuns, err = count(`SELECT COUNT(*) FROM test_runs WHERE user_id=$1 AND project_id=$2 AND started_at >= $3 AND started_at < $4`, userID, projectID, start, end); err != nil {
			return Summary{}, err
		}
		if sum.FailedTestRuns, err = count(`SELECT COUNT(*) FROM test_runs WHERE user_id=$1 AND project_id=$2 AND started_at >= $3 AND started_at < $4 AND status IN ('failed','errored')`, userID, projectID, start, end); err != nil {
			return Summary{}, err
		}
		if sum.Investigations, err = count(`SELECT COUNT(*) FROM investigations WHERE user_id=$1 AND project_id=$2 AND created_at >= $3 AND created_at < $4`, userID, projectID, start, end); err != nil {
			return Summary{}, err
		}
		if sum.OpenInvestigations, err = count(`SELECT COUNT(*) FROM investigations WHERE user_id=$1 AND project_id=$2 AND status='open'`, userID, projectID); err != nil {
			return Summary{}, err
		}
		if sum.Decisions, err = count(`SELECT COUNT(*) FROM decisions WHERE user_id=$1 AND project_id=$2 AND created_at >= $3 AND created_at < $4`, userID, projectID, start, end); err != nil {
			return Summary{}, err
		}
		if sum.Notes, err = count(`SELECT COUNT(*) FROM notes WHERE user_id=$1 AND project_id=$2 AND created_at >= $3 AND created_at < $4`, userID, projectID, start, end); err != nil {
			return Summary{}, err
		}
	} else {
		if sum.Sessions, err = count(`SELECT COUNT(*) FROM development_sessions WHERE user_id=$1 AND started_at >= $2 AND started_at < $3`, userID, start, end); err != nil {
			return Summary{}, err
		}
		if sum.Events, err = count(`SELECT COUNT(*) FROM activity_events WHERE user_id=$1 AND occurred_at >= $2 AND occurred_at < $3`, userID, start, end); err != nil {
			return Summary{}, err
		}
		if sum.Commits, err = count(`SELECT COUNT(*) FROM git_commits WHERE user_id=$1 AND committed_at >= $2 AND committed_at < $3`, userID, start, end); err != nil {
			return Summary{}, err
		}
		if sum.TestRuns, err = count(`SELECT COUNT(*) FROM test_runs WHERE user_id=$1 AND started_at >= $2 AND started_at < $3`, userID, start, end); err != nil {
			return Summary{}, err
		}
		if sum.FailedTestRuns, err = count(`SELECT COUNT(*) FROM test_runs WHERE user_id=$1 AND started_at >= $2 AND started_at < $3 AND status IN ('failed','errored')`, userID, start, end); err != nil {
			return Summary{}, err
		}
		if sum.Investigations, err = count(`SELECT COUNT(*) FROM investigations WHERE user_id=$1 AND created_at >= $2 AND created_at < $3`, userID, start, end); err != nil {
			return Summary{}, err
		}
		if sum.OpenInvestigations, err = count(`SELECT COUNT(*) FROM investigations WHERE user_id=$1 AND status='open'`, userID); err != nil {
			return Summary{}, err
		}
		if sum.Decisions, err = count(`SELECT COUNT(*) FROM decisions WHERE user_id=$1 AND created_at >= $2 AND created_at < $3`, userID, start, end); err != nil {
			return Summary{}, err
		}
		if sum.Notes, err = count(`SELECT COUNT(*) FROM notes WHERE user_id=$1 AND created_at >= $2 AND created_at < $3`, userID, start, end); err != nil {
			return Summary{}, err
		}
	}
	return sum, nil
}

func (s Summary) Markdown() string {
	scope := "Global"
	if s.ProjectID != "" {
		scope = "Project " + s.ProjectID
	}
	return fmt.Sprintf("# %s Report\n\n- Scope: %s\n- Period: %s to %s\n- Sessions: %d\n- Events: %d\n- Commits: %d\n- Test runs: %d (failed/errored: %d)\n- Investigations: %d (open: %d)\n- Decisions: %d\n- Notes: %d\n",
		title(s.ReportType), scope, s.PeriodStart.Format(time.RFC3339), s.PeriodEnd.Format(time.RFC3339), s.Sessions, s.Events, s.Commits, s.TestRuns, s.FailedTestRuns, s.Investigations, s.OpenInvestigations, s.Decisions, s.Notes)
}

func title(v string) string {
	if v == "daily" {
		return "Daily"
	}
	if v == "weekly" {
		return "Weekly"
	}
	if v == "project" {
		return "Project"
	}
	return "Engineering"
}
