package streaming

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"devtrail/backend/internal/platform/httpserver"
	"devtrail/backend/internal/user"
)

type Handler struct {
	hub         *Hub
	requireAuth func(http.HandlerFunc) http.HandlerFunc
}

func NewHandler(hub *Hub, requireAuth func(http.HandlerFunc) http.HandlerFunc) *Handler {
	return &Handler{hub: hub, requireAuth: requireAuth}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/stream/activity", h.requireAuth(h.Stream))
}

func (h *Handler) Stream(w http.ResponseWriter, r *http.Request) {
	u, ok := user.UserFromContext(r.Context())
	if !ok { httpserver.Error(w, 401, "unauthorized", "authentication required", nil); return }
	flusher, ok := w.(http.Flusher)
	if !ok { httpserver.Error(w, 500, "stream_unsupported", "streaming unsupported", nil); return }
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	ch, cancel := h.hub.Subscribe(u.ID)
	defer cancel()

	fmt.Fprintf(w, "event: connected\ndata: {\"ok\":true}\n\n")
	flusher.Flush()
	keepalive := time.NewTicker(25 * time.Second)
	defer keepalive.Stop()
	for {
		select {
		case <-r.Context().Done():
			return
		case ev, ok := <-ch:
			if !ok { return }
			b, _ := json.Marshal(ev)
			fmt.Fprintf(w, "id: %s\nevent: activity\ndata: %s\n\n", ev.ID, b)
			flusher.Flush()
		case <-keepalive.C:
			fmt.Fprintf(w, ": keepalive\n\n")
			flusher.Flush()
		}
	}
}
