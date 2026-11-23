package gitactivity

import "time"

type Commit struct {
	ID             string    `json:"id"`
	UserID         string    `json:"user_id"`
	ProjectID      string    `json:"project_id"`
	SessionID      *string   `json:"session_id,omitempty"`
	CommitHash     string    `json:"commit_hash"`
	Branch         string    `json:"branch,omitempty"`
	MessageSummary string    `json:"message_summary,omitempty"`
	AuthorName     string    `json:"author_name,omitempty"`
	AuthorEmail    string    `json:"author_email,omitempty"`
	CommittedAt    time.Time `json:"committed_at"`
	FilesChanged   *int      `json:"files_changed,omitempty"`
	Insertions     *int      `json:"insertions,omitempty"`
	Deletions      *int      `json:"deletions,omitempty"`
	EventID        *string   `json:"event_id,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

type CreateRequest struct {
	ProjectID      string    `json:"project_id"`
	SessionID      *string   `json:"session_id"`
	CommitHash     string    `json:"commit_hash"`
	Branch         string    `json:"branch"`
	MessageSummary string    `json:"message_summary"`
	AuthorName     string    `json:"author_name"`
	AuthorEmail    string    `json:"author_email"`
	CommittedAt    time.Time `json:"committed_at"`
	FilesChanged   *int      `json:"files_changed"`
	Insertions     *int      `json:"insertions"`
	Deletions      *int      `json:"deletions"`
	EventID        *string   `json:"event_id"`
}
