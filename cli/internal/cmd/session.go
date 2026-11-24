package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"devtrail/cli/internal/localdb"
)

func Session(args []string) error {
	if len(args) < 1 { return fmt.Errorf("usage: devtrail session <start|pause|resume|end>") }
	switch args[0] {
	case "start": return sessionStart(args[1:])
	case "pause": return sessionTransition("paused", "session.paused")
	case "resume": return sessionTransition("active", "session.resumed")
	case "end": return sessionTransition("ended", "session.ended")
	default: return fmt.Errorf("unknown session command: %s", args[0])
	}
}

func sessionStart(args []string) error {
	var projectID, title string
	for i := 0; i < len(args); i++ {
		if args[i] == "--project" && i+1 < len(args) { projectID = args[i+1]; i++ }
		if args[i] == "--title" && i+1 < len(args) { title = args[i+1]; i++ }
	}
	db, _, err := openDB()
	if err != nil { return err }
	defer db.Close()
	ctx := context.Background()
	if projectID == "" {
		projects, _ := db.ListProjects(ctx)
		if len(projects) == 1 { projectID = projects[0].ID } else { return fmt.Errorf("provide --project <id>") }
	}
	s := localdb.Session{ID: uuidv4(), ProjectID: projectID, Title: title, Status: "active", StartedAt: time.Now().UTC()}
	if err := db.CreateSession(ctx, s); err != nil { return err }
	_ = db.SetKV(ctx, "active_session_id", s.ID)
	payload, _ := json.Marshal(map[string]any{"title": title, "project_id": projectID})
	_ = db.InsertEvent(ctx, localdb.Event{EventID: uuidv4(), ProjectID: projectID, SessionID: &s.ID, EventType: "session.started", OccurredAt: s.StartedAt, Source: "local_agent", Payload: payload, SchemaVersion: 1})
	fmt.Println("Session started:", s.ID)
	return nil
}

func sessionTransition(status, eventType string) error {
	db, _, err := openDB()
	if err != nil { return err }
	defer db.Close()
	ctx := context.Background()
	id, ok, _ := db.GetKV(ctx, "active_session_id")
	if !ok || id == "" { return fmt.Errorf("no active session") }
	s, err := db.GetSession(ctx, id)
	if err != nil { return err }
	var ended *time.Time
	if status == "ended" { t := time.Now().UTC(); ended = &t; _ = db.SetKV(ctx, "active_session_id", "") }
	if status == "active" { _ = db.SetKV(ctx, "active_session_id", id) }
	if err := db.UpdateSessionStatus(ctx, id, status, ended); err != nil { return err }
	payload, _ := json.Marshal(map[string]any{"status": status})
	_ = db.InsertEvent(ctx, localdb.Event{EventID: uuidv4(), ProjectID: s.ProjectID, SessionID: &s.ID, EventType: eventType, OccurredAt: time.Now().UTC(), Source: "local_agent", Payload: payload, SchemaVersion: 1})
	fmt.Println("Session", status+":", id)
	return nil
}

func Sessions(args []string) error {
	db, _, err := openDB()
	if err != nil { return err }
	defer db.Close()
	sessions, err := db.ListSessions(context.Background())
	if err != nil { return err }
	if len(sessions) == 0 { fmt.Println("No sessions."); return nil }
	for _, s := range sessions {
		fmt.Printf("- %s [%s] %s project=%s\n", s.ID, s.Status, s.Title, s.ProjectID)
	}
	return nil
}
