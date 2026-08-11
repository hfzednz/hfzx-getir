package hub

import (
	"encoding/json"
	"sync"
)

// Hub fans out tracking/ops events to subscribers (WS/SSE).
type Hub struct {
	mu   sync.RWMutex
	subs map[string]map[chan []byte]struct{}
}

func New() *Hub {
	return &Hub{subs: make(map[string]map[chan []byte]struct{})}
}

func (h *Hub) Subscribe(topic string) chan []byte {
	ch := make(chan []byte, 16)
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.subs[topic] == nil {
		h.subs[topic] = make(map[chan []byte]struct{})
	}
	h.subs[topic][ch] = struct{}{}
	return ch
}

func (h *Hub) Unsubscribe(topic string, ch chan []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if m := h.subs[topic]; m != nil {
		delete(m, ch)
		close(ch)
	}
}

func (h *Hub) Publish(topic string, payload any) int {
	b, _ := json.Marshal(payload)
	h.mu.RLock()
	defer h.mu.RUnlock()
	n := 0
	for ch := range h.subs[topic] {
		select {
		case ch <- b:
			n++
		default:
		}
	}
	return n
}

func (h *Hub) Topics() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.subs)
}
