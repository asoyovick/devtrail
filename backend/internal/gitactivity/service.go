package gitactivity

import (
	"context"
	"fmt"
	"strings"
)

type Service struct{ store Store }

func NewService(store Store) *Service { return &Service{store: store} }
func (s *Service) Create(ctx context.Context, userID string, req CreateRequest) (Commit, error) {
	req.ProjectID = strings.TrimSpace(req.ProjectID)
	req.CommitHash = strings.TrimSpace(req.CommitHash)
	if req.ProjectID == "" || req.CommitHash == "" || req.CommittedAt.IsZero() {
		return Commit{}, fmt.Errorf("project_id, commit_hash, and committed_at are required")
	}
	return s.store.Create(ctx, userID, req)
}
func (s *Service) List(ctx context.Context, userID, projectID string) ([]Commit, error) {
	return s.store.ListByUser(ctx, userID, projectID)
}
