package note

import (
	"context"
	"fmt"
	"strings"
)

type Service struct{ store Store }

func NewService(store Store) *Service { return &Service{store: store} }

func (s *Service) Create(ctx context.Context, userID string, req CreateRequest) (Note, error) {
	req.Title = strings.TrimSpace(req.Title)
	req.Body = strings.TrimSpace(req.Body)
	if req.Title == "" || req.Body == "" {
		return Note{}, fmt.Errorf("title and body are required")
	}
	if req.Tags == nil {
		req.Tags = []string{}
	}
	return s.store.Create(ctx, userID, req)
}
func (s *Service) List(ctx context.Context, userID, projectID string) ([]Note, error) {
	return s.store.ListByUser(ctx, userID, projectID)
}
func (s *Service) Get(ctx context.Context, userID, id string) (Note, error) {
	return s.store.GetByID(ctx, userID, id)
}
func (s *Service) Update(ctx context.Context, userID, id string, req UpdateRequest) (Note, error) {
	return s.store.Update(ctx, userID, id, req)
}
func (s *Service) Delete(ctx context.Context, userID, id string) error {
	return s.store.Delete(ctx, userID, id)
}
