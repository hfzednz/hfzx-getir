package app_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/routing-service/internal/app"
	"github.com/nexora/routing-service/internal/app/memory"
	"github.com/nexora/routing-service/internal/domain"
)

type testEnv struct {
	Deps   *app.Deps
	Store  *memory.Store
	Clock  *memory.Clock
	Tenant uuid.UUID
}

func testDeps(t *testing.T) *testEnv {
	t.Helper()
	store := memory.NewStore()
	repos := memory.NewRepos(store)
	clock := &memory.Clock{T: time.Date(2026, 8, 6, 14, 0, 0, 0, time.UTC)}
	deps := &app.Deps{
		Routes: repos.Routes, Outbox: repos.Outbox,
		Maps:      app.HaversineMapsClient{},
		Traffic:   app.FixedTrafficClient{Value: 1},
		Weather:   app.FixedWeatherClient{Value: 1},
		Publisher: &memory.EventPublisher{S: store},
		Clock:     clock,
		IDs:       memory.IDGen{},
	}
	return &testEnv{
		Deps: deps, Store: store, Clock: clock,
		Tenant: uuid.MustParse("11111111-1111-1111-1111-111111111111"),
	}
}

func (e *testEnv) createMultiStop(t *testing.T) domain.Route {
	t.Helper()
	// Origin (warehouse) in Istanbul, then three stops deliberately out of order:
	// nearest to origin should be stop A, then B, then C (farthest).
	route, err := e.Deps.CreateRoute(context.Background(), app.CreateRouteInput{
		TenantID: e.Tenant,
		Waypoints: []app.WaypointInput{
			{Kind: domain.WaypointWarehouse, Lat: 41.0082, Lon: 28.9784, Label: "WH"},
			{Kind: domain.WaypointStop, Lat: 41.0500, Lon: 29.0000, Label: "C"}, // farthest north
			{Kind: domain.WaypointStop, Lat: 41.0150, Lon: 28.9850, Label: "A"}, // nearest
			{Kind: domain.WaypointStop, Lat: 41.0300, Lon: 28.9900, Label: "B"}, // middle
		},
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	return route
}

func TestOptimizeOrdersStops(t *testing.T) {
	env := testDeps(t)
	route := env.createMultiStop(t)

	opt, err := env.Deps.Optimize(context.Background(), app.OptimizeInput{
		TenantID: env.Tenant, RouteID: route.ID,
	})
	if err != nil {
		t.Fatalf("optimize: %v", err)
	}
	if opt.Status != domain.RouteStatusOptimized {
		t.Fatalf("status=%s", opt.Status)
	}
	if len(opt.Waypoints) != 4 {
		t.Fatalf("waypoints=%d", len(opt.Waypoints))
	}
	// Expect WH → A → B → C
	labels := []string{opt.Waypoints[0].Label, opt.Waypoints[1].Label, opt.Waypoints[2].Label, opt.Waypoints[3].Label}
	want := []string{"WH", "A", "B", "C"}
	for i := range want {
		if labels[i] != want[i] {
			t.Fatalf("order=%v want=%v", labels, want)
		}
		if opt.Waypoints[i].Sequence != i {
			t.Fatalf("sequence[%d]=%d", i, opt.Waypoints[i].Sequence)
		}
	}
}

func TestETAIncreasesWithTrafficFactor(t *testing.T) {
	env := testDeps(t)
	route := env.createMultiStop(t)
	baseDur := route.DurationSeconds
	if baseDur <= 0 {
		t.Fatalf("base duration=%v", baseDur)
	}

	_, err := env.Deps.UpdateTrafficHint(context.Background(), app.UpdateTrafficHintInput{
		TenantID: env.Tenant, RegionKey: "ist-center",
		Lat: 41.0082, Lon: 28.9784, RadiusM: 5000, Factor: 2.0,
	})
	if err != nil {
		t.Fatalf("hint: %v", err)
	}

	recalc, err := env.Deps.RecalculateETA(context.Background(), app.RecalculateETAInput{
		TenantID: env.Tenant, RouteID: route.ID, Reason: "traffic",
	})
	if err != nil {
		t.Fatalf("recalc: %v", err)
	}
	if recalc.TrafficFactor < 2 {
		t.Fatalf("trafficFactor=%v want >= 2", recalc.TrafficFactor)
	}
	if recalc.DurationSeconds <= baseDur {
		t.Fatalf("duration=%v should be > base=%v", recalc.DurationSeconds, baseDur)
	}
}

func TestRecalcAfterMove(t *testing.T) {
	env := testDeps(t)
	route := env.createMultiStop(t)
	baseDist := route.DistanceMeters

	// Move origin closer to first stop A — remaining distance should shrink.
	lat, lon := 41.0140, 28.9840
	moved, err := env.Deps.RecalculateETA(context.Background(), app.RecalculateETAInput{
		TenantID: env.Tenant, RouteID: route.ID,
		CurrentLat: &lat, CurrentLon: &lon, Reason: "courier_moved",
	})
	if err != nil {
		t.Fatalf("recalc: %v", err)
	}
	if moved.Waypoints[0].Lat != lat || moved.Waypoints[0].Lon != lon {
		t.Fatalf("origin not updated: %+v", moved.Waypoints[0])
	}
	if moved.DistanceMeters >= baseDist {
		t.Fatalf("distance=%v should be < base=%v after move closer", moved.DistanceMeters, baseDist)
	}
	if moved.ETAAt == nil {
		t.Fatal("ETAAt nil")
	}
}
