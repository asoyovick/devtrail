package project

import (
	"context"
	"fmt"
	"strings"
)

type Service struct {
	store Store
}

func NewService(store Store) *Service { return &Service{store: store} }

func (s *Service) Create(ctx context.Context, userID string, req CreateRequest) (Project, error) {
	req.Name = strings.TrimSpace(req.Name)
	req.Description = strings.TrimSpace(req.Description)
	req.RepoRootHash = strings.TrimSpace(req.RepoRootHash)
	req.DefaultBranch = strings.TrimSpace(req.DefaultBranch)
	if req.Name == "" {
		return Project{}, fmt.Errorf("name is required")
	}
	return s.store.Create(ctx, userID, req)
}

func (s *Service) List(ctx context.Context, userID string) ([]Project, error) {
	return s.store.ListByUser(ctx, userID)
}

func (s *Service) Get(ctx context.Context, userID, projectID string) (Project, error) {
	return s.store.GetByID(ctx, userID, projectID)
}

func (s *Service) Update(ctx context.Context, userID, projectID string, req UpdateRequest) (Project, error) {
	if req.Name != nil && strings.TrimSpace(*req.Name) == "" {
		return Project{}, fmt.Errorf("name is required")
	}
	if req.Status != nil {
		status := strings.TrimSpace(*req.Status)
		if status != "active" && status != "archived" {
			return Project{}, fmt.Errorf("invalid status")
		}
		req.Status = &status
	}
	return s.store.Update(ctx, userID, projectID, req)
}
