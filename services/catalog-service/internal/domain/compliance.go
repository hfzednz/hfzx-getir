package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

const maxAgeRestriction = 99

// ProductCompliance holds regulatory / compliance flags for a product.
type ProductCompliance struct {
	ID                   uuid.UUID
	ProductID            uuid.UUID
	TenantID             uuid.UUID
	AgeRestriction       int
	IsHazardous          bool
	HazardClass          string
	IsPharmacy           bool
	RequiresPrescription bool
	IsFood               bool
	IsOrganic            bool
	IsHalal              bool
	IsVegan              bool
	IsGlutenFree         bool
	RestrictedCountries  []string
	AllowedCountries     []string
	Certificates         []map[string]any
	Metadata             map[string]any
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// Validate checks structural invariants.
func (c ProductCompliance) Validate() error {
	if c.ID == uuid.Nil {
		return fmt.Errorf("%w: compliance id required", ErrInvalidArgument)
	}
	if c.ProductID == uuid.Nil {
		return fmt.Errorf("%w: product_id required", ErrInvalidArgument)
	}
	if c.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if c.AgeRestriction < 0 || c.AgeRestriction > maxAgeRestriction {
		return fmt.Errorf("%w: age_restriction out of range", ErrInvalidArgument)
	}
	if c.RequiresPrescription && !c.IsPharmacy {
		return fmt.Errorf("%w: prescription requires is_pharmacy", ErrInvariant)
	}
	for _, cc := range c.RestrictedCountries {
		if len(cc) != 2 {
			return fmt.Errorf("%w: restricted country %q must be ISO-3166 alpha-2", ErrInvalidArgument, cc)
		}
	}
	for _, cc := range c.AllowedCountries {
		if len(cc) != 2 {
			return fmt.Errorf("%w: allowed country %q must be ISO-3166 alpha-2", ErrInvalidArgument, cc)
		}
	}
	if c.Certificates == nil {
		return fmt.Errorf("%w: certificates required (may be empty)", ErrInvalidArgument)
	}
	return nil
}
