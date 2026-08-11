package domain

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

// SupplierStatus is the lifecycle of a supplier.
type SupplierStatus string

const (
	SupplierStatusActive   SupplierStatus = "active"
	SupplierStatusInactive SupplierStatus = "inactive"
	SupplierStatusArchived SupplierStatus = "archived"
)

func (s SupplierStatus) Valid() bool {
	switch s {
	case SupplierStatusActive, SupplierStatusInactive, SupplierStatusArchived:
		return true
	default:
		return false
	}
}

const (
	maxSupplierCodeLen = 64
	maxSupplierNameLen = 200
)

// Supplier is a supplier master record (no settlement contracts).
type Supplier struct {
	ID           uuid.UUID
	TenantID     uuid.UUID
	Code         string
	Name         string
	ContactEmail string
	ContactPhone string
	CountryCode  string
	ExternalRef  string
	Metadata     map[string]any
	Status       SupplierStatus
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time
}

// Validate checks structural invariants.
func (s Supplier) Validate() error {
	if s.ID == uuid.Nil {
		return fmt.Errorf("%w: supplier id required", ErrInvalidArgument)
	}
	if s.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if strings.TrimSpace(s.Code) == "" {
		return fmt.Errorf("%w: supplier code required", ErrInvalidArgument)
	}
	if utf8.RuneCountInString(s.Code) > maxSupplierCodeLen {
		return fmt.Errorf("%w: supplier code too long", ErrInvalidArgument)
	}
	if strings.TrimSpace(s.Name) == "" {
		return fmt.Errorf("%w: supplier name required", ErrInvalidArgument)
	}
	if utf8.RuneCountInString(s.Name) > maxSupplierNameLen {
		return fmt.Errorf("%w: supplier name too long", ErrInvalidArgument)
	}
	if !s.Status.Valid() {
		return fmt.Errorf("%w: invalid supplier status %q", ErrInvalidArgument, s.Status)
	}
	if cc := s.CountryCode; cc != "" && len(cc) != 2 {
		return fmt.Errorf("%w: country_code must be ISO-3166 alpha-2", ErrInvalidArgument)
	}
	return nil
}

// SupplierProduct links a supplier to a product/variant with metadata only.
// CostHintMinor is a supplier quote in minor units — NOT a sellable price.
type SupplierProduct struct {
	ID               uuid.UUID
	SupplierID       uuid.UUID
	ProductID        uuid.UUID
	VariantID        *uuid.UUID
	TenantID         uuid.UUID
	SupplierSKU      string
	CostHintMinor    *int64 // supplier quote; labeled as quote, not sell price
	CostHintCurrency string
	LeadTimeDays     *int
	MOQ              *int
	Metadata         map[string]any
	IsPreferred      bool
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// Validate checks structural invariants.
func (sp SupplierProduct) Validate() error {
	if sp.ID == uuid.Nil {
		return fmt.Errorf("%w: supplier_product id required", ErrInvalidArgument)
	}
	if sp.SupplierID == uuid.Nil {
		return fmt.Errorf("%w: supplier_id required", ErrInvalidArgument)
	}
	if sp.ProductID == uuid.Nil {
		return fmt.Errorf("%w: product_id required", ErrInvalidArgument)
	}
	if sp.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if sp.CostHintMinor != nil && *sp.CostHintMinor < 0 {
		return fmt.Errorf("%w: cost_hint_minor must be >= 0", ErrInvalidArgument)
	}
	if sp.CostHintMinor != nil && sp.CostHintCurrency != "" && len(sp.CostHintCurrency) != 3 {
		return fmt.Errorf("%w: cost_hint_currency must be ISO-4217", ErrInvalidArgument)
	}
	if sp.LeadTimeDays != nil && *sp.LeadTimeDays < 0 {
		return fmt.Errorf("%w: lead_time_days must be >= 0", ErrInvalidArgument)
	}
	if sp.MOQ != nil && *sp.MOQ < 0 {
		return fmt.Errorf("%w: moq must be >= 0", ErrInvalidArgument)
	}
	if sp.VariantID != nil && *sp.VariantID == uuid.Nil {
		return fmt.Errorf("%w: variant_id must not be nil UUID", ErrInvalidArgument)
	}
	return nil
}
