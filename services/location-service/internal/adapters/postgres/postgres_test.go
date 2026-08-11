package postgres

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/location-service/internal/app/ports"
	"github.com/nexora/location-service/internal/domain"
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

func TestJSONComponentsRoundTrip(t *testing.T) {
	c := JSONComponents{Country: "TR", City: "Istanbul"}
	v, err := c.Value()
	if err != nil {
		t.Fatal(err)
	}
	var out JSONComponents
	if err := out.Scan(v); err != nil {
		t.Fatal(err)
	}
	if out.City != "Istanbul" {
		t.Fatalf("got %#v", out)
	}
}

func TestJSONGeocodeRoundTrip(t *testing.T) {
	g := JSONGeocode{PlaceID: "p1", Lat: 41.0, Lng: 29.0, Confidence: 0.9}
	v, err := g.Value()
	if err != nil {
		t.Fatal(err)
	}
	var out JSONGeocode
	if err := out.Scan(v); err != nil {
		t.Fatal(err)
	}
	if out.PlaceID != "p1" {
		t.Fatalf("got %#v", out)
	}
}

func TestNullHelpers(t *testing.T) {
	now := time.Now().UTC()
	if nullTime(nil) != nil || nullTime(&time.Time{}) != nil {
		t.Fatal("null time")
	}
	if nullTime(&now) != now {
		t.Fatal("time value")
	}
	id := uuid.New()
	if nullUUID(nil) != nil {
		t.Fatal("nil pointer")
	}
	if nullUUID(&id) != id {
		t.Fatal("uuid value")
	}
}

func TestInterfaceCompliance(t *testing.T) {
	var (
		_ ports.AddressRepo      = (*AddressRepo)(nil)
		_ ports.POIRepo          = (*POIRepo)(nil)
		_ ports.HistoryRepo      = (*HistoryRepo)(nil)
		_ ports.CacheRepo        = (*CacheRepo)(nil)
		_ ports.HeatRepo         = (*HeatRepo)(nil)
		_ ports.OutboxRepository = (*OutboxRepo)(nil)
	)
	_ = domain.POIKindWarehouse
	b, _ := json.Marshal(domain.HeatCell{GridCell: "x"})
	if len(b) == 0 {
		t.Fatal("marshal")
	}
	repos := NewRepos(nil)
	if repos.Addresses == nil || repos.Outbox == nil {
		t.Fatal("NewRepos")
	}
}

func TestMapUniqueViolationNil(t *testing.T) {
	if mapUniqueViolation(nil) != nil {
		t.Fatal("nil")
	}
}
