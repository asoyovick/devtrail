package streaming

import (
	"sync"
	"time"
)

type Event struct {
	ID        string         `json:"id"`
	Type      string         `json:"type"`
	ProjectID *string        `json:"project_id,omitempty"`
	SessionID *string        `json:"session_id,omitempty"`
	Source    string         `json:"source"`
	OccurredAt time.Time     `json:"occurred_at"`
	Payload   any            `json:"payload,omitempty"`
}

type Hub struct {
	mu    sync.RWMutex
	subs  map[string]map[chan Event]struct{}
}

func NewHub() *Hub { return &Hub{subs: map[string]map[chan Event]struct{}{}} }

func (h *Hub) Subscribe(userID string) (chan Event, func()) {
	ch := make(chan Event, 16)
	h.mu.Lock()
	if _, ok := h.subs[userID]; !ok { h.subs[userID] = map[chan Event]struct{}{} }
	h.subs[userID][ch] = struct{}{}
	h.mu.Unlock()
	return ch, func() {
		h.mu.Lock()
		if m, ok := h.subs[userID]; ok {
			delete(m, ch)
			if len(m) == 0 { delete(h.subs, userID) }
		}
		h.mu.Unlock()
		close(ch)
	}
}

func (h *Hub) Publish(userID string, ev Event) {
	h.mu.RLock()
	m := h.subs[userID]
	for ch := range m {
		select { case ch <- ev: default: }
	}
	h.mu.RUnlock()
}
