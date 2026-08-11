package postgres

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/crm-service/internal/app/ports"
)

func TestOpenRequiresURL(t *testing.T) {
	if _, err := Open(""); err == nil {
		t.Fatal("expected error for empty DATABASE_URL")
	}
}

func TestJSONMapRoundTrip(t *testing.T) {
	m := JSONMap{"a": "b", "n": float64(1)}
	v, err := m.Value()
	if err != nil {
		t.Fatal(err)
	}
	var out JSONMap
	if err := out.Scan(v); err != nil {
		t.Fatal(err)
	}
	if out["a"] != "b" {
		t.Fatalf("got %#v", out)
	}
}

func TestTextArrayRoundTrip(t *testing.T) {
	a := TextArray{"vip", "refund"}
	v, err := a.Value()
	if err != nil {
		t.Fatal(err)
	}
	var out TextArray
	if err := out.Scan(v); err != nil {
		t.Fatal(err)
	}
	if len(out) != 2 || out[0] != "vip" {
		t.Fatalf("got %#v", out)
	}
}

func TestUUIDArrayRoundTrip(t *testing.T) {
	id := uuid.New()
	a := UUIDArray{id}
	v, err := a.Value()
	if err != nil {
		t.Fatal(err)
	}
	var out UUIDArray
	if err := out.Scan(v); err != nil {
		t.Fatal(err)
	}
	if len(out) != 1 || out[0] != id {
		t.Fatalf("got %#v", out)
	}
}

func TestNullHelpers(t *testing.T) {
	id := uuid.New()
	if nullUUID(nil) != nil {
		t.Fatal("nil pointer")
	}
	if nullUUID(&id) != id {
		t.Fatal("uuid value")
	}
	now := time.Now().UTC()
	if nullTime(nil) != nil || nullTime(&time.Time{}) != nil {
		t.Fatal("null time")
	}
	if nullTime(&now) != now {
		t.Fatal("time value")
	}
	n := uuid.NullUUID{UUID: id, Valid: true}
	if got := scanNullUUID(n); got == nil || *got != id {
		t.Fatal("scanNullUUID")
	}
}

func TestInterfaceCompliance(t *testing.T) {
	var (
		_ ports.TicketRepo       = (*TicketRepo)(nil)
		_ ports.ChatRepo         = (*ChatRepo)(nil)
		_ ports.AgentRepo        = (*AgentRepo)(nil)
		_ ports.KBRepo           = (*KBRepo)(nil)
		_ ports.CaseRepo         = (*CaseRepo)(nil)
		_ ports.FeedbackRepo     = (*FeedbackRepo)(nil)
		_ ports.SLARepo          = (*SLARepo)(nil)
		_ ports.OutboxRepository = (*OutboxRepo)(nil)
	)
	b, _ := json.Marshal(map[string]any{"ok": true})
	if len(b) == 0 {
		t.Fatal("marshal")
	}
	repos := NewRepos(nil)
	if repos.Tickets == nil || repos.Outbox == nil || repos.KB == nil {
		t.Fatal("NewRepos")
	}
}

func TestMapUniqueViolationNil(t *testing.T) {
	if mapUniqueViolation(nil) != nil {
		t.Fatal("nil")
	}
}
