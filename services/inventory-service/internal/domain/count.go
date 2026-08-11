package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// CountSessionType classifies an inventory count session.
type CountSessionType string

const (
	CountSessionTypeCycle CountSessionType = "cycle"
	CountSessionTypeFull  CountSessionType = "full"
	CountSessionTypeBlind CountSessionType = "blind"
	CountSessionTypeAudit CountSessionType = "audit"
	CountSessionTypeSpot  CountSessionType = "spot"
)

func (t CountSessionType) Valid() bool {
	switch t {
	case CountSessionTypeCycle, CountSessionTypeFull, CountSessionTypeBlind,
		CountSessionTypeAudit, CountSessionTypeSpot:
		return true
	default:
		return false
	}
}

// CountSessionStatus is the lifecycle of a count session.
type CountSessionStatus string

const (
	CountSessionStatusDraft           CountSessionStatus = "draft"
	CountSessionStatusInProgress      CountSessionStatus = "in_progress"
	CountSessionStatusSubmitted       CountSessionStatus = "submitted"
	CountSessionStatusPendingApproval CountSessionStatus = "pending_approval"
	CountSessionStatusApproved        CountSessionStatus = "approved"
	CountSessionStatusRejected        CountSessionStatus = "rejected"
	CountSessionStatusCancelled       CountSessionStatus = "cancelled"
)

func (s CountSessionStatus) Valid() bool {
	switch s {
	case CountSessionStatusDraft, CountSessionStatusInProgress, CountSessionStatusSubmitted,
		CountSessionStatusPendingApproval, CountSessionStatusApproved,
		CountSessionStatusRejected, CountSessionStatusCancelled:
		return true
	default:
		return false
	}
}

var countTransitions = map[CountSessionStatus][]CountSessionStatus{
	CountSessionStatusDraft: {
		CountSessionStatusInProgress, CountSessionStatusCancelled,
	},
	CountSessionStatusInProgress: {
		CountSessionStatusSubmitted, CountSessionStatusCancelled,
	},
	CountSessionStatusSubmitted: {
		CountSessionStatusPendingApproval, CountSessionStatusApproved, CountSessionStatusRejected,
	},
	CountSessionStatusPendingApproval: {
		CountSessionStatusApproved, CountSessionStatusRejected,
	},
	CountSessionStatusApproved:  {},
	CountSessionStatusRejected:  {},
	CountSessionStatusCancelled: {},
}

// CanTransitionTo reports whether from → to is allowed.
func (s CountSessionStatus) CanTransitionTo(to CountSessionStatus) bool {
	if !s.Valid() || !to.Valid() {
		return false
	}
	if s == to {
		return true
	}
	for _, next := range countTransitions[s] {
		if next == to {
			return true
		}
	}
	return false
}

// CountSession is a cycle/full/blind/audit/spot count header.
type CountSession struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	WarehouseID uuid.UUID
	LocationID  *uuid.UUID
	Type        CountSessionType
	Status      CountSessionStatus
	StartedBy   *uuid.UUID
	SubmittedBy *uuid.UUID
	ApprovedBy  *uuid.UUID
	Notes       string
	Lines       []CountLine
	Metadata    map[string]any
	CreatedAt   time.Time
	UpdatedAt   time.Time
	StartedAt   *time.Time
	SubmittedAt *time.Time
	ApprovedAt  *time.Time
	CancelledAt *time.Time
}

// CountLine compares counted qty vs system qty for a variant/location/lot.
type CountLine struct {
	ID         uuid.UUID
	SessionID  uuid.UUID
	VariantID  uuid.UUID
	SKUCode    string
	LocationID *uuid.UUID
	LotID      *uuid.UUID
	SystemQty  int64
	CountedQty *int64
	Variance   *int64
	Approved   *bool
	Notes      string
	Metadata   map[string]any
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// Validate checks structural invariants.
func (s CountSession) Validate() error {
	if s.ID == uuid.Nil {
		return fmt.Errorf("%w: count session id required", ErrInvalidArgument)
	}
	if s.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if s.WarehouseID == uuid.Nil {
		return fmt.Errorf("%w: warehouse_id required", ErrInvalidArgument)
	}
	if !s.Type.Valid() {
		return fmt.Errorf("%w: invalid count type %q", ErrInvalidArgument, s.Type)
	}
	if !s.Status.Valid() {
		return fmt.Errorf("%w: invalid count status %q", ErrInvalidArgument, s.Status)
	}
	for i, line := range s.Lines {
		if err := line.Validate(); err != nil {
			return fmt.Errorf("line[%d]: %w", i, err)
		}
	}
	return nil
}

// Validate checks a count line.
func (l CountLine) Validate() error {
	if l.ID == uuid.Nil {
		return fmt.Errorf("%w: count line id required", ErrInvalidArgument)
	}
	if l.VariantID == uuid.Nil {
		return fmt.Errorf("%w: variant_id required", ErrInvalidArgument)
	}
	if l.SystemQty < 0 {
		return fmt.Errorf("%w: system_qty cannot be negative", ErrInvalidArgument)
	}
	if l.CountedQty != nil {
		if *l.CountedQty < 0 {
			return fmt.Errorf("%w: counted_qty cannot be negative", ErrInvalidArgument)
		}
		expected := *l.CountedQty - l.SystemQty
		if l.Variance == nil || *l.Variance != expected {
			return fmt.Errorf("%w: variance must equal counted_qty - system_qty", ErrInvariant)
		}
	}
	return nil
}

// SetCounted sets counted qty and computes variance.
func (l *CountLine) SetCounted(qty int64) error {
	if qty < 0 {
		return fmt.Errorf("%w: counted_qty", ErrNegativeQuantity)
	}
	l.CountedQty = &qty
	v := qty - l.SystemQty
	l.Variance = &v
	l.UpdatedAt = time.Now().UTC()
	return nil
}

// HasVariance reports whether the line has a non-zero variance.
func (l CountLine) HasVariance() bool {
	return l.Variance != nil && *l.Variance != 0
}

// TransitionTo moves the session along the allowed state machine.
func (s *CountSession) TransitionTo(next CountSessionStatus, actorID *uuid.UUID) error {
	if !s.Status.CanTransitionTo(next) {
		return fmt.Errorf("%w: %s → %s", ErrInvalidTransition, s.Status, next)
	}
	now := time.Now().UTC()
	s.Status = next
	s.UpdatedAt = now
	switch next {
	case CountSessionStatusInProgress:
		s.StartedBy = actorID
		s.StartedAt = &now
	case CountSessionStatusSubmitted:
		s.SubmittedBy = actorID
		s.SubmittedAt = &now
	case CountSessionStatusApproved:
		s.ApprovedBy = actorID
		s.ApprovedAt = &now
	case CountSessionStatusCancelled:
		s.CancelledAt = &now
	}
	return nil
}
