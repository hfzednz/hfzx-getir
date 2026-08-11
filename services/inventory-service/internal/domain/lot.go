package domain

import (
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
)

// LotStatus is the quality/disposition state of a lot/batch.
type LotStatus string

const (
	LotStatusGood       LotStatus = "good"
	LotStatusDamaged    LotStatus = "damaged"
	LotStatusQuarantine LotStatus = "quarantine"
	LotStatusExpired    LotStatus = "expired"
)

func (s LotStatus) Valid() bool {
	switch s {
	case LotStatusGood, LotStatusDamaged, LotStatusQuarantine, LotStatusExpired:
		return true
	default:
		return false
	}
}

// Allocatable reports whether the lot can be reserved/picked.
func (s LotStatus) Allocatable() bool {
	return s == LotStatusGood
}

// Lot is a batch under a stock balance (FEFO/FIFO ranked by expiry).
type Lot struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	BalanceID   uuid.UUID
	WarehouseID uuid.UUID
	VariantID   uuid.UUID
	LotCode     string
	Qty         int64
	ExpiryDate  *time.Time
	MfgDate     *time.Time
	Status      LotStatus
	ReceivedAt  *time.Time
	Metadata    map[string]any
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Validate checks structural invariants.
func (l Lot) Validate() error {
	if l.ID == uuid.Nil {
		return fmt.Errorf("%w: lot id required", ErrInvalidArgument)
	}
	if l.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if l.BalanceID == uuid.Nil {
		return fmt.Errorf("%w: balance_id required", ErrInvalidArgument)
	}
	if l.WarehouseID == uuid.Nil {
		return fmt.Errorf("%w: warehouse_id required", ErrInvalidArgument)
	}
	if l.VariantID == uuid.Nil {
		return fmt.Errorf("%w: variant_id required", ErrInvalidArgument)
	}
	if l.LotCode == "" {
		return fmt.Errorf("%w: lot_code required", ErrInvalidArgument)
	}
	if l.Qty < 0 {
		return fmt.Errorf("%w: lot qty cannot be negative", ErrInvariant)
	}
	if !l.Status.Valid() {
		return fmt.Errorf("%w: invalid lot status %q", ErrInvalidArgument, l.Status)
	}
	if l.MfgDate != nil && l.ExpiryDate != nil && l.MfgDate.After(*l.ExpiryDate) {
		return fmt.Errorf("%w: mfg_date after expiry_date", ErrInvariant)
	}
	return nil
}

// IsExpiredAt reports whether the lot is past expiry at asOf (date-granular).
func (l Lot) IsExpiredAt(asOf time.Time) bool {
	if l.Status == LotStatusExpired {
		return true
	}
	if l.ExpiryDate == nil {
		return false
	}
	asOfDay := time.Date(asOf.Year(), asOf.Month(), asOf.Day(), 0, 0, 0, 0, time.UTC)
	exp := time.Date(l.ExpiryDate.Year(), l.ExpiryDate.Month(), l.ExpiryDate.Day(), 0, 0, 0, 0, time.UTC)
	return asOfDay.After(exp)
}

// Deduct reduces lot quantity (pick / consume).
func (l *Lot) Deduct(qty int64) error {
	if qty <= 0 {
		return fmt.Errorf("%w: deduct qty", ErrNegativeQuantity)
	}
	if !l.Status.Allocatable() {
		return fmt.Errorf("%w: status %s", ErrLotNotAllocatable, l.Status)
	}
	if qty > l.Qty {
		return fmt.Errorf("%w: deduct %d exceeds lot qty %d", ErrInsufficientStock, qty, l.Qty)
	}
	l.Qty -= qty
	l.UpdatedAt = time.Now().UTC()
	return nil
}

// SortLotsFEFO sorts lots in-place by earliest expiry first (nil expiry last),
// then by received_at, then by lot_code for stability.
func SortLotsFEFO(lots []Lot) {
	sort.SliceStable(lots, func(i, j int) bool {
		return compareLotsFEFO(lots[i], lots[j]) < 0
	})
}

func compareLotsFEFO(a, b Lot) int {
	switch {
	case a.ExpiryDate == nil && b.ExpiryDate == nil:
		// fall through
	case a.ExpiryDate == nil:
		return 1
	case b.ExpiryDate == nil:
		return -1
	default:
		if a.ExpiryDate.Before(*b.ExpiryDate) {
			return -1
		}
		if a.ExpiryDate.After(*b.ExpiryDate) {
			return 1
		}
	}
	switch {
	case a.ReceivedAt == nil && b.ReceivedAt == nil:
		// fall through
	case a.ReceivedAt == nil:
		return 1
	case b.ReceivedAt == nil:
		return -1
	default:
		if a.ReceivedAt.Before(*b.ReceivedAt) {
			return -1
		}
		if a.ReceivedAt.After(*b.ReceivedAt) {
			return 1
		}
	}
	if a.LotCode < b.LotCode {
		return -1
	}
	if a.LotCode > b.LotCode {
		return 1
	}
	return 0
}

// LotAllocation is a FEFO pick result for a single lot.
type LotAllocation struct {
	LotID uuid.UUID
	Qty   int64
}

// AllocateLotsFEFO allocates qty across allocatable lots by earliest expiry first.
// Returns ErrInsufficientStock when total good qty is less than requested.
func AllocateLotsFEFO(lots []Lot, qty int64, asOf time.Time) ([]LotAllocation, error) {
	if qty <= 0 {
		return nil, fmt.Errorf("%w: allocate qty", ErrNegativeQuantity)
	}
	candidates := make([]Lot, 0, len(lots))
	for _, l := range lots {
		if !l.Status.Allocatable() || l.Qty <= 0 || l.IsExpiredAt(asOf) {
			continue
		}
		candidates = append(candidates, l)
	}
	SortLotsFEFO(candidates)

	remaining := qty
	out := make([]LotAllocation, 0, len(candidates))
	for _, l := range candidates {
		if remaining == 0 {
			break
		}
		take := l.Qty
		if take > remaining {
			take = remaining
		}
		out = append(out, LotAllocation{LotID: l.ID, Qty: take})
		remaining -= take
	}
	if remaining > 0 {
		return nil, fmt.Errorf("%w: shortfall %d", ErrInsufficientStock, remaining)
	}
	return out, nil
}
