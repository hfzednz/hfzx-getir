package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// PrincipalKind classifies the identity subject.
type PrincipalKind string

const (
	PrincipalKindUser           PrincipalKind = "user"
	PrincipalKindServiceAccount PrincipalKind = "service_account"
	PrincipalKindRobot          PrincipalKind = "robot"
	PrincipalKindGuest          PrincipalKind = "guest"
)

func (k PrincipalKind) Valid() bool {
	switch k {
	case PrincipalKindUser, PrincipalKindServiceAccount, PrincipalKindRobot, PrincipalKindGuest:
		return true
	default:
		return false
	}
}

// PrincipalStatus is the lifecycle state of a principal.
type PrincipalStatus string

const (
	PrincipalStatusActive    PrincipalStatus = "active"
	PrincipalStatusLocked    PrincipalStatus = "locked"
	PrincipalStatusSuspended PrincipalStatus = "suspended"
	PrincipalStatusDeleted   PrincipalStatus = "deleted"
)

func (s PrincipalStatus) Valid() bool {
	switch s {
	case PrincipalStatusActive, PrincipalStatusLocked, PrincipalStatusSuspended, PrincipalStatusDeleted:
		return true
	default:
		return false
	}
}

// CanAuthenticate reports whether the principal may start an auth flow.
func (s PrincipalStatus) CanAuthenticate() bool {
	return s == PrincipalStatusActive
}

// Principal is a canonical identity subject within a tenant.
type Principal struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	Kind        PrincipalKind
	Status      PrincipalStatus
	DisplayName string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}

// Validate checks structural invariants.
func (p Principal) Validate() error {
	if p.ID == uuid.Nil {
		return fmt.Errorf("%w: principal id required", ErrInvalidArgument)
	}
	if p.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if !p.Kind.Valid() {
		return fmt.Errorf("%w: invalid principal kind %q", ErrInvalidArgument, p.Kind)
	}
	if !p.Status.Valid() {
		return fmt.Errorf("%w: invalid principal status %q", ErrInvalidArgument, p.Status)
	}
	if p.Status == PrincipalStatusDeleted && p.DeletedAt == nil {
		return fmt.Errorf("%w: deleted principal requires deleted_at", ErrInvariant)
	}
	if p.DeletedAt != nil && p.Status != PrincipalStatusDeleted {
		return fmt.Errorf("%w: deleted_at set but status is %s", ErrInvariant, p.Status)
	}
	return nil
}

// IsActive is a convenience for authz gates.
func (p Principal) IsActive() bool {
	return p.Status == PrincipalStatusActive && p.DeletedAt == nil
}
