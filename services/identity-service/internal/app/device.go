package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/identity-service/internal/app/ports"
	"github.com/nexora/identity-service/internal/domain"
)

// RegisterDeviceInput registers or updates a device fingerprint for a principal.
type RegisterDeviceInput struct {
	PrincipalID uuid.UUID
	Fingerprint string
	Platform    string
	Name        string
	AuthCtx     ports.AuthContext
}

// RegisterDevice creates a device or refreshes last-seen if fingerprint exists.
func (d *Deps) RegisterDevice(ctx context.Context, in RegisterDeviceInput) (domain.Device, error) {
	if in.PrincipalID == uuid.Nil || in.Fingerprint == "" {
		return domain.Device{}, fmt.Errorf("%w: principal_id and fingerprint required", domain.ErrInvalidArgument)
	}
	now := d.now()
	if existing, err := d.Devices.FindByFingerprint(ctx, in.PrincipalID, in.Fingerprint); err == nil {
		if existing.IsRevoked() {
			return domain.Device{}, domain.ErrDeviceRevoked
		}
		_ = existing.See(now)
		if in.Name != "" {
			existing.Name = in.Name
		}
		if in.Platform != "" {
			existing.Platform = in.Platform
		}
		if err := d.Devices.Update(ctx, existing); err != nil {
			return domain.Device{}, err
		}
		return existing, nil
	}
	dev := domain.Device{
		ID: d.newID(), PrincipalID: in.PrincipalID, Fingerprint: in.Fingerprint,
		Platform: in.Platform, Name: in.Name, LastSeenAt: now, CreatedAt: now, UpdatedAt: now,
	}
	if err := dev.Validate(); err != nil {
		return domain.Device{}, err
	}
	if err := d.Devices.Create(ctx, dev); err != nil {
		return domain.Device{}, err
	}
	d.appendAudit(ctx, domain.AuditEvent{
		ActorID: &in.PrincipalID, ActorKind: "user",
		Action: "device.register", ResourceType: "device", ResourceID: dev.ID.String(),
		Outcome: domain.AuditOutcomeSuccess, IP: in.AuthCtx.IP,
	})
	return dev, nil
}

// TrustDevice marks a device as trusted.
func (d *Deps) TrustDevice(ctx context.Context, principalID, deviceID uuid.UUID) (domain.Device, error) {
	dev, err := d.Devices.GetByID(ctx, deviceID)
	if err != nil {
		return domain.Device{}, err
	}
	if dev.PrincipalID != principalID {
		return domain.Device{}, domain.ErrForbidden
	}
	if err := dev.Trust(d.now()); err != nil {
		return domain.Device{}, err
	}
	if err := d.Devices.Update(ctx, dev); err != nil {
		return domain.Device{}, err
	}
	d.appendAudit(ctx, domain.AuditEvent{
		ActorID: &principalID, ActorKind: "user",
		Action: "device.trust", ResourceType: "device", ResourceID: dev.ID.String(),
		Outcome: domain.AuditOutcomeSuccess,
	})
	return dev, nil
}

// RevokeDevice permanently disables a device.
func (d *Deps) RevokeDevice(ctx context.Context, principalID, deviceID uuid.UUID) error {
	dev, err := d.Devices.GetByID(ctx, deviceID)
	if err != nil {
		return err
	}
	if dev.PrincipalID != principalID {
		return domain.ErrForbidden
	}
	dev.Revoke(d.now())
	if err := d.Devices.Update(ctx, dev); err != nil {
		return err
	}
	d.appendAudit(ctx, domain.AuditEvent{
		ActorID: &principalID, ActorKind: "user",
		Action: "device.revoke", ResourceType: "device", ResourceID: dev.ID.String(),
		Outcome: domain.AuditOutcomeSuccess,
	})
	return nil
}

// ListDevices returns devices for a principal.
func (d *Deps) ListDevices(ctx context.Context, principalID uuid.UUID) ([]domain.Device, error) {
	if principalID == uuid.Nil {
		return nil, fmt.Errorf("%w: principal_id required", domain.ErrInvalidArgument)
	}
	return d.Devices.ListByPrincipal(ctx, principalID)
}
