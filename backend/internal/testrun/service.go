package testrun

import (
	"context"
	"fmt"
	"strings"
)

type Service struct{ store Store }

func NewService(store Store) *Service { return &Service{store: store} }

func (s *Service) Create(ctx context.Context, userID string, req CreateRequest) (TestRun, error) {
	req.ProjectID = strings.TrimSpace(req.ProjectID)
	req.Command = strings.TrimSpace(req.Command)
	req.Status = strings.TrimSpace(req.Status)
	if req.ProjectID == "" || req.Command == "" || req.Status == "" || req.StartedAt.IsZero() {
		return TestRun{}, fmt.Errorf("project_id, command, status, and started_at are required")
	}
	switch req.Status {
	case "passed", "failed", "errored", "skipped", "unknown":
	default:
		return TestRun{}, fmt.Errorf("invalid status")
	}
	return s.store.Create(ctx, userID, req)
}
func (s *Service) List(ctx context.Context, userID, projectID string) ([]TestRun, error) {
	return s.store.ListByUser(ctx, userID, projectID)
}
func (s *Service) Get(ctx context.Context, userID, id string) (TestRun, error) {
	return s.store.GetByID(ctx, userID, id)
}
