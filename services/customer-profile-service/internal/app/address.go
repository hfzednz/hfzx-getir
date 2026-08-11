package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/customer-profile-service/internal/app/ports"
	"github.com/nexora/customer-profile-service/internal/domain"
)

// AddAddressInput creates a new address.
type AddAddressInput struct {
	ProfileID   uuid.UUID
	Label       domain.AddressLabel
	CustomLabel string
	Line1       string
	Building    string
	Apartment   string
	Entrance    string
	Floor       string
	Door        string
	Notes       string
	Lat         float64
	Lng         float64
	CityID      *uuid.UUID
	IsDefault   bool
	TraceID     string
}

// UpdateAddressInput patches an address.
type UpdateAddressInput struct {
	AddressID   uuid.UUID
	Label       *domain.AddressLabel
	CustomLabel *string
	Line1       *string
	Building    *string
	Apartment   *string
	Entrance    *string
	Floor       *string
	Door        *string
	Notes       *string
	Lat         *float64
	Lng         *float64
	TraceID     string
}

func (d *Deps) publishAddress(ctx context.Context, eventType string, p domain.CustomerProfile, addressID uuid.UUID, extra map[string]any, traceID string) {
	payload := map[string]any{
		"eventId":     d.newID().String(),
		"eventType":   eventType,
		"occurredAt":  d.now(),
		"tenantId":    p.TenantID,
		"principalId": p.PrincipalID,
		"profileId":   p.ID,
		"addressId":   addressID,
		"traceId":     traceID,
	}
	for k, v := range extra {
		payload[k] = v
	}
	d.publish(ctx, ports.TopicAddressEvents, addressID.String(), payload)
}

