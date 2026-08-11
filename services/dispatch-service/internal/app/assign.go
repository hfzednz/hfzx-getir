package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/dispatch-service/internal/app/ports"
	"github.com/nexora/dispatch-service/internal/domain"
)

// CreateDispatchInput creates a queued dispatch job.
type CreateDispatchInput struct {
	TenantID        uuid.UUID
	OrderID         uuid.UUID
	FulfillmentID   uuid.UUID
	WarehouseID     uuid.UUID
	Pickup          domain.Point
	Dropoff         domain.Point
	RequiredVehicle domain.VehicleType
	City            string
}

// CreateDispatch persists a queued dispatch and emits DispatchCreated.
func (d *Deps) CreateDispatch(ctx context.Context, in CreateDispatchInput) (domain.Dispatch, error) {
	if in.TenantID == uuid.Nil || in.OrderID == uuid.Nil {
		return domain.Dispatch{}, fmt.Errorf("%w: tenant_id and order_id required", domain.ErrInvalidArgument)
	}
	if d.Geofence != nil && in.City != "" {
		ok, err := d.Geofence.CheckServiceability(ctx, in.TenantID, in.City, in.Dropoff)
		if err != nil {
			return domain.Dispatch{}, err
		}
		if !ok {
			return domain.Dispatch{}, fmt.Errorf("%w: dropoff not serviceable", domain.ErrInvariant)
		}
	}
	now := d.now()
	disp := domain.Dispatch{
		ID: d.newID(), TenantID: in.TenantID,
		OrderID: in.OrderID, FulfillmentID: in.FulfillmentID, WarehouseID: in.WarehouseID,
		Status: domain.StatusQueued, Pickup: in.Pickup, Dropoff: in.Dropoff,
		RequiredVehicle: in.RequiredVehicle, CreatedAt: now, UpdatedAt: now,
	}
	if disp.RequiredVehicle == "" {
		disp.RequiredVehicle = domain.VehicleBike
	}
	if err := disp.Validate(); err != nil {
		return domain.Dispatch{}, err
	}
	if err := d.Dispatches.Create(ctx, disp); err != nil {
		return domain.Dispatch{}, err
	}
	d.recordEvent(ctx, disp, domain.EventDispatchCreated, "", domain.StatusQueued, nil)
	d.emit(ctx, disp.TenantID, disp.ID, domain.EventDispatchCreated, map[string]any{
		"orderId": disp.OrderID.String(), "status": string(disp.Status),
	})
	return disp, nil
}

func (d *Deps) assignCourier(ctx context.Context, disp domain.Dispatch, courier domain.CourierSnapshot, strategy string, dist float64) (domain.Dispatch, error) {
	from := disp.Status
	if err := disp.Transition(domain.StatusAssigned); err != nil {
		return domain.Dispatch{}, err
	}
	cid := courier.CourierPrincipalID
	disp.CourierPrincipalID = &cid
	disp.UpdatedAt = d.now()

	if d.Routing != nil {
		res, err := d.Routing.CreateRoute(ctx, ports.RouteRequest{
			TenantID: disp.TenantID, DispatchID: disp.ID,
			Waypoints: []domain.Point{disp.Pickup, disp.Dropoff},
		})
		if err == nil {
			routeID := res.RouteID
			eta := res.ETASeconds
			disp.RouteID = &routeID
			disp.ETASeconds = &eta
		}
	}
	if err := d.Dispatches.Update(ctx, disp); err != nil {
		return domain.Dispatch{}, err
	}
	_ = d.Couriers.AdjustLoad(ctx, disp.TenantID, cid, 1)
	distCopy := dist
	_ = d.Dispatches.AppendAttempt(ctx, domain.AssignmentAttempt{
		ID: d.newID(), TenantID: disp.TenantID, DispatchID: disp.ID,
		CourierPrincipalID: &cid, Strategy: strategy, Success: true,
		DistanceM: &distCopy, CreatedAt: d.now(),
	})
	if d.Tracking != nil {
		_ = d.Tracking.SubscribeDispatch(ctx, disp.TenantID, disp.ID, cid)
	}
	eventType := domain.EventCourierAssigned
	if strategy == "reassign" {
		eventType = domain.EventCourierReassigned
	}
	d.recordEvent(ctx, disp, eventType, from, domain.StatusAssigned, map[string]any{
		"courierPrincipalId": cid.String(), "strategy": strategy,
	})
	d.emit(ctx, disp.TenantID, disp.ID, eventType, map[string]any{
		"courierPrincipalId": cid.String(), "strategy": strategy, "distanceM": dist,
	})
	return disp, nil
}

// AutoAssignInput runs nearest-available assignment.
type AutoAssignInput struct {
	TenantID   uuid.UUID
	DispatchID uuid.UUID
}

