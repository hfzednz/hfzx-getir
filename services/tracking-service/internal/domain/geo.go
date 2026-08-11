package domain

import "math"

const (
	// EarthRadiusMeters is the mean Earth radius used by Haversine.
	EarthRadiusMeters = 6371000.0
	// DefaultArrivalThresholdM is the default distance for arrival detection.
	DefaultArrivalThresholdM = 50.0
	// DefaultHistoryCap is the max location history rows retained per courier.
	DefaultHistoryCap = 100
)

// HaversineMeters returns great-circle distance in meters between two WGS84 points.
func HaversineMeters(lat1, lon1, lat2, lon2 float64) float64 {
	φ1 := lat1 * math.Pi / 180
	φ2 := lat2 * math.Pi / 180
	Δφ := (lat2 - lat1) * math.Pi / 180
	Δλ := (lon2 - lon1) * math.Pi / 180

	a := math.Sin(Δφ/2)*math.Sin(Δφ/2) +
		math.Cos(φ1)*math.Cos(φ2)*math.Sin(Δλ/2)*math.Sin(Δλ/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return EarthRadiusMeters * c
}

// ValidLatLon reports whether coordinates are within WGS84 bounds.
func ValidLatLon(lat, lon float64) bool {
	return lat >= -90 && lat <= 90 && lon >= -180 && lon <= 180
}

// WithinThreshold reports whether distance between points is <= threshold meters.
func WithinThreshold(lat1, lon1, lat2, lon2, thresholdM float64) bool {
	if thresholdM <= 0 {
		thresholdM = DefaultArrivalThresholdM
	}
	return HaversineMeters(lat1, lon1, lat2, lon2) <= thresholdM
}
