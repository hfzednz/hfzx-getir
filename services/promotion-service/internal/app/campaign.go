package app

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/promotion-service/internal/domain"
)

// CreateCampaignInput creates a draft/scheduled campaign.
type CreateCampaignInput struct {
	TenantID    uuid.UUID
	Name        string
	Description string
	StartsAt    *time.Time
	EndsAt      *time.Time
}

// CreateCampaign creates a new campaign in draft (or scheduled if starts_at in future).
func (d *Deps) CreateCampaign(ctx context.Context, in CreateCampaignInput) (domain.Campaign, error) {
	if in.TenantID == uuid.Nil {
		return domain.Campaign{}, fmt.Errorf("%w: tenant_id required", domain.ErrInvalidArgument)
	}
	if in.Name == "" {
		return domain.Campaign{}, fmt.Errorf("%w: name required", domain.ErrInvalidArgument)
	}
	now := d.now()
	status := domain.CampaignDraft
	if in.StartsAt != nil && in.StartsAt.After(now) {
		status = domain.CampaignScheduled
	}
	c := domain.Campaign{
		ID:          d.newID(),
		TenantID:    in.TenantID,
		Name:        in.Name,
		Description: in.Description,
		Status:      status,
		StartsAt:    in.StartsAt,
		EndsAt:      in.EndsAt,
		CreatedAt:   now,
		UpdatedAt:   now,
		Version:     1,
	}
	if err := c.Validate(); err != nil {
		return domain.Campaign{}, err
	}
	if err := d.Campaigns.Create(ctx, c); err != nil {
		return domain.Campaign{}, err
	}
	d.emit(ctx, c.TenantID, c.ID, domain.EventCampaignCreated, map[string]any{
		"name": c.Name, "status": string(c.Status),
	})
	return c, nil
}

// ActivateCampaign transitions a campaign to active.
func (d *Deps) ActivateCampaign(ctx context.Context, tenantID, campaignID uuid.UUID) (domain.Campaign, error) {
	c, err := d.Campaigns.GetByID(ctx, tenantID, campaignID)
	if err != nil {
		return domain.Campaign{}, err
	}
	if !c.CanTransition(domain.CampaignActive) {
		return domain.Campaign{}, fmt.Errorf("%w: %s -> active", domain.ErrInvalidTransition, c.Status)
	}
	now := d.now()
	c.Status = domain.CampaignActive
	c.UpdatedAt = now
	c.Version++
	if err := d.Campaigns.Update(ctx, c); err != nil {
		return domain.Campaign{}, err
	}
	d.emit(ctx, c.TenantID, c.ID, domain.EventCampaignActivated, nil)
	return c, nil
}

// PauseCampaign pauses an active campaign.
func (d *Deps) PauseCampaign(ctx context.Context, tenantID, campaignID uuid.UUID) (domain.Campaign, error) {
	c, err := d.Campaigns.GetByID(ctx, tenantID, campaignID)
	if err != nil {
		return domain.Campaign{}, err
	}
	if !c.CanTransition(domain.CampaignPaused) {
		return domain.Campaign{}, fmt.Errorf("%w: %s -> paused", domain.ErrInvalidTransition, c.Status)
	}
	c.Status = domain.CampaignPaused
	c.UpdatedAt = d.now()
	c.Version++
	if err := d.Campaigns.Update(ctx, c); err != nil {
		return domain.Campaign{}, err
	}
	d.emit(ctx, c.TenantID, c.ID, domain.EventCampaignPaused, nil)
	return c, nil
}

// ExpireCampaign marks a campaign expired.
func (d *Deps) ExpireCampaign(ctx context.Context, tenantID, campaignID uuid.UUID) (domain.Campaign, error) {
	c, err := d.Campaigns.GetByID(ctx, tenantID, campaignID)
	if err != nil {
		return domain.Campaign{}, err
	}
	if !c.CanTransition(domain.CampaignExpired) {
		return domain.Campaign{}, fmt.Errorf("%w: %s -> expired", domain.ErrInvalidTransition, c.Status)
	}
	c.Status = domain.CampaignExpired
	c.UpdatedAt = d.now()
	c.Version++
	if err := d.Campaigns.Update(ctx, c); err != nil {
		return domain.Campaign{}, err
	}
	d.emit(ctx, c.TenantID, c.ID, domain.EventCampaignExpired, nil)
	return c, nil
}

// GetCampaign returns a campaign by id.
func (d *Deps) GetCampaign(ctx context.Context, tenantID, campaignID uuid.UUID) (domain.Campaign, error) {
	return d.Campaigns.GetByID(ctx, tenantID, campaignID)
}

// ListCampaigns lists campaigns for admin.
func (d *Deps) ListCampaigns(ctx context.Context, tenantID uuid.UUID, status *domain.CampaignStatus, limit, offset int) ([]domain.Campaign, error) {
	if limit <= 0 {
		limit = 50
	}
	return d.Campaigns.List(ctx, tenantID, status, limit, offset)
}
