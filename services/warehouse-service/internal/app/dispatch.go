package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/nexora/warehouse-service/internal/app/ports"
	"github.com/nexora/warehouse-service/internal/domain"
)

// DispatchVerifyCmd verifies a packed unit before handoff.
type DispatchVerifyCmd struct {
	TenantID       uuid.UUID
	DispatchUnitID uuid.UUID
	TrackingCode   string
	ActorID        *uuid.UUID
}

// DispatchVerify confirms package identity against label.
func (d *Deps) DispatchVerify(ctx context.Context, in DispatchVerifyCmd) (domain.DispatchUnit, error) {
	unit, err := d.Dispatches.GetByID(ctx, in.TenantID, in.DispatchUnitID)
	if err != nil {
		return domain.DispatchUnit{}, err
	}
	if unit.Status != domain.DispatchStatusQueued {
		return domain.DispatchUnit{}, domain.ErrInvalidTransition
	}
	if in.TrackingCode != "" && in.TrackingCode != unit.TrackingCode {
		return domain.DispatchUnit{}, domain.ErrBarcodeMismatch
	}
	now := d.now()
	unit.Status = domain.DispatchStatusVerified
	unit.VerifiedAt = &now
	unit.UpdatedAt = now
	if err := d.Dispatches.Update(ctx, unit); err != nil {
		return domain.DispatchUnit{}, err
	}

	d.publishEvent(ctx, domain.EventDispatchStarted, unit.TenantID, unit.WarehouseID, unit.FulfillmentID, map[string]any{
		"dispatchUnitId": unit.ID,
	})
	return unit, nil
}

// HandoffConfirmCmd confirms courier handoff and consumes inventory.
type HandoffConfirmCmd struct {
	TenantID       uuid.UUID
	DispatchUnitID uuid.UUID
	CourierRef     string
	ActorID        *uuid.UUID
}

// HandoffConfirm completes dispatch and calls Inventory.Consume.
func (d *Deps) HandoffConfirm(ctx context.Context, in HandoffConfirmCmd) (domain.DispatchUnit, error) {
	unit, err := d.Dispatches.GetByID(ctx, in.TenantID, in.DispatchUnitID)
	if err != nil {
		return domain.DispatchUnit{}, err
	}
	if unit.Status != domain.DispatchStatusVerified && unit.Status != domain.DispatchStatusQueued {
		return domain.DispatchUnit{}, domain.ErrInvalidTransition
	}

	fo, err := d.Fulfillments.GetByID(ctx, in.TenantID, unit.FulfillmentID)
	if err != nil {
		return domain.DispatchUnit{}, err
	}
	if fo.ReservationID != nil && d.Inventory != nil {
		if err := d.Inventory.Consume(ctx, ports.ConsumeRequest{
			TenantID: in.TenantID, ReservationID: *fo.ReservationID,
			IdempotencyKey: "consume:" + fo.ID.String(),
		}); err != nil {
			return domain.DispatchUnit{}, err
		}
	}

	now := d.now()
	unit.Status = domain.DispatchStatusHandedOff
	unit.CourierRef = in.CourierRef
	unit.HandedOffAt = &now
	unit.UpdatedAt = now
	if err := d.Dispatches.Update(ctx, unit); err != nil {
		return domain.DispatchUnit{}, err
	}

	task, err := d.Tasks.GetByID(ctx, in.TenantID, unit.TaskID)
	if err == nil {
		from := task.Status
		task.Status = domain.TaskStatusCompleted
		task.CompletedAt = &now
		task.UpdatedAt = now
		appendTaskHistory(&task, now, "handoff", in.ActorID, from, domain.TaskStatusCompleted, in.CourierRef)
		_ = d.Tasks.Update(ctx, task)
	}

	fo.Status = domain.FulfillmentStatusDispatched
	fo.CourierRef = in.CourierRef
	fo.UpdatedAt = now
	_ = d.Fulfillments.Update(ctx, fo)

	d.publishEvent(ctx, domain.EventDispatchCompleted, fo.TenantID, fo.WarehouseID, fo.ID, map[string]any{
		"dispatchUnitId": unit.ID, "courierRef": in.CourierRef,
	})
	d.publishEvent(ctx, domain.EventCourierAssigned, fo.TenantID, fo.WarehouseID, fo.ID, map[string]any{
		"courierRef": in.CourierRef,
	})
	return unit, nil
}

// ListDispatchQueue returns queued dispatch units.
func (d *Deps) ListDispatchQueue(ctx context.Context, tenantID, warehouseID uuid.UUID, limit, offset int) ([]domain.DispatchUnit, int, error) {
	return d.Dispatches.ListQueued(ctx, tenantID, warehouseID, limit, offset)
}
