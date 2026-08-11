package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/warehouse-service/internal/app/ports"
	"github.com/nexora/warehouse-service/internal/domain"
)

// ClaimPickTaskCmd claims the next or specific pick task.
type ClaimPickTaskCmd struct {
	TenantID    uuid.UUID
	WarehouseID uuid.UUID
	TaskID      *uuid.UUID
	PickerID    uuid.UUID
}

// ClaimPickTask assigns a queued pick task to a picker.
func (d *Deps) ClaimPickTask(ctx context.Context, in ClaimPickTaskCmd) (domain.Task, error) {
	if in.TenantID == uuid.Nil || in.WarehouseID == uuid.Nil || in.PickerID == uuid.Nil {
		return domain.Task{}, domain.ErrInvalidArgument
	}

	var task domain.Task
	var err error
	if in.TaskID != nil {
		task, err = d.Tasks.GetByID(ctx, in.TenantID, *in.TaskID)
		if err != nil {
			return domain.Task{}, err
		}
	} else {
		status := domain.TaskStatusQueued
		tt := domain.TaskTypePick
		list, _, err := d.Tasks.List(ctx, ports.TaskFilter{
			TenantID: in.TenantID, WarehouseID: in.WarehouseID,
			Type: &tt, Status: &status, Limit: 1,
		})
		if err != nil {
			return domain.Task{}, err
		}
		if len(list) == 0 {
			return domain.Task{}, domain.ErrNotFound
		}
		task = list[0]
	}

	if task.Type != domain.TaskTypePick || task.Status != domain.TaskStatusQueued {
		return domain.Task{}, domain.ErrTaskNotClaimable
	}

	now := d.now()
	from := task.Status
	task.Status = domain.TaskStatusClaimed
	task.AssigneeID = &in.PickerID
	task.ClaimedAt = &now
	task.UpdatedAt = now
	appendTaskHistory(&task, now, "claimed", &in.PickerID, from, domain.TaskStatusClaimed, "")
	if err := d.Tasks.Update(ctx, task); err != nil {
		return domain.Task{}, err
	}

	session, err := d.Picks.GetSessionByTaskID(ctx, in.TenantID, task.ID)
	if err == nil {
		session.PickerID = &in.PickerID
		session.UpdatedAt = now
		_ = d.Picks.UpdateSession(ctx, session)
	}

	d.publishEvent(ctx, domain.EventTaskAssigned, task.TenantID, task.WarehouseID, task.FulfillmentID, map[string]any{
		"taskId": task.ID, "assigneeId": in.PickerID,
	})
	return task, nil
}

// StartPickCmd starts picking on a claimed task.
type StartPickCmd struct {
	TenantID uuid.UUID
	TaskID   uuid.UUID
	PickerID uuid.UUID
}

// StartPick moves task to in_progress and fulfillment to picking.
func (d *Deps) StartPick(ctx context.Context, in StartPickCmd) (domain.PickSession, error) {
	task, err := d.Tasks.GetByID(ctx, in.TenantID, in.TaskID)
	if err != nil {
		return domain.PickSession{}, err
	}
	if task.Type != domain.TaskTypePick {
		return domain.PickSession{}, domain.ErrInvalidArgument
	}
	if task.Status != domain.TaskStatusClaimed && task.Status != domain.TaskStatusInProgress {
		return domain.PickSession{}, domain.ErrInvalidTransition
	}
	if task.AssigneeID == nil || *task.AssigneeID != in.PickerID {
		return domain.PickSession{}, domain.ErrForbidden
	}

	now := d.now()
	from := task.Status
	task.Status = domain.TaskStatusInProgress
	task.UpdatedAt = now
	appendTaskHistory(&task, now, "started", &in.PickerID, from, domain.TaskStatusInProgress, "")
	if err := d.Tasks.Update(ctx, task); err != nil {
		return domain.PickSession{}, err
	}

	session, err := d.Picks.GetSessionByTaskID(ctx, in.TenantID, task.ID)
	if err != nil {
		return domain.PickSession{}, err
	}
	session.StartedAt = &now
	session.PickerID = &in.PickerID
	session.UpdatedAt = now
	if err := d.Picks.UpdateSession(ctx, session); err != nil {
		return domain.PickSession{}, err
	}

	fo, err := d.Fulfillments.GetByID(ctx, in.TenantID, task.FulfillmentID)
	if err != nil {
		return domain.PickSession{}, err
	}
	fo.Status = domain.FulfillmentStatusPicking
	fo.UpdatedAt = now
	_ = d.Fulfillments.Update(ctx, fo)

	d.publishEvent(ctx, domain.EventPickingStarted, fo.TenantID, fo.WarehouseID, fo.ID, map[string]any{
		"taskId": task.ID, "sessionId": session.ID,
	})
	return session, nil
}

// ScanPickLineCmd validates a scanned barcode against expected pick line.
type ScanPickLineCmd struct {
	TenantID  uuid.UUID
	SessionID uuid.UUID
	LineID    uuid.UUID // pick line id
	Barcode   string
	Qty       int64
	PickerID  uuid.UUID
}

