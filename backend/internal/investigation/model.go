package investigation

import "time"

type Investigation struct {
	ID                   string     `json:"id"`
	UserID               string     `json:"user_id"`
	ProjectID            string     `json:"project_id"`
	SessionID            *string    `json:"session_id,omitempty"`
	Title                string     `json:"title"`
	ProblemDescription   string     `json:"problem_description"`
	ObservedBehavior     string     `json:"observed_behavior,omitempty"`
	InvestigationSteps   string     `json:"investigation_steps,omitempty"`
	AttemptedApproaches  string     `json:"attempted_approaches,omitempty"`
	RootCause            string     `json:"root_cause,omitempty"`
	Resolution           string     `json:"resolution,omitempty"`
	LessonsLearned       string     `json:"lessons_learned,omitempty"`
	PreventionStrategies string     `json:"prevention_strategies,omitempty"`
	Status               string     `json:"status"`
	Severity             string     `json:"severity,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
	ResolvedAt           *time.Time `json:"resolved_at,omitempty"`
}

type CreateRequest struct {
	ProjectID            string  `json:"project_id"`
	SessionID            *string `json:"session_id"`
	Title                string  `json:"title"`
	ProblemDescription   string  `json:"problem_description"`
	ObservedBehavior     string  `json:"observed_behavior"`
	InvestigationSteps   string  `json:"investigation_steps"`
	AttemptedApproaches  string  `json:"attempted_approaches"`
	RootCause            string  `json:"root_cause"`
	Resolution           string  `json:"resolution"`
	LessonsLearned       string  `json:"lessons_learned"`
	PreventionStrategies string  `json:"prevention_strategies"`
	Severity             string  `json:"severity"`
}

type UpdateRequest struct {
	Title                *string `json:"title"`
	ProblemDescription   *string `json:"problem_description"`
	ObservedBehavior     *string `json:"observed_behavior"`
	InvestigationSteps   *string `json:"investigation_steps"`
	AttemptedApproaches  *string `json:"attempted_approaches"`
	RootCause            *string `json:"root_cause"`
	Resolution           *string `json:"resolution"`
	LessonsLearned       *string `json:"lessons_learned"`
	PreventionStrategies *string `json:"prevention_strategies"`
	Status               *string `json:"status"`
	Severity             *string `json:"severity"`
}
