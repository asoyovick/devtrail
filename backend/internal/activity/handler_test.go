package activity

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"devtrail/backend/internal/user"
)

type fakeStore struct{ events map[string]Event }

func newFakeStore() *fakeStore { return &fakeStore{events: map[string]Event{}} }

func (s *fakeStore) Ingest(_ context.Context, userID string, e IngestEvent) (string, error) {
	if _, ok := s.events[e.EventID]; ok {
		return "duplicate", nil
	}
	s.events[e.EventID] = Event{ID: "1", EventID: e.EventID, UserID: userID, EventType: e.EventType, OccurredAt: e.OccurredAt, Source: e.Source, Payload: e.Payload, SchemaVersion: e.SchemaVersion}
	return "accepted", nil
}

func (s *fakeStore) ListByUser(_ context.Context, userID, projectID string, limit int) ([]Event, error) {
	out := []Event{}
	for _, e := range s.events {
		if e.UserID == userID {
			out = append(out, e)
		}
	}
	if len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

func fakeRequireAuth(userID string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r.WithContext(user.WithUser(r.Context(), user.User{ID: userID})))
		}
	}
}

func TestIngestAndList(t *testing.T) {
	store := newFakeStore()
	h := NewHandler(NewService(store), fakeRequireAuth("user-1"))
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	payload := `{"events":[{"event_id":"11111111-1111-1111-1111-111111111111","event_type":"session.started","occurred_at":"2026-11-23T10:00:00Z","source":"cli","payload":{"ok":true},"schema_version":1}]}`
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/api/v1/sync/events", bytes.NewBufferString(payload)))
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	var resp IngestResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(resp.Results) != 1 || resp.Results[0].Status != "accepted" {
		t.Fatalf("unexpected resp: %+v", resp)
	}

	// duplicate
	rr2 := httptest.NewRecorder()
	mux.ServeHTTP(rr2, httptest.NewRequest(http.MethodPost, "/api/v1/sync/events", bytes.NewBufferString(payload)))
	var resp2 IngestResponse
	_ = json.Unmarshal(rr2.Body.Bytes(), &resp2)
	if len(resp2.Results) != 1 || resp2.Results[0].Status != "duplicate" {
		t.Fatalf("unexpected dup resp: %+v", resp2)
	}

	listRR := httptest.NewRecorder()
	mux.ServeHTTP(listRR, httptest.NewRequest(http.MethodGet, "/api/v1/activity", nil))
	if listRR.Code != http.StatusOK {
		t.Fatalf("list status=%d", listRR.Code)
	}
}
