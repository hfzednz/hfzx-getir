package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/dispatch-service/internal/domain"
)

func (d *Deps) transition(ctx context.Context, tenantID, id uuid.UUID, to domain.DispatchStatus, eventType string, extra map[string]any) (domain.Dispatch, error) {
	disp, err := d.Dispatches.Get(ctx, tenantID, id)
	if err != nil {
		return domain.Dispatch{}, err
	}
	from := disp.Status
	if err := disp.Transition(to); err != nil {
		return domain.Dispatch{}, err
	}
	disp.UpdatedAt = d.now()
	if err := d.Dispatches.Update(ctx, disp); err != nil {
		return domain.Dispatch{}, err
	}
	d.recordEvent(ctx, disp, eventType, from, to, extra)
	d.emit(ctx, disp.TenantID, disp.ID, eventType, extra)
	return disp, nil
}

// StartPickup moves assigned → pickup_started.
func (d *Deps) StartPickup(ctx context.Context, tenantID, id uuid.UUID) (domain.Dispatch, error) {
	return d.transition(ctx, tenantID, id, domain.StatusPickupStarted, domain.EventPickupStarted, nil)
}

// CompletePickup moves pickup_started → picked_up.
func (d *Deps) CompletePickup(ctx context.Context, tenantID, id uuid.UUID) (domain.Dispatch, error) {
	return d.transition(ctx, tenantID, id, domain.StatusPickedUp, domain.EventPickupCompleted, nil)
}

// StartTransit moves picked_up → in_transit.
func (d *Deps) StartTransit(ctx context.Context, tenantID, id uuid.UUID) (domain.Dispatch, error) {
	return d.transition(ctx, tenantID, id, domain.StatusInTransit, domain.EventDeliveryStarted, nil)
}

// Arrive moves in_transit → arrived.
func (d *Deps) Arrive(ctx context.Context, tenantID, id uuid.UUID) (domain.Dispatch, error) {
	return d.transition(ctx, tenantID, id, domain.StatusArrived, domain.EventDeliveryArrived, nil)
}

// CompleteDeliveryInput records POD and completes delivery.
type CompleteDeliveryInput struct {
	TenantID    uuid.UUID
	DispatchID  uuid.UUID
	PODType     domain.PODType
	PODReference string
}

// CompleteDelivery moves arrived → delivered with POD.
func (d *Deps) CompleteDelivery(ctx context.Context, in CompleteDeliveryInput) (domain.Dispatch, error) {
	if !in.PODType.Valid() {
		return domain.Dispatch{}, fmt.Errorf("%w: invalid pod type %q", domain.ErrInvalidArgument, in.PODType)
	}
	if in.PODReference == "" {
		return domain.Dispatch{}, fmt.Errorf("%w: pod reference required", domain.ErrInvalidArgument)
	}
	disp, err := d.Dispatches.Get(ctx, in.TenantID, in.DispatchID)
	if err != nil {
		return domain.Dispatch{}, err
	}
	from := disp.Status
	if err := disp.Transition(domain.StatusDelivered); err != nil {
		return domain.Dispatch{}, err
	}
	pod := in.PODType
	disp.PODType = &pod
	disp.PODReference = in.PODReference
	disp.UpdatedAt = d.now()
	if err := d.Dispatches.Update(ctx, disp); err != nil {
		return domain.Dispatch{}, err
	}
	if disp.CourierPrincipalID != nil {
		_ = d.Couriers.AdjustLoad(ctx, disp.TenantID, *disp.CourierPrincipalID, -1)
	}
	extra := map[string]any{"podType": string(in.PODType), "podReference": in.PODReference}
	d.recordEvent(ctx, disp, domain.EventDeliveryCompleted, from, domain.StatusDelivered, extra)
	d.emit(ctx, disp.TenantID, disp.ID, domain.EventDeliveryCompleted, extra)
	return disp, nil
}

// FailDeliveryInput fails a delivery with optional requeue.
type FailDeliveryInput struct {
	TenantID   uuid.UUID
	DispatchID uuid.UUID
	Reason     domain.FailReason
	Note       string
	Requeue    bool
}

// FailDelivery moves to failed; optionally requeues for reassignment.
func (d *Deps) FailDelivery(ctx context.Context, in FailDeliveryInput) (domain.Dispatch, error) {
	if !in.Reason.Valid() {
		return domain.Dispatch{}, fmt.Errorf("%w: invalid fail reason %q", domain.ErrInvalidArgument, in.Reason)
	}
	disp, err := d.Dispatches.Get(ctx, in.TenantID, in.DispatchID)
	if err != nil {
		return domain.Dispatch{}, err
	}
	from := disp.Status
	if err := disp.Transition(domain.StatusFailed); err != nil {
		return domain.Dispatch{}, err
	}
	reason := in.Reason
	disp.FailReason = &reason
	disp.FailNote = in.Note
	if disp.CourierPrincipalID != nil {
		_ = d.Couriers.AdjustLoad(ctx, disp.TenantID, *disp.CourierPrincipalID, -1)
		disp.CourierPrincipalID = nil
	}
	disp.UpdatedAt = d.now()
	if err := d.Dispatches.Update(ctx, disp); err != nil {
		return domain.Dispatch{}, err
	}
	extra := map[string]any{"reason": string(in.Reason), "note": in.Note, "requeue": in.Requeue}
	d.recordEvent(ctx, disp, domain.EventDeliveryFailed, from, domain.StatusFailed, extra)
	d.emit(ctx, disp.TenantID, disp.ID, domain.EventDeliveryFailed, extra)

	if in.Requeue {
		if err := disp.Transition(domain.StatusQueued); err != nil {
			return domain.Dispatch{}, err
		}
		disp.FailReason = nil
		disp.FailNote = ""
		disp.UpdatedAt = d.now()
		if err := d.Dispatches.Update(ctx, disp); err != nil {
			return domain.Dispatch{}, err
		}
		d.recordEvent(ctx, disp, "DispatchRequeued", domain.StatusFailed, domain.StatusQueued, nil)
	}
	return disp, nil
}

