package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/nexora/location-service/internal/domain"
)

// IngestHistoryInput records a location history ping.
type IngestHistoryInput struct {
	TenantID    uuid.UUID
	SubjectType domain.SubjectType
	SubjectID   string
	Lat         float64
	Lng         float64
}

// IngestHistory appends a capped history row (not live GPS SoT).
func (d *Deps) IngestHistory(ctx context.Context, in IngestHistoryInput) (domain.LocationHistory, error) {
	if in.TenantID == uuid.Nil {
		return domain.LocationHistory{}, fmt.Errorf("%w: tenant_id required", domain.ErrInvalidArgument)
	}
	h := domain.LocationHistory{
		ID: d.newID(), TenantID: in.TenantID,
		SubjectType: in.SubjectType, SubjectID: strings.TrimSpace(in.SubjectID),
		Lat: in.Lat, Lng: in.Lng, RecordedAt: d.now(),
	}
	if err := h.Validate(); err != nil {
		return domain.LocationHistory{}, err
	}
	if d.History == nil {
		return domain.LocationHistory{}, fmt.Errorf("%w: history repo not configured", domain.ErrInvariant)
	}
	if err := d.History.Ingest(ctx, h); err != nil {
		return domain.LocationHistory{}, err
	}
	d.emit(ctx, h.TenantID, h.ID, domain.EventLocationUpdated, map[string]any{
		"subjectType": string(h.SubjectType), "subjectId": h.SubjectID,
	})
	return h, nil
}

// GetHistoryInput lists history for a subject.
type GetHistoryInput struct {
	TenantID    uuid.UUID
	SubjectType domain.SubjectType
	SubjectID   string
	Limit       int
}

// GetHistory returns capped location history for a subject.
func (d *Deps) GetHistory(ctx context.Context, in GetHistoryInput) ([]domain.LocationHistory, error) {
	if in.TenantID == uuid.Nil {
		return nil, fmt.Errorf("%w: tenant_id required", domain.ErrInvalidArgument)
	}
	if !in.SubjectType.Valid() {
		return nil, fmt.Errorf("%w: invalid subject_type", domain.ErrInvalidArgument)
	}
	if strings.TrimSpace(in.SubjectID) == "" {
		return nil, fmt.Errorf("%w: subject_id required", domain.ErrInvalidArgument)
	}
	if d.History == nil {
		return nil, fmt.Errorf("%w: history repo not configured", domain.ErrInvariant)
	}
	return d.History.List(ctx, in.TenantID, in.SubjectType, in.SubjectID, in.Limit)
}
