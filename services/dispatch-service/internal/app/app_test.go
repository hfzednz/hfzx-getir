package app_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/dispatch-service/internal/app"
	"github.com/nexora/dispatch-service/internal/app/memory"
	"github.com/nexora/dispatch-service/internal/domain"
)

func testDeps(t *testing.T) (*app.Deps, *memory.Store) {
	t.Helper()
	store := memory.NewStore()
	repos := memory.NewRepos(store)
	clock := &memory.Clock{T: time.Date(2026, 8, 6, 14, 0, 0, 0, time.UTC)}
	return &app.Deps{
		Dispatches: repos.Dispatches,
		Couriers:   repos.Couriers,
		Vehicles:   repos.Vehicles,
		Outbox:     repos.Outbox,
		Routing:    &memory.RoutingClient{ETA: 600},
		Tracking:   &memory.TrackingClient{},
		Geofence:   &memory.GeofenceClient{OK: true},
		Publisher:  &memory.EventPublisher{S: store},
		Clock:      clock,
		IDs:        memory.IDGen{},
	}, store
}

func tenant() uuid.UUID {
	return uuid.MustParse("11111111-1111-1111-1111-111111111111")
}

func seedCourier(t *testing.T, d *app.Deps, id uuid.UUID, lat, lng float64, load, cap int) {
	t.Helper()
	_, err := d.UpsertCourierSnapshot(context.Background(), app.UpsertCourierInput{
		TenantID: tenant(), CourierPrincipalID: id,
		Available: true, Lat: lat, Lng: lng,
		CurrentLoad: load, MaxCapacity: cap, Rating: 4.8,
		VehicleType: domain.VehicleBike, OnShift: true,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func createQueued(t *testing.T, d *app.Deps, pickup domain.Point) domain.Dispatch {
	t.Helper()
	disp, err := d.CreateDispatch(context.Background(), app.CreateDispatchInput{
		TenantID: tenant(), OrderID: uuid.New(), FulfillmentID: uuid.New(),
		WarehouseID: uuid.New(), Pickup: pickup, Dropoff: domain.Point{Lat: 41.1, Lng: 29.1},
		RequiredVehicle: domain.VehicleBike, City: "Istanbul",
	})
	if err != nil {
		t.Fatal(err)
	}
	return disp
}

func TestAutoAssignPicksNearestAvailable(t *testing.T) {
	d, _ := testDeps(t)
	near := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	far := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	seedCourier(t, d, near, 41.010, 29.000, 0, 3)
	seedCourier(t, d, far, 41.200, 29.200, 0, 3)

	disp := createQueued(t, d, domain.Point{Lat: 41.011, Lng: 29.001})
	out, err := d.AutoAssign(context.Background(), app.AutoAssignInput{
		TenantID: tenant(), DispatchID: disp.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if out.Status != domain.StatusAssigned {
		t.Fatalf("status=%s", out.Status)
	}
	if out.CourierPrincipalID == nil || *out.CourierPrincipalID != near {
		t.Fatalf("expected nearest courier %s, got %v", near, out.CourierPrincipalID)
	}
}

func TestAutoAssignSkipsCapacityFull(t *testing.T) {
	d, store := testDeps(t)
	full := uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc")
	free := uuid.MustParse("dddddddd-dddd-dddd-dddd-dddddddddddd")
	seedCourier(t, d, full, 41.010, 29.000, 2, 2) // nearest but full
	seedCourier(t, d, free, 41.050, 29.050, 0, 2)

	disp := createQueued(t, d, domain.Point{Lat: 41.011, Lng: 29.001})
	out, err := d.AutoAssign(context.Background(), app.AutoAssignInput{
		TenantID: tenant(), DispatchID: disp.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if out.CourierPrincipalID == nil || *out.CourierPrincipalID != free {
		t.Fatalf("expected free courier, got %v", out.CourierPrincipalID)
	}
	foundSkip := false
	for _, a := range store.Attempts() {
		if a.Reason == "capacity_full" {
			foundSkip = true
			break
		}
	}
	if !foundSkip {
		t.Fatal("expected capacity_full attempt recorded")
	}
}

func TestReassign(t *testing.T) {
	d, _ := testDeps(t)
	a := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	b := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	seedCourier(t, d, a, 41.01, 29.00, 0, 3)
	seedCourier(t, d, b, 41.02, 29.02, 0, 3)
	disp := createQueued(t, d, domain.Point{Lat: 41.01, Lng: 29.00})
	assigned, err := d.ManualAssign(context.Background(), app.ManualAssignInput{
		TenantID: tenant(), DispatchID: disp.ID, CourierPrincipalID: a,
	})
	if err != nil {
		t.Fatal(err)
	}
	out, err := d.Reassign(context.Background(), app.ReassignInput{
		TenantID: tenant(), DispatchID: assigned.ID, CourierPrincipalID: b,
	})
	if err != nil {
		t.Fatal(err)
	}
	if out.CourierPrincipalID == nil || *out.CourierPrincipalID != b {
		t.Fatalf("expected courier b, got %v", out.CourierPrincipalID)
	}
}

func TestPODComplete(t *testing.T) {
	d, _ := testDeps(t)
	c := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	seedCourier(t, d, c, 41.01, 29.00, 0, 3)
	disp := createQueued(t, d, domain.Point{Lat: 41.01, Lng: 29.00})
	ctx := context.Background()
	disp, _ = d.ManualAssign(ctx, app.ManualAssignInput{TenantID: tenant(), DispatchID: disp.ID, CourierPrincipalID: c})
	disp, _ = d.StartPickup(ctx, tenant(), disp.ID)
	disp, _ = d.CompletePickup(ctx, tenant(), disp.ID)
	disp, _ = d.StartTransit(ctx, tenant(), disp.ID)
	disp, _ = d.Arrive(ctx, tenant(), disp.ID)
	out, err := d.CompleteDelivery(ctx, app.CompleteDeliveryInput{
		TenantID: tenant(), DispatchID: disp.ID, PODType: domain.PODOTP, PODReference: "1234",
	})
	if err != nil {
		t.Fatal(err)
	}
	if out.Status != domain.StatusDelivered || out.PODType == nil || *out.PODType != domain.PODOTP {
		t.Fatalf("unexpected %+v", out)
	}
}

func TestFailRequeueOptional(t *testing.T) {
	d, _ := testDeps(t)
	c := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	seedCourier(t, d, c, 41.01, 29.00, 0, 3)
	disp := createQueued(t, d, domain.Point{Lat: 41.01, Lng: 29.00})
	ctx := context.Background()
	disp, _ = d.ManualAssign(ctx, app.ManualAssignInput{TenantID: tenant(), DispatchID: disp.ID, CourierPrincipalID: c})
	out, err := d.FailDelivery(ctx, app.FailDeliveryInput{
		TenantID: tenant(), DispatchID: disp.ID,
		Reason: domain.FailCustomerUnavailable, Note: "not home", Requeue: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if out.Status != domain.StatusQueued {
		t.Fatalf("expected requeued, got %s", out.Status)
	}
	if out.CourierPrincipalID != nil {
		t.Fatal("courier should be cleared")
	}
}

func TestIllegalTransition(t *testing.T) {
	d, _ := testDeps(t)
	disp := createQueued(t, d, domain.Point{Lat: 41.01, Lng: 29.00})
	_, err := d.StartTransit(context.Background(), tenant(), disp.ID)
	if !errors.Is(err, domain.ErrInvalidTransition) {
		t.Fatalf("want invalid transition, got %v", err)
	}
}
