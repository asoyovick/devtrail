package activity

import (
	"encoding/json"
	"time"
)

type Event struct {
	ID            string          `json:"id"`
	EventID       string          `json:"event_id"`
	UserID        string          `json:"user_id"`
	ProjectID     *string         `json:"project_id,omitempty"`
	SessionID     *string         `json:"session_id,omitempty"`
	EventType     string          `json:"event_type"`
	OccurredAt    time.Time       `json:"occurred_at"`
	Source        string          `json:"source"`
	Payload       json.RawMessage `json:"payload"`
	SchemaVersion int             `json:"schema_version"`
	IngestedAt    time.Time       `json:"ingested_at"`
}

type IngestEvent struct {
	EventID       string          `json:"event_id"`
	ProjectID     *string         `json:"project_id"`
	SessionID     *string         `json:"session_id"`
	EventType     string          `json:"event_type"`
	OccurredAt    time.Time       `json:"occurred_at"`
	Source        string          `json:"source"`
	Payload       json.RawMessage `json:"payload"`
	SchemaVersion int             `json:"schema_version"`
}

type IngestRequest struct {
	Events []IngestEvent `json:"events"`
}

type IngestResult struct {
	EventID string `json:"event_id"`
	Status  string `json:"status"`
	Reason  string `json:"reason,omitempty"`
}

type IngestResponse struct {
	Results []IngestResult `json:"results"`
}
