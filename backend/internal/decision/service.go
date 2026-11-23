package decision

import (
	"context"
	"fmt"
	"strings"
)

type Service struct{ store Store }

func NewService(store Store) *Service { return &Service{store: store} }

func (s *Service) Create(ctx context.Context, userID string, req CreateRequest) (Decision, error) {
	req.ProjectID = strings.TrimSpace(req.ProjectID)
	req.Title = strings.TrimSpace(req.Title)
	req.Context = strings.TrimSpace(req.Context)
	req.Problem = strings.TrimSpace(req.Problem)
	req.SelectedApproach = strings.TrimSpace(req.SelectedApproach)
	req.Reasoning = strings.TrimSpace(req.Reasoning)
	if req.ProjectID == "" || req.Title == "" || req.Context == "" || req.Problem == "" || req.SelectedApproach == "" || req.Reasoning == "" {
		return Decision{}, fmt.Errorf("missing required fields")
	}
	if req.Status != "" {
		st := strings.TrimSpace(req.Status)
		if st != "proposed" && st != "accepted" && st != "superseded" && st != "deprecated" {
			return Decision{}, fmt.Errorf("invalid status")
		}
		req.Status = st
	}
	return s.store.Create(ctx, userID, req)
}
func (s *Service) List(ctx context.Context, userID, projectID string) ([]Decision, error) {
	return s.store.ListByUser(ctx, userID, projectID)
}
func (s *Service) Get(ctx context.Context, userID, id string) (Decision, error) {
	return s.store.GetByID(ctx, userID, id)
}
func (s *Service) Update(ctx context.Context, userID, id string, req UpdateRequest) (Decision, error) {
	return s.store.Update(ctx, userID, id, req)
}
