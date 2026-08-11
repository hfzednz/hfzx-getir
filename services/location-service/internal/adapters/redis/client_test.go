package redis

import (
	"testing"

	"github.com/google/uuid"
	"github.com/nexora/location-service/internal/domain"
)

func TestOpenRequiresURL(t *testing.T) {
	if _, err := Open("", nil); err == nil {
		t.Fatal("expected error for empty REDIS_URL")
	}
}

func TestKeyHelpers(t *testing.T) {
	tid := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	pid := uuid.MustParse("22222222-2222-2222-2222-222222222222")

	if got := geoKey(tid); got != "location:poi:geo:11111111-1111-1111-1111-111111111111" {
		t.Fatalf("geoKey=%q", got)
	}
	if got := geoKindKey(tid, domain.POIKindWarehouse); got != "location:poi:geo:11111111-1111-1111-1111-111111111111:warehouse" {
		t.Fatalf("geoKindKey=%q", got)
	}
	if got := docKey(pid); got != "location:poi:doc:22222222-2222-2222-2222-222222222222" {
		t.Fatalf("docKey=%q", got)
	}
	if got := idsKey(tid); got != "location:poi:ids:11111111-1111-1111-1111-111111111111" {
		t.Fatalf("idsKey=%q", got)
	}
}

func TestPoiDocRoundTrip(t *testing.T) {
	p := domain.POI{
		ID: uuid.New(), TenantID: uuid.New(), Kind: domain.POIKindStore,
		RefID: "r1", Name: "N", Lat: 41.0, Lng: 29.0, Active: true,
		Meta: map[string]any{"k": "v"},
	}
	d := toDoc(p)
	out := fromDoc(d)
	if out.ID != p.ID || out.Kind != p.Kind || out.Lat != p.Lat || out.Lng != p.Lng || !out.Active {
		t.Fatalf("roundtrip mismatch: %#v", out)
	}
}
