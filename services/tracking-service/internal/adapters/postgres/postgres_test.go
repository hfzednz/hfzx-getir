package postgres

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/tracking-service/internal/app/ports"
	"github.com/nexora/tracking-service/internal/domain"
)

func TestOpenRequiresURL(t *testing.T) {
	if _, err := Open(""); err == nil {
		t.Fatal("expected error for empty DATABASE_URL")
	}
}

func TestJSONMapRoundTrip(t *testing.T) {
	m := JSONMap{"zone": "A", "n": float64(1)}
	v, err := m.Value()
	if err != nil {
		t.Fatal(err)
	}
	var out JSONMap
	if err := out.Scan(v); err != nil {
		t.Fatal(err)
	}
	if out["zone"] != "A" {
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
	f := 12.5
	if nullFloat(nil) != nil {
		t.Fatal("nil float")
	}
	if nullFloat(&f) != f {
		t.Fatal("float value")
	}
}

func TestInterfaceCompliance(t *testing.T) {
	var (
		_ ports.LocationRepo     = (*LocationRepo)(nil)
		_ ports.TimelineRepo     = (*TimelineRepo)(nil)
		_ ports.OutboxRepository = (*OutboxRepo)(nil)
	)
	b, _ := json.Marshal(domain.CourierLocation{CourierID: uuid.New(), Lat: 1, Lon: 2})
	if len(b) == 0 {
		t.Fatal("marshal")
	}
	repos := NewRepos(nil)
	if repos.Locations == nil || repos.Timelines == nil || repos.Outbox == nil {
		t.Fatal("NewRepos")
	}
}

func TestMapUniqueViolationNil(t *testing.T) {
	if mapUniqueViolation(nil) != nil {
		t.Fatal("nil")
	}
}
