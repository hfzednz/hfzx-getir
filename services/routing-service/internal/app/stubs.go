package app

import (
	"context"

	"github.com/nexora/routing-service/internal/app/ports"
	"github.com/nexora/routing-service/internal/domain"
)

// HaversineMapsClient builds a distance matrix from Haversine (Google Maps stub).
type HaversineMapsClient struct {
	SpeedMPS float64
}

func (c HaversineMapsClient) DistanceMatrix(_ context.Context, req ports.DistanceMatrixRequest) (ports.DistanceMatrixResult, error) {
	speed := c.SpeedMPS
	if speed <= 0 {
		speed = domain.DefaultSpeedMPS
	}
	dist := make([][]float64, len(req.Origins))
	dur := make([][]float64, len(req.Origins))
	for i, o := range req.Origins {
		dist[i] = make([]float64, len(req.Destinations))
		dur[i] = make([]float64, len(req.Destinations))
		for j, d := range req.Destinations {
			m := domain.HaversineMeters(o.Lat, o.Lon, d.Lat, d.Lon)
			dist[i][j] = m
			dur[i][j] = domain.ETASeconds(m, speed, 1, 1)
		}
	}
	return ports.DistanceMatrixResult{DistancesMeters: dist, DurationsSeconds: dur}, nil
}

// FixedTrafficClient returns a constant traffic factor.
type FixedTrafficClient struct{ Value float64 }

func (c FixedTrafficClient) Factor(_ context.Context, _ ports.TrafficFactorRequest) (float64, error) {
	if c.Value <= 0 {
		return 1, nil
	}
	return c.Value, nil
}

// FixedWeatherClient returns a constant weather factor.
type FixedWeatherClient struct{ Value float64 }

func (c FixedWeatherClient) Factor(_ context.Context, _ ports.WeatherFactorRequest) (float64, error) {
	if c.Value <= 0 {
		return 1, nil
	}
	return c.Value, nil
}
