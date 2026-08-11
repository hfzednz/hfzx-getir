package app_test

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/location-service/internal/app"
	"github.com/nexora/location-service/internal/app/memory"
	"github.com/nexora/location-service/internal/domain"
)

type testEnv struct {
	Deps     *app.Deps
	Store    *memory.Store
	Maps     *app.MockMapsProvider
	Geofence *app.MemoryGeofenceClient
	Routing  *app.MemoryRoutingClient
	Tenant   uuid.UUID
}

func testDeps(t *testing.T) *testEnv {
	t.Helper()
	store := memory.NewStore()
	repos := memory.NewRepos(store)
	clock := &memory.Clock{T: time.Now().UTC()}
	maps := &app.MockMapsProvider{}
	geo := app.NewMemoryGeofenceClient()
	routing := &app.MemoryRoutingClient{}
	deps := &app.Deps{
		Addresses: repos.Addresses,
		POIs:      repos.POIs,
		History:   repos.History,
		Cache:     repos.Cache,
		Heat:      repos.Heat,
		Outbox:    repos.Outbox,
		Maps:      maps,
		Geofence:  geo,
		Routing:   routing,
		Publisher: &memory.EventPublisher{S: store},
		Clock:     clock,
		IDs:       memory.IDGen{},
	}
	return &testEnv{
		Deps: deps, Store: store, Maps: maps, Geofence: geo, Routing: routing,
		Tenant: uuid.MustParse("11111111-1111-1111-1111-111111111111"),
	}
}

func TestGeocodeCachedOnSecondCall(t *testing.T) {
	env := testDeps(t)
	ctx := context.Background()
	in := app.ForwardGeocodeInput{TenantID: env.Tenant, Query: "Taksim Square"}

	r1, err := env.Deps.ForwardGeocode(ctx, in)
	if err != nil {
		t.Fatalf("first: %v", err)
	}
	if r1.Cached {
		t.Fatal("first call should not be cached")
	}
	if env.Maps.GeocodeCalls != 1 {
		t.Fatalf("provider calls=%d", env.Maps.GeocodeCalls)
	}

	r2, err := env.Deps.ForwardGeocode(ctx, in)
	if err != nil {
		t.Fatalf("second: %v", err)
	}
	if !r2.Cached {
		t.Fatal("second call should be cached")
	}
	if env.Maps.GeocodeCalls != 1 {
		t.Fatalf("provider should not be called again, calls=%d", env.Maps.GeocodeCalls)
	}
	if r2.PlaceID != r1.PlaceID {
		t.Fatalf("place mismatch %q vs %q", r2.PlaceID, r1.PlaceID)
	}
}

func TestReverseReturnsAddress(t *testing.T) {
	env := testDeps(t)
	res, err := env.Deps.ReverseGeocode(context.Background(), app.ReverseGeocodeInput{
		TenantID: env.Tenant, Lat: 41.0082, Lng: 28.9784,
	})
	if err != nil {
		t.Fatalf("reverse: %v", err)
	}
	if res.Formatted == "" {
		t.Fatal("expected formatted address")
	}
	if res.Components.City == "" {
		t.Fatal("expected city component")
	}
	if env.Maps.ReverseCalls != 1 {
		t.Fatalf("reverse calls=%d", env.Maps.ReverseCalls)
	}
}

func TestNearbyFindsPOIWithinRadiusExcludesFar(t *testing.T) {
	env := testDeps(t)
	ctx := context.Background()
	near, err := env.Deps.UpsertPOI(ctx, app.UpsertPOIInput{
		TenantID: env.Tenant, Kind: domain.POIKindStore, RefID: "near-1",
		Name: "Near", Lat: 41.0082, Lng: 28.9784,
	})
	if err != nil {
		t.Fatalf("upsert near: %v", err)
	}
	_, err = env.Deps.UpsertPOI(ctx, app.UpsertPOIInput{
		TenantID: env.Tenant, Kind: domain.POIKindStore, RefID: "far-1",
		Name: "Far", Lat: 41.1000, Lng: 29.1000,
	})
	if err != nil {
		t.Fatalf("upsert far: %v", err)
	}

	hits, err := env.Deps.Nearby(ctx, app.NearbyInput{
		TenantID: env.Tenant, Lat: 41.0082, Lng: 28.9784, RadiusM: 500, Limit: 10,
	})
	if err != nil {
		t.Fatalf("nearby: %v", err)
	}
	if len(hits) != 1 || hits[0].ID != near.ID {
		t.Fatalf("hits=%v want only near", hits)
	}
}

