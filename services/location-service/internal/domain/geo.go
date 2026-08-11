package domain

import (
	"fmt"
	"math"
)

const (
	// EarthRadiusMeters is the mean Earth radius used by Haversine.
	EarthRadiusMeters = 6371000.0
)

// LatLng is a WGS84 geographic point.
type LatLng struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

// Validate checks WGS84 bounds.
func (p LatLng) Validate() error {
	if !ValidLatLng(p.Lat, p.Lng) {
		return fmt.Errorf("%w: lat/lng out of range", ErrInvalidArgument)
	}
	return nil
}

// ValidLatLng reports whether coordinates are within WGS84 bounds.
func ValidLatLng(lat, lng float64) bool {
	return lat >= -90 && lat <= 90 && lng >= -180 && lng <= 180
}

// HaversineDistanceMeters returns great-circle distance in meters.
func HaversineDistanceMeters(a, b LatLng) float64 {
	φ1 := a.Lat * math.Pi / 180
	φ2 := b.Lat * math.Pi / 180
	Δφ := (b.Lat - a.Lat) * math.Pi / 180
	Δλ := (b.Lng - a.Lng) * math.Pi / 180

	x := math.Sin(Δφ/2)*math.Sin(Δφ/2) +
		math.Cos(φ1)*math.Cos(φ2)*math.Sin(Δλ/2)*math.Sin(Δλ/2)
	c := 2 * math.Atan2(math.Sqrt(x), math.Sqrt(1-x))
	return EarthRadiusMeters * c
}

// BBox is an axis-aligned geographic bounding box.
type BBox struct {
	MinLat float64 `json:"minLat"`
	MinLng float64 `json:"minLng"`
	MaxLat float64 `json:"maxLat"`
	MaxLng float64 `json:"maxLng"`
}

// Validate checks bbox invariants.
func (b BBox) Validate() error {
	if !ValidLatLng(b.MinLat, b.MinLng) || !ValidLatLng(b.MaxLat, b.MaxLng) {
		return fmt.Errorf("%w: bbox lat/lng out of range", ErrInvalidArgument)
	}
	if b.MinLat > b.MaxLat || b.MinLng > b.MaxLng {
		return fmt.Errorf("%w: bbox min must be <= max", ErrInvalidArgument)
	}
	return nil
}

// Contains reports whether p is inside the bbox (inclusive).
func (b BBox) Contains(p LatLng) bool {
	return p.Lat >= b.MinLat && p.Lat <= b.MaxLat &&
		p.Lng >= b.MinLng && p.Lng <= b.MaxLng
}