// AddAddress creates an address, optionally validating zone and setting default.
func (d *Deps) AddAddress(ctx context.Context, in AddAddressInput) (domain.Address, error) {
	p, err := d.requireActiveProfile(ctx, in.ProfileID)
	if err != nil {
		return domain.Address{}, err
	}
	label := in.Label
	if label == "" {
		label = domain.AddressLabelHome
	}
	now := d.now()
	a := domain.Address{
		ID:          d.newID(),
		ProfileID:   in.ProfileID,
		TenantID:    p.TenantID,
		Label:       label,
		CustomLabel: in.CustomLabel,
		Line1:       in.Line1,
		Building:    in.Building,
		Apartment:   in.Apartment,
		Entrance:    in.Entrance,
		Floor:       in.Floor,
		Door:        in.Door,
		Notes:       in.Notes,
		Lat:         in.Lat,
		Lng:         in.Lng,
		CityID:      in.CityID,
		IsDefault:   in.IsDefault,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := a.Validate(); err != nil {
		return domain.Address{}, err
	}

	if d.Zones != nil {
		_, ok, zerr := d.Zones.ValidateZone(ctx, p.TenantID, a.Lat, a.Lng)
		if zerr != nil {
			return domain.Address{}, zerr
		}
		if ok {
			t := now
			a.ZoneValidatedAt = &t
		}
	}

	if a.IsDefault {
		if err := d.Addresses.ClearDefault(ctx, in.ProfileID); err != nil {
			return domain.Address{}, err
		}
	} else {
		existing, _ := d.Addresses.ListByProfile(ctx, in.ProfileID)
		active := 0
		for _, e := range existing {
			if e.DeletedAt == nil {
				active++
			}
		}
		if active == 0 {
			a.IsDefault = true
		}
	}

	if err := d.Addresses.Create(ctx, a); err != nil {
		return domain.Address{}, err
	}
	d.publishAddress(ctx, domain.EventAddressAdded, p, a.ID, nil, in.TraceID)
	return a, nil
}

// UpdateAddress patches address fields.
func (d *Deps) UpdateAddress(ctx context.Context, in UpdateAddressInput) (domain.Address, error) {
	a, err := d.Addresses.GetByID(ctx, in.AddressID)
	if err != nil {
		return domain.Address{}, err
	}
	p, err := d.requireActiveProfile(ctx, a.ProfileID)
	if err != nil {
		return domain.Address{}, err
	}
	if in.Label != nil {
		a.Label = *in.Label
	}
	if in.CustomLabel != nil {
		a.CustomLabel = *in.CustomLabel
	}
	if in.Line1 != nil {
		a.Line1 = *in.Line1
	}
	if in.Building != nil {
		a.Building = *in.Building
	}
	if in.Apartment != nil {
		a.Apartment = *in.Apartment
	}
	if in.Entrance != nil {
		a.Entrance = *in.Entrance
	}
	if in.Floor != nil {
		a.Floor = *in.Floor
	}
	if in.Door != nil {
		a.Door = *in.Door
	}
	if in.Notes != nil {
		a.Notes = *in.Notes
	}
	if in.Lat != nil {
		a.Lat = *in.Lat
	}
	if in.Lng != nil {
		a.Lng = *in.Lng
	}
	a.UpdatedAt = d.now()
	if err := a.Validate(); err != nil {
		return domain.Address{}, err
	}
	if err := d.Addresses.Update(ctx, a); err != nil {
		return domain.Address{}, err
	}
	d.publishAddress(ctx, domain.EventAddressUpdated, p, a.ID, nil, in.TraceID)
	return a, nil
}

// DeleteAddress soft-removes an address.
func (d *Deps) DeleteAddress(ctx context.Context, addressID uuid.UUID, traceID string) error {
	a, err := d.Addresses.GetByID(ctx, addressID)
	if err != nil {
		return err
	}
	p, err := d.requireActiveProfile(ctx, a.ProfileID)
	if err != nil {
		return err
	}
	wasDefault := a.IsDefault
	if err := d.Addresses.Delete(ctx, addressID); err != nil {
		return err
	}
	if wasDefault {
		remaining, _ := d.Addresses.ListByProfile(ctx, a.ProfileID)
		for i := range remaining {
			if remaining[i].DeletedAt == nil {
				remaining[i].IsDefault = true
				remaining[i].UpdatedAt = d.now()
				_ = d.Addresses.Update(ctx, remaining[i])
				break
			}
		}
	}
	d.publishAddress(ctx, domain.EventAddressRemoved, p, a.ID, nil, traceID)
	return nil
}

// SetDefaultAddress marks one address as default for the profile.
func (d *Deps) SetDefaultAddress(ctx context.Context, profileID, addressID uuid.UUID, traceID string) (domain.Address, error) {
	p, err := d.requireActiveProfile(ctx, profileID)
	if err != nil {
		return domain.Address{}, err
	}
	a, err := d.Addresses.GetByID(ctx, addressID)
	if err != nil {
		return domain.Address{}, err
	}
	if a.ProfileID != profileID {
		return domain.Address{}, fmt.Errorf("%w: address does not belong to profile", domain.ErrForbidden)
	}
	if err := d.Addresses.ClearDefault(ctx, profileID); err != nil {
		return domain.Address{}, err
	}
	a.IsDefault = true
	a.UpdatedAt = d.now()
	if err := d.Addresses.Update(ctx, a); err != nil {
		return domain.Address{}, err
	}
	d.publishAddress(ctx, domain.EventAddressUpdated, p, a.ID, map[string]any{"isDefault": true}, traceID)
	return a, nil
}

// ListAddresses returns non-deleted addresses for a profile.
func (d *Deps) ListAddresses(ctx context.Context, profileID uuid.UUID) ([]domain.Address, error) {
	if _, err := d.requireActiveProfile(ctx, profileID); err != nil {
		return nil, err
	}
	addrs, err := d.Addresses.ListByProfile(ctx, profileID)
	if err != nil {
		return nil, err
	}
	out := make([]domain.Address, 0, len(addrs))
	for _, a := range addrs {
		if a.DeletedAt == nil {
			out = append(out, a)
		}
	}
	return out, nil
}

// ValidateAddressZone re-validates an address against the geofence port.
func (d *Deps) ValidateAddressZone(ctx context.Context, addressID uuid.UUID) (domain.Address, error) {
	a, err := d.Addresses.GetByID(ctx, addressID)
	if err != nil {
		return domain.Address{}, err
	}
	if d.Zones == nil {
		return a, fmt.Errorf("%w: zone validator not configured", domain.ErrInvariant)
	}
	_, ok, zerr := d.Zones.ValidateZone(ctx, a.TenantID, a.Lat, a.Lng)
	if zerr != nil {
		return domain.Address{}, zerr
	}
	now := d.now()
	a.UpdatedAt = now
	if ok {
		a.ZoneValidatedAt = &now
	} else {
		a.ZoneValidatedAt = nil
	}
	if err := d.Addresses.Update(ctx, a); err != nil {
		return domain.Address{}, err
	}
	if !ok {
		return a, domain.ErrZoneInvalid
	}
	return a, nil
}
