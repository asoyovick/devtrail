package localdb

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

type DB struct { SQL *sql.DB }

type Project struct {
	ID        string
	Name      string
	Path      string
	RepoRoot  string
	RemoteURL string
	CreatedAt time.Time
}

type Event struct {
	EventID       string
	ProjectID     string
	SessionID     *string
	EventType     string
	OccurredAt    time.Time
	Source        string
	Payload       json.RawMessage
	SchemaVersion int
	SyncStatus    string
}

func Open(path string) (*DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil { return nil, err }
	db.SetMaxOpenConns(1)
	if err := db.Ping(); err != nil { return nil, err }
	d := &DB{SQL: db}
	if err := d.migrate(context.Background()); err != nil { db.Close(); return nil, err }
	return d, nil
}

func (d *DB) Close() error { return d.SQL.Close() }

func (d *DB) migrate(ctx context.Context) error {
	stmts := []string{
		`PRAGMA journal_mode=WAL;`,
		`CREATE TABLE IF NOT EXISTS projects (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			path TEXT NOT NULL UNIQUE,
			repo_root TEXT,
			remote_url TEXT,
			created_at TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS sessions (
			id TEXT PRIMARY KEY,
			project_id TEXT NOT NULL,
			title TEXT,
			status TEXT NOT NULL,
			started_at TEXT NOT NULL,
			ended_at TEXT,
			created_at TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS events (
			event_id TEXT PRIMARY KEY,
			project_id TEXT,
			session_id TEXT,
			event_type TEXT NOT NULL,
			occurred_at TEXT NOT NULL,
			source TEXT NOT NULL,
			payload_json TEXT NOT NULL,
			schema_version INTEGER NOT NULL DEFAULT 1,
			sync_status TEXT NOT NULL DEFAULT 'pending',
			sync_attempts INTEGER NOT NULL DEFAULT 0,
			last_error TEXT,
			created_at TEXT NOT NULL
		);`,
		`CREATE INDEX IF NOT EXISTS idx_events_sync_status ON events(sync_status, created_at);`,
		`CREATE INDEX IF NOT EXISTS idx_events_project_occurred ON events(project_id, occurred_at DESC);`,
		`CREATE TABLE IF NOT EXISTS kv (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL
		);`,
	}
	for _, s := range stmts {
		if _, err := d.SQL.ExecContext(ctx, s); err != nil { return fmt.Errorf("migrate: %w", err) }
	}
	return nil
}

func (d *DB) UpsertProject(ctx context.Context, p Project) error {
	_, err := d.SQL.ExecContext(ctx, `INSERT INTO projects(id,name,path,repo_root,remote_url,created_at) VALUES(?,?,?,?,?,?)
		ON CONFLICT(path) DO UPDATE SET name=excluded.name, repo_root=excluded.repo_root, remote_url=excluded.remote_url`,
		p.ID, p.Name, p.Path, p.RepoRoot, p.RemoteURL, p.CreatedAt.UTC().Format(time.RFC3339Nano))
	return err
}

func (d *DB) ListProjects(ctx context.Context) ([]Project, error) {
	rows, err := d.SQL.QueryContext(ctx, `SELECT id,name,path,repo_root,remote_url,created_at FROM projects ORDER BY created_at DESC`)
	if err != nil { return nil, err }
	defer rows.Close()
	var out []Project
	for rows.Next() {
		var p Project; var created string
		if err := rows.Scan(&p.ID,&p.Name,&p.Path,&p.RepoRoot,&p.RemoteURL,&created); err != nil { return nil, err }
		p.CreatedAt, _ = time.Parse(time.RFC3339Nano, created)
		out = append(out, p)
	}
	return out, rows.Err()
}

func (d *DB) GetProjectByPath(ctx context.Context, path string) (Project, error) {
	var p Project; var created string
	err := d.SQL.QueryRowContext(ctx, `SELECT id,name,path,repo_root,remote_url,created_at FROM projects WHERE path=?`, path).Scan(&p.ID,&p.Name,&p.Path,&p.RepoRoot,&p.RemoteURL,&created)
	if err != nil { return Project{}, err }
	p.CreatedAt, _ = time.Parse(time.RFC3339Nano, created)
	return p, nil
}

func (d *DB) InsertEvent(ctx context.Context, e Event) error {
	payload := string(e.Payload)
	if payload == "" { payload = "{}" }
	_, err := d.SQL.ExecContext(ctx, `INSERT OR IGNORE INTO events(event_id,project_id,session_id,event_type,occurred_at,source,payload_json,schema_version,sync_status,created_at) VALUES(?,?,?,?,?,?,?,?, 'pending', ?)`,
		e.EventID, e.ProjectID, e.SessionID, e.EventType, e.OccurredAt.UTC().Format(time.RFC3339Nano), e.Source, payload, e.SchemaVersion, time.Now().UTC().Format(time.RFC3339Nano))
	return err
}

