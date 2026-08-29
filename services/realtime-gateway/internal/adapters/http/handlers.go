package httpadapter

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/nexora/realtime-gateway/internal/authz"
	"github.com/nexora/realtime-gateway/internal/hub"
)

func NewServer(addr string, h *hub.Hub) *http.Server {
	return NewServerAuth(addr, h, os.Getenv("SSE_TICKET_SECRET"), os.Getenv("REALTIME_PUBLISH_TOKEN"))
}

func NewServerAuth(addr string, h *hub.Hub, ticketSecret, publishToken string) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /ready", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ready"})
	})
	mux.HandleFunc("POST /v1/realtime/publish", func(w http.ResponseWriter, r *http.Request) {
		if publishToken == "" || r.Header.Get("X-Realtime-Publish-Token") != publishToken {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
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
		ticket := r.URL.Query().Get("ticket")
		if topic == "" {
			http.Error(w, "topic required", 400)
			return
		}
		claims, err := authz.ParseSSETicket(ticketSecret, ticket)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		if claims.Topic != topic {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		hdrTid := r.Header.Get("X-Tenant-Id")
		if hdrTid != "" && claims.Tenant != "" && hdrTid != claims.Tenant {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "stream unsupported", 500)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("X-Accel-Buffering", "no")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, ": connected\n\n")
		flusher.Flush()
		ch := h.Subscribe(topic)
		defer h.Unsubscribe(topic, ch)
		ctx := r.Context()
		ping := time.NewTicker(15 * time.Second)
		defer ping.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ping.C:
				_, _ = fmt.Fprint(w, ": ping\n\n")
				flusher.Flush()
			case msg := <-ch:
				fmt.Fprintf(w, "data: %s\n\n", msg)
				flusher.Flush()
			}
		}
	})
	return &http.Server{Addr: addr, Handler: mux, ReadHeaderTimeout: 5 * time.Second}
}
