package app_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/tracking-service/internal/app"
	"github.com/nexora/tracking-service/internal/app/memory"
	"github.com/nexora/tracking-service/internal/domain"
)

type testEnv struct {
	Deps      *app.Deps
	Store     *memory.Store
	Clock     *memory.Clock
	Tenant    uuid.UUID
	CourierID uuid.UUID
	OrderID   uuid.UUID
}

func testDeps(t *testing.T) *testEnv {
	t.Helper()
	store := memory.NewStore()
	repos := memory.NewRepos(store)
	clock := &memory.Clock{T: time.Date(2026, 8, 6, 14, 0, 0, 0, time.UTC)}
	deps := &app.Deps{
		Locations: repos.Locations, Timelines: repos.Timelines, Outbox: repos.Outbox,
		Geofence: app.NoopGeofenceClient{},
		Publisher: &memory.EventPublisher{S: store},
		Clock: clock, IDs: memory.IDGen{},
		HistoryCap: 5, ArrivalThresh: 50,
	}
	return &testEnv{
		Deps: deps, Store: store, Clock: clock,
		Tenant:    uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		CourierID: uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc"),
		OrderID:   uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
	}
}

func TestIngestUpdatesLatest(t *testing.T) {
	env := testDeps(t)
	_, err := env.Deps.IngestLocation(context.Background(), app.IngestLocationInput{
		TenantID: env.Tenant, CourierID: env.CourierID,
		Lat: 41.01, Lon: 28.98, AccuracyM: 10,
	})
	if err != nil {
		t.Fatalf("ingest1: %v", err)
	}
	loc, err := env.Deps.IngestLocation(context.Background(), app.IngestLocationInput{
		TenantID: env.Tenant, CourierID: env.CourierID,
		Lat: 41.02, Lon: 28.99, AccuracyM: 8,
	})
	if err != nil {
		t.Fatalf("ingest2: %v", err)
	}
	if loc.Lat != 41.02 || loc.Lon != 28.99 {
		t.Fatalf("latest=%v,%v", loc.Lat, loc.Lon)
	}
	live, err := env.Deps.GetLiveCourier(context.Background(), app.GetLiveCourierInput{
		TenantID: env.Tenant, CourierID: env.CourierID,
	})
	if err != nil {
		t.Fatalf("live: %v", err)
	}
	if live.Lat != 41.02 {
		t.Fatalf("live lat=%v", live.Lat)
	}
}

func TestArrivalWhenWithinThreshold(t *testing.T) {
	env := testDeps(t)
	dropLat, dropLon := 41.0082, 28.9784
	_, err := env.Deps.IngestLocation(context.Background(), app.IngestLocationInput{
		TenantID: env.Tenant, CourierID: env.CourierID,
		Lat: dropLat + 0.0001, Lon: dropLon, AccuracyM: 5, // ~11m away
	})
	if err != nil {
		t.Fatalf("ingest: %v", err)
	}
	res, err := env.Deps.DetectArrival(context.Background(), app.DetectArrivalInput{
		TenantID: env.Tenant, CourierID: env.CourierID, OrderID: env.OrderID,
		DropoffLat: dropLat, DropoffLon: dropLon, ThresholdM: 50,
	})
	if err != nil {
		t.Fatalf("detect: %v", err)
	}
	if !res.Arrived {
		t.Fatalf("expected arrived, dist=%v", res.DistanceMeters)
	}
	if res.TimelineEventID == nil {
		t.Fatal("expected timeline event")
	}

	// Far away — not arrived
	_, err = env.Deps.IngestLocation(context.Background(), app.IngestLocationInput{
		TenantID: env.Tenant, CourierID: env.CourierID,
		Lat: 41.05, Lon: 29.05, AccuracyM: 5,
	})
	if err != nil {
		t.Fatalf("ingest far: %v", err)
	}
	res2, err := env.Deps.DetectArrival(context.Background(), app.DetectArrivalInput{
		TenantID: env.Tenant, CourierID: env.CourierID, OrderID: env.OrderID,
		DropoffLat: dropLat, DropoffLon: dropLon, ThresholdM: 50,
	})
	if err != nil {
		t.Fatalf("detect2: %v", err)
	}
	if res2.Arrived {
		t.Fatalf("should not arrive, dist=%v", res2.DistanceMeters)
	}
}

func TestTimelineAppend(t *testing.T) {
	env := testDeps(t)
	ev, err := env.Deps.AppendTimeline(context.Background(), app.AppendTimelineInput{
		TenantID: env.Tenant, OrderID: env.OrderID, CourierID: &env.CourierID,
		Type: domain.TimelineCustom, Message: "picked up",
	})
	if err != nil {
		t.Fatalf("append: %v", err)
	}
	if ev.ID == uuid.Nil {
		t.Fatal("empty id")
	}
	list, err := env.Deps.GetOrderTimeline(context.Background(), app.GetOrderTimelineInput{
		TenantID: env.Tenant, OrderID: env.OrderID,
	})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 1 || list[0].Message != "picked up" {
		t.Fatalf("timeline=%+v", list)
	}
}

func TestHistoryCapped(t *testing.T) {
	env := testDeps(t)
	for i := 0; i < 8; i++ {
		_, err := env.Deps.IngestLocation(context.Background(), app.IngestLocationInput{
			TenantID: env.Tenant, CourierID: env.CourierID,
			Lat: 41.0 + float64(i)*0.001, Lon: 28.98, AccuracyM: 5,
		})
		if err != nil {
			t.Fatalf("ingest %d: %v", i, err)
		}
	}
	n := env.Store.HistoryLen(env.Tenant, env.CourierID)
	if n != 5 {
		t.Fatalf("history len=%d want 5", n)
	}
}
