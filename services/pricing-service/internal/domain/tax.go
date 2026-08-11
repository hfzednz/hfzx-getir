package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// TaxRule is a display tax rate (basis points). Pricing does not own fiscal ledgers.
type TaxRule struct {
	ID         uuid.UUID
	TenantID   uuid.UUID
	Code       string
	Name       string
	RateBps    int // e.g. 1800 = 18.00%
	Inclusive  bool
	RegionID   *uuid.UUID
	Active     bool
	Priority   int
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// Validate checks tax rule invariants.
func (t TaxRule) Validate() error {
	if t.ID == uuid.Nil {
		return fmt.Errorf("%w: tax_rule id required", ErrInvalidArgument)
	}
	if t.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if strings.TrimSpace(t.Code) == "" {
		return fmt.Errorf("%w: code required", ErrInvalidArgument)
	}
	if t.RateBps < 0 || t.RateBps > 100000 {
		return fmt.Errorf("%w: rate_bps out of range", ErrInvalidArgument)
	}
	return nil
}

// TaxOn computes tax minor units on a taxable base (exclusive).
func (t TaxRule) TaxOn(baseMinor int64) int64 {
	if baseMinor <= 0 || t.RateBps <= 0 {
		return 0
	}
	return baseMinor * int64(t.RateBps) / 10000
}

// SelectTaxRule picks the best active rule for an optional region (priority DESC, then region match).
func SelectTaxRule(rules []TaxRule, regionID *uuid.UUID) (TaxRule, bool) {
	var best TaxRule
	found := false
	for _, r := range rules {
		if !r.Active {
			continue
		}
		if r.RegionID != nil {
			if regionID == nil || *r.RegionID != *regionID {
				continue
			}
		}
		if !found || r.Priority > best.Priority || (r.Priority == best.Priority && r.RegionID != nil && best.RegionID == nil) {
			best = r
			found = true
		}
	}
	return best, found
}
