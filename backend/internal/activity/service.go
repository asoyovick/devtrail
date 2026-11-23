package activity

import (
	"context"
	"strings"
)

type Service struct{ store Store }

func NewService(store Store) *Service { return &Service{store: store} }

func (s *Service) IngestBatch(ctx context.Context, userID string, req IngestRequest) (IngestResponse, error) {
	results := make([]IngestResult, 0, len(req.Events))
	for _, e := range req.Events {
		e.EventID = strings.TrimSpace(e.EventID)
		e.EventType = strings.TrimSpace(e.EventType)
		e.Source = strings.TrimSpace(e.Source)
		if e.EventID == "" || e.EventType == "" || e.OccurredAt.IsZero() {
			results = append(results, IngestResult{EventID: e.EventID, Status: "rejected", Reason: "missing required fields"})
			continue
		}
		if e.Source != "cli" && e.Source != "backend" && e.Source != "web" {
			results = append(results, IngestResult{EventID: e.EventID, Status: "rejected", Reason: "invalid source"})
			continue
		}
		status, err := s.store.Ingest(ctx, userID, e)
		if err != nil {
			results = append(results, IngestResult{EventID: e.EventID, Status: "rejected", Reason: "storage error"})
			continue
		}
		results = append(results, IngestResult{EventID: e.EventID, Status: status})
	}
	return IngestResponse{Results: results}, nil
}

func (s *Service) List(ctx context.Context, userID, projectID string, limit int) ([]Event, error) {
	return s.store.ListByUser(ctx, userID, projectID, limit)
}
