package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/promotion-service/internal/domain"
)

// CreatePromotionInput creates a promotion + optional rule.
type CreatePromotionInput struct {
	TenantID         uuid.UUID
	CampaignID       uuid.UUID
	Name             string
	Type             domain.PromotionType
	PercentOff       int
	FixedOffMinor    int64
	BuyQty           int
	GetQty           int
	ThresholdMinor   int64
	GiftVariantID    string
	MaxDiscountMinor int64
	Priority         int
	Rule             *CreateRuleInput
}

// CreateRuleInput is the optional rule attached at promotion create.
type CreateRuleInput struct {
	Priority            int
	StackGroup          string
	Stackable           bool
	ExcludePromotionIDs []uuid.UUID
	VariantIDs          []string
	CategoryIDs         []string
	BrandIDs            []string
	SegmentIDs          []string
	GlobalLimit         int
	PerUserLimit        int
	PerOrderLimit       int
	PerDeviceLimit      int
	MinQty              int
}

// CreatePromotionResult returns the promotion and rule.
type CreatePromotionResult struct {
	Promotion domain.Promotion
	Rule      domain.Rule
}

// CreatePromotion creates a promotion under a campaign, with a default rule if omitted.
func (d *Deps) CreatePromotion(ctx context.Context, in CreatePromotionInput) (CreatePromotionResult, error) {
	if in.TenantID == uuid.Nil || in.CampaignID == uuid.Nil {
		return CreatePromotionResult{}, fmt.Errorf("%w: tenant_id and campaign_id required", domain.ErrInvalidArgument)
	}
	if _, err := d.Campaigns.GetByID(ctx, in.TenantID, in.CampaignID); err != nil {
		return CreatePromotionResult{}, err
	}
	now := d.now()
	p := domain.Promotion{
		ID:               d.newID(),
		TenantID:         in.TenantID,
		CampaignID:       in.CampaignID,
		Name:             in.Name,
		Type:             in.Type,
		PercentOff:       in.PercentOff,
		FixedOffMinor:    in.FixedOffMinor,
		BuyQty:           in.BuyQty,
		GetQty:           in.GetQty,
		ThresholdMinor:   in.ThresholdMinor,
		GiftVariantID:    in.GiftVariantID,
		MaxDiscountMinor: in.MaxDiscountMinor,
		Priority:         in.Priority,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if err := p.Validate(); err != nil {
		return CreatePromotionResult{}, err
	}
	if err := d.Promotions.Create(ctx, p); err != nil {
		return CreatePromotionResult{}, err
	}

	ri := in.Rule
	if ri == nil {
		ri = &CreateRuleInput{Priority: in.Priority}
	}
	priority := ri.Priority
	if priority == 0 {
		priority = in.Priority
	}
	rule := domain.Rule{
		ID:                  d.newID(),
		TenantID:            in.TenantID,
		PromotionID:         p.ID,
		Priority:            priority,
		StackGroup:          ri.StackGroup,
		Stackable:           ri.Stackable,
		ExcludePromotionIDs: ri.ExcludePromotionIDs,
		VariantIDs:          ri.VariantIDs,
		CategoryIDs:         ri.CategoryIDs,
		BrandIDs:            ri.BrandIDs,
		SegmentIDs:          ri.SegmentIDs,
		GlobalLimit:         ri.GlobalLimit,
		PerUserLimit:        ri.PerUserLimit,
		PerOrderLimit:       ri.PerOrderLimit,
		PerDeviceLimit:      ri.PerDeviceLimit,
		MinQty:              ri.MinQty,
		CreatedAt:           now,
		UpdatedAt:           now,
	}
	if err := rule.Validate(); err != nil {
		return CreatePromotionResult{}, err
	}
	if err := d.Rules.Create(ctx, rule); err != nil {
		return CreatePromotionResult{}, err
	}
	return CreatePromotionResult{Promotion: p, Rule: rule}, nil
}

// GetPromotion returns a promotion by id.
func (d *Deps) GetPromotion(ctx context.Context, tenantID, id uuid.UUID) (domain.Promotion, error) {
	return d.Promotions.GetByID(ctx, tenantID, id)
}

// ListPromotionsByCampaign lists promotions for a campaign.
func (d *Deps) ListPromotionsByCampaign(ctx context.Context, tenantID, campaignID uuid.UUID) ([]domain.Promotion, error) {
	return d.Promotions.ListByCampaign(ctx, tenantID, campaignID)
}
