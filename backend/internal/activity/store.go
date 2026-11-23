package activity

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Store interface {
	Ingest(ctx context.Context, userID string, e IngestEvent) (string, error)
	ListByUser(ctx context.Context, userID, projectID string, limit int) ([]Event, error)
}

type PGStore struct{ pool *pgxpool.Pool }

func NewPGStore(pool *pgxpool.Pool) *PGStore { return &PGStore{pool: pool} }

func (s *PGStore) Ingest(ctx context.Context, userID string, e IngestEvent) (string, error) {
	payload := e.Payload
	if len(payload) == 0 || strings.TrimSpace(string(payload)) == "" {
		payload = json.RawMessage(`{}`)
	}
	version := e.SchemaVersion
	if version <= 0 {
		version = 1
	}
	tag, err := s.pool.Exec(ctx, `
		INSERT INTO activity_events (event_id, user_id, project_id, session_id, event_type, occurred_at, source, payload, schema_version)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		ON CONFLICT (event_id) DO NOTHING
	`, e.EventID, userID, e.ProjectID, e.SessionID, e.EventType, e.OccurredAt, e.Source, payload, version)
	if err != nil {
		return "rejected", fmt.Errorf("ingest event: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return "duplicate", nil
	}
	return "accepted", nil
}

func (s *PGStore) ListByUser(ctx context.Context, userID, projectID string, limit int) ([]Event, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	query := `
		SELECT id, event_id, user_id, project_id, session_id, event_type, occurred_at, source, payload, schema_version, ingested_at
		FROM activity_events WHERE user_id = $1`
	args := []any{userID}
	if strings.TrimSpace(projectID) != "" {
		query += " AND project_id = $2"
		args = append(args, projectID)
		query += " ORDER BY occurred_at DESC LIMIT $3"
		args = append(args, limit)
	} else {
		query += " ORDER BY occurred_at DESC LIMIT $2"
		args = append(args, limit)
	}
	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list events: %w", err)
	}
	defer rows.Close()
	out := []Event{}
	for rows.Next() {
		var e Event
		if err := rows.Scan(&e.ID, &e.EventID, &e.UserID, &e.ProjectID, &e.SessionID, &e.EventType, &e.OccurredAt, &e.Source, &e.Payload, &e.SchemaVersion, &e.IngestedAt); err != nil {
			return nil, fmt.Errorf("scan event: %w", err)
		}
		out = append(out, e)
	}
	return out, rows.Err()
}
