package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// TaskType classifies warehouse work items.
type TaskType string

const (
	TaskTypePick        TaskType = "pick"
	TaskTypePack        TaskType = "pack"
	TaskTypeDispatch    TaskType = "dispatch"
	TaskTypeQC          TaskType = "qc"
	TaskTypeReplenish   TaskType = "replenish"
	TaskTypeClean       TaskType = "clean"
	TaskTypeMaintenance TaskType = "maintenance"
	TaskTypeEmergency   TaskType = "emergency"
)

func (t TaskType) Valid() bool {
	switch t {
	case TaskTypePick, TaskTypePack, TaskTypeDispatch, TaskTypeQC,
		TaskTypeReplenish, TaskTypeClean, TaskTypeMaintenance, TaskTypeEmergency:
		return true
	default:
		return false
	}
}

// TaskStatus is the lifecycle of a work task.
type TaskStatus string

const (
	TaskStatusQueued     TaskStatus = "queued"
	TaskStatusClaimed    TaskStatus = "claimed"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusCancelled  TaskStatus = "cancelled"
	TaskStatusEscalated  TaskStatus = "escalated"
)

func (s TaskStatus) Valid() bool {
	switch s {
	case TaskStatusQueued, TaskStatusClaimed, TaskStatusInProgress,
		TaskStatusCompleted, TaskStatusCancelled, TaskStatusEscalated:
		return true
	default:
		return false
	}
}

func (s TaskStatus) IsTerminal() bool {
	return s == TaskStatusCompleted || s == TaskStatusCancelled
}

var taskTransitions = map[TaskStatus][]TaskStatus{
	TaskStatusQueued: {
		TaskStatusClaimed, TaskStatusCancelled, TaskStatusEscalated,
	},
	TaskStatusClaimed: {
		TaskStatusInProgress, TaskStatusCancelled, TaskStatusEscalated, TaskStatusQueued,
	},
	TaskStatusInProgress: {
		TaskStatusCompleted, TaskStatusCancelled, TaskStatusEscalated,
	},
	TaskStatusEscalated: {
		TaskStatusClaimed, TaskStatusCancelled, TaskStatusInProgress,
	},
}

// TaskHistoryEntry is an append-only history record on a task.
type TaskHistoryEntry struct {
	At         time.Time
	Action     string
	ActorID    *uuid.UUID
	FromStatus TaskStatus
	ToStatus   TaskStatus
	Note       string
}

// Task is a unit of warehouse work in the priority queue.
type Task struct {
	ID            uuid.UUID
	TenantID      uuid.UUID
	WarehouseID   uuid.UUID
	FulfillmentID uuid.UUID
	StationID     *uuid.UUID
	Type          TaskType
	Status        TaskStatus
	AssigneeID    *uuid.UUID
	Priority      int
	WaveID        *uuid.UUID
	BatchID       *uuid.UUID
	SLADeadline   *time.Time
	ClaimedAt     *time.Time
	StartedAt     *time.Time
	CompletedAt   *time.Time
	CancelledAt   *time.Time
	EscalatedAt    *time.Time
	EscalationNote string
	History        []TaskHistoryEntry
	Metadata       map[string]any
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// Validate checks structural invariants.
func (t Task) Validate() error {
	if t.ID == uuid.Nil {
		return fmt.Errorf("%w: task id required", ErrInvalidArgument)
	}
	if t.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if t.WarehouseID == uuid.Nil {
		return fmt.Errorf("%w: warehouse_id required", ErrInvalidArgument)
	}
	if !t.Type.Valid() {
		return fmt.Errorf("%w: invalid task type %q", ErrInvalidArgument, t.Type)
	}
	if !t.Status.Valid() {
		return fmt.Errorf("%w: invalid task status %q", ErrInvalidArgument, t.Status)
	}
	return nil
}

// CanTransitionTo reports whether status allows moving to next.
func (t Task) CanTransitionTo(next TaskStatus) bool {
	if !t.Status.Valid() || !next.Valid() {
		return false
	}
	if t.Status == next {
		return true
	}
	if t.Status.IsTerminal() {
		return false
	}
	for _, s := range taskTransitions[t.Status] {
		if s == next {
			return true
		}
	}
	return false
}

func (t *Task) applyStatus(next TaskStatus) error {
	if t.Status.IsTerminal() {
		return fmt.Errorf("%w: status %s", ErrAlreadyTerminal, t.Status)
	}
	if !t.CanTransitionTo(next) {
		return fmt.Errorf("%w: %s → %s", ErrInvalidTransition, t.Status, next)
	}
	now := time.Now().UTC()
	from := t.Status
	t.Status = next
	t.UpdatedAt = now
	switch next {
	case TaskStatusClaimed:
		t.ClaimedAt = &now
	case TaskStatusInProgress:
		t.StartedAt = &now
	case TaskStatusCompleted:
		t.CompletedAt = &now
	case TaskStatusCancelled:
		t.CancelledAt = &now
	case TaskStatusEscalated:
		t.EscalatedAt = &now
	case TaskStatusQueued:
		t.AssigneeID = nil
		t.ClaimedAt = nil
	}
	t.History = append(t.History, TaskHistoryEntry{
		At: now, Action: string(next), FromStatus: from, ToStatus: next,
	})
	return nil
}

// Claim assigns the task to an employee. Only allowed from queued.
func (t *Task) Claim(assigneeID uuid.UUID) error {
	if assigneeID == uuid.Nil {
		return fmt.Errorf("%w: assignee_id required", ErrInvalidArgument)
	}
	if t.Status != TaskStatusQueued {
		return fmt.Errorf("%w: status %s (claim only from queued)", ErrTaskNotClaimable, t.Status)
	}
	if err := t.applyStatus(TaskStatusClaimed); err != nil {
		return err
	}
	t.AssigneeID = &assigneeID
	return nil
}

// Start moves claimed → in_progress.
func (t *Task) Start() error {
	return t.applyStatus(TaskStatusInProgress)
}

// Complete moves in_progress → completed.
func (t *Task) Complete() error {
	return t.applyStatus(TaskStatusCompleted)
}

// Cancel moves to cancelled from a non-terminal status.
func (t *Task) Cancel() error {
	return t.applyStatus(TaskStatusCancelled)
}

// Escalate marks the task escalated (SLA / exception).
func (t *Task) Escalate() error {
	return t.applyStatus(TaskStatusEscalated)
}

// Reassign changes assignee on a claimed, in-progress, or escalated task.
func (t *Task) Reassign(assigneeID uuid.UUID) error {
	if assigneeID == uuid.Nil {
		return fmt.Errorf("%w: assignee_id required", ErrInvalidArgument)
	}
	switch t.Status {
	case TaskStatusClaimed, TaskStatusEscalated, TaskStatusInProgress:
		t.AssigneeID = &assigneeID
		t.UpdatedAt = time.Now().UTC()
		if t.Status == TaskStatusEscalated {
			return t.applyStatus(TaskStatusClaimed)
		}
		return nil
	default:
		return fmt.Errorf("%w: cannot reassign from %s", ErrInvalidTransition, t.Status)
	}
}

// Requeue returns a claimed task to the queue.
func (t *Task) Requeue() error {
	if t.Status != TaskStatusClaimed {
		return fmt.Errorf("%w: requeue only from claimed (got %s)", ErrInvalidTransition, t.Status)
	}
	return t.applyStatus(TaskStatusQueued)
}
