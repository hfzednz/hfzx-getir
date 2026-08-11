package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// PickLineStatus is the status of a line within a pick session.
type PickLineStatus string

const (
	PickLineStatusPending  PickLineStatus = "pending"
	PickLineStatusPartial  PickLineStatus = "partial"
	PickLineStatusComplete PickLineStatus = "complete"
	PickLineStatusShort    PickLineStatus = "short"
)

func (s PickLineStatus) Valid() bool {
	switch s {
	case PickLineStatusPending, PickLineStatusPartial, PickLineStatusComplete, PickLineStatusShort:
		return true
	default:
		return false
	}
}

// PickRouteStep is one ordered stop in a pick route (opaque location_code).
type PickRouteStep struct {
	LineID       uuid.UUID `json:"lineId"`
	LocationCode string    `json:"locationCode"`
	Seq          int       `json:"seq"`
	VariantID    uuid.UUID `json:"variantId,omitempty"`
	Qty          int       `json:"qty,omitempty"`
}

// PickLine is a pick-session line with scan progress.
type PickLine struct {
	ID           uuid.UUID
	SessionID    uuid.UUID
	LineID       uuid.UUID // fulfillment line id
	VariantID    uuid.UUID
	SKUCode      string
	Barcode      string // expected barcode (opaque)
	LocationCode string
	QtyRequired  int64
	QtyPicked    int64
	Sequence     int
	Status       PickLineStatus
}

// QtyRemaining returns units still to pick.
func (l PickLine) QtyRemaining() int64 {
	return l.QtyRequired - l.QtyPicked
}

