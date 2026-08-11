package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Device is a known client endpoint for a principal.
type Device struct {
	ID          uuid.UUID
	PrincipalID uuid.UUID
	Fingerprint string
	Platform    string
	Name        string
	Trusted     bool
	TrustedAt   *time.Time
	LastSeenAt  time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
	RevokedAt   *time.Time
}

func (d Device) IsRevoked() bool {
	return d.RevokedAt != nil
}

func (d Device) IsTrusted() bool {
	return d.Trusted && d.RevokedAt == nil
}

func (d Device) Validate() error {
	if d.ID == uuid.Nil {
		return fmt.Errorf("%w: device id required", ErrInvalidArgument)
	}
	if d.PrincipalID == uuid.Nil {
		return fmt.Errorf("%w: principal_id required", ErrInvalidArgument)
	}
	if d.Fingerprint == "" {
		return fmt.Errorf("%w: fingerprint required", ErrInvalidArgument)
	}
	if d.Trusted && d.TrustedAt == nil {
		return fmt.Errorf("%w: trusted device requires trusted_at", ErrInvariant)
	}
	return nil
}

// Trust marks the device as explicitly trusted.
func (d *Device) Trust(now time.Time) error {
	if d.IsRevoked() {
		return ErrDeviceRevoked
	}
	d.Trusted = true
	d.TrustedAt = &now
	d.UpdatedAt = now
	return nil
}

// Revoke permanently untrusts and disables the device.
func (d *Device) Revoke(now time.Time) {
	if d.RevokedAt != nil {
		return
	}
	d.Trusted = false
	d.RevokedAt = &now
	d.UpdatedAt = now
}

// See updates last-seen for an active device.
func (d *Device) See(now time.Time) error {
	if d.IsRevoked() {
		return ErrDeviceRevoked
	}
	d.LastSeenAt = now
	d.UpdatedAt = now
	return nil
}
