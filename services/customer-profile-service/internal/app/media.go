package app

import (
	"context"
	"fmt"
	"io"

	"github.com/google/uuid"
	"github.com/nexora/customer-profile-service/internal/app/ports"
	"github.com/nexora/customer-profile-service/internal/domain"
)

// SetAvatarInput uploads a new avatar via MediaStore.
type SetAvatarInput struct {
	ProfileID   uuid.UUID
	Filename    string
	ContentType string
	Body        io.Reader
	TraceID     string
}

// SetAvatar stores an avatar and updates the profile media refs.
func (d *Deps) SetAvatar(ctx context.Context, in SetAvatarInput) (domain.CustomerProfile, error) {
	p, err := d.requireActiveProfile(ctx, in.ProfileID)
	if err != nil {
		return domain.CustomerProfile{}, err
	}
	if d.Media == nil {
		return domain.CustomerProfile{}, fmt.Errorf("%w: media store not configured", domain.ErrInvariant)
	}
	if in.Body == nil {
		return domain.CustomerProfile{}, fmt.Errorf("%w: avatar body required", domain.ErrInvalidArgument)
	}
	url, version, err := d.Media.PutAvatar(ctx, p.TenantID, p.ID, in.Filename, in.ContentType, in.Body)
	if err != nil {
		return domain.CustomerProfile{}, err
	}
	p.AvatarURL = url
	p.UpdatedAt = d.now()
	if err := p.Validate(); err != nil {
		return domain.CustomerProfile{}, err
	}
	if err := d.Profiles.Update(ctx, p); err != nil {
		return domain.CustomerProfile{}, err
	}
	d.publish(ctx, ports.TopicMediaEvents, p.ID.String(), map[string]any{
		"eventId": d.newID().String(), "eventType": domain.EventAvatarUpdated,
		"occurredAt": d.now(), "tenantId": p.TenantID, "principalId": p.PrincipalID,
		"profileId": p.ID, "avatarUrl": url, "version": version, "traceId": in.TraceID,
	})
	return p, nil
}

// DeleteAvatar removes the avatar from MediaStore and clears profile refs.
func (d *Deps) DeleteAvatar(ctx context.Context, profileID uuid.UUID, traceID string) (domain.CustomerProfile, error) {
	p, err := d.requireActiveProfile(ctx, profileID)
	if err != nil {
		return domain.CustomerProfile{}, err
	}
	if d.Media != nil {
		if err := d.Media.DeleteAvatar(ctx, p.TenantID, p.ID); err != nil {
			return domain.CustomerProfile{}, err
		}
	}
	p.AvatarURL = ""
	p.UpdatedAt = d.now()
	if err := d.Profiles.Update(ctx, p); err != nil {
		return domain.CustomerProfile{}, err
	}
	d.publish(ctx, ports.TopicMediaEvents, p.ID.String(), map[string]any{
		"eventId": d.newID().String(), "eventType": domain.EventAvatarUpdated,
		"occurredAt": d.now(), "tenantId": p.TenantID, "principalId": p.PrincipalID,
		"profileId": p.ID, "avatarUrl": "", "version": 0, "deleted": true, "traceId": traceID,
	})
	return p, nil
}
