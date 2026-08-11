package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ConfidenceScore is a geocode/address confidence in [0,1].
type ConfidenceScore float64

// Valid reports whether the score is in range.
func (c ConfidenceScore) Valid() bool {
	return c >= 0 && c <= 1
}

// AddressComponents holds structured address parts.
type AddressComponents struct {
	Country      string `json:"country,omitempty"`
	City         string `json:"city,omitempty"`
	District     string `json:"district,omitempty"`
	Neighborhood string `json:"neighborhood,omitempty"`
	Street       string `json:"street,omitempty"`
	PostalCode   string `json:"postalCode,omitempty"`
}

// NormalizedAddress is the enriched address aggregate.
type NormalizedAddress struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	Line1       string
	Building    string
	Entrance    string
	Floor       string
	Apt         string
	Landmark    string
	Lat         float64
	Lng         float64
	PlaceID     string
	Confidence  ConfidenceScore
	RiskScore   float64
	Components  AddressComponents
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Validate checks address invariants.
func (a NormalizedAddress) Validate() error {
	if a.ID == uuid.Nil {
		return fmt.Errorf("%w: address id required", ErrInvalidArgument)
	}
	if a.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if strings.TrimSpace(a.Line1) == "" && a.PlaceID == "" {
		return fmt.Errorf("%w: line1 or place_id required", ErrInvalidArgument)
	}
	if !ValidLatLng(a.Lat, a.Lng) {
		return fmt.Errorf("%w: lat/lng out of range", ErrInvalidArgument)
	}
	if !a.Confidence.Valid() {
		return fmt.Errorf("%w: confidence out of range", ErrInvalidArgument)
	}
	if a.RiskScore < 0 || a.RiskScore > 1 {
		return fmt.Errorf("%w: risk_score out of range", ErrInvalidArgument)
	}
	return nil
}

// Normalize trims and collapses whitespace on textual fields.
func Normalize(a NormalizedAddress) NormalizedAddress {
	a.Line1 = collapseWS(a.Line1)
	a.Building = collapseWS(a.Building)
	a.Entrance = collapseWS(a.Entrance)
	a.Floor = collapseWS(a.Floor)
	a.Apt = collapseWS(a.Apt)
	a.Landmark = collapseWS(a.Landmark)
	a.PlaceID = strings.TrimSpace(a.PlaceID)
	a.Components.Country = collapseWS(a.Components.Country)
	a.Components.City = collapseWS(a.Components.City)
	a.Components.District = collapseWS(a.Components.District)
	a.Components.Neighborhood = collapseWS(a.Components.Neighborhood)
	a.Components.Street = collapseWS(a.Components.Street)
	a.Components.PostalCode = collapseWS(a.Components.PostalCode)
	return a
}

func collapseWS(s string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(s)), " ")
}

// DeliveryFeasibility is the geofence serviceability outcome.
type DeliveryFeasibility struct {
	Serviceable bool       `json:"serviceable"`
	ZoneID      *uuid.UUID `json:"zoneId,omitempty"`
	Reason      string     `json:"reason,omitempty"`
	Score       float64    `json:"score"`
}