// AutoAssign picks the nearest available courier by haversine + capacity.
func (d *Deps) AutoAssign(ctx context.Context, in AutoAssignInput) (domain.Dispatch, error) {
	disp, err := d.Dispatches.Get(ctx, in.TenantID, in.DispatchID)
	if err != nil {
		return domain.Dispatch{}, err
	}
	if disp.Status != domain.StatusQueued && disp.Status != domain.StatusFailed {
		return domain.Dispatch{}, fmt.Errorf("%w: auto-assign requires queued/failed, got %s", domain.ErrInvalidTransition, disp.Status)
	}
	pool, err := d.Couriers.ListAvailable(ctx, in.TenantID)
	if err != nil {
		return domain.Dispatch{}, err
	}
	// Record capacity-full skips as failed attempts for observability.
	for _, c := range pool {
		if c.Available && c.OnShift && c.CurrentLoad >= c.MaxCapacity {
			_ = d.Dispatches.AppendAttempt(ctx, domain.AssignmentAttempt{
				ID: d.newID(), TenantID: in.TenantID, DispatchID: disp.ID,
				CourierPrincipalID: &c.CourierPrincipalID, Strategy: "nearest",
				Success: false, Reason: "capacity_full", CreatedAt: d.now(),
			})
		}
	}
	best, dist, err := domain.SelectNearestCourier(disp.Pickup, disp.RequiredVehicle, pool)
	if err != nil {
		_ = d.Dispatches.AppendAttempt(ctx, domain.AssignmentAttempt{
			ID: d.newID(), TenantID: in.TenantID, DispatchID: disp.ID,
			Strategy: "nearest", Success: false, Reason: err.Error(), CreatedAt: d.now(),
		})
		return domain.Dispatch{}, err
	}
	return d.assignCourier(ctx, disp, best, "nearest", dist)
}

// ManualAssignInput assigns a specific courier.
type ManualAssignInput struct {
	TenantID           uuid.UUID
	DispatchID         uuid.UUID
	CourierPrincipalID uuid.UUID
}

// ManualAssign assigns the given courier if capacity allows.
func (d *Deps) ManualAssign(ctx context.Context, in ManualAssignInput) (domain.Dispatch, error) {
	disp, err := d.Dispatches.Get(ctx, in.TenantID, in.DispatchID)
	if err != nil {
		return domain.Dispatch{}, err
	}
	if disp.Status != domain.StatusQueued && disp.Status != domain.StatusFailed {
		return domain.Dispatch{}, fmt.Errorf("%w: manual-assign requires queued/failed", domain.ErrInvalidTransition)
	}
	c, err := d.Couriers.Get(ctx, in.TenantID, in.CourierPrincipalID)
	if err != nil {
		return domain.Dispatch{}, err
	}
	if !c.HasCapacity() {
		return domain.Dispatch{}, domain.ErrCapacityFull
	}
	if disp.RequiredVehicle != "" && c.VehicleType != disp.RequiredVehicle {
		return domain.Dispatch{}, fmt.Errorf("%w: vehicle type mismatch", domain.ErrInvariant)
	}
	dist := domain.HaversineMeters(disp.Pickup, domain.Point{Lat: c.Lat, Lng: c.Lng})
	return d.assignCourier(ctx, disp, c, "manual", dist)
}

// ReassignInput moves a job to another courier.
type ReassignInput struct {
	TenantID              uuid.UUID
	DispatchID            uuid.UUID
	CourierPrincipalID    uuid.UUID
}

// Reassign unloads the previous courier and assigns a new one.
func (d *Deps) Reassign(ctx context.Context, in ReassignInput) (domain.Dispatch, error) {
	disp, err := d.Dispatches.Get(ctx, in.TenantID, in.DispatchID)
	if err != nil {
		return domain.Dispatch{}, err
	}
	if disp.Status != domain.StatusAssigned && disp.Status != domain.StatusPickupStarted {
		return domain.Dispatch{}, fmt.Errorf("%w: reassign requires assigned/pickup_started", domain.ErrInvalidTransition)
	}
	prev := disp.CourierPrincipalID
	c, err := d.Couriers.Get(ctx, in.TenantID, in.CourierPrincipalID)
	if err != nil {
		return domain.Dispatch{}, err
	}
	if !c.HasCapacity() {
		return domain.Dispatch{}, domain.ErrCapacityFull
	}
	if prev != nil {
		_ = d.Couriers.AdjustLoad(ctx, in.TenantID, *prev, -1)
	}
	// Force status back through assign path.
	disp.Status = domain.StatusQueued
	disp.CourierPrincipalID = nil
	dist := domain.HaversineMeters(disp.Pickup, domain.Point{Lat: c.Lat, Lng: c.Lng})
	return d.assignCourier(ctx, disp, c, "reassign", dist)
}