func TestNearestWarehouse(t *testing.T) {
	env := testDeps(t)
	ctx := context.Background()
	_, err := env.Deps.UpsertPOI(ctx, app.UpsertPOIInput{
		TenantID: env.Tenant, Kind: domain.POIKindWarehouse, RefID: "wh-far",
		Name: "Far WH", Lat: 41.05, Lng: 29.05,
	})
	if err != nil {
		t.Fatalf("far: %v", err)
	}
	closeWH, err := env.Deps.UpsertPOI(ctx, app.UpsertPOIInput{
		TenantID: env.Tenant, Kind: domain.POIKindWarehouse, RefID: "wh-near",
		Name: "Near WH", Lat: 41.0090, Lng: 28.9790,
	})
	if err != nil {
		t.Fatalf("near: %v", err)
	}
	_, err = env.Deps.UpsertPOI(ctx, app.UpsertPOIInput{
		TenantID: env.Tenant, Kind: domain.POIKindStore, RefID: "store-1",
		Name: "Store", Lat: 41.0083, Lng: 28.9785,
	})
	if err != nil {
		t.Fatalf("store: %v", err)
	}

	hits, err := env.Deps.NearestOfKind(ctx, app.NearestOfKindInput{
		TenantID: env.Tenant, Kind: domain.POIKindWarehouse,
		Lat: 41.0082, Lng: 28.9784, Limit: 1,
	})
	if err != nil {
		t.Fatalf("nearest: %v", err)
	}
	if len(hits) != 1 || hits[0].ID != closeWH.ID {
		t.Fatalf("got=%v want near warehouse", hits)
	}
}

func TestValidateAddressCallsGeofenceServiceability(t *testing.T) {
	env := testDeps(t)
	env.Geofence.Serviceable = true
	env.Geofence.ZoneID = uuid.MustParse("22222222-2222-2222-2222-222222222222")

	res, err := env.Deps.ValidateAddress(context.Background(), app.ValidateAddressInput{
		TenantID: env.Tenant, Line1: "Istiklal Cad.", Lat: 41.034, Lng: 28.985,
	})
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	if env.Geofence.Calls != 1 {
		t.Fatalf("geofence calls=%d", env.Geofence.Calls)
	}
	if !res.Feasibility.Serviceable {
		t.Fatal("expected serviceable")
	}
	if math.Abs(env.Geofence.LastLat-41.034) > 1e-9 || math.Abs(env.Geofence.LastLng-28.985) > 1e-9 {
		t.Fatalf("geofence coords lat=%v lng=%v", env.Geofence.LastLat, env.Geofence.LastLng)
	}
}

func TestHaversineKnownDistanceApprox(t *testing.T) {
	// ~1 degree latitude ≈ 111.19 km
	a := domain.LatLng{Lat: 41.0, Lng: 29.0}
	b := domain.LatLng{Lat: 42.0, Lng: 29.0}
	d := domain.HaversineDistanceMeters(a, b)
	if d < 110000 || d > 112500 {
		t.Fatalf("distance=%v want ~111km", d)
	}
}

func TestHistoryCapped(t *testing.T) {
	env := testDeps(t)
	ctx := context.Background()
	subj := "device-1"
	for i := 0; i < domain.MaxHistoryPerSubject+25; i++ {
		_, err := env.Deps.IngestHistory(ctx, app.IngestHistoryInput{
			TenantID: env.Tenant, SubjectType: domain.SubjectDevice, SubjectID: subj,
			Lat: 41.0 + float64(i)*0.0001, Lng: 29.0,
		})
		if err != nil {
			t.Fatalf("ingest %d: %v", i, err)
		}
	}
	rows, err := env.Deps.GetHistory(ctx, app.GetHistoryInput{
		TenantID: env.Tenant, SubjectType: domain.SubjectDevice, SubjectID: subj, Limit: 1000,
	})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(rows) != domain.MaxHistoryPerSubject {
		t.Fatalf("len=%d want %d", len(rows), domain.MaxHistoryPerSubject)
	}
}

func TestHeatmapReturnsCells(t *testing.T) {
	env := testDeps(t)
	ctx := context.Background()
	_, err := env.Deps.UpsertHeatCell(ctx, app.UpsertHeatCellInput{
		TenantID: env.Tenant, GridCell: "grid:41.0,29.0", DemandScore: 12, Density: 3.5,
	})
	if err != nil {
		t.Fatalf("upsert: %v", err)
	}
	_, err = env.Deps.UpsertHeatCell(ctx, app.UpsertHeatCellInput{
		TenantID: env.Tenant, GridCell: "grid:41.1,29.1", DemandScore: 8, Density: 2,
	})
	if err != nil {
		t.Fatalf("upsert2: %v", err)
	}
	cells, err := env.Deps.DemandHeatmap(ctx, app.DemandHeatmapInput{TenantID: env.Tenant, Limit: 10})
	if err != nil {
		t.Fatalf("heatmap: %v", err)
	}
	if len(cells) != 2 {
		t.Fatalf("cells=%d", len(cells))
	}
}

func TestProxyRouteCallsRoutingClient(t *testing.T) {
	env := testDeps(t)
	res, err := env.Deps.ProxyRoute(context.Background(), app.ProxyRouteInput{
		TenantID: env.Tenant,
		Origin:   domain.LatLng{Lat: 41.0082, Lng: 28.9784},
		Dest:     domain.LatLng{Lat: 41.035, Lng: 28.985},
	})
	if err != nil {
		t.Fatalf("proxy: %v", err)
	}
	if env.Routing.CreateRouteCalls != 1 {
		t.Fatalf("routing calls=%d", env.Routing.CreateRouteCalls)
	}
	if res.RouteID == "" || res.DistanceMeters <= 0 {
		t.Fatalf("result=%+v", res)
	}
}
