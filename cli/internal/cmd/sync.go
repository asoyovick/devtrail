package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"devtrail/cli/internal/config"
	"devtrail/cli/internal/localdb"
)

func Sync(args []string) error {
	db, cfg, err := openDB()
	if err != nil { return err }
	defer db.Close()
	if err := syncPending(context.Background(), db, cfg); err != nil { return err }
	fmt.Println("Sync complete.")
	return nil
}

func syncPending(ctx context.Context, db *localdb.DB, cfg config.Config) error {
	if cfg.AuthToken == "" {
		return fmt.Errorf("not authenticated; run: devtrail login --email <email> --password <password>")
	}
	events, err := db.ListPendingEvents(ctx, 100)
	if err != nil { return err }
	if len(events) == 0 { return nil }
	// ensure project mappings
	mapped := make([]map[string]any, 0, len(events))
	ids := make([]string, 0, len(events))
	for _, e := range events {
		backendProjectID, err := ensureBackendProject(ctx, db, cfg, e.ProjectID)
		if err != nil { return fmt.Errorf("map project %s: %w", e.ProjectID, err) }
		var sid *string
		if e.SessionID != nil && *e.SessionID != "" { sid = e.SessionID }
		mapped = append(mapped, map[string]any{
			"event_id": e.EventID,
			"project_id": backendProjectID,
			"session_id": sid,
			"event_type": e.EventType,
			"occurred_at": e.OccurredAt.Format(time.RFC3339Nano),
			"source": e.Source,
			"payload": json.RawMessage(e.Payload),
			"schema_version": e.SchemaVersion,
		})
		ids = append(ids, e.EventID)
	}
	body, _ := json.Marshal(map[string]any{"events": mapped})
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, cfg.APIBaseURL+"/api/v1/sync/events", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+cfg.AuthToken)
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil { return err }
	defer res.Body.Close()
	respBody, _ := io.ReadAll(res.Body)
	if res.StatusCode >= 300 {
		return fmt.Errorf("sync failed (%d): %s", res.StatusCode, string(respBody))
	}
	var out struct { Results []struct{ EventID, Status, Reason string } `json:"results"` }
	if err := json.Unmarshal(respBody, &out); err == nil {
		accepted := []string{}
		for _, r := range out.Results {
			if r.Status == "accepted" || r.Status == "duplicate" { accepted = append(accepted, r.EventID) } else { _ = db.MarkEventError(ctx, r.EventID, r.Reason) }
		}
		if len(accepted) > 0 { _ = db.MarkEventsSynced(ctx, accepted) }
		fmt.Printf("Synced %d/%d events\n", len(accepted), len(events))
		return nil
	}
	_ = db.MarkEventsSynced(ctx, ids)
	fmt.Printf("Synced %d events\n", len(events))
	return nil
}

func ensureBackendProject(ctx context.Context, db *localdb.DB, cfg config.Config, localProjectID string) (string, error) {
	if localProjectID == "" { return "", nil }
	key := "backend_project_id:" + localProjectID
	if v, ok, _ := db.GetKV(ctx, key); ok && v != "" { return v, nil }
	// find local project
	projects, err := db.ListProjects(ctx)
	if err != nil { return "", err }
	var name string
	for _, p := range projects { if p.ID == localProjectID { name = p.Name; break } }
	if name == "" { name = "DevTrail Project" }
	// create backend project
	body, _ := json.Marshal(map[string]any{"name": name, "description": "Synced from DevTrail local agent"})
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, cfg.APIBaseURL+"/api/v1/projects", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+cfg.AuthToken)
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil { return "", err }
	defer res.Body.Close()
	respBody, _ := io.ReadAll(res.Body)
	if res.StatusCode >= 300 { return "", fmt.Errorf("create project: %s", string(respBody)) }
	var p struct { ID string `json:"id"` }
	if err := json.Unmarshal(respBody, &p); err != nil { return "", err }
	_ = db.SetKV(ctx, key, p.ID)
	return p.ID, nil
}
