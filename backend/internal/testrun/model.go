package testrun

import "time"

type TestRun struct {
	ID             string     `json:"id"`
	UserID         string     `json:"user_id"`
	ProjectID      string     `json:"project_id"`
	SessionID      *string    `json:"session_id,omitempty"`
	Command        string     `json:"command"`
	Status         string     `json:"status"`
	StartedAt      time.Time  `json:"started_at"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
	DurationMs     *int64     `json:"duration_ms,omitempty"`
	TotalTests     *int       `json:"total_tests,omitempty"`
	PassedTests    *int       `json:"passed_tests,omitempty"`
	FailedTests    *int       `json:"failed_tests,omitempty"`
	FailureSummary string     `json:"failure_summary,omitempty"`
	EventID        *string    `json:"event_id,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

type CreateRequest struct {
	ProjectID      string     `json:"project_id"`
	SessionID      *string    `json:"session_id"`
	Command        string     `json:"command"`
	Status         string     `json:"status"`
	StartedAt      time.Time  `json:"started_at"`
	CompletedAt    *time.Time `json:"completed_at"`
	DurationMs     *int64     `json:"duration_ms"`
	TotalTests     *int       `json:"total_tests"`
	PassedTests    *int       `json:"passed_tests"`
	FailedTests    *int       `json:"failed_tests"`
	FailureSummary string     `json:"failure_summary"`
	EventID        *string    `json:"event_id"`
}
