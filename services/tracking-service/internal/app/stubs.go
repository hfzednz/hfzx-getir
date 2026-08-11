package app

import (
	"context"

	"github.com/nexora/tracking-service/internal/app/ports"
)

// NoopGeofenceClient returns no zone hits.
type NoopGeofenceClient struct{}

func (NoopGeofenceClient) Check(_ context.Context, _ ports.GeofenceCheckRequest) (ports.GeofenceCheckResult, error) {
	return ports.GeofenceCheckResult{}, nil
}
