package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// PointEntryKind classifies point ledger rows.
type PointEntryKind string

const (
	PointEarn   PointEntryKind = "earn"
	PointRedeem PointEntryKind = "redeem"
	PointExpire PointEntryKind = "expire"
	PointAdjust PointEntryKind = "adjust"
	PointGrant  PointEntryKind = "grant"
)

// Valid reports whether the point entry kind is recognized.
func (k PointEntryKind) Valid() bool {
	switch k {
	case PointEarn, PointRedeem, PointExpire, PointAdjust, PointGrant:
		return true
	default:
		return false
	}
}

// Account is a loyalty account for an opaque principal.
type Account struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	PrincipalID uuid.UUID
	Points      int64
	TierPoints  int64
	XP          int64
	Level       int
	Active      bool
	Version     int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Validate checks account invariants.
func (a Account) Validate() error {
	if a.ID == uuid.Nil {
		return fmt.Errorf("%w: account id required", ErrInvalidArgument)
	}
	if a.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if a.PrincipalID == uuid.Nil {
		return fmt.Errorf("%w: principal_id required", ErrInvalidArgument)
	}
	if a.Points < 0 || a.TierPoints < 0 || a.XP < 0 {
		return fmt.Errorf("%w: points/xp must be non-negative", ErrInvariant)
	}
	return nil
}

// PointLedgerEntry is an append-only points mutation.
type PointLedgerEntry struct {
	ID             uuid.UUID
	TenantID       uuid.UUID
	AccountID      uuid.UUID
	Kind           PointEntryKind
	Points         int64
	BalanceAfter   int64
	OrderID        *uuid.UUID
	Reference      string
	IdempotencyKey string
	Metadata       map[string]any
	CreatedAt      time.Time
}
