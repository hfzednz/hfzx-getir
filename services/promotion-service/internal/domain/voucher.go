package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// VoucherStatus is the voucher lifecycle.
type VoucherStatus string

const (
	VoucherIssued  VoucherStatus = "issued"
	VoucherRedeemed VoucherStatus = "redeemed"
	VoucherExpired VoucherStatus = "expired"
	VoucherVoid    VoucherStatus = "void"
)

// Valid reports whether the voucher status is recognized.
func (s VoucherStatus) Valid() bool {
	switch s {
	case VoucherIssued, VoucherRedeemed, VoucherExpired, VoucherVoid:
		return true
	default:
		return false
	}
}

// Voucher is a personal credit/discount instrument.
type Voucher struct {
	ID            uuid.UUID
	TenantID      uuid.UUID
	PromotionID   *uuid.UUID // optional binding
	Code          string
	PrincipalID   uuid.UUID
	Status        VoucherStatus
	ValueMinor    int64
	Currency      string
	RemainingMinor int64
	StartsAt      *time.Time
	EndsAt        *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// Validate checks voucher invariants.
func (v Voucher) Validate() error {
	if v.ID == uuid.Nil {
		return fmt.Errorf("%w: voucher id required", ErrInvalidArgument)
	}
	if v.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if v.PrincipalID == uuid.Nil {
		return fmt.Errorf("%w: principal_id required", ErrInvalidArgument)
	}
	if strings.TrimSpace(v.Code) == "" {
		return fmt.Errorf("%w: voucher code required", ErrInvalidArgument)
	}
	if !v.Status.Valid() {
		return fmt.Errorf("%w: invalid voucher status %q", ErrInvalidArgument, v.Status)
	}
	if v.ValueMinor < 0 || v.RemainingMinor < 0 {
		return fmt.Errorf("%w: voucher amounts must be non-negative", ErrNegativeMoney)
	}
	if len(v.Currency) != 3 {
		return fmt.Errorf("%w: currency must be ISO-4217", ErrInvalidArgument)
	}
	return nil
}

// IsRedeemableAt reports whether the voucher can be redeemed at now.
func (v Voucher) IsRedeemableAt(now time.Time) bool {
	if v.Status != VoucherIssued {
		return false
	}
	if v.RemainingMinor <= 0 {
		return false
	}
	if v.StartsAt != nil && now.Before(*v.StartsAt) {
		return false
	}
	if v.EndsAt != nil && !now.Before(*v.EndsAt) {
		return false
	}
	return true
}

// VoucherRedemption records a voucher spend (idempotent by key).
type VoucherRedemption struct {
	ID             uuid.UUID
	TenantID       uuid.UUID
	VoucherID      uuid.UUID
	PrincipalID    uuid.UUID
	IdempotencyKey string
	OrderRef       string
	AmountMinor    int64
	Currency       string
	RedeemedAt     time.Time
	CreatedAt      time.Time
}
