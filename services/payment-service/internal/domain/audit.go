package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// AuditEntry records a money mutation for compliance.
type AuditEntry struct {
	ID         uuid.UUID
	TenantID   uuid.UUID
	IntentID   *uuid.UUID
	Action     string
	ActorID    *uuid.UUID
	ActorType  string
	AmountMinor int64
	Currency   string
	Detail     map[string]any
	CreatedAt  time.Time
}

// Validate checks audit entry invariants.
func (a AuditEntry) Validate() error {
	if a.ID == uuid.Nil {
		return fmt.Errorf("%w: audit id required", ErrInvalidArgument)
	}
	if a.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if a.Action == "" {
		return fmt.Errorf("%w: action required", ErrInvalidArgument)
	}
	return nil
}
