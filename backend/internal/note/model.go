package note

import "time"

type Note struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	ProjectID *string   `json:"project_id,omitempty"`
	SessionID *string   `json:"session_id,omitempty"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	Tags      []string  `json:"tags"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateRequest struct {
	ProjectID *string  `json:"project_id"`
	SessionID *string  `json:"session_id"`
	Title     string   `json:"title"`
	Body      string   `json:"body"`
	Tags      []string `json:"tags"`
}

type UpdateRequest struct {
	Title *string   `json:"title"`
	Body  *string   `json:"body"`
	Tags  *[]string `json:"tags"`
}
