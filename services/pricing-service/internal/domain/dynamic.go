package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// DynamicKind is percent or fixed adjustment.
type DynamicKind string

const (
	DynamicKindPercent DynamicKind = "percent"
	DynamicKindFixed   DynamicKind = "fixed"
)

// Valid reports whether the kind is recognized.
func (k DynamicKind) Valid() bool {
	switch k {
	case DynamicKindPercent, DynamicKindFixed:
		return true
	default:
		return false
	}
}

// DynamicTrigger is when a dynamic rule fires.
type DynamicTrigger string

const (
	TriggerTimeOfDay     DynamicTrigger = "time_of_day"
	TriggerInventoryHint DynamicTrigger = "inventory_hint"
)

// Valid reports whether the trigger is recognized.
func (t DynamicTrigger) Valid() bool {
	switch t {
	case TriggerTimeOfDay, TriggerInventoryHint:
		return true
	default:
		return false
	}
}

// DynamicRule adjusts resolved unit prices at quote time.
type DynamicRule struct {
	ID                 uuid.UUID
	TenantID           uuid.UUID
	Code               string
	Kind               DynamicKind
	Trigger            DynamicTrigger
	AdjustmentBps      int   // percent kind (can be negative for discount bump)
	AdjustmentMinor    int64 // fixed kind (can be negative)
	StartHour          int   // inclusive 0-23 for time_of_day
	EndHour            int   // exclusive 0-24 for time_of_day
	InventoryThreshold int   // fire when hint available qty <= threshold
	Active             bool
	Priority           int
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// Validate checks dynamic rule invariants.
func (r DynamicRule) Validate() error {
	if r.ID == uuid.Nil {
		return fmt.Errorf("%w: dynamic_rule id required", ErrInvalidArgument)
	}
	if r.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if strings.TrimSpace(r.Code) == "" {
		return fmt.Errorf("%w: code required", ErrInvalidArgument)
	}
	if !r.Kind.Valid() {
		return fmt.Errorf("%w: invalid kind %q", ErrInvalidArgument, r.Kind)
	}
	if !r.Trigger.Valid() {
		return fmt.Errorf("%w: invalid trigger %q", ErrInvalidArgument, r.Trigger)
	}
	if r.Trigger == TriggerTimeOfDay {
		if r.StartHour < 0 || r.StartHour > 23 || r.EndHour < 1 || r.EndHour > 24 || r.StartHour >= r.EndHour {
			return fmt.Errorf("%w: invalid time window [%d,%d)", ErrInvalidArgument, r.StartHour, r.EndHour)
		}
	}
	return nil
}

// AppliesAt reports whether the rule should fire given time and optional inventory qty.
func (r DynamicRule) AppliesAt(at time.Time, inventoryQty *int) bool {
	if !r.Active {
		return false
	}
	switch r.Trigger {
	case TriggerTimeOfDay:
		h := at.UTC().Hour()
		return h >= r.StartHour && h < r.EndHour
	case TriggerInventoryHint:
		if inventoryQty == nil {
			return false
		}
		return *inventoryQty <= r.InventoryThreshold
	default:
		return false
	}
}

// ApplyToUnit adjusts a unit price (minor units). Result clamped to >= 0.
func (r DynamicRule) ApplyToUnit(unitMinor int64) int64 {
	var out int64
	switch r.Kind {
	case DynamicKindPercent:
		out = unitMinor + unitMinor*int64(r.AdjustmentBps)/10000
	case DynamicKindFixed:
		out = unitMinor + r.AdjustmentMinor
	default:
		return unitMinor
	}
	if out < 0 {
		return 0
	}
	return out
}

// SelectDynamicRules returns active applicable rules sorted by priority DESC.
func SelectDynamicRules(rules []DynamicRule, at time.Time, inventoryQty *int) []DynamicRule {
	var out []DynamicRule
	for _, r := range rules {
		if r.AppliesAt(at, inventoryQty) {
			out = append(out, r)
		}
	}
	// insertion sort by priority DESC
	for i := 1; i < len(out); i++ {
		j := i
		for j > 0 && out[j].Priority > out[j-1].Priority {
			out[j], out[j-1] = out[j-1], out[j]
			j--
		}
	}
	return out
}
