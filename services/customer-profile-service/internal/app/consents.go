package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/customer-profile-service/internal/app/ports"
	"github.com/nexora/customer-profile-service/internal/domain"
)

// SetConsentInput grants or revokes a consent channel.
type SetConsentInput struct {
	ProfileID uuid.UUID
	Channel   domain.ConsentChannel
	Granted   bool
	Source    string
	TraceID   string
}

// SetConsent upserts consent and emits ConsentChanged.
func (d *Deps) SetConsent(ctx context.Context, in SetConsentInput) (domain.Consent, error) {
	p, err := d.requireActiveProfile(ctx, in.ProfileID)
	if err != nil {
		return domain.Consent{}, err
	}
	if in.Channel == "" {
		return domain.Consent{}, fmt.Errorf("%w: channel required", domain.ErrInvalidArgument)
	}
	now := d.now()
	c, err := d.Consents.Get(ctx, in.ProfileID, in.Channel)
	if err != nil && err != domain.ErrNotFound {
		return domain.Consent{}, err
	}
	if err == domain.ErrNotFound {
		c = domain.Consent{
			ID:        d.newID(),
			ProfileID: in.ProfileID,
			TenantID:  p.TenantID,
			Channel:   in.Channel,
			CreatedAt: now,
		}
	}
	c.Granted = in.Granted
	c.Source = in.Source
	c.RecordedAt = now
	c.UpdatedAt = now
	c.TenantID = p.TenantID
	if err := c.Validate(); err != nil {
		return domain.Consent{}, err
	}
	if err := d.Consents.Upsert(ctx, c); err != nil {
		return domain.Consent{}, err
	}
	d.publish(ctx, ports.TopicConsentEvents, p.ID.String(), map[string]any{
		"eventId": d.newID().String(), "eventType": domain.EventConsentChanged,
		"occurredAt": now, "tenantId": p.TenantID, "principalId": p.PrincipalID,
		"profileId": p.ID, "channel": string(c.Channel), "granted": c.Granted, "traceId": in.TraceID,
	})
	return c, nil
}

// ListConsents returns all consents for a profile.
func (d *Deps) ListConsents(ctx context.Context, profileID uuid.UUID) ([]domain.Consent, error) {
	if _, err := d.requireActiveProfile(ctx, profileID); err != nil {
		return nil, err
	}
	return d.Consents.List(ctx, profileID)
}

// CheckConsent reports whether a channel is currently granted (gRPC helper).
func (d *Deps) CheckConsent(ctx context.Context, profileID uuid.UUID, channel domain.ConsentChannel) (bool, error) {
	if profileID == uuid.Nil {
		return false, fmt.Errorf("%w: profile_id required", domain.ErrInvalidArgument)
	}
	if channel == "" {
		return false, fmt.Errorf("%w: channel required", domain.ErrInvalidArgument)
	}
	c, err := d.Consents.Get(ctx, profileID, channel)
	if err == domain.ErrNotFound {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return c.Granted, nil
}
