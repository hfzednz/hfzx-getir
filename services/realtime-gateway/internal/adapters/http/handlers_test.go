package httpadapter

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/nexora/realtime-gateway/internal/authz"
	"github.com/nexora/realtime-gateway/internal/hub"
)

func TestSSERequiresTicket(t *testing.T) {
	h := hub.New()
	ts := httptest.NewServer(NewServerAuth(":0", h, "sse-secret", "pub-token").Handler)
	defer ts.Close()
	resp, err := http.Get(ts.URL + "/v1/realtime/sse?topic=order:test")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("unauthenticated sse=%d", resp.StatusCode)
	}
}

func TestSSEOwnOrderTicket(t *testing.T) {
	h := hub.New()
	secret := "sse-secret"
	ts := httptest.NewServer(NewServerAuth(":0", h, secret, "pub-token").Handler)
	defer ts.Close()
	ticket, err := authz.IssueSSETicket(secret, "11111111-1111-1111-1111-111111111111", "cust-a", "order:1", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, ts.URL+"/v1/realtime/sse?topic=order:1&ticket="+ticket, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("status=%d", resp.StatusCode)
	}
	buf := make([]byte, 64)
	n, _ := resp.Body.Read(buf)
	if !strings.Contains(string(buf[:n]), ": connected") {
		t.Fatalf("chunk=%q", buf[:n])
	}
}

func TestSSEWrongTopicDenied(t *testing.T) {
	secret := "sse-secret"
	ts := httptest.NewServer(NewServerAuth(":0", hub.New(), secret, "pub-token").Handler)
	defer ts.Close()
	ticket, _ := authz.IssueSSETicket(secret, "t1", "cust-a", "order:1", time.Minute)
	resp, err := http.Get(ts.URL + "/v1/realtime/sse?topic=order:other&ticket=" + ticket)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("got %d", resp.StatusCode)
	}
}

func TestPublishRequiresToken(t *testing.T) {
	ts := httptest.NewServer(NewServerAuth(":0", hub.New(), "sse-secret", "pub-token").Handler)
	defer ts.Close()
	resp, err := http.Post(ts.URL+"/v1/realtime/publish", "application/json", strings.NewReader(`{"topic":"order:1"}`))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("got %d", resp.StatusCode)
	}
}

func TestSSEDeliversPublishedEvent(t *testing.T) {
	h := hub.New()
	secret := "sse-secret"
	ts := httptest.NewServer(NewServerAuth(":0", h, secret, "pub-token").Handler)
	defer ts.Close()
	ticket, _ := authz.IssueSSETicket(secret, "t1", "cust-a", "order:1", time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, ts.URL+"/v1/realtime/sse?topic=order:1&ticket="+ticket, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	go func() {
		for i := 0; i < 40; i++ {
			if h.Publish("order:1", map[string]any{"eventType": "PickingStarted", "status": "picking"}) > 0 {
				return
			}
			time.Sleep(25 * time.Millisecond)
		}
	}()

	buf := make([]byte, 512)
	var got []byte
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			got = append(got, buf[:n]...)
			if strings.Contains(string(got), "PickingStarted") {
				return
			}
		}
		if err != nil {
			t.Fatalf("read=%q err=%v", got, err)
		}
	}
}
