package domain

import (
	"sort"
	"time"

	"github.com/google/uuid"
)

// CartLine is an opaque cart line for evaluation (no catalog titles).
type CartLine struct {
	LineID        string
	VariantID     string
	CategoryID    string
	BrandID       string
	Quantity      int
	UnitPriceMinor int64
}

// LineTotalMinor returns qty * unit price.
func (l CartLine) LineTotalMinor() int64 {
	return l.UnitPriceMinor * int64(l.Quantity)
}

// EvaluateContext is the input for cart promotion evaluation.
type EvaluateContext struct {
	TenantID     uuid.UUID
	PrincipalID  uuid.UUID
	DeviceID     string
	Currency     string
	Lines        []CartLine
	CouponCodes  []string
	ShippingMinor int64
	SegmentIDs   []string
	OrderRef     string
	Now          time.Time
}

// DiscountLine is one applied discount result.
type DiscountLine struct {
	PromotionID   uuid.UUID
	CampaignID    uuid.UUID
	Type          PromotionType
	AmountMinor   int64
	Currency      string
	Description   string
	StackGroup    string
	Priority      int
	AppliedLineIDs []string
	CouponCode    string
}

// EvaluateResult holds conflict-resolved discounts.
type EvaluateResult struct {
	Discounts      []DiscountLine
	TotalDiscountMinor int64
	Currency       string
	ShippingDiscountMinor int64
}

// Candidate is an internal eligible promotion before conflict resolution.
type Candidate struct {
	Promotion Promotion
	Rule      Rule
	Campaign  Campaign
	Discount  DiscountLine
}

// ResolveConflicts applies priority DESC, stack groups, and exclusions.
func ResolveConflicts(cands []Candidate) []DiscountLine {
	if len(cands) == 0 {
		return nil
	}
	sort.SliceStable(cands, func(i, j int) bool {
		if cands[i].Discount.Priority != cands[j].Discount.Priority {
			return cands[i].Discount.Priority > cands[j].Discount.Priority
		}
		return cands[i].Promotion.CreatedAt.Before(cands[j].Promotion.CreatedAt)
	})

	excluded := map[uuid.UUID]struct{}{}
	stackTaken := map[string]uuid.UUID{}
	var winners []DiscountLine

	for _, c := range cands {
		pid := c.Promotion.ID
		if _, skip := excluded[pid]; skip {
			continue
		}
		sg := c.Rule.StackGroup
		if sg != "" && !c.Rule.Stackable {
			if winner, taken := stackTaken[sg]; taken && winner != pid {
				continue
			}
		}
		// Apply exclusions from this winner.
		for _, ex := range c.Rule.ExcludePromotionIDs {
			excluded[ex] = struct{}{}
		}
		// Drop already-accepted winners that this exclusion targets.
		filtered := winners[:0]
		for _, w := range winners {
			if _, drop := excluded[w.PromotionID]; drop {
				continue
			}
			filtered = append(filtered, w)
		}
		winners = filtered

		if sg != "" && !c.Rule.Stackable {
			stackTaken[sg] = pid
		}
		winners = append(winners, c.Discount)
	}
	return winners
}

// ComputeDiscount calculates the discount amount for a promotion against lines.
func ComputeDiscount(p Promotion, matched []CartLine, shippingMinor int64, currency string) (amount int64, shipDisc int64, appliedIDs []string) {
	appliedIDs = make([]string, 0, len(matched))
	var subtotal int64
	for _, l := range matched {
		subtotal += l.LineTotalMinor()
		appliedIDs = append(appliedIDs, l.LineID)
	}

	switch p.Type {
	case PromoPercent:
		amount = subtotal * int64(p.PercentOff) / 100
		if p.MaxDiscountMinor > 0 && amount > p.MaxDiscountMinor {
			amount = p.MaxDiscountMinor
		}
	case PromoFixed:
		amount = p.FixedOffMinor
		if amount > subtotal {
			amount = subtotal
		}
	case PromoBOGO:
		// Buy BuyQty get GetQty free (cheapest free units).
		amount = bogoDiscount(matched, p.BuyQty, p.GetQty)
	case PromoBXGY:
		amount = bogoDiscount(matched, p.BuyQty, p.GetQty)
	case PromoMultibuy:
		amount = multibuyDiscount(matched, p.BuyQty, p.GetQty, p.FixedOffMinor, p.PercentOff)
	case PromoThreshold:
		if subtotal < p.ThresholdMinor {
			return 0, 0, nil
		}
		if p.FixedOffMinor > 0 {
			amount = p.FixedOffMinor
			if amount > subtotal {
				amount = subtotal
			}
		} else if p.PercentOff > 0 {
			amount = subtotal * int64(p.PercentOff) / 100
		}
	case PromoFreeShip:
		shipDisc = shippingMinor
		amount = 0
	case PromoBundle:
		if subtotal < p.ThresholdMinor {
			return 0, 0, nil
		}
		amount = p.FixedOffMinor
		if amount > subtotal {
			amount = subtotal
		}
	case PromoGift:
		// Gift does not reduce line money; amount 0, metadata elsewhere.
		amount = 0
	}
	_ = currency
	if amount < 0 {
		amount = 0
	}
	return amount, shipDisc, appliedIDs
}

func bogoDiscount(lines []CartLine, buyQty, getQty int) int64 {
	if buyQty <= 0 || getQty <= 0 {
		return 0
	}
	// Expand unit prices.
	var units []int64
	for _, l := range lines {
		for i := 0; i < l.Quantity; i++ {
			units = append(units, l.UnitPriceMinor)
		}
	}
	if len(units) == 0 {
		return 0
	}
	sort.Slice(units, func(i, j int) bool { return units[i] < units[j] })
	sets := len(units) / (buyQty + getQty)
	if sets == 0 {
		// Classic BOGO: buy N get M — if qty >= buy+get use sets; else if qty > buyQty free min(get, qty-buy)
		if len(units) > buyQty {
			free := getQty
			if rem := len(units) - buyQty; rem < free {
				free = rem
			}
			var disc int64
			for i := 0; i < free; i++ {
				disc += units[i]
			}
			return disc
		}
		return 0
	}
	freeCount := sets * getQty
	var disc int64
	for i := 0; i < freeCount && i < len(units); i++ {
		disc += units[i]
	}
	return disc
}

func multibuyDiscount(lines []CartLine, buyQty, getQty int, fixedOff int64, percentOff int) int64 {
	var qty int
	var subtotal int64
	for _, l := range lines {
		qty += l.Quantity
		subtotal += l.LineTotalMinor()
	}
	if qty < buyQty {
		return 0
	}
	if fixedOff > 0 {
		sets := qty / buyQty
		d := fixedOff * int64(sets)
		if d > subtotal {
			return subtotal
		}
		return d
	}
	if percentOff > 0 {
		return subtotal * int64(percentOff) / 100
	}
	_ = getQty
	return 0
}
