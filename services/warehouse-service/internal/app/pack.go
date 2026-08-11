package app

import (
	"context"
	"fmt"
	"math"

	"github.com/google/uuid"
	"github.com/nexora/warehouse-service/internal/domain"
)

// ClaimPackCmd claims a pack task at a station.
type ClaimPackCmd struct {
	TenantID  uuid.UUID
	TaskID    uuid.UUID
	PackerID  uuid.UUID
	StationID uuid.UUID
}

// ClaimPack assigns pack task and station.
func (d *Deps) ClaimPack(ctx context.Context, in ClaimPackCmd) (domain.PackSession, error) {
	task, err := d.Tasks.GetByID(ctx, in.TenantID, in.TaskID)
	if err != nil {
		return domain.PackSession{}, err
	}
	if task.Type != domain.TaskTypePack || task.Status != domain.TaskStatusQueued {
		return domain.PackSession{}, domain.ErrTaskNotClaimable
	}

	now := d.now()
	from := task.Status
	task.Status = domain.TaskStatusClaimed
	task.AssigneeID = &in.PackerID
	task.StationID = &in.StationID
	task.ClaimedAt = &now
	task.UpdatedAt = now
	appendTaskHistory(&task, now, "claimed", &in.PackerID, from, domain.TaskStatusClaimed, "")
	if err := d.Tasks.Update(ctx, task); err != nil {
		return domain.PackSession{}, err
	}

	pack, err := d.Packs.GetByTaskID(ctx, in.TenantID, task.ID)
	if err != nil {
		return domain.PackSession{}, err
	}
	pack.Status = domain.PackSessionStatusClaimed
	pack.PackerID = &in.PackerID
	pack.StationID = &in.StationID
	pack.UpdatedAt = now
	if err := d.Packs.Update(ctx, pack); err != nil {
		return domain.PackSession{}, err
	}

	fo, _ := d.Fulfillments.GetByID(ctx, in.TenantID, task.FulfillmentID)
	fo.Status = domain.FulfillmentStatusPacking
	fo.UpdatedAt = now
	_ = d.Fulfillments.Update(ctx, fo)

	task.Status = domain.TaskStatusInProgress
	task.UpdatedAt = now
	appendTaskHistory(&task, now, "started", &in.PackerID, domain.TaskStatusClaimed, domain.TaskStatusInProgress, "")
	_ = d.Tasks.Update(ctx, task)

	d.publishEvent(ctx, domain.EventPackingStarted, fo.TenantID, fo.WarehouseID, fo.ID, map[string]any{
		"taskId": task.ID, "packSessionId": pack.ID,
	})
	return pack, nil
}

// VerifyWeightCmd checks actual pack weight within tolerance.
type VerifyWeightCmd struct {
	TenantID      uuid.UUID
	PackSessionID uuid.UUID
	ActualWeightG int64
	PackerID      uuid.UUID
}

// VerifyWeight validates weight against expected ± tolerance.
func (d *Deps) VerifyWeight(ctx context.Context, in VerifyWeightCmd) (domain.PackSession, error) {
	pack, err := d.Packs.GetByID(ctx, in.TenantID, in.PackSessionID)
	if err != nil {
		return domain.PackSession{}, err
	}
	if pack.Status != domain.PackSessionStatusClaimed && pack.Status != domain.PackSessionStatusVerified {
		return domain.PackSession{}, domain.ErrInvalidTransition
	}
	tol := pack.WeightTolerance
	if tol <= 0 {
		tol = d.weightTol()
	}
	diff := int64(math.Abs(float64(in.ActualWeightG - pack.ExpectedWeightG)))
	if diff > tol {
		return domain.PackSession{}, fmt.Errorf("%w: expected %d±%d got %d", domain.ErrWeightMismatch, pack.ExpectedWeightG, tol, in.ActualWeightG)
	}
	pack.ActualWeightG = &in.ActualWeightG
	pack.Status = domain.PackSessionStatusVerified
	pack.UpdatedAt = d.now()
	if err := d.Packs.Update(ctx, pack); err != nil {
		return domain.PackSession{}, err
	}
	return pack, nil
}

// SealPackCmd seals a weight-verified pack.
type SealPackCmd struct {
	TenantID      uuid.UUID
	PackSessionID uuid.UUID
	PackerID      uuid.UUID
}

// Seal marks the pack as sealed.
func (d *Deps) Seal(ctx context.Context, in SealPackCmd) (domain.PackSession, error) {
	pack, err := d.Packs.GetByID(ctx, in.TenantID, in.PackSessionID)
	if err != nil {
		return domain.PackSession{}, err
	}
	if pack.Status != domain.PackSessionStatusVerified {
		return domain.PackSession{}, domain.ErrInvalidTransition
	}
	now := d.now()
	pack.Status = domain.PackSessionStatusSealed
	pack.SealedAt = &now
	pack.UpdatedAt = now
	if err := d.Packs.Update(ctx, pack); err != nil {
		return domain.PackSession{}, err
	}
	return pack, nil
}