func (d *DB) ListPendingEvents(ctx context.Context, limit int) ([]Event, error) {
	rows, err := d.SQL.QueryContext(ctx, `SELECT event_id,project_id,session_id,event_type,occurred_at,source,payload_json,schema_version,sync_status FROM events WHERE sync_status='pending' ORDER BY created_at ASC LIMIT ?`, limit)
	if err != nil { return nil, err }
	defer rows.Close()
	var out []Event
	for rows.Next() {
		var e Event; var occurred string; var sessionID sql.NullString; var payload string
		if err := rows.Scan(&e.EventID,&e.ProjectID,&sessionID,&e.EventType,&occurred,&e.Source,&payload,&e.SchemaVersion,&e.SyncStatus); err != nil { return nil, err }
		e.Payload = json.RawMessage(payload)
		e.OccurredAt, _ = time.Parse(time.RFC3339Nano, occurred)
		if sessionID.Valid { e.SessionID = &sessionID.String }
		out = append(out, e)
	}
	return out, rows.Err()
}

func (d *DB) MarkEventsSynced(ctx context.Context, ids []string) error {
	tx, err := d.SQL.BeginTx(ctx, nil)
	if err != nil { return err }
	defer tx.Rollback()
	stmt, err := tx.PrepareContext(ctx, `UPDATE events SET sync_status='synced', last_error=NULL WHERE event_id=?`)
	if err != nil { return err }
	defer stmt.Close()
	for _, id := range ids { if _, err := stmt.ExecContext(ctx, id); err != nil { return err } }
	return tx.Commit()
}

func (d *DB) MarkEventError(ctx context.Context, id string, msg string) error {
	_, err := d.SQL.ExecContext(ctx, `UPDATE events SET sync_attempts=sync_attempts+1, last_error=? WHERE event_id=?`, msg, id)
	return err
}

func (d *DB) GetKV(ctx context.Context, key string) (string, bool, error) {
	var v string
	err := d.SQL.QueryRowContext(ctx, `SELECT value FROM kv WHERE key=?`, key).Scan(&v)
	if err == sql.ErrNoRows { return "", false, nil }
	return v, err == nil, err
}

func (d *DB) SetKV(ctx context.Context, key, value string) error {
	_, err := d.SQL.ExecContext(ctx, `INSERT INTO kv(key,value) VALUES(?,?) ON CONFLICT(key) DO UPDATE SET value=excluded.value`, key, value)
	return err
}

type Session struct {
	ID        string
	ProjectID string
	Title     string
	Status    string
	StartedAt time.Time
	EndedAt   *time.Time
}

func (d *DB) CreateSession(ctx context.Context, s Session) error {
	_, err := d.SQL.ExecContext(ctx, `INSERT INTO sessions(id,project_id,title,status,started_at,created_at) VALUES(?,?,?,?,?,?)`,
		s.ID, s.ProjectID, s.Title, s.Status, s.StartedAt.UTC().Format(time.RFC3339Nano), time.Now().UTC().Format(time.RFC3339Nano))
	return err
}

func (d *DB) UpdateSessionStatus(ctx context.Context, id, status string, endedAt *time.Time) error {
	var endVal any
	if endedAt != nil { endVal = endedAt.UTC().Format(time.RFC3339Nano) }
	_, err := d.SQL.ExecContext(ctx, `UPDATE sessions SET status=?, ended_at=? WHERE id=?`, status, endVal, id)
	return err
}

func (d *DB) GetSession(ctx context.Context, id string) (Session, error) {
	var s Session; var started, ended sql.NullString
	err := d.SQL.QueryRowContext(ctx, `SELECT id,project_id,title,status,started_at,ended_at FROM sessions WHERE id=?`, id).Scan(&s.ID,&s.ProjectID,&s.Title,&s.Status,&started,&ended)
	if err != nil { return Session{}, err }
	s.StartedAt, _ = time.Parse(time.RFC3339Nano, started.String)
	if ended.Valid { t, _ := time.Parse(time.RFC3339Nano, ended.String); s.EndedAt = &t }
	return s, nil
}

func (d *DB) ListSessions(ctx context.Context) ([]Session, error) {
	rows, err := d.SQL.QueryContext(ctx, `SELECT id,project_id,title,status,started_at,ended_at FROM sessions ORDER BY started_at DESC LIMIT 50`)
	if err != nil { return nil, err }
	defer rows.Close()
	var out []Session
	for rows.Next() {
		var s Session; var started string; var ended sql.NullString
		if err := rows.Scan(&s.ID,&s.ProjectID,&s.Title,&s.Status,&started,&ended); err != nil { return nil, err }
		s.StartedAt, _ = time.Parse(time.RFC3339Nano, started)
		if ended.Valid { t, _ := time.Parse(time.RFC3339Nano, ended.String); s.EndedAt = &t }
		out = append(out, s)
	}
	return out, rows.Err()
}
