package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// OrderLine is a priced line snapshot on the order aggregate.
// Variant/SKU/warehouse are opaque catalog/warehouse refs.
type OrderLine struct {
	ID             uuid.UUID
	OrderID        uuid.UUID
	TenantID       uuid.UUID
	VariantID      uuid.UUID
	SKUCode        string
	TitleSnapshot  string
	Qty            int
	UnitPriceMinor int64
	DiscountsMinor int64
	TaxMinor       int64
	LineTotalMinor int64
	WarehouseID    *uuid.UUID
	SortOrder      int
	Metadata       map[string]any
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// Validate checks structural invariants for a line.
func (l OrderLine) Validate() error {
	if l.ID == uuid.Nil {
		return fmt.Errorf("%w: order line id required", ErrInvalidArgument)
	}
	if l.OrderID == uuid.Nil {
		return fmt.Errorf("%w: order_id required", ErrInvalidArgument)
	}
	if l.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if l.VariantID == uuid.Nil {
		return fmt.Errorf("%w: variant_id required", ErrInvalidArgument)
	}
	if l.Qty <= 0 {
		return fmt.Errorf("%w: qty must be > 0", ErrInvalidArgument)
	}
	if l.UnitPriceMinor < 0 || l.DiscountsMinor < 0 || l.TaxMinor < 0 || l.LineTotalMinor < 0 {
		return fmt.Errorf("%w: money fields must be non-negative", ErrNegativeMoney)
	}
	return nil
}

// UnitPrice returns the unit price as Money in the given currency.
func (l OrderLine) UnitPrice(currency string) (Money, error) {
	return NewMoney(l.UnitPriceMinor, currency)
}

// LineTotal returns the line total as Money in the given currency.
func (l OrderLine) LineTotal(currency string) (Money, error) {
	return NewMoney(l.LineTotalMinor, currency)
}

// ComputeLineTotalMinor recomputes qty*unit - discounts + tax in minor units.
func ComputeLineTotalMinor(qty int, unitPriceMinor, discountsMinor, taxMinor int64) (int64, error) {
	if qty <= 0 {
		return 0, fmt.Errorf("%w: qty must be > 0", ErrInvalidArgument)
	}
	if unitPriceMinor < 0 || discountsMinor < 0 || taxMinor < 0 {
		return 0, fmt.Errorf("%w: money fields must be non-negative", ErrNegativeMoney)
	}
	gross := unitPriceMinor * int64(qty)
	if unitPriceMinor > 0 && gross/unitPriceMinor != int64(qty) {
		return 0, fmt.Errorf("%w: line total overflow", ErrInvariant)
	}
	if discountsMinor > gross {
		return 0, fmt.Errorf("%w: discounts exceed gross", ErrInvariant)
	}
	total := gross - discountsMinor + taxMinor
	if total < 0 {
		return 0, fmt.Errorf("%w: line total negative", ErrNegativeMoney)
	}
	return total, nil
}
