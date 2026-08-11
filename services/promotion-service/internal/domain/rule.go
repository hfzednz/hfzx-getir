package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Rule governs eligibility, stacking, and exclusions for a promotion.
type Rule struct {
	ID           uuid.UUID
	TenantID     uuid.UUID
	PromotionID  uuid.UUID
	Priority     int
	StackGroup   string
	Stackable    bool
	// ExcludePromotionIDs blocks these promotions when this one wins.
	ExcludePromotionIDs []uuid.UUID
	// Opaque targeting ids (never resolved to catalog titles).
	VariantIDs  []string
	CategoryIDs []string
	BrandIDs    []string
	SegmentIDs  []string
	// Usage limits (0 = unlimited).
	GlobalLimit   int
	PerUserLimit  int
	PerOrderLimit int
	PerDeviceLimit int
	MinQty        int
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// Validate checks rule invariants.
func (r Rule) Validate() error {
	if r.ID == uuid.Nil {
		return fmt.Errorf("%w: rule id required", ErrInvalidArgument)
	}
	if r.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if r.PromotionID == uuid.Nil {
		return fmt.Errorf("%w: promotion_id required", ErrInvalidArgument)
	}
	return nil
}

// MatchesLine reports whether a cart line matches opaque targeting filters.
// Empty filters mean "match all".
func (r Rule) MatchesLine(variantID, categoryID, brandID string) bool {
	if len(r.VariantIDs) > 0 && !containsStr(r.VariantIDs, variantID) {
		return false
	}
	if len(r.CategoryIDs) > 0 && !containsStr(r.CategoryIDs, categoryID) {
		return false
	}
	if len(r.BrandIDs) > 0 && !containsStr(r.BrandIDs, brandID) {
		return false
	}
	return true
}

// MatchesSegments reports whether the user has at least one required segment
// (or no segment filter is set).
func (r Rule) MatchesSegments(userSegments []string) bool {
	if len(r.SegmentIDs) == 0 {
		return true
	}
	for _, s := range r.SegmentIDs {
		if containsStr(userSegments, s) {
			return true
		}
	}
	return false
}

func containsStr(ss []string, v string) bool {
	for _, s := range ss {
		if s == v {
			return true
		}
	}
	return false
}
