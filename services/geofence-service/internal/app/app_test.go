package app_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/geofence-service/internal/app"
	"github.com/nexora/geofence-service/internal/app/memory"
	"github.com/nexora/geofence-service/internal/domain"
)

func testDeps(t *testing.T) (*app.Deps, *memory.Store) {
	t.Helper()
	store := memory.NewStore()
	repos := memory.NewRepos(store)
	clock := &memory.Clock{T: time.Date(2026, 8, 6, 14, 0, 0, 0, time.UTC)}
	return &app.Deps{
		Zones: repos.Zones, Outbox: repos.Outbox,
		Publisher: &memory.EventPublisher{S: store},
		Clock:     clock,
		IDs:       memory.IDGen{},
	}, store
}

func tenant() uuid.UUID {
	return uuid.MustParse("11111111-1111-1111-1111-111111111111")
}

func square() []domain.Point {
	// Axis-aligned square around (0,0): lat/lng roughly [-1,1] x [-1,1]
	return []domain.Point{
		{Lat: -1, Lng: -1},
		{Lat: -1, Lng: 1},
		{Lat: 1, Lng: 1},
		{Lat: 1, Lng: -1},
	}
}

func TestPointInsideSquarePolygon(t *testing.T) {
	d, _ := testDeps(t)
	ctx := context.Background()
	z, err := d.CreateZone(ctx, app.CreateZoneInput{
		TenantID: tenant(), Name: "square", City: "Istanbul",
		Kind: domain.ZoneKindDelivery, Vertices: square(),
	})
	if err != nil {
		t.Fatal(err)
	}
	res, err := d.Contains(ctx, app.ContainsInput{
		TenantID: tenant(), ZoneID: z.ID, Point: domain.Point{Lat: 0, Lng: 0},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Inside {
		t.Fatal("expected point inside square")
	}
}

func TestPointOutsideSquarePolygon(t *testing.T) {
	d, _ := testDeps(t)
	ctx := context.Background()
	z, err := d.CreateZone(ctx, app.CreateZoneInput{
		TenantID: tenant(), Name: "square", City: "Istanbul",
		Kind: domain.ZoneKindDelivery, Vertices: square(),
	})
	if err != nil {
		t.Fatal(err)
	}
	res, err := d.Contains(ctx, app.ContainsInput{
		TenantID: tenant(), ZoneID: z.ID, Point: domain.Point{Lat: 5, Lng: 5},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Inside {
		t.Fatal("expected point outside square")
	}
}

func TestPointInRadius(t *testing.T) {
	d, _ := testDeps(t)
	ctx := context.Background()
	lat, lng, r := 41.0, 29.0, 500.0
	z, err := d.CreateZone(ctx, app.CreateZoneInput{
		TenantID: tenant(), Name: "hub", City: "Istanbul",
		Kind: domain.ZoneKindWarehouse, CenterLat: &lat, CenterLng: &lng, RadiusM: &r,
	})
	if err != nil {
		t.Fatal(err)
	}
	inside, err := d.Contains(ctx, app.ContainsInput{
		TenantID: tenant(), ZoneID: z.ID, Point: domain.Point{Lat: 41.001, Lng: 29.0},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !inside.Inside {
		t.Fatal("expected nearby point inside radius")
	}
	outside, err := d.Contains(ctx, app.ContainsInput{
		TenantID: tenant(), ZoneID: z.ID, Point: domain.Point{Lat: 42.0, Lng: 30.0},
	})
	if err != nil {
		t.Fatal(err)
	}
	if outside.Inside {
		t.Fatal("expected far point outside radius")
	}
}

func TestRestrictedBlocksServiceability(t *testing.T) {
	d, _ := testDeps(t)
	ctx := context.Background()
	_, err := d.CreateZone(ctx, app.CreateZoneInput{
		TenantID: tenant(), Name: "delivery", City: "Istanbul",
		Kind: domain.ZoneKindDelivery, Vertices: square(),
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = d.CreateZone(ctx, app.CreateZoneInput{
		TenantID: tenant(), Name: "no-fly", City: "Istanbul",
		Kind: domain.ZoneKindRestricted, Vertices: square(),
	})
	if err != nil {
		t.Fatal(err)
	}
	res, err := d.CheckServiceability(ctx, app.ServiceabilityInput{
		TenantID: tenant(), City: "Istanbul", Point: domain.Point{Lat: 0, Lng: 0},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Serviceable || !res.Blocked || res.Reason != "restricted_zone" {
		t.Fatalf("expected restricted block, got %+v", res)
	}
}

func TestServiceableWhenDeliveryOnly(t *testing.T) {
	d, _ := testDeps(t)
	ctx := context.Background()
	_, err := d.CreateZone(ctx, app.CreateZoneInput{
		TenantID: tenant(), Name: "delivery", City: "Istanbul",
		Kind: domain.ZoneKindDelivery, Vertices: square(),
	})
	if err != nil {
		t.Fatal(err)
	}
	res, err := d.CheckServiceability(ctx, app.ServiceabilityInput{
		TenantID: tenant(), City: "Istanbul", Point: domain.Point{Lat: 0, Lng: 0},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Serviceable || res.Blocked {
		t.Fatalf("expected serviceable, got %+v", res)
	}
}

func TestDeleteZoneNotFound(t *testing.T) {
	d, _ := testDeps(t)
	err := d.DeleteZone(context.Background(), tenant(), uuid.New())
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("want not found, got %v", err)
	}
}

func TestDomainRayCastingHelpers(t *testing.T) {
	verts := square()
	if !domain.PointInPolygon(domain.Point{Lat: 0, Lng: 0}, verts) {
		t.Fatal("center should be inside")
	}
	if domain.PointInPolygon(domain.Point{Lat: 10, Lng: 10}, verts) {
		t.Fatal("far point should be outside")
	}
	c := domain.Point{Lat: 0, Lng: 0}
	if !domain.PointInRadius(domain.Point{Lat: 0.001, Lng: 0}, c, 200) {
		t.Fatal("near point should be in radius")
	}
}
