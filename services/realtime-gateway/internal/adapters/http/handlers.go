package httpadapter

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/nexora/realtime-gateway/internal/hub"
)

func NewServer(addr string, h *hub.Hub) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /ready", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ready"})
	})
	mux.HandleFunc("POST /v1/realtime/publish", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Topic   string         `json:"topic"`
			Payload map[string]any `json:"payload"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		n := h.Publish(body.Topic, body.Payload)
		_ = json.NewEncoder(w).Encode(map[string]any{"delivered": n})
	})
	mux.HandleFunc("GET /v1/realtime/sse", func(w http.ResponseWriter, r *http.Request) {
		topic := r.URL.Query().Get("topic")
		if topic == "" {
			http.Error(w, "topic required", 400)
			return
		}
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "stream unsupported", 500)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		ch := h.Subscribe(topic)
		defer h.Unsubscribe(topic, ch)
		ctx := r.Context()
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-ch:
				fmt.Fprintf(w, "data: %s\n\n", msg)
				flusher.Flush()
			}
		}
	})
	return &http.Server{Addr: addr, Handler: mux, ReadHeaderTimeout: 5 * time.Second}
}
