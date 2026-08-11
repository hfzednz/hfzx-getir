package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// TransferStatus is the lifecycle of an inter/intra warehouse transfer.
type TransferStatus string

const (
	TransferStatusDraft           TransferStatus = "draft"
	TransferStatusPendingApproval TransferStatus = "pending_approval"
	TransferStatusApproved        TransferStatus = "approved"
	TransferStatusInTransit       TransferStatus = "in_transit"
	TransferStatusCompleted       TransferStatus = "completed"
	TransferStatusCancelled       TransferStatus = "cancelled"
)

func (s TransferStatus) Valid() bool {
	switch s {
	case TransferStatusDraft, TransferStatusPendingApproval, TransferStatusApproved,
		TransferStatusInTransit, TransferStatusCompleted, TransferStatusCancelled:
		return true
	default:
		return false
	}
}

var transferTransitions = map[TransferStatus][]TransferStatus{
	TransferStatusDraft: {
		TransferStatusPendingApproval, TransferStatusApproved, TransferStatusCancelled,
	},
	TransferStatusPendingApproval: {
		TransferStatusApproved, TransferStatusCancelled,
	},
	TransferStatusApproved: {
		TransferStatusInTransit, TransferStatusCancelled,
	},
	TransferStatusInTransit: {
		TransferStatusCompleted, TransferStatusCancelled,
	},
	TransferStatusCompleted: {},
	TransferStatusCancelled: {},
}

// CanTransitionTo reports whether from → to is allowed.
func (s TransferStatus) CanTransitionTo(to TransferStatus) bool {
	if !s.Valid() || !to.Valid() {
		return false
	}
	if s == to {
		return true
	}
	for _, next := range transferTransitions[s] {
		if next == to {
			return true
		}
	}
	return false
}

// Transfer moves stock between warehouses and/or locations.
type Transfer struct {
	ID              uuid.UUID
	TenantID        uuid.UUID
	Code            string
	FromWarehouseID uuid.UUID
	ToWarehouseID   uuid.UUID
	FromLocationID  *uuid.UUID
	ToLocationID    *uuid.UUID
	Status          TransferStatus
	RequestedBy     *uuid.UUID
	ApprovedBy      *uuid.UUID
	Reason          string
	Lines           []TransferLine
	Metadata        map[string]any
	CreatedAt       time.Time
	UpdatedAt       time.Time
	ApprovedAt      *time.Time
	ShippedAt       *time.Time
	CompletedAt     *time.Time
	CancelledAt     *time.Time
}

// TransferLine is a variant quantity on a transfer.
type TransferLine struct {
	ID           uuid.UUID
	TransferID   uuid.UUID
	VariantID    uuid.UUID
	SKUCode      string
	LotID        *uuid.UUID
	QtyRequested int64
	QtyShipped   int64
	QtyReceived  int64
	Metadata     map[string]any
	CreatedAt    time.Time
}

// Validate checks structural invariants.
func (t Transfer) Validate() error {
	if t.ID == uuid.Nil {
		return fmt.Errorf("%w: transfer id required", ErrInvalidArgument)
	}
	if t.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if t.FromWarehouseID == uuid.Nil || t.ToWarehouseID == uuid.Nil {
		return fmt.Errorf("%w: from/to warehouse_id required", ErrInvalidArgument)
	}
	if t.FromWarehouseID == t.ToWarehouseID {
		sameLoc := (t.FromLocationID == nil && t.ToLocationID == nil) ||
			(t.FromLocationID != nil && t.ToLocationID != nil && *t.FromLocationID == *t.ToLocationID)
		if sameLoc {
			return fmt.Errorf("%w: transfer source and destination must differ", ErrInvariant)
		}
	}
	if !t.Status.Valid() {
		return fmt.Errorf("%w: invalid transfer status %q", ErrInvalidArgument, t.Status)
	}
	for i, line := range t.Lines {
		if err := line.Validate(); err != nil {
			return fmt.Errorf("line[%d]: %w", i, err)
		}
	}
	return nil
}

// Validate checks a transfer line.
func (l TransferLine) Validate() error {
	if l.ID == uuid.Nil {
		return fmt.Errorf("%w: transfer line id required", ErrInvalidArgument)
	}
	if l.VariantID == uuid.Nil {
		return fmt.Errorf("%w: variant_id required", ErrInvalidArgument)
	}
	if l.QtyRequested <= 0 {
		return fmt.Errorf("%w: qty_requested", ErrNegativeQuantity)
	}
	if l.QtyShipped < 0 || l.QtyReceived < 0 {
		return fmt.Errorf("%w: shipped/received cannot be negative", ErrInvariant)
	}
	if l.QtyShipped > l.QtyRequested {
		return fmt.Errorf("%w: qty_shipped exceeds qty_requested", ErrInvariant)
	}
	if l.QtyReceived > l.QtyShipped {
		return fmt.Errorf("%w: qty_received exceeds qty_shipped", ErrInvariant)
	}
	return nil
}

// TransitionTo moves the transfer along the allowed state machine.
func (t *Transfer) TransitionTo(next TransferStatus, actorID *uuid.UUID) error {
	if !t.Status.CanTransitionTo(next) {
		return fmt.Errorf("%w: %s → %s", ErrInvalidTransition, t.Status, next)
	}
	now := time.Now().UTC()
	t.Status = next
	t.UpdatedAt = now
	switch next {
	case TransferStatusApproved:
		t.ApprovedBy = actorID
		t.ApprovedAt = &now
	case TransferStatusInTransit:
		t.ShippedAt = &now
	case TransferStatusCompleted:
		t.CompletedAt = &now
	case TransferStatusCancelled:
		t.CancelledAt = &now
	}
	return nil
}
