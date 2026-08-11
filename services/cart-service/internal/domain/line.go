package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// DefaultMaxQty is the default line-level available qty cap.
const DefaultMaxQty = 99

// CartLine is a cart line with opaque variant_id and max qty rule.
type CartLine struct {
	ID              uuid.UUID
	CartID          uuid.UUID
	TenantID        uuid.UUID
	VariantID       uuid.UUID // opaque — catalog owns content
	Qty             int
	MaxQty          int // available qty rule at line level (not stock ledger)
	Notes           string
	Addons          []LineAddon
	ReplacementPref string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// LineAddon is an opaque addon attached to a line.
type LineAddon struct {
	Code  string         `json:"code"`
	Qty   int            `json:"qty"`
	Meta  map[string]any `json:"meta,omitempty"`
}

// Validate checks line invariants including qty <= max_qty.
func (l CartLine) Validate() error {
	if l.ID == uuid.Nil {
		return fmt.Errorf("%w: line id required", ErrInvalidArgument)
	}
	if l.CartID == uuid.Nil {
		return fmt.Errorf("%w: cart_id required", ErrInvalidArgument)
	}
	if l.VariantID == uuid.Nil {
		return fmt.Errorf("%w: variant_id required", ErrInvalidArgument)
	}
	if l.Qty <= 0 {
		return fmt.Errorf("%w: qty must be > 0", ErrInvalidArgument)
	}
	if l.MaxQty <= 0 {
		return fmt.Errorf("%w: max_qty must be > 0", ErrInvalidArgument)
	}
	if l.Qty > l.MaxQty {
		return fmt.Errorf("%w: qty %d > max %d", ErrMaxQtyExceeded, l.Qty, l.MaxQty)
	}
	return nil
}

// ClampQty returns qty clamped to [1, MaxQty].
func (l CartLine) ClampQty(qty int) int {
	if qty < 1 {
		return 1
	}
	if qty > l.MaxQty {
		return l.MaxQty
	}
	return qty
}

// SetQty sets qty enforcing max available rule.
func (l *CartLine) SetQty(qty int) error {
	if qty <= 0 {
		return fmt.Errorf("%w: qty must be > 0", ErrInvalidArgument)
	}
	if qty > l.MaxQty {
		return fmt.Errorf("%w: qty %d > max %d", ErrMaxQtyExceeded, qty, l.MaxQty)
	}
	l.Qty = qty
	return nil
}
