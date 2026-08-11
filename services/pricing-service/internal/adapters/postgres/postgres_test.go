package postgres

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/pricing-service/internal/app/ports"
)

func TestOpenRequiresURL(t *testing.T) {
	if _, err := Open(""); err == nil {
		t.Fatal("expected error for empty DATABASE_URL")
	}
}

func TestJSONMapRoundTrip(t *testing.T) {
	m := JSONMap{"k": "v", "n": float64(1)}
	v, err := m.Value()
	if err != nil {
		t.Fatal(err)
	}
	var out JSONMap
	if err := out.Scan(v); err != nil {
		t.Fatal(err)
	}
	if out["k"] != "v" {
		t.Fatalf("got %#v", out)
	}
	nilMap := JSONMap(nil)
	v2, err := nilMap.Value()
	if err != nil {
		t.Fatal(err)
	}
	var empty JSONMap
	if err := empty.Scan(v2); err != nil {
		t.Fatal(err)
	}
	if err := empty.Scan(nil); err != nil {
		t.Fatal(err)
	}
}

func TestNullHelpers(t *testing.T) {
	id := uuid.New()
	if nullUUID(nil) != nil {
		t.Fatal("nil uuid")
	}
	nilID := uuid.Nil
	if nullUUID(&nilID) != nil {
		t.Fatal("nil uuid value")
	}
	if nullUUID(&id) != id {
		t.Fatal("uuid")
	}
	now := time.Now().UTC()
	if nullTime(nil) != nil {
		t.Fatal("nil time")
	}
	if got := nullTime(&now); got != now {
		t.Fatalf("time got %v", got)
	}
	zero := time.Time{}
	if nullTime(&zero) != nil {
		t.Fatal("zero time")
	}
	if scanNullUUID(uuid.NullUUID{}) != nil {
		t.Fatal("invalid null uuid")
	}
	if scanNullUUID(uuid.NullUUID{UUID: id, Valid: true}) == nil {
		t.Fatal("valid null uuid")
	}
}

func TestMapUniqueViolationNil(t *testing.T) {
	if mapUniqueViolation(nil) != nil {
		t.Fatal("nil")
	}
}

func TestInterfaceCompliance(t *testing.T) {
	var (
		_ ports.PriceRepo         = (*PriceRepo)(nil)
		_ ports.TaxRepo           = (*TaxRepo)(nil)
		_ ports.DynamicRepo       = (*DynamicRepo)(nil)
		_ ports.QuoteAuditRepo    = (*QuoteAuditRepo)(nil)
		_ ports.OutboxRepository  = (*OutboxRepo)(nil)
	)
	repos := NewRepos(nil)
	if repos.Prices == nil || repos.Taxes == nil || repos.Dynamics == nil || repos.Audits == nil || repos.Outbox == nil {
		t.Fatal("NewRepos missing adapters")
	}
}
