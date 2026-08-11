package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// MergePolicy controls guest→auth cart merge conflict handling.
type MergePolicy string

const (
	// MergePolicySumQty sums quantities for the same variant (default).
	MergePolicySumQty MergePolicy = "sum_qty"
	// MergePolicyPreferGuest keeps guest qty when both have the variant.
	MergePolicyPreferGuest MergePolicy = "prefer_guest"
	// MergePolicyPreferAuth keeps auth qty when both have the variant.
	MergePolicyPreferAuth MergePolicy = "prefer_auth"
)

// DefaultMergePolicy is qty-sum on login.
const DefaultMergePolicy = MergePolicySumQty

// MergeCarts merges guest into auth (target). Guest is marked merged.
// Same variant_id lines are combined per policy; qty is clamped to max_qty.
func MergeCarts(target, guest Cart, policy MergePolicy, now time.Time, newLineID func() uuid.UUID) (Cart, Cart, error) {
	if target.TenantID != guest.TenantID {
		return Cart{}, Cart{}, fmt.Errorf("%w: tenant mismatch", ErrConflict)
	}
	if err := target.RequireActive(); err != nil {
		return Cart{}, Cart{}, err
	}
	if err := guest.RequireActive(); err != nil {
		return Cart{}, Cart{}, err
	}
	if policy == "" {
		policy = DefaultMergePolicy
	}

	byVariant := make(map[uuid.UUID]int)
	lines := make([]CartLine, 0, len(target.Lines)+len(guest.Lines))
	for i, l := range target.Lines {
		cp := l
		cp.CartID = target.ID
		lines = append(lines, cp)
		byVariant[l.VariantID] = i
	}

	for _, gl := range guest.Lines {
		if idx, ok := byVariant[gl.VariantID]; ok {
			tl := &lines[idx]
			switch policy {
			case MergePolicyPreferGuest:
				tl.Qty = gl.Qty
				if gl.MaxQty < tl.MaxQty {
					tl.MaxQty = gl.MaxQty
				}
			case MergePolicyPreferAuth:
				// keep target qty
				if gl.MaxQty < tl.MaxQty {
					tl.MaxQty = gl.MaxQty
				}
			default: // sum
				tl.Qty += gl.Qty
				if gl.MaxQty < tl.MaxQty {
					tl.MaxQty = gl.MaxQty
				}
			}
			if tl.Qty > tl.MaxQty {
				tl.Qty = tl.MaxQty
			}
			if gl.Notes != "" && tl.Notes == "" {
				tl.Notes = gl.Notes
			}
			tl.UpdatedAt = now
		} else {
			nl := gl
			nl.ID = newLineID()
			nl.CartID = target.ID
			nl.TenantID = target.TenantID
			nl.UpdatedAt = now
			byVariant[gl.VariantID] = len(lines)
			lines = append(lines, nl)
		}
	}

	// Merge coupons (union by code).
	couponSet := make(map[string]AppliedCoupon)
	for _, c := range target.Coupons {
		couponSet[strings.ToUpper(c.Code)] = c
	}
	for _, c := range guest.Coupons {
		key := strings.ToUpper(c.Code)
		if _, ok := couponSet[key]; !ok {
			couponSet[key] = c
		}
	}
	coupons := make([]AppliedCoupon, 0, len(couponSet))
	for _, c := range couponSet {
		coupons = append(coupons, c)
	}

	target.Lines = lines
	target.Coupons = coupons
	target.Quote = nil // invalidate quote after merge
	target.UpdatedAt = now
	target.Version++

	guest.Status = CartStatusMerged
	guest.MergedIntoID = &target.ID
	guest.MergedAt = &now
	guest.UpdatedAt = now
	guest.Version++
	guest.Lines = nil
	guest.Coupons = nil
	guest.Quote = nil

	return target, guest, nil
}
