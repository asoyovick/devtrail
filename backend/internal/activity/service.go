package activity

import (
	"context"
	"encoding/json"
	"strings"
	"time"
)

type PublishedEvent struct {
	ID         string          `json:"id"`
	Type       string          `json:"type"`
	ProjectID  *string         `json:"project_id,omitempty"`
	SessionID  *string         `json:"session_id,omitempty"`
	Source     string          `json:"source"`
	OccurredAt time.Time       `json:"occurred_at"`
	Payload    json.RawMessage `json:"payload,omitempty"`
}

type Publisher interface {
	Publish(userID string, ev PublishedEvent)
}

type Service struct {
	store     Store
	publisher Publisher
}

func NewService(store Store) *Service { return &Service{store: store} }
func (s *Service) SetPublisher(p Publisher) { s.publisher = p }

var allowedSources = map[string]bool{"cli": true, "backend": true, "web": true, "local_agent": true, "github": true, "system": true}

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
		if !allowedSources[e.Source] {
			results = append(results, IngestResult{EventID: e.EventID, Status: "rejected", Reason: "invalid source"})
			continue
		}
		status, err := s.store.Ingest(ctx, userID, e)
		if err != nil {
			results = append(results, IngestResult{EventID: e.EventID, Status: "rejected", Reason: "storage error"})
			continue
		}
		results = append(results, IngestResult{EventID: e.EventID, Status: status})
		if status == "accepted" && s.publisher != nil {
			s.publisher.Publish(userID, PublishedEvent{ID: e.EventID, Type: e.EventType, ProjectID: e.ProjectID, SessionID: e.SessionID, Source: e.Source, OccurredAt: e.OccurredAt, Payload: e.Payload})
		}
	}
	return IngestResponse{Results: results}, nil
}

func (s *Service) List(ctx context.Context, userID, projectID string, limit int) ([]Event, error) {
	return s.store.ListByUser(ctx, userID, projectID, limit)
}
