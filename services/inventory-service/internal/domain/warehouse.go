package domain

import (
	"fmt"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

// WarehouseStatus is the lifecycle of a warehouse site.
type WarehouseStatus string

const (
	WarehouseStatusActive      WarehouseStatus = "active"
	WarehouseStatusInactive    WarehouseStatus = "inactive"
	WarehouseStatusMaintenance WarehouseStatus = "maintenance"
	WarehouseStatusClosed      WarehouseStatus = "closed"
)

func (s WarehouseStatus) Valid() bool {
	switch s {
	case WarehouseStatusActive, WarehouseStatusInactive,
		WarehouseStatusMaintenance, WarehouseStatusClosed:
		return true
	default:
		return false
	}
}

const maxWarehouseCodeLen = 64
const maxWarehouseNameLen = 200

// Warehouse is a physical or logical stock site owned by a tenant.
type Warehouse struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	Code      string
	Name      string
	RegionID  *uuid.UUID
	Timezone  string
	Status    WarehouseStatus
	Metadata  map[string]any
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

// Validate checks structural invariants.
func (w Warehouse) Validate() error {
	if w.ID == uuid.Nil {
		return fmt.Errorf("%w: warehouse id required", ErrInvalidArgument)
	}
	if w.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if w.Code == "" {
		return fmt.Errorf("%w: warehouse code required", ErrInvalidArgument)
	}
	if utf8.RuneCountInString(w.Code) > maxWarehouseCodeLen {
		return fmt.Errorf("%w: warehouse code too long", ErrInvalidArgument)
	}
	if w.Name == "" {
		return fmt.Errorf("%w: warehouse name required", ErrInvalidArgument)
	}
	if utf8.RuneCountInString(w.Name) > maxWarehouseNameLen {
		return fmt.Errorf("%w: warehouse name too long", ErrInvalidArgument)
	}
	if w.Timezone == "" {
		return fmt.Errorf("%w: timezone required", ErrInvalidArgument)
	}
	if !w.Status.Valid() {
		return fmt.Errorf("%w: invalid warehouse status %q", ErrInvalidArgument, w.Status)
	}
	if w.DeletedAt != nil && w.Status != WarehouseStatusClosed {
		return fmt.Errorf("%w: deleted warehouse should be closed", ErrInvariant)
	}
	return nil
}

// IsOperable reports whether the warehouse can accept stock mutations.
func (w Warehouse) IsOperable() bool {
	return w.Status == WarehouseStatusActive && w.DeletedAt == nil
}
