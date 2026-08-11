package domain

import (
	"fmt"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

// VariantStatus is the lifecycle of a product variant.
type VariantStatus string

const (
	VariantStatusDraft    VariantStatus = "draft"
	VariantStatusActive   VariantStatus = "active"
	VariantStatusInactive VariantStatus = "inactive"
	VariantStatusArchived VariantStatus = "archived"
	VariantStatusDeleted  VariantStatus = "deleted"
)

func (s VariantStatus) Valid() bool {
	switch s {
	case VariantStatusDraft, VariantStatusActive, VariantStatusInactive,
		VariantStatusArchived, VariantStatusDeleted:
		return true
	default:
		return false
	}
}

var variantTransitions = map[VariantStatus][]VariantStatus{
	VariantStatusDraft: {
		VariantStatusActive, VariantStatusArchived, VariantStatusDeleted,
	},
	VariantStatusActive: {
		VariantStatusInactive, VariantStatusArchived, VariantStatusDeleted,
	},
	VariantStatusInactive: {
		VariantStatusActive, VariantStatusArchived, VariantStatusDeleted,
	},
	VariantStatusArchived: {
		VariantStatusDraft, VariantStatusDeleted,
	},
	VariantStatusDeleted: {},
}

// CanTransitionTo reports whether from → to is allowed for variants.
func (s VariantStatus) CanTransitionTo(to VariantStatus) bool {
	if !s.Valid() || !to.Valid() {
		return false
	}
	for _, next := range variantTransitions[s] {
		if next == to {
			return true
		}
	}
	return false
}

// ValidateVariantStatusTransition returns ErrInvalidTransition when disallowed.
func ValidateVariantStatusTransition(from, to VariantStatus) error {
	if from == to {
		return nil
	}
	if !from.CanTransitionTo(to) {
		return fmt.Errorf("%w: %s → %s", ErrInvalidTransition, from, to)
	}
	return nil
}

const maxVariantNameLen = 200

// Variant is a sellable option combination under a product.
// Stock quantities are owned by inventory-service.
type Variant struct {
	ID           uuid.UUID
	ProductID    uuid.UUID
	TenantID     uuid.UUID
	SKUCode      string
	Name         string
	OptionValues map[string]any
	Status       VariantStatus
	SortOrder    int
	BarcodeHint  string
	Metadata     map[string]any
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time
}

// Validate checks structural invariants.
func (v Variant) Validate() error {
	if v.ID == uuid.Nil {
		return fmt.Errorf("%w: variant id required", ErrInvalidArgument)
	}
	if v.ProductID == uuid.Nil {
		return fmt.Errorf("%w: product_id required", ErrInvalidArgument)
	}
	if v.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if !v.Status.Valid() {
		return fmt.Errorf("%w: invalid variant status %q", ErrInvalidArgument, v.Status)
	}
	if utf8.RuneCountInString(v.SKUCode) > maxSKUCodeLen {
		return fmt.Errorf("%w: sku_code too long", ErrInvalidArgument)
	}
	if utf8.RuneCountInString(v.Name) > maxVariantNameLen {
		return fmt.Errorf("%w: variant name too long", ErrInvalidArgument)
	}
	if v.OptionValues == nil {
		return fmt.Errorf("%w: option_values required", ErrInvalidArgument)
	}
	if v.Status == VariantStatusDeleted && v.DeletedAt == nil {
		return fmt.Errorf("%w: deleted variant requires deleted_at", ErrInvariant)
	}
	if v.DeletedAt != nil && v.Status != VariantStatusDeleted {
		return fmt.Errorf("%w: deleted_at set but status is %s", ErrInvariant, v.Status)
	}
	return nil
}
