package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/nexora/location-service/internal/domain"
)

// ValidateAddressInput validates and scores an address via geofence serviceability.
type ValidateAddressInput struct {
	TenantID uuid.UUID
	Line1    string
	Lat      float64
	Lng      float64
	PlaceID  string
}

// ValidateAddressResult combines normalized address + feasibility.
type ValidateAddressResult struct {
	Address     domain.NormalizedAddress
	Feasibility domain.DeliveryFeasibility
}

// ValidateAddress geocodes if needed, normalizes, and checks geofence serviceability.
func (d *Deps) ValidateAddress(ctx context.Context, in ValidateAddressInput) (ValidateAddressResult, error) {
	if in.TenantID == uuid.Nil {
		return ValidateAddressResult{}, fmt.Errorf("%w: tenant_id required", domain.ErrInvalidArgument)
	}
	lat, lng := in.Lat, in.Lng
	placeID := strings.TrimSpace(in.PlaceID)
	line1 := strings.TrimSpace(in.Line1)
	conf := domain.ConfidenceScore(0.5)
	comps := domain.AddressComponents{}

	needForward := line1 != "" && (!domain.ValidLatLng(lat, lng) || (lat == 0 && lng == 0))
	needReverse := line1 == "" && placeID == "" && domain.ValidLatLng(lat, lng) && !(lat == 0 && lng == 0)

	switch {
	case needForward:
		geo, err := d.ForwardGeocode(ctx, ForwardGeocodeInput{TenantID: in.TenantID, Query: line1})
		if err != nil {
			return ValidateAddressResult{}, err
		}
		lat, lng = geo.Lat, geo.Lng
		if placeID == "" {
			placeID = geo.PlaceID
		}
		if line1 == "" {
			line1 = geo.Formatted
		}
		conf = geo.Confidence
		comps = geo.Components
	case needReverse:
		geo, err := d.ReverseGeocode(ctx, ReverseGeocodeInput{TenantID: in.TenantID, Lat: lat, Lng: lng})
		if err != nil {
			return ValidateAddressResult{}, err
		}
		if placeID == "" {
			placeID = geo.PlaceID
		}
		line1 = geo.Formatted
		conf = geo.Confidence
		comps = geo.Components
	case line1 == "" && placeID == "":
		return ValidateAddressResult{}, fmt.Errorf("%w: line1 or lat/lng required", domain.ErrInvalidArgument)
	}
	if !domain.ValidLatLng(lat, lng) {
		return ValidateAddressResult{}, fmt.Errorf("%w: lat/lng out of range", domain.ErrInvalidArgument)
	}

	if d.Geofence == nil {
		return ValidateAddressResult{}, fmt.Errorf("%w: geofence client not configured", domain.ErrInvariant)
	}
	feas, err := d.Geofence.CheckServiceability(ctx, in.TenantID, lat, lng)
	if err != nil {
		return ValidateAddressResult{}, err
	}

	now := d.now()
	addr := domain.Normalize(domain.NormalizedAddress{
		ID: d.newID(), TenantID: in.TenantID, Line1: line1,
		Lat: lat, Lng: lng, PlaceID: placeID, Confidence: conf,
		RiskScore: riskFromFeasibility(feas), Components: comps,
		CreatedAt: now, UpdatedAt: now,
	})
	if err := addr.Validate(); err != nil {
		return ValidateAddressResult{}, err
	}
	if d.Addresses != nil {
		_ = d.Addresses.Upsert(ctx, addr)
	}
	if d.Search != nil {
		_ = d.Search.IndexAddress(ctx, addr)
	}
	d.emit(ctx, addr.TenantID, addr.ID, domain.EventAddressValidated, map[string]any{
		"serviceable": feas.Serviceable, "confidence": float64(addr.Confidence),
	})
	return ValidateAddressResult{Address: addr, Feasibility: feas}, nil
}

