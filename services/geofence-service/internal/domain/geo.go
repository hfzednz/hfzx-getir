package domain

import "math"

// Point is a WGS84 coordinate.
type Point struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

// PointInPolygon reports whether p is inside a closed polygon using ray casting.
// Vertices should be ordered; the polygon is closed automatically.
func PointInPolygon(p Point, vertices []Point) bool {
	n := len(vertices)
	if n < 3 {
		return false
	}
	inside := false
	j := n - 1
	for i := 0; i < n; i++ {
		yi, xi := vertices[i].Lat, vertices[i].Lng
		yj, xj := vertices[j].Lat, vertices[j].Lng
		intersect := ((yi > p.Lat) != (yj > p.Lat)) &&
			(p.Lng < (xj-xi)*(p.Lat-yi)/(yj-yi+1e-18)+xi)
		if intersect {
			inside = !inside
		}
		j = i
	}
	return inside
}

// PointInRadius reports whether p is within radiusMeters of center (haversine).
func PointInRadius(p, center Point, radiusMeters float64) bool {
	if radiusMeters < 0 {
		return false
	}
	return HaversineMeters(p, center) <= radiusMeters
}

// HaversineMeters returns great-circle distance in meters.
func HaversineMeters(a, b Point) float64 {
	const earth = 6371000.0
	lat1 := a.Lat * math.Pi / 180
	lat2 := b.Lat * math.Pi / 180
	dLat := (b.Lat - a.Lat) * math.Pi / 180
	dLng := (b.Lng - a.Lng) * math.Pi / 180
	h := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1)*math.Cos(lat2)*math.Sin(dLng/2)*math.Sin(dLng/2)
	return 2 * earth * math.Asin(math.Min(1, math.Sqrt(h)))
}
