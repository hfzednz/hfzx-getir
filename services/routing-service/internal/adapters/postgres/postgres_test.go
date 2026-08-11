package postgres

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/routing-service/internal/app/ports"
	"github.com/nexora/routing-service/internal/domain"
)

func TestOpenRequiresURL(t *testing.T) {
	if _, err := Open(""); err == nil {
		t.Fatal("expected error for empty DATABASE_URL")
	}
}

func TestJSONMapRoundTrip(t *testing.T) {
	m := JSONMap{"a": "b"}
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

func TestWaypointsJSONRoundTrip(t *testing.T) {
	id := uuid.New()
	w := WaypointsJSON{{ID: id, Sequence: 0, Kind: domain.WaypointWarehouse, Lat: 41.0, Lon: 29.0, Label: "wh"}}
	v, err := w.Value()
	if err != nil {
		t.Fatal(err)
	}
	var out WaypointsJSON
	if err := out.Scan(v); err != nil {
		t.Fatal(err)
	}
	if len(out) != 1 || out[0].ID != id || out[0].Kind != domain.WaypointWarehouse {
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
}

func TestInterfaceCompliance(t *testing.T) {
	var (
		_ ports.RouteRepo        = (*RouteRepo)(nil)
		_ ports.OutboxRepository = (*OutboxRepo)(nil)
	)
	b, _ := json.Marshal(domain.Waypoint{ID: uuid.New(), Sequence: 1, Kind: domain.WaypointStop})
	if len(b) == 0 {
		t.Fatal("marshal")
	}
	repos := NewRepos(nil)
	if repos.Routes == nil || repos.Outbox == nil {
		t.Fatal("NewRepos")
	}
}

func TestMapUniqueViolationNil(t *testing.T) {
	if mapUniqueViolation(nil) != nil {
		t.Fatal("nil")
	}
}
