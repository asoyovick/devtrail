package session

import (
	"context"
	"fmt"
	"strings"
)

type Service struct{ store Store }

func NewService(store Store) *Service { return &Service{store: store} }

func (s *Service) Start(ctx context.Context, userID string, req StartRequest) (Session, error) {
	req.ProjectID = strings.TrimSpace(req.ProjectID)
	req.Title = strings.TrimSpace(req.Title)
	req.Source = strings.TrimSpace(req.Source)
	if req.ProjectID == "" {
		return Session{}, fmt.Errorf("project_id is required")
	}
	if req.Source == "" {
		req.Source = "web"
	}
	if req.Source != "web" && req.Source != "cli" {
		return Session{}, fmt.Errorf("invalid source")
	}
	return s.store.Create(ctx, userID, req)
}

func (s *Service) Get(ctx context.Context, userID, sessionID string) (Session, error) {
	return s.store.GetByID(ctx, userID, sessionID)
}

func (s *Service) List(ctx context.Context, userID, projectID string) ([]Session, error) {
	return s.store.ListByUser(ctx, userID, projectID)
}

func (s *Service) Pause(ctx context.Context, userID, sessionID string) (Session, error) {
	return s.store.SetStatus(ctx, userID, sessionID, "paused")
}
func (s *Service) Resume(ctx context.Context, userID, sessionID string) (Session, error) {
	return s.store.SetStatus(ctx, userID, sessionID, "active")
}
func (s *Service) End(ctx context.Context, userID, sessionID string) (Session, error) {
	return s.store.SetStatus(ctx, userID, sessionID, "ended")
}
