package report

import "time"

type Summary struct {
	ReportType         string    `json:"report_type"`
	PeriodStart        time.Time `json:"period_start"`
	PeriodEnd          time.Time `json:"period_end"`
	ProjectID          string    `json:"project_id,omitempty"`
	Sessions           int       `json:"sessions"`
	Events             int       `json:"events"`
	Commits            int       `json:"commits"`
	TestRuns           int       `json:"test_runs"`
	FailedTestRuns     int       `json:"failed_test_runs"`
	Investigations     int       `json:"investigations"`
	OpenInvestigations int       `json:"open_investigations"`
	Decisions          int       `json:"decisions"`
	Notes              int       `json:"notes"`
}
