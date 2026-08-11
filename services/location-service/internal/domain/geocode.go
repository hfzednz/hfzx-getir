package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// GeocodeResult is a provider geocode / reverse / autocomplete hit.
type GeocodeResult struct {
	PlaceID     string             `json:"placeId,omitempty"`
	Formatted   string             `json:"formatted,omitempty"`
	Lat         float64            `json:"lat"`
	Lng         float64            `json:"lng"`
	Confidence  ConfidenceScore    `json:"confidence"`
	Components  AddressComponents  `json:"components,omitempty"`
	Provider    string             `json:"provider,omitempty"`
	Cached      bool               `json:"cached,omitempty"`
}

// Validate checks geocode result invariants.
func (g GeocodeResult) Validate() error {
	if !ValidLatLng(g.Lat, g.Lng) {
		return fmt.Errorf("%w: lat/lng out of range", ErrInvalidArgument)
	}
	if !g.Confidence.Valid() {
		return fmt.Errorf("%w: confidence out of range", ErrInvalidArgument)
	}
	return nil
}

// OfflineManifest describes a downloadable offline map region package.
type OfflineManifest struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	Region    string
	Version   string
	URL       string
	SizeBytes int64
	UpdatedAt time.Time
}

// Validate checks offline manifest invariants.
func (m OfflineManifest) Validate() error {
	if m.ID == uuid.Nil {
		return fmt.Errorf("%w: manifest id required", ErrInvalidArgument)
	}
	if m.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if strings.TrimSpace(m.Region) == "" {
		return fmt.Errorf("%w: region required", ErrInvalidArgument)
	}
	if strings.TrimSpace(m.Version) == "" {
		return fmt.Errorf("%w: version required", ErrInvalidArgument)
	}
	if strings.TrimSpace(m.URL) == "" {
		return fmt.Errorf("%w: url required", ErrInvalidArgument)
	}
	if m.SizeBytes < 0 {
		return fmt.Errorf("%w: size_bytes must be >= 0", ErrInvalidArgument)
	}
	return nil
}

// ProviderConfig toggles map SDK providers per tenant (tile SDKs on clients).
type ProviderConfig struct {
	TenantID uuid.UUID
	Google   bool
	Apple    bool
	Mapbox   bool
	OSM      bool
	UpdatedAt time.Time
}