func riskFromFeasibility(f domain.DeliveryFeasibility) float64 {
	if f.Serviceable {
		return 0.1
	}
	return 0.9
}

// NormalizeAddressInput normalizes textual address fields.
type NormalizeAddressInput struct {
	TenantID   uuid.UUID
	Address    domain.NormalizedAddress
}

// NormalizeAddress applies Normalize helpers and persists.
func (d *Deps) NormalizeAddress(ctx context.Context, in NormalizeAddressInput) (domain.NormalizedAddress, error) {
	if in.TenantID == uuid.Nil {
		return domain.NormalizedAddress{}, fmt.Errorf("%w: tenant_id required", domain.ErrInvalidArgument)
	}
	a := in.Address
	a.TenantID = in.TenantID
	if a.ID == uuid.Nil {
		a.ID = d.newID()
	}
	now := d.now()
	if a.CreatedAt.IsZero() {
		a.CreatedAt = now
	}
	a.UpdatedAt = now
	a = domain.Normalize(a)
	if err := a.Validate(); err != nil {
		return domain.NormalizedAddress{}, err
	}
	if d.Addresses != nil {
		if err := d.Addresses.Upsert(ctx, a); err != nil {
			return domain.NormalizedAddress{}, err
		}
	}
	if d.Search != nil {
		_ = d.Search.IndexAddress(ctx, a)
	}
	return a, nil
}

// EnrichAddressInput enriches an address via reverse/forward geocode.
type EnrichAddressInput struct {
	TenantID uuid.UUID
	Address  domain.NormalizedAddress
}

// EnrichAddress fills place_id/components/confidence from MapsProvider.
func (d *Deps) EnrichAddress(ctx context.Context, in EnrichAddressInput) (domain.NormalizedAddress, error) {
	if in.TenantID == uuid.Nil {
		return domain.NormalizedAddress{}, fmt.Errorf("%w: tenant_id required", domain.ErrInvalidArgument)
	}
	a := in.Address
	a.TenantID = in.TenantID
	if a.ID == uuid.Nil {
		a.ID = d.newID()
	}

	var geo domain.GeocodeResult
	var err error
	if domain.ValidLatLng(a.Lat, a.Lng) && !(a.Lat == 0 && a.Lng == 0) {
		geo, err = d.ReverseGeocode(ctx, ReverseGeocodeInput{TenantID: in.TenantID, Lat: a.Lat, Lng: a.Lng})
	} else if strings.TrimSpace(a.Line1) != "" {
		geo, err = d.ForwardGeocode(ctx, ForwardGeocodeInput{TenantID: in.TenantID, Query: a.Line1})
	} else {
		return domain.NormalizedAddress{}, fmt.Errorf("%w: line1 or lat/lng required", domain.ErrInvalidArgument)
	}
	if err != nil {
		return domain.NormalizedAddress{}, err
	}

	a.Lat, a.Lng = geo.Lat, geo.Lng
	if a.PlaceID == "" {
		a.PlaceID = geo.PlaceID
	}
	if a.Line1 == "" {
		a.Line1 = geo.Formatted
	}
	a.Confidence = geo.Confidence
	a.Components = geo.Components
	now := d.now()
	if a.CreatedAt.IsZero() {
		a.CreatedAt = now
	}
	a.UpdatedAt = now
	a = domain.Normalize(a)
	if err := a.Validate(); err != nil {
		return domain.NormalizedAddress{}, err
	}
	if d.Addresses != nil {
		if err := d.Addresses.Upsert(ctx, a); err != nil {
			return domain.NormalizedAddress{}, err
		}
	}
	if d.Search != nil {
		_ = d.Search.IndexAddress(ctx, a)
	}
	d.emit(ctx, a.TenantID, a.ID, domain.EventAddressCreated, map[string]any{
		"placeId": a.PlaceID, "confidence": float64(a.Confidence),
	})
	return a, nil
}