// BatchCreateInput creates a batch of new dispatches.
type BatchCreateInput struct {
	TenantID uuid.UUID
	Label    string
	Items    []CreateDispatchInput
}

// BatchCreate creates multiple queued dispatches under one batch.
func (d *Deps) BatchCreate(ctx context.Context, in BatchCreateInput) (domain.Batch, []domain.Dispatch, error) {
	if len(in.Items) == 0 {
		return domain.Batch{}, nil, fmt.Errorf("%w: batch items required", domain.ErrInvalidArgument)
	}
	ids := make([]uuid.UUID, 0, len(in.Items))
	created := make([]domain.Dispatch, 0, len(in.Items))
	batchID := d.newID()
	for _, item := range in.Items {
		item.TenantID = in.TenantID
		disp, err := d.CreateDispatch(ctx, item)
		if err != nil {
			return domain.Batch{}, nil, err
		}
		disp.BatchID = &batchID
		disp.UpdatedAt = d.now()
		if err := d.Dispatches.Update(ctx, disp); err != nil {
			return domain.Batch{}, nil, err
		}
		ids = append(ids, disp.ID)
		created = append(created, disp)
	}
	batch := domain.Batch{
		ID: batchID, TenantID: in.TenantID, Label: in.Label,
		DispatchIDs: ids, CreatedAt: d.now(),
	}
	if err := d.Dispatches.CreateBatch(ctx, batch); err != nil {
		return domain.Batch{}, nil, err
	}
	return batch, created, nil
}

// UpsertVehicleInput upserts a fleet vehicle.
type UpsertVehicleInput struct {
	TenantID uuid.UUID
	ID       *uuid.UUID
	Plate    string
	Type     domain.VehicleType
	Capacity int
	Active   *bool
}

// UpsertVehicle creates or updates a fleet vehicle.
func (d *Deps) UpsertVehicle(ctx context.Context, in UpsertVehicleInput) (domain.Vehicle, error) {
	now := d.now()
	active := true
	if in.Active != nil {
		active = *in.Active
	}
	var v domain.Vehicle
	if in.ID != nil {
		existing, err := d.Vehicles.Get(ctx, in.TenantID, *in.ID)
		if err != nil {
			return domain.Vehicle{}, err
		}
		v = existing
		v.Plate = in.Plate
		v.Type = in.Type
		v.Capacity = in.Capacity
		v.Active = active
		v.UpdatedAt = now
	} else {
		v = domain.Vehicle{
			ID: d.newID(), TenantID: in.TenantID, Plate: in.Plate,
			Type: in.Type, Capacity: in.Capacity, Active: active,
			CreatedAt: now, UpdatedAt: now,
		}
	}
	if err := v.Validate(); err != nil {
		return domain.Vehicle{}, err
	}
	if err := d.Vehicles.Upsert(ctx, v); err != nil {
		return domain.Vehicle{}, err
	}
	return v, nil
}

// UpsertCourierInput upserts a courier availability snapshot.
type UpsertCourierInput struct {
	TenantID           uuid.UUID
	CourierPrincipalID uuid.UUID
	Available          bool
	Lat                float64
	Lng                float64
	CurrentLoad        int
	MaxCapacity        int
	Rating             float64
	VehicleType        domain.VehicleType
	OnShift            bool
}

// UpsertCourierSnapshot updates courier pool state.
func (d *Deps) UpsertCourierSnapshot(ctx context.Context, in UpsertCourierInput) (domain.CourierSnapshot, error) {
	if in.CourierPrincipalID == uuid.Nil {
		return domain.CourierSnapshot{}, fmt.Errorf("%w: courier_principal_id required", domain.ErrInvalidArgument)
	}
	c := domain.CourierSnapshot{
		ID: d.newID(), TenantID: in.TenantID, CourierPrincipalID: in.CourierPrincipalID,
		Available: in.Available, Lat: in.Lat, Lng: in.Lng,
		CurrentLoad: in.CurrentLoad, MaxCapacity: in.MaxCapacity,
		Rating: in.Rating, VehicleType: in.VehicleType, OnShift: in.OnShift,
		UpdatedAt: d.now(),
	}
	if existing, err := d.Couriers.Get(ctx, in.TenantID, in.CourierPrincipalID); err == nil {
		c.ID = existing.ID
	}
	if err := d.Couriers.Upsert(ctx, c); err != nil {
		return domain.CourierSnapshot{}, err
	}
	return c, nil
}

// AdminListInput lists dispatches for admin.
type AdminListInput struct {
	TenantID uuid.UUID
	Status   domain.DispatchStatus
	Limit    int
	Offset   int
}

// AdminList returns paginated dispatches.
func (d *Deps) AdminList(ctx context.Context, in AdminListInput) ([]domain.Dispatch, int, error) {
	if in.Limit <= 0 {
		in.Limit = 50
	}
	return d.Dispatches.List(ctx, in.TenantID, in.Status, in.Limit, in.Offset)
}

// GetDispatch returns a dispatch by id.
func (d *Deps) GetDispatch(ctx context.Context, tenantID, id uuid.UUID) (domain.Dispatch, error) {
	return d.Dispatches.Get(ctx, tenantID, id)
}

// ListVehicles returns paginated vehicles.
func (d *Deps) ListVehicles(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]domain.Vehicle, int, error) {
	if limit <= 0 {
		limit = 50
	}
	return d.Vehicles.List(ctx, tenantID, limit, offset)
}
