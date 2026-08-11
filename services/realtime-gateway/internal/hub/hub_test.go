package hub_test

import (
	"testing"

	"github.com/nexora/realtime-gateway/internal/hub"
)

func TestFanout(t *testing.T) {
	h := hub.New()
	ch := h.Subscribe("order:1")
	n := h.Publish("order:1", map[string]any{"status": "moving", "lat": 41.0})
	if n != 1 {
		t.Fatalf("delivered=%d", n)
	}
	msg := <-ch
	if len(msg) == 0 {
		t.Fatal("empty")
	}
	h.Unsubscribe("order:1", ch)
}
