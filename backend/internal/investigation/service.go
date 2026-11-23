package investigation

import (
	"context"
	"fmt"
	"strings"
)

type Service struct{ store Store }

func NewService(store Store) *Service { return &Service{store: store} }

func (s *Service) Create(ctx context.Context, userID string, req CreateRequest) (Investigation, error) {
	req.Title = strings.TrimSpace(req.Title)
	req.ProblemDescription = strings.TrimSpace(req.ProblemDescription)
	req.ProjectID = strings.TrimSpace(req.ProjectID)
	if req.ProjectID == "" || req.Title == "" || req.ProblemDescription == "" {
		return Investigation{}, fmt.Errorf("project_id, title, and problem_description are required")
	}
	return s.store.Create(ctx, userID, req)
}

func (s *Service) List(ctx context.Context, userID, projectID string) ([]Investigation, error) {
	return s.store.ListByUser(ctx, userID, projectID)
}
func (s *Service) Get(ctx context.Context, userID, id string) (Investigation, error) {
	return s.store.GetByID(ctx, userID, id)
}
func (s *Service) Update(ctx context.Context, userID, id string, req UpdateRequest) (Investigation, error) {
	return s.store.Update(ctx, userID, id, req)
}
