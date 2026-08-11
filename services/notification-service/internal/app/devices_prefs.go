package app

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/nexora/notification-service/internal/domain"
)

// RegisterDeviceInput registers a push token.
type RegisterDeviceInput struct {
	TenantID    uuid.UUID
	PrincipalID uuid.UUID
	Platform    domain.DevicePlatform
	Token       string
	Locale      string
}

// RegisterDevice upserts a device push token.
func (d *Deps) RegisterDevice(ctx context.Context, in RegisterDeviceInput) (domain.Device, error) {
	if in.TenantID == uuid.Nil || in.PrincipalID == uuid.Nil || strings.TrimSpace(in.Token) == "" {
		return domain.Device{}, fmt.Errorf("%w: tenant_id, principal_id, token required", domain.ErrInvalidArgument)
	}
	platform := in.Platform
	if platform == "" {
		platform = domain.PlatformFCM
	}
	locale := in.Locale
	if locale == "" {
		locale = "en"
	}
	now := d.now()
	dev := domain.Device{
		ID: d.newID(), TenantID: in.TenantID, PrincipalID: in.PrincipalID,
		Platform: platform, Token: in.Token, Locale: locale, Active: true,
		CreatedAt: now, UpdatedAt: now,
	}
	return d.Devices.Upsert(ctx, dev)
}

// SetPreferencesInput updates channel opt-outs and quiet hours.
type SetPreferencesInput struct {
	TenantID      uuid.UUID
	PrincipalID   uuid.UUID
	ChannelOptOut map[string]bool
	QuietStart    *int
	QuietEnd      *int
}

// SetPreferences upserts preferences.
func (d *Deps) SetPreferences(ctx context.Context, in SetPreferencesInput) (domain.Preference, error) {
	if in.TenantID == uuid.Nil || in.PrincipalID == uuid.Nil {
		return domain.Preference{}, fmt.Errorf("%w: tenant_id and principal_id required", domain.ErrInvalidArgument)
	}
	now := d.now()
	pref, err := d.Preferences.Get(ctx, in.TenantID, in.PrincipalID)
	if err != nil {
		pref = domain.DefaultPreference(in.TenantID, in.PrincipalID, now)
		pref.ID = d.newID()
	}
	if pref.ChannelOptOut == nil {
		pref.ChannelOptOut = map[domain.Channel]bool{}
	}
	if in.ChannelOptOut != nil {
		for k, v := range in.ChannelOptOut {
			ch := domain.Channel(k)
			if !ch.Valid() {
				return domain.Preference{}, fmt.Errorf("%w: invalid channel %s", domain.ErrInvalidArgument, k)
			}
			pref.ChannelOptOut[ch] = v
		}
	}
	if in.QuietStart != nil {
		pref.QuietStart = *in.QuietStart
	}
	if in.QuietEnd != nil {
		pref.QuietEnd = *in.QuietEnd
	}
	pref.UpdatedAt = now
	if err := d.Preferences.Upsert(ctx, pref); err != nil {
		return domain.Preference{}, err
	}
	return pref, nil
}

// GetPreferences returns preferences (defaults if missing).
func (d *Deps) GetPreferences(ctx context.Context, tenantID, principalID uuid.UUID) (domain.Preference, error) {
	if tenantID == uuid.Nil || principalID == uuid.Nil {
		return domain.Preference{}, fmt.Errorf("%w: tenant_id and principal_id required", domain.ErrInvalidArgument)
	}
	pref, err := d.Preferences.Get(ctx, tenantID, principalID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.DefaultPreference(tenantID, principalID, d.now()), nil
		}
		return domain.Preference{}, err
	}
	return pref, nil
}
