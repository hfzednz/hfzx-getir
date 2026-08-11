package domain

import (
	"fmt"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

// BundleCompositionType distinguishes static vs dynamic kits.
type BundleCompositionType string

const (
	BundleCompositionStatic  BundleCompositionType = "static"
	BundleCompositionDynamic BundleCompositionType = "dynamic"
)

func (t BundleCompositionType) Valid() bool {
	switch t {
	case BundleCompositionStatic, BundleCompositionDynamic:
		return true
	default:
		return false
	}
}

const maxBundleNameLen = 200

// Bundle is the composition header for a bundle/kit/pack product.
// Qty on items is BOM composition, not warehouse stock.
type Bundle struct {
	ID          uuid.UUID
	ProductID   uuid.UUID
	TenantID    uuid.UUID
	Composition BundleCompositionType
	Name        string
	Metadata    map[string]any
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Validate checks structural invariants.
func (b Bundle) Validate() error {
	if b.ID == uuid.Nil {
		return fmt.Errorf("%w: bundle id required", ErrInvalidArgument)
	}
	if b.ProductID == uuid.Nil {
		return fmt.Errorf("%w: product_id required", ErrInvalidArgument)
	}
	if b.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if !b.Composition.Valid() {
		return fmt.Errorf("%w: invalid bundle composition %q", ErrInvalidArgument, b.Composition)
	}
	if utf8.RuneCountInString(b.Name) > maxBundleNameLen {
		return fmt.Errorf("%w: bundle name too long", ErrInvalidArgument)
	}
	return nil
}

// BundleItem is a component variant and composition quantity.
type BundleItem struct {
	ID                  uuid.UUID
	BundleID            uuid.UUID
	ComponentVariantID  uuid.UUID
	Qty                 int
	IsOptional          bool
	SortOrder           int
	Metadata            map[string]any
	CreatedAt           time.Time
}

// Validate checks structural invariants.
func (i BundleItem) Validate() error {
	if i.ID == uuid.Nil {
		return fmt.Errorf("%w: bundle_item id required", ErrInvalidArgument)
	}
	if i.BundleID == uuid.Nil {
		return fmt.Errorf("%w: bundle_id required", ErrInvalidArgument)
	}
	if i.ComponentVariantID == uuid.Nil {
		return fmt.Errorf("%w: component_variant_id required", ErrInvalidArgument)
	}
	if i.Qty <= 0 {
		return fmt.Errorf("%w: qty must be > 0", ErrInvalidArgument)
	}
	return nil
}