// PickSession is an active pick run bound to a pick task.
type PickSession struct {
	ID            uuid.UUID
	TenantID      uuid.UUID
	TaskID        uuid.UUID
	WarehouseID   uuid.UUID
	FulfillmentID uuid.UUID
	PickerID      *uuid.UUID
	Strategy      PickStrategy
	Route         []PickRouteStep
	Lines         []PickLine
	StartedAt     *time.Time
	CompletedAt   *time.Time
	Metadata      map[string]any
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// PickScan records a single scan attempt against a fulfillment/pick line.
type PickScan struct {
	ID          uuid.UUID
	SessionID   uuid.UUID
	LineID      uuid.UUID
	ScannedCode string
	Qty         int64
	OK          bool
	Reason      string
	ScannedBy   *uuid.UUID
	ScannedAt   time.Time
	Metadata    map[string]any
	CreatedAt   time.Time
}

// Validate checks pick session invariants.
func (s PickSession) Validate() error {
	if s.ID == uuid.Nil {
		return fmt.Errorf("%w: pick session id required", ErrInvalidArgument)
	}
	if s.TaskID == uuid.Nil {
		return fmt.Errorf("%w: task_id required", ErrInvalidArgument)
	}
	if s.WarehouseID == uuid.Nil {
		return fmt.Errorf("%w: warehouse_id required", ErrInvalidArgument)
	}
	if s.Strategy != "" && !s.Strategy.Valid() {
		return fmt.Errorf("%w: invalid pick strategy %q", ErrInvalidArgument, s.Strategy)
	}
	return nil
}

// Start marks the session started.
func (s *PickSession) Start() error {
	if s.StartedAt != nil {
		return fmt.Errorf("%w: pick session already started", ErrConflict)
	}
	if s.CompletedAt != nil {
		return fmt.Errorf("%w: pick session already completed", ErrAlreadyTerminal)
	}
	now := time.Now().UTC()
	s.StartedAt = &now
	s.UpdatedAt = now
	return nil
}

// Complete marks the session completed.
func (s *PickSession) Complete() error {
	if s.StartedAt == nil {
		return fmt.Errorf("%w: pick session not started", ErrInvariant)
	}
	if s.CompletedAt != nil {
		return fmt.Errorf("%w: pick session already completed", ErrAlreadyTerminal)
	}
	for _, l := range s.Lines {
		if l.QtyPicked < l.QtyRequired {
			return fmt.Errorf("%w: line %s", ErrRemainingQty, l.ID)
		}
	}
	now := time.Now().UTC()
	s.CompletedAt = &now
	s.UpdatedAt = now
	return nil
}

// SetRoute replaces the ordered pick route (from AI OptimizePickRoute port).
func (s *PickSession) SetRoute(steps []PickRouteStep) error {
	if s.CompletedAt != nil {
		return fmt.Errorf("%w: cannot change route on completed session", ErrAlreadyTerminal)
	}
	for i := range steps {
		if steps[i].LineID == uuid.Nil {
			return fmt.Errorf("%w: route step[%d] line_id required", ErrInvalidArgument, i)
		}
		if steps[i].Seq <= 0 {
			steps[i].Seq = i + 1
		}
	}
	s.Route = steps
	s.UpdatedAt = time.Now().UTC()
	return nil
}

// SortRouteByLocation groups/sorts by location_code then seq (simple zone heuristic).
func SortRouteByLocation(steps []PickRouteStep) []PickRouteStep {
	out := make([]PickRouteStep, len(steps))
	copy(out, steps)
	for i := 1; i < len(out); i++ {
		j := i
		for j > 0 {
			a, b := out[j-1], out[j]
			if a.LocationCode < b.LocationCode ||
				(a.LocationCode == b.LocationCode && a.Seq <= b.Seq) {
				break
			}
			out[j-1], out[j] = b, a
			j--
		}
	}
	for i := range out {
		out[i].Seq = i + 1
	}
	return out
}

// ValidatePickScan checks expected barcode against scanned value and remaining qty.
// Returns nil when the scan is accepted; otherwise a domain error.
func ValidatePickScan(expectedBarcode, scanned string, qtyRemaining int) error {
	expectedBarcode = strings.TrimSpace(expectedBarcode)
	scanned = strings.TrimSpace(scanned)

	if scanned == "" {
		return fmt.Errorf("%w: scanned code empty", ErrScanRejected)
	}
	if qtyRemaining <= 0 {
		return fmt.Errorf("%w: no quantity remaining", ErrOverpick)
	}
	if expectedBarcode == "" {
		// No expected barcode configured — accept any non-empty scan (shelf/RFID stub).
		return nil
	}
	if !strings.EqualFold(expectedBarcode, scanned) {
		return fmt.Errorf("%w: expected %q got %q", ErrBarcodeMismatch, expectedBarcode, scanned)
	}
	return nil
}

// ApplyPickLineScan validates and increments a pick-session line.
func ApplyPickLineScan(line *PickLine, scanned string, qty int64) (*PickScan, error) {
	if line == nil {
		return nil, fmt.Errorf("%w: line required", ErrInvalidArgument)
	}
	if qty <= 0 {
		return nil, fmt.Errorf("%w: scan qty must be positive", ErrInvalidArgument)
	}
	remaining := line.QtyRemaining()
	if qty > remaining {
		scan := &PickScan{
			ID: uuid.New(), LineID: line.ID, ScannedCode: scanned, Qty: qty,
			OK: false, Reason: "overpick", ScannedAt: time.Now().UTC(), CreatedAt: time.Now().UTC(),
		}
		return scan, fmt.Errorf("%w: want %d remaining %d", ErrOverpick, qty, remaining)
	}
	if err := ValidatePickScan(line.Barcode, scanned, int(remaining)); err != nil {
		reason := "rejected"
		switch {
		case errors.Is(err, ErrBarcodeMismatch):
			reason = "barcode_mismatch"
		case errors.Is(err, ErrOverpick):
			reason = "overpick"
		case errors.Is(err, ErrScanRejected):
			reason = "empty_scan"
		}
		scan := &PickScan{
			ID: uuid.New(), LineID: line.ID, ScannedCode: scanned, Qty: qty,
			OK: false, Reason: reason, ScannedAt: time.Now().UTC(), CreatedAt: time.Now().UTC(),
		}
		return scan, err
	}
	line.QtyPicked += qty
	if line.QtyPicked >= line.QtyRequired {
		line.Status = PickLineStatusComplete
	} else {
		line.Status = PickLineStatusPartial
	}
	scan := &PickScan{
		ID: uuid.New(), LineID: line.ID, ScannedCode: scanned, Qty: qty,
		OK: true, ScannedAt: time.Now().UTC(), CreatedAt: time.Now().UTC(),
	}
	return scan, nil
}

// ApplyScan validates and increments a fulfillment line qty_picked.
func ApplyScan(line *FulfillmentLine, scanned string, qty int64) (*PickScan, error) {
	if line == nil {
		return nil, fmt.Errorf("%w: line required", ErrInvalidArgument)
	}
	if qty <= 0 {
		return nil, fmt.Errorf("%w: scan qty must be positive", ErrInvalidArgument)
	}
	remaining := line.QtyRemaining()
	if qty > remaining {
		scan := &PickScan{
			ID: uuid.New(), LineID: line.ID, ScannedCode: scanned, Qty: qty,
			OK: false, Reason: "overpick", ScannedAt: time.Now().UTC(), CreatedAt: time.Now().UTC(),
		}
		return scan, fmt.Errorf("%w: want %d remaining %d", ErrOverpick, qty, remaining)
	}
	if err := ValidatePickScan(line.ExpectedBarcode(), scanned, int(remaining)); err != nil {
		reason := "rejected"
		switch {
		case errors.Is(err, ErrBarcodeMismatch):
			reason = "barcode_mismatch"
		case errors.Is(err, ErrOverpick):
			reason = "overpick"
		case errors.Is(err, ErrScanRejected):
			reason = "empty_scan"
		}
		scan := &PickScan{
			ID: uuid.New(), LineID: line.ID, ScannedCode: scanned, Qty: qty,
			OK: false, Reason: reason, ScannedAt: time.Now().UTC(), CreatedAt: time.Now().UTC(),
		}
		return scan, err
	}
	line.QtyPicked += qty
	line.UpdatedAt = time.Now().UTC()
	scan := &PickScan{
		ID: uuid.New(), LineID: line.ID, ScannedCode: scanned, Qty: qty,
		OK: true, ScannedAt: time.Now().UTC(), CreatedAt: time.Now().UTC(),
	}
	return scan, nil
}

// StrategyForFlags picks a default strategy from fulfillment flags.
func StrategyForFlags(express, vip bool, priority int) PickStrategy {
	switch {
	case express:
		return PickStrategyExpress
	case vip || priority >= 100:
		return PickStrategyPriority
	default:
		return PickStrategySingle
	}
}
