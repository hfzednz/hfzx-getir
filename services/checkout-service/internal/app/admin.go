package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/nexora/checkout-service/internal/app/ports"
	"github.com/nexora/checkout-service/internal/domain"
)

// ListSessionsInput filters admin explorer listings.
type ListSessionsInput struct {
	TenantID    uuid.UUID
	PrincipalID *uuid.UUID
	Status      *domain.SessionStatus
	Query       string
	Limit       int
	Offset      int
}

// ListSessions returns paginated sessions for admin explorer.
func (d *Deps) ListSessions(ctx context.Context, in ListSessionsInput) ([]domain.Session, int, error) {
	limit := in.Limit
	if limit <= 0 {
		limit = 50
	}
	return d.Sessions.List(ctx, ports.SessionFilter{
		TenantID: in.TenantID, PrincipalID: in.PrincipalID,
		Status: in.Status, Query: in.Query, Limit: limit, Offset: in.Offset,
	})
}

// AbandonmentMetrics is admin abandonment / funnel metrics.
type AbandonmentMetrics struct {
	ByStatus       map[string]int `json:"byStatus"`
	AbandonedCount int            `json:"abandonedCount"`
	CompletedCount int            `json:"completedCount"`
	BlockedCount   int            `json:"blockedCount"`
	StartedCount   int            `json:"startedCount"`
	ReadyCount     int            `json:"readyCount"`
	AbandonRate    float64        `json:"abandonRate"`
}

// Metrics returns abandonment and status funnel metrics for a tenant.
func (d *Deps) Metrics(ctx context.Context, tenantID uuid.UUID) (AbandonmentMetrics, error) {
	counts, err := d.Sessions.CountByStatus(ctx, tenantID)
	if err != nil {
		return AbandonmentMetrics{}, err
	}
	m := AbandonmentMetrics{ByStatus: map[string]int{}}
	total := 0
	for st, n := range counts {
		m.ByStatus[string(st)] = n
		total += n
		switch st {
		case domain.StatusAbandoned:
			m.AbandonedCount = n
		case domain.StatusCompleted:
			m.CompletedCount = n
		case domain.StatusBlocked:
			m.BlockedCount = n
		case domain.StatusStarted:
			m.StartedCount = n
		case domain.StatusReady:
			m.ReadyCount = n
		}
	}
	denom := m.AbandonedCount + m.CompletedCount
	if denom > 0 {
		m.AbandonRate = float64(m.AbandonedCount) / float64(denom)
	}
	_ = total
	return m, nil
}