// GenerateLabelCmd creates a shipping label and dispatch unit.
type GenerateLabelCmd struct {
	TenantID      uuid.UUID
	PackSessionID uuid.UUID
	PackerID      uuid.UUID
	Format        string
}

// GenerateLabel creates label metadata, completes pack, queues dispatch.
func (d *Deps) GenerateLabel(ctx context.Context, in GenerateLabelCmd) (domain.Label, domain.DispatchUnit, error) {
	pack, err := d.Packs.GetByID(ctx, in.TenantID, in.PackSessionID)
	if err != nil {
		return domain.Label{}, domain.DispatchUnit{}, err
	}
	if pack.Status != domain.PackSessionStatusSealed {
		return domain.Label{}, domain.DispatchUnit{}, domain.ErrInvalidTransition
	}

	now := d.now()
	format := in.Format
	if format == "" {
		format = "zpl"
	}
	tracking := "NXR-" + d.newID().String()[:8]
	label := domain.Label{
		ID: d.newID(), TenantID: pack.TenantID, FulfillmentID: pack.FulfillmentID,
		PackSessionID: pack.ID, TrackingCode: tracking, Barcode: tracking,
		Format: format, PrintIntent: "print", Payload: map[string]any{"tracking": tracking},
		CreatedAt: now,
	}
	if err := d.Labels.Create(ctx, label); err != nil {
		return domain.Label{}, domain.DispatchUnit{}, err
	}

	pack.Status = domain.PackSessionStatusLabeled
	pack.LabeledAt = &now
	pack.LabelID = &label.ID
	pack.UpdatedAt = now
	_ = d.Packs.Update(ctx, pack)

	task, err := d.Tasks.GetByID(ctx, in.TenantID, pack.TaskID)
	if err != nil {
		return domain.Label{}, domain.DispatchUnit{}, err
	}
	from := task.Status
	task.Status = domain.TaskStatusCompleted
	task.CompletedAt = &now
	task.UpdatedAt = now
	appendTaskHistory(&task, now, "completed", &in.PackerID, from, domain.TaskStatusCompleted, "labeled")
	_ = d.Tasks.Update(ctx, task)

	fo, err := d.Fulfillments.GetByID(ctx, in.TenantID, pack.FulfillmentID)
	if err != nil {
		return domain.Label{}, domain.DispatchUnit{}, err
	}
	fo.Status = domain.FulfillmentStatusPacked
	fo.UpdatedAt = now
	_ = d.Fulfillments.Update(ctx, fo)

	d.publishEvent(ctx, domain.EventLabelGenerated, fo.TenantID, fo.WarehouseID, fo.ID, map[string]any{
		"labelId": label.ID, "trackingCode": tracking,
	})
	d.publishEvent(ctx, domain.EventPackingCompleted, fo.TenantID, fo.WarehouseID, fo.ID, map[string]any{
		"packSessionId": pack.ID,
	})

	dispatchTask := domain.Task{
		ID: d.newID(), TenantID: fo.TenantID, WarehouseID: fo.WarehouseID,
		FulfillmentID: fo.ID, Type: domain.TaskTypeDispatch, Status: domain.TaskStatusQueued,
		Priority: fo.Priority, CreatedAt: now, UpdatedAt: now,
	}
	appendTaskHistory(&dispatchTask, now, "created", &in.PackerID, "", domain.TaskStatusQueued, "dispatch queued")
	_ = d.Tasks.Create(ctx, dispatchTask)

	unit := domain.DispatchUnit{
		ID: d.newID(), TenantID: fo.TenantID, WarehouseID: fo.WarehouseID,
		FulfillmentID: fo.ID, TaskID: dispatchTask.ID, LabelID: &label.ID,
		TrackingCode: tracking, Status: domain.DispatchStatusQueued,
		CreatedAt: now, UpdatedAt: now,
	}
	if err := d.Dispatches.Create(ctx, unit); err != nil {
		return domain.Label{}, domain.DispatchUnit{}, err
	}

	fo.Status = domain.FulfillmentStatusDispatchQueued
	fo.UpdatedAt = d.now()
	_ = d.Fulfillments.Update(ctx, fo)

	pack.Status = domain.PackSessionStatusCompleted
	pack.UpdatedAt = d.now()
	_ = d.Packs.Update(ctx, pack)

	return label, unit, nil
}
