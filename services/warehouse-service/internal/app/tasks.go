package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/nexora/warehouse-service/internal/app/ports"
	"github.com/nexora/warehouse-service/internal/domain"
)

// ReassignTaskCmd moves a task to another assignee.
type ReassignTaskCmd struct {
	TenantID     uuid.UUID
	TaskID       uuid.UUID
	NewAssignee  uuid.UUID
	ActorID      *uuid.UUID
	Note         string
}

// ReassignTask reassigns a non-terminal task.
func (d *Deps) ReassignTask(ctx context.Context, in ReassignTaskCmd) (domain.Task, error) {
	task, err := d.Tasks.GetByID(ctx, in.TenantID, in.TaskID)
	if err != nil {
		return domain.Task{}, err
	}
	if task.Status == domain.TaskStatusCompleted || task.Status == domain.TaskStatusCancelled {
		return domain.Task{}, domain.ErrInvalidTransition
	}
	now := d.now()
	from := task.Status
	task.AssigneeID = &in.NewAssignee
	if task.Status == domain.TaskStatusQueued {
		task.Status = domain.TaskStatusClaimed
		task.ClaimedAt = &now
	}
	task.UpdatedAt = now
	appendTaskHistory(&task, now, "reassigned", in.ActorID, from, task.Status, in.Note)
	if err := d.Tasks.Update(ctx, task); err != nil {
		return domain.Task{}, err
	}
	d.publishEvent(ctx, domain.EventTaskReassigned, task.TenantID, task.WarehouseID, task.FulfillmentID, map[string]any{
		"taskId": task.ID, "assigneeId": in.NewAssignee,
	})
	return task, nil
}

// CancelTaskCmd cancels a task.
type CancelTaskCmd struct {
	TenantID uuid.UUID
	TaskID   uuid.UUID
	Reason   string
	ActorID  *uuid.UUID
}

// CancelTask cancels a non-completed task.
func (d *Deps) CancelTask(ctx context.Context, in CancelTaskCmd) (domain.Task, error) {
	task, err := d.Tasks.GetByID(ctx, in.TenantID, in.TaskID)
	if err != nil {
		return domain.Task{}, err
	}
	if task.Status == domain.TaskStatusCompleted || task.Status == domain.TaskStatusCancelled {
		return domain.Task{}, domain.ErrInvalidTransition
	}
	now := d.now()
	from := task.Status
	task.Status = domain.TaskStatusCancelled
	task.UpdatedAt = now
	appendTaskHistory(&task, now, "cancelled", in.ActorID, from, domain.TaskStatusCancelled, in.Reason)
	if err := d.Tasks.Update(ctx, task); err != nil {
		return domain.Task{}, err
	}
	d.publishEvent(ctx, domain.EventTaskCancelled, task.TenantID, task.WarehouseID, task.FulfillmentID, map[string]any{
		"taskId": task.ID, "reason": in.Reason,
	})
	return task, nil
}

// EscalateTaskCmd escalates a stuck task.
type EscalateTaskCmd struct {
	TenantID uuid.UUID
	TaskID   uuid.UUID
	Note     string
	ActorID  *uuid.UUID
}

// EscalateTask marks a task escalated with a note.
func (d *Deps) EscalateTask(ctx context.Context, in EscalateTaskCmd) (domain.Task, error) {
	task, err := d.Tasks.GetByID(ctx, in.TenantID, in.TaskID)
	if err != nil {
		return domain.Task{}, err
	}
	if task.Status == domain.TaskStatusCompleted || task.Status == domain.TaskStatusCancelled {
		return domain.Task{}, domain.ErrInvalidTransition
	}
	now := d.now()
	from := task.Status
	task.Status = domain.TaskStatusEscalated
	task.EscalationNote = in.Note
	task.UpdatedAt = now
	appendTaskHistory(&task, now, "escalated", in.ActorID, from, domain.TaskStatusEscalated, in.Note)
	if err := d.Tasks.Update(ctx, task); err != nil {
		return domain.Task{}, err
	}
	d.publishEvent(ctx, domain.EventTaskEscalated, task.TenantID, task.WarehouseID, task.FulfillmentID, map[string]any{
		"taskId": task.ID, "note": in.Note,
	})
	return task, nil
}

// ListTasks lists tasks for a warehouse queue.
func (d *Deps) ListTasks(ctx context.Context, f ports.TaskFilter) ([]domain.Task, int, error) {
	return d.Tasks.List(ctx, f)
}

// GetTask returns a task by id.
func (d *Deps) GetTask(ctx context.Context, tenantID, id uuid.UUID) (domain.Task, error) {
	return d.Tasks.GetByID(ctx, tenantID, id)
}
