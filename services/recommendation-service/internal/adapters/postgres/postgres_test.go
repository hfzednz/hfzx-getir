package postgres

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/recommendation-service/internal/app/ports"
	"github.com/nexora/recommendation-service/internal/domain"
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
	a := TextArray{"electronics", "gift"}
	v, err := a.Value()
	if err != nil {
		t.Fatal(err)
	}
	var out TextArray
	if err := out.Scan(v); err != nil {
		t.Fatal(err)
	}
	if len(out) != 2 || out[0] != "electronics" {
		t.Fatalf("got %#v", out)
	}
}

func TestNullHelpers(t *testing.T) {
	if nullUUIDValue(uuid.Nil) != nil {
		t.Fatal("nil uuid")
	}
	id := uuid.New()
	if nullUUIDValue(id) != id {
		t.Fatal("uuid value")
	}
	now := time.Now().UTC()
	if nullTime(nil) != nil || nullTime(&time.Time{}) != nil {
		t.Fatal("null time")
	}
	if nullTime(&now) != now {
		t.Fatal("time value")
	}
	if scanUUIDOrNil(uuid.NullUUID{}) != uuid.Nil {
		t.Fatal("scanUUIDOrNil")
	}
}

func TestInterfaceCompliance(t *testing.T) {
	var (
		_ ports.FeatureRepo      = (*FeatureRepo)(nil)
		_ ports.SignalRepo       = (*SignalRepo)(nil)
		_ ports.CoOccurRepo      = (*CoOccurRepo)(nil)
		_ ports.OutboxRepository = (*OutboxRepo)(nil)
	)
	_ = domain.SignalPurchase
	b, _ := json.Marshal(domain.CoOccurrence{Count: 1})
	if len(b) == 0 {
		t.Fatal("marshal")
	}
	repos := NewRepos(nil)
	if repos.Features == nil || repos.Outbox == nil {
		t.Fatal("NewRepos")
	}
}

func TestMapUniqueViolationNil(t *testing.T) {
	if mapUniqueViolation(nil) != nil {
		t.Fatal("nil")
	}
}
