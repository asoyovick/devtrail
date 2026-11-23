package project

import "time"

type Project struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	Name          string    `json:"name"`
	Description   string    `json:"description,omitempty"`
	RepoRootHash  string    `json:"repo_root_hash,omitempty"`
	DefaultBranch string    `json:"default_branch,omitempty"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type CreateRequest struct {
	Name          string `json:"name"`
	Description   string `json:"description"`
	RepoRootHash  string `json:"repo_root_hash"`
	DefaultBranch string `json:"default_branch"`
}

type UpdateRequest struct {
	Name          *string `json:"name"`
	Description   *string `json:"description"`
	DefaultBranch *string `json:"default_branch"`
	Status        *string `json:"status"`
}
