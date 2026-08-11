package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// PromotionType classifies discount mechanics.
type PromotionType string

const (
	PromoPercent   PromotionType = "percent"
	PromoFixed     PromotionType = "fixed"
	PromoBOGO      PromotionType = "bogo"
	PromoBXGY      PromotionType = "bxgy"
	PromoBundle    PromotionType = "bundle"
	PromoThreshold PromotionType = "threshold"
	PromoFreeShip  PromotionType = "free_ship"
	PromoGift      PromotionType = "gift"
	PromoMultibuy  PromotionType = "multibuy"
)

// Valid reports whether the promotion type is recognized.
func (t PromotionType) Valid() bool {
	switch t {
	case PromoPercent, PromoFixed, PromoBOGO, PromoBXGY, PromoBundle,
		PromoThreshold, PromoFreeShip, PromoGift, PromoMultibuy:
		return true
	default:
		return false
	}
}

// Promotion is a discount definition bound to a campaign.
type Promotion struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	CampaignID  uuid.UUID
	Name        string
	Type        PromotionType
	// PercentOff is 0-10000 representing basis points of percent (1000 = 10%).
	// For percent type, PercentOff is 1-100 (whole percent) for simplicity.
	PercentOff int
	// FixedOffMinor is fixed discount in minor units (currency from evaluate context).
	FixedOffMinor int64
	// BuyQty / GetQty for BOGO, BXGY, multibuy.
	BuyQty int
	GetQty int
	// ThresholdMinor is minimum cart/line subtotal for threshold/bundle.
	ThresholdMinor int64
	// GiftVariantID is an opaque variant id for gift promotions.
	GiftVariantID string
	// MaxDiscountMinor caps percent discounts (0 = no cap).
	MaxDiscountMinor int64
	Priority         int
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// Validate checks promotion invariants.
func (p Promotion) Validate() error {
	if p.ID == uuid.Nil {
		return fmt.Errorf("%w: promotion id required", ErrInvalidArgument)
	}
	if p.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if p.CampaignID == uuid.Nil {
		return fmt.Errorf("%w: campaign_id required", ErrInvalidArgument)
	}
	if p.Name == "" {
		return fmt.Errorf("%w: promotion name required", ErrInvalidArgument)
	}
	if !p.Type.Valid() {
		return fmt.Errorf("%w: invalid promotion type %q", ErrInvalidArgument, p.Type)
	}
	switch p.Type {
	case PromoPercent:
		if p.PercentOff <= 0 || p.PercentOff > 100 {
			return fmt.Errorf("%w: percent_off must be 1-100", ErrInvalidArgument)
		}
	case PromoFixed:
		if p.FixedOffMinor <= 0 {
			return fmt.Errorf("%w: fixed_off_minor must be > 0", ErrInvalidArgument)
		}
	case PromoBOGO, PromoBXGY, PromoMultibuy:
		if p.BuyQty <= 0 || p.GetQty <= 0 {
			return fmt.Errorf("%w: buy_qty and get_qty must be > 0", ErrInvalidArgument)
		}
	case PromoThreshold:
		if p.ThresholdMinor <= 0 {
			return fmt.Errorf("%w: threshold_minor must be > 0", ErrInvalidArgument)
		}
		if p.FixedOffMinor <= 0 && p.PercentOff <= 0 {
			return fmt.Errorf("%w: threshold needs fixed or percent off", ErrInvalidArgument)
		}
	}
	return nil
}
