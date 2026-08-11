package domain

import (
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

// ProductKind classifies the master product.
type ProductKind string

const (
	ProductKindStandard     ProductKind = "standard"
	ProductKindBundle       ProductKind = "bundle"
	ProductKindKit          ProductKind = "kit"
	ProductKindPack         ProductKind = "pack"
	ProductKindSubscription ProductKind = "subscription"
	ProductKindDigital      ProductKind = "digital"
	ProductKindGift         ProductKind = "gift"
	ProductKindSeasonal     ProductKind = "seasonal"
	ProductKindLimited      ProductKind = "limited"
)

func (k ProductKind) Valid() bool {
	switch k {
	case ProductKindStandard, ProductKindBundle, ProductKindKit, ProductKindPack,
		ProductKindSubscription, ProductKindDigital, ProductKindGift,
		ProductKindSeasonal, ProductKindLimited:
		return true
	default:
		return false
	}
}

// ProductStatus is the publish / approval lifecycle state.
type ProductStatus string

const (
	ProductStatusDraft         ProductStatus = "draft"
	ProductStatusPendingReview ProductStatus = "pending_review"
	ProductStatusApproved      ProductStatus = "approved"
	ProductStatusPublished     ProductStatus = "published"
	ProductStatusHidden        ProductStatus = "hidden"
	ProductStatusArchived      ProductStatus = "archived"
	ProductStatusDeleted       ProductStatus = "deleted"
	ProductStatusScheduled     ProductStatus = "scheduled"
)

func (s ProductStatus) Valid() bool {
	switch s {
	case ProductStatusDraft, ProductStatusPendingReview, ProductStatusApproved,
		ProductStatusPublished, ProductStatusHidden, ProductStatusArchived,
		ProductStatusDeleted, ProductStatusScheduled:
		return true
	default:
		return false
	}
}

// Allowed product status transitions (happy path + side states).
var productTransitions = map[ProductStatus][]ProductStatus{
	ProductStatusDraft: {
		ProductStatusPendingReview, ProductStatusScheduled, ProductStatusArchived, ProductStatusDeleted,
	},
	ProductStatusPendingReview: {
		ProductStatusApproved, ProductStatusDraft, ProductStatusDeleted,
	},
	ProductStatusApproved: {
		ProductStatusPublished, ProductStatusScheduled, ProductStatusDraft, ProductStatusArchived, ProductStatusDeleted,
	},
	ProductStatusPublished: {
		ProductStatusHidden, ProductStatusArchived, ProductStatusDeleted, ProductStatusPendingReview,
	},
	ProductStatusHidden: {
		ProductStatusPublished, ProductStatusArchived, ProductStatusDeleted,
	},
	ProductStatusScheduled: {
		ProductStatusPublished, ProductStatusDraft, ProductStatusArchived, ProductStatusDeleted,
	},
	ProductStatusArchived: {
		ProductStatusDraft, ProductStatusDeleted,
	},
	ProductStatusDeleted: {},
}

// CanTransitionTo reports whether from → to is a valid product status change.
func (s ProductStatus) CanTransitionTo(to ProductStatus) bool {
	if !s.Valid() || !to.Valid() {
		return false
	}
	for _, next := range productTransitions[s] {
		if next == to {
			return true
		}
	}
	return false
}

// ValidateProductStatusTransition returns ErrInvalidTransition when disallowed.
func ValidateProductStatusTransition(from, to ProductStatus) error {
	if from == to {
		return nil
	}
	if !from.CanTransitionTo(to) {
		return fmt.Errorf("%w: %s → %s", ErrInvalidTransition, from, to)
	}
	return nil
}

var slugRegexp = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

const (
	maxSlugLen        = 200
	maxSKUCodeLen     = 64
	maxExternalRefLen = 128
)

// ValidateSlug checks kebab-case slug format used across catalog entities.
func ValidateSlug(slug string) error {
	slug = strings.TrimSpace(slug)
	if slug == "" {
		return fmt.Errorf("%w: slug required", ErrInvalidSlug)
	}
	if utf8.RuneCountInString(slug) > maxSlugLen {
		return fmt.Errorf("%w: slug too long", ErrInvalidSlug)
	}
	if !slugRegexp.MatchString(slug) {
		return fmt.Errorf("%w: must be lowercase kebab-case", ErrInvalidSlug)
	}
	return nil
}

// Product is the master product aggregate root.
// Catalog owns product truth only — no stock quantities, no sell prices, no orders.
type Product struct {
	ID               uuid.UUID
	TenantID         uuid.UUID
	BrandID          *uuid.UUID
	Kind             ProductKind
	Status           ProductStatus
	Slug             string
	SKUCode          string
	ExternalRef      string
	GTINBase         string
	ManufacturerSKU  string
	Metadata         map[string]any
	ScheduledAt      *time.Time
	PublishedAt      *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        *time.Time
}

// Validate checks structural invariants.
func (p Product) Validate() error {
	if p.ID == uuid.Nil {
		return fmt.Errorf("%w: product id required", ErrInvalidArgument)
	}
	if p.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if !p.Kind.Valid() {
		return fmt.Errorf("%w: invalid product kind %q", ErrInvalidArgument, p.Kind)
	}
	if !p.Status.Valid() {
		return fmt.Errorf("%w: invalid product status %q", ErrInvalidArgument, p.Status)
	}
	if err := ValidateSlug(p.Slug); err != nil {
		return err
	}
	if utf8.RuneCountInString(p.SKUCode) > maxSKUCodeLen {
		return fmt.Errorf("%w: sku_code too long", ErrInvalidArgument)
	}
	if utf8.RuneCountInString(p.ExternalRef) > maxExternalRefLen {
		return fmt.Errorf("%w: external_ref too long", ErrInvalidArgument)
	}
	if p.Status == ProductStatusScheduled && p.ScheduledAt == nil {
		return fmt.Errorf("%w: scheduled status requires scheduled_at", ErrInvariant)
	}
	if p.Status == ProductStatusDeleted && p.DeletedAt == nil {
		return fmt.Errorf("%w: deleted product requires deleted_at", ErrInvariant)
	}
	if p.DeletedAt != nil && p.Status != ProductStatusDeleted {
		return fmt.Errorf("%w: deleted_at set but status is %s", ErrInvariant, p.Status)
	}
	if p.BrandID != nil && *p.BrandID == uuid.Nil {
		return fmt.Errorf("%w: brand_id must not be nil UUID", ErrInvalidArgument)
	}
	return nil
}

// IsPublished reports whether the product is customer-visible.
func (p Product) IsPublished() bool {
	return p.Status == ProductStatusPublished && p.DeletedAt == nil
}

// IsEditable reports whether normal author edits are allowed.
func (p Product) IsEditable() bool {
	switch p.Status {
	case ProductStatusDraft, ProductStatusPendingReview, ProductStatusApproved, ProductStatusScheduled:
		return p.DeletedAt == nil
	default:
		return false
	}
}

// TransitionTo applies a validated status change (caller persists).
func (p *Product) TransitionTo(to ProductStatus, now time.Time) error {
	if err := ValidateProductStatusTransition(p.Status, to); err != nil {
		return err
	}
	p.Status = to
	p.UpdatedAt = now
	switch to {
	case ProductStatusPublished:
		p.PublishedAt = &now
		p.ScheduledAt = nil
	case ProductStatusDeleted:
		p.DeletedAt = &now
	case ProductStatusDraft, ProductStatusPendingReview, ProductStatusApproved:
		p.DeletedAt = nil
	}
	return nil
}
