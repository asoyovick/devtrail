package session

import "time"

type Session struct {
	ID              string     `json:"id"`
	UserID          string     `json:"user_id"`
	ProjectID       string     `json:"project_id"`
	Title           string     `json:"title,omitempty"`
	Status          string     `json:"status"`
	StartedAt       time.Time  `json:"started_at"`
	EndedAt         *time.Time `json:"ended_at,omitempty"`
	DurationSeconds *int       `json:"duration_seconds,omitempty"`
	Source          string     `json:"source"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type StartRequest struct {
	ProjectID string `json:"project_id"`
	Title     string `json:"title"`
	Source    string `json:"source"`
}