// ScanPickLine validates barcode and increments picked qty.
func (d *Deps) ScanPickLine(ctx context.Context, in ScanPickLineCmd) (domain.PickSession, error) {
	if in.Barcode == "" || in.Qty <= 0 {
		return domain.PickSession{}, domain.ErrInvalidArgument
	}
	session, err := d.Picks.GetSessionByID(ctx, in.TenantID, in.SessionID)
	if err != nil {
		return domain.PickSession{}, err
	}
	task, err := d.Tasks.GetByID(ctx, in.TenantID, session.TaskID)
	if err != nil {
		return domain.PickSession{}, err
	}
	if task.Status != domain.TaskStatusInProgress {
		return domain.PickSession{}, domain.ErrInvalidTransition
	}

	found := false
	for i := range session.Lines {
		pl := &session.Lines[i]
		if pl.ID != in.LineID {
			continue
		}
		found = true
		if pl.Barcode != in.Barcode {
			return domain.PickSession{}, domain.ErrBarcodeMismatch
		}
		remaining := pl.QtyRequired - pl.QtyPicked
		if in.Qty > remaining {
			return domain.PickSession{}, fmt.Errorf("%w: qty exceeds remaining %d", domain.ErrInvalidArgument, remaining)
		}
		pl.QtyPicked += in.Qty
		if pl.QtyPicked >= pl.QtyRequired {
			pl.Status = domain.PickLineStatusComplete
		} else {
			pl.Status = domain.PickLineStatusPartial
		}
		break
	}
	if !found {
		return domain.PickSession{}, domain.ErrNotFound
	}

	session.UpdatedAt = d.now()
	if err := d.Picks.UpdateSession(ctx, session); err != nil {
		return domain.PickSession{}, err
	}
	return session, nil
}

// CompletePickCmd finishes a pick session and queues a pack task.
type CompletePickCmd struct {
	TenantID  uuid.UUID
	TaskID    uuid.UUID
	PickerID  uuid.UUID
}

// CompletePick requires all lines fully picked, then creates pack task.
func (d *Deps) CompletePick(ctx context.Context, in CompletePickCmd) (domain.Task, error) {
	task, err := d.Tasks.GetByID(ctx, in.TenantID, in.TaskID)
	if err != nil {
		return domain.Task{}, err
	}
	if task.Type != domain.TaskTypePick || task.Status != domain.TaskStatusInProgress {
		return domain.Task{}, domain.ErrInvalidTransition
	}

	session, err := d.Picks.GetSessionByTaskID(ctx, in.TenantID, task.ID)
	if err != nil {
		return domain.Task{}, err
	}
	for _, pl := range session.Lines {
		if pl.QtyPicked < pl.QtyRequired {
			return domain.Task{}, domain.ErrRemainingQty
		}
	}

	now := d.now()
	session.CompletedAt = &now
	session.UpdatedAt = now
	_ = d.Picks.UpdateSession(ctx, session)

	from := task.Status
	task.Status = domain.TaskStatusCompleted
	task.CompletedAt = &now
	task.UpdatedAt = now
	appendTaskHistory(&task, now, "completed", &in.PickerID, from, domain.TaskStatusCompleted, "")
	if err := d.Tasks.Update(ctx, task); err != nil {
		return domain.Task{}, err
	}

	fo, err := d.Fulfillments.GetByID(ctx, in.TenantID, task.FulfillmentID)
	if err != nil {
		return domain.Task{}, err
	}
	for i := range fo.Lines {
		for _, pl := range session.Lines {
			if pl.LineID == fo.Lines[i].ID {
				fo.Lines[i].QtyPicked = pl.QtyPicked
			}
		}
	}
	fo.Status = domain.FulfillmentStatusPicked
	fo.UpdatedAt = now
	_ = d.Fulfillments.Update(ctx, fo)

	d.publishEvent(ctx, domain.EventPickingCompleted, fo.TenantID, fo.WarehouseID, fo.ID, map[string]any{
		"taskId": task.ID,
	})
	d.publishEvent(ctx, domain.EventTaskCompleted, fo.TenantID, fo.WarehouseID, fo.ID, map[string]any{
		"taskId": task.ID, "type": task.Type,
	})

	packTask := domain.Task{
		ID: d.newID(), TenantID: fo.TenantID, WarehouseID: fo.WarehouseID,
		FulfillmentID: fo.ID, Type: domain.TaskTypePack, Status: domain.TaskStatusQueued,
		Priority: fo.Priority, CreatedAt: now, UpdatedAt: now,
	}
	appendTaskHistory(&packTask, now, "created", &in.PickerID, "", domain.TaskStatusQueued, "pack queued")
	if err := d.Tasks.Create(ctx, packTask); err != nil {
		return domain.Task{}, err
	}

	pack := domain.PackSession{
		ID: d.newID(), TenantID: fo.TenantID, WarehouseID: fo.WarehouseID,
		FulfillmentID: fo.ID, TaskID: packTask.ID, Status: domain.PackSessionStatusQueued,
		ExpectedWeightG: int64(len(fo.Lines)) * 500,
		WeightTolerance: d.weightTol(),
		CreatedAt: now, UpdatedAt: now,
	}
	if err := d.Packs.Create(ctx, pack); err != nil {
		return domain.Task{}, err
	}

	fo.Status = domain.FulfillmentStatusPackQueued
	fo.UpdatedAt = d.now()
	_ = d.Fulfillments.Update(ctx, fo)

	return packTask, nil
}
