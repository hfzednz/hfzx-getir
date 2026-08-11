package domain

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

// BrandStatus is the lifecycle of a brand.
type BrandStatus string

const (
	BrandStatusActive   BrandStatus = "active"
	BrandStatusInactive BrandStatus = "inactive"
	BrandStatusArchived BrandStatus = "archived"
)

func (s BrandStatus) Valid() bool {
	switch s {
	case BrandStatusActive, BrandStatusInactive, BrandStatusArchived:
		return true
	default:
		return false
	}
}

const (
	maxBrandNameLen = 200
	maxURLLen       = 2048
)

// Brand is a manufacturer / brand master record.
type Brand struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	Name        string
	Slug        string
	Description string
	LogoURL     string
	WebsiteURL  string
	CountryCode string
	ExternalRef string
	Metadata    map[string]any
	Status      BrandStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}

// Validate checks structural invariants.
func (b Brand) Validate() error {
	if b.ID == uuid.Nil {
		return fmt.Errorf("%w: brand id required", ErrInvalidArgument)
	}
	if b.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if strings.TrimSpace(b.Name) == "" {
		return fmt.Errorf("%w: brand name required", ErrInvalidArgument)
	}
	if utf8.RuneCountInString(b.Name) > maxBrandNameLen {
		return fmt.Errorf("%w: brand name too long", ErrInvalidArgument)
	}
	if err := ValidateSlug(b.Slug); err != nil {
		return err
	}
	if !b.Status.Valid() {
		return fmt.Errorf("%w: invalid brand status %q", ErrInvalidArgument, b.Status)
	}
	if utf8.RuneCountInString(b.LogoURL) > maxURLLen {
		return fmt.Errorf("%w: logo_url too long", ErrInvalidArgument)
	}
	if utf8.RuneCountInString(b.WebsiteURL) > maxURLLen {
		return fmt.Errorf("%w: website_url too long", ErrInvalidArgument)
	}
	if cc := b.CountryCode; cc != "" && len(cc) != 2 {
		return fmt.Errorf("%w: country_code must be ISO-3166 alpha-2", ErrInvalidArgument)
	}
	if utf8.RuneCountInString(b.ExternalRef) > maxExternalRefLen {
		return fmt.Errorf("%w: external_ref too long", ErrInvalidArgument)
	}
	return nil
}
