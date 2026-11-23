package decision

import "time"

type Decision struct {
	ID               string    `json:"id"`
	UserID           string    `json:"user_id"`
	ProjectID        string    `json:"project_id"`
	Title            string    `json:"title"`
	Context          string    `json:"context"`
	Problem          string    `json:"problem"`
	Alternatives     string    `json:"alternatives,omitempty"`
	SelectedApproach string    `json:"selected_approach"`
	Reasoning        string    `json:"reasoning"`
	Consequences     string    `json:"consequences,omitempty"`
	Status           string    `json:"status"`
	DecidedAt        time.Time `json:"decided_at"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type CreateRequest struct {
	ProjectID        string `json:"project_id"`
	Title            string `json:"title"`
	Context          string `json:"context"`
	Problem          string `json:"problem"`
	Alternatives     string `json:"alternatives"`
	SelectedApproach string `json:"selected_approach"`
	Reasoning        string `json:"reasoning"`
	Consequences     string `json:"consequences"`
	Status           string `json:"status"`
}

type UpdateRequest struct {
	Title            *string `json:"title"`
	Context          *string `json:"context"`
	Problem          *string `json:"problem"`
	Alternatives     *string `json:"alternatives"`
	SelectedApproach *string `json:"selected_approach"`
	Reasoning        *string `json:"reasoning"`
	Consequences     *string `json:"consequences"`
	Status           *string `json:"status"`
}
