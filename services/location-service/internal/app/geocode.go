package app

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/location-service/internal/domain"
)

const geocodeCacheTTL = 24 * time.Hour

// ForwardGeocodeInput is the forward geocode request.
type ForwardGeocodeInput struct {
	TenantID uuid.UUID
	Query    string
}

// ForwardGeocode resolves a query via cache then MapsProvider.
func (d *Deps) ForwardGeocode(ctx context.Context, in ForwardGeocodeInput) (domain.GeocodeResult, error) {
	if in.TenantID == uuid.Nil {
		return domain.GeocodeResult{}, fmt.Errorf("%w: tenant_id required", domain.ErrInvalidArgument)
	}
	q := strings.TrimSpace(in.Query)
	if q == "" {
		return domain.GeocodeResult{}, fmt.Errorf("%w: query required", domain.ErrInvalidArgument)
	}
	hash := queryHash("fwd", in.TenantID.String(), q)
	if d.Cache != nil {
		if cached, ok, err := d.Cache.GetGeocode(ctx, hash); err == nil && ok {
			cached.Cached = true
			return cached, nil
		}
	}
	if d.Maps == nil {
		return domain.GeocodeResult{}, fmt.Errorf("%w: maps provider not configured", domain.ErrInvariant)
	}
	res, err := d.Maps.Geocode(ctx, q)
	if err != nil {
		return domain.GeocodeResult{}, err
	}
	if err := res.Validate(); err != nil {
		return domain.GeocodeResult{}, err
	}
	if d.Cache != nil {
		_ = d.Cache.SetGeocode(ctx, hash, res, d.now().Add(geocodeCacheTTL))
	}
	res.Cached = false
	return res, nil
}

// ReverseGeocodeInput is the reverse geocode request.
type ReverseGeocodeInput struct {
	TenantID uuid.UUID
	Lat      float64
	Lng      float64
}

// ReverseGeocode resolves coordinates to an address (cached).
func (d *Deps) ReverseGeocode(ctx context.Context, in ReverseGeocodeInput) (domain.GeocodeResult, error) {
	if in.TenantID == uuid.Nil {
		return domain.GeocodeResult{}, fmt.Errorf("%w: tenant_id required", domain.ErrInvalidArgument)
	}
	if !domain.ValidLatLng(in.Lat, in.Lng) {
		return domain.GeocodeResult{}, fmt.Errorf("%w: lat/lng out of range", domain.ErrInvalidArgument)
	}
	hash := queryHash("rev", in.TenantID.String(), fmt.Sprintf("%.6f,%.6f", in.Lat, in.Lng))
	if d.Cache != nil {
		if cached, ok, err := d.Cache.GetGeocode(ctx, hash); err == nil && ok {
			cached.Cached = true
			return cached, nil
		}
	}
	if d.Maps == nil {
		return domain.GeocodeResult{}, fmt.Errorf("%w: maps provider not configured", domain.ErrInvariant)
	}
	res, err := d.Maps.Reverse(ctx, in.Lat, in.Lng)
	if err != nil {
		return domain.GeocodeResult{}, err
	}
	if err := res.Validate(); err != nil {
		return domain.GeocodeResult{}, err
	}
	if d.Cache != nil {
		_ = d.Cache.SetGeocode(ctx, hash, res, d.now().Add(geocodeCacheTTL))
	}
	return res, nil
}

// AutocompleteInput is the autocomplete request.
type AutocompleteInput struct {
	TenantID uuid.UUID
	Query    string
	Limit    int
}

// Autocomplete returns place suggestions (cached by query).
func (d *Deps) Autocomplete(ctx context.Context, in AutocompleteInput) ([]domain.GeocodeResult, error) {
	if in.TenantID == uuid.Nil {
		return nil, fmt.Errorf("%w: tenant_id required", domain.ErrInvalidArgument)
	}
	q := strings.TrimSpace(in.Query)
	if q == "" {
		return nil, fmt.Errorf("%w: query required", domain.ErrInvalidArgument)
	}
	limit := in.Limit
	if limit <= 0 {
		limit = 5
	}
	hash := queryHash("ac", in.TenantID.String(), fmt.Sprintf("%s|%d", q, limit))
	if d.Cache != nil {
		if cached, ok, err := d.Cache.GetGeocode(ctx, hash); err == nil && ok {
			cached.Cached = true
			return []domain.GeocodeResult{cached}, nil
		}
	}
	if d.Maps == nil {
		return nil, fmt.Errorf("%w: maps provider not configured", domain.ErrInvariant)
	}
	res, err := d.Maps.Autocomplete(ctx, q, limit)
	if err != nil {
		return nil, err
	}
	if d.Cache != nil && len(res) > 0 {
		_ = d.Cache.SetGeocode(ctx, hash, res[0], d.now().Add(geocodeCacheTTL))
	}
	return res, nil
}

func queryHash(parts ...string) string {
	h := sha256.Sum256([]byte(strings.Join(parts, "|")))
	return hex.EncodeToString(h[:])
}
