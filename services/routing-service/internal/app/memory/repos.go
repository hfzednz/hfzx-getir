package memory

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/routing-service/internal/app/ports"
	"github.com/nexora/routing-service/internal/domain"
)

// Repos bundles in-memory port implementations.
type Repos struct {
	Routes *RouteRepo
	Outbox *OutboxRepo
}

// NewRepos wires all memory repositories to a shared store.
func NewRepos(s *Store) *Repos {
	return &Repos{
		Routes: &RouteRepo{S: s},
		Outbox: &OutboxRepo{S: s},
	}
}

// RouteRepo is an in-memory RouteRepo.
type RouteRepo struct{ S *Store }

func (r *RouteRepo) SaveRoute(_ context.Context, route domain.Route) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	cp := route
	cp.Waypoints = append([]domain.Waypoint(nil), route.Waypoints...)
	cp.Legs = append([]domain.RouteLeg(nil), route.Legs...)
	r.S.routes[route.ID] = cp
	return nil
}

func (r *RouteRepo) GetRoute(_ context.Context, tenantID, routeID uuid.UUID) (domain.Route, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	route, ok := r.S.routes[routeID]
	if !ok || route.TenantID != tenantID {
		return domain.Route{}, fmt.Errorf("%w: route", domain.ErrNotFound)
	}
	cp := route
	cp.Waypoints = append([]domain.Waypoint(nil), route.Waypoints...)
	cp.Legs = append([]domain.RouteLeg(nil), route.Legs...)
	return cp, nil
}

func (r *RouteRepo) ListRoutes(_ context.Context, tenantID uuid.UUID, limit int) ([]domain.Route, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	if limit <= 0 {
		limit = 100
	}
	out := make([]domain.Route, 0)
	for _, route := range r.S.routes {
		if route.TenantID != tenantID {
			continue
		}
		cp := route
		cp.Waypoints = append([]domain.Waypoint(nil), route.Waypoints...)
		cp.Legs = append([]domain.RouteLeg(nil), route.Legs...)
		out = append(out, cp)
		if len(out) >= limit {
			break
		}
	}
	return out, nil
}

func (r *RouteRepo) SaveETASnapshot(_ context.Context, s domain.ETASnapshot) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.etas = append(r.S.etas, s)
	return nil
}

func (r *RouteRepo) ListETASnapshots(_ context.Context, tenantID, routeID uuid.UUID, limit int) ([]domain.ETASnapshot, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	if limit <= 0 {
		limit = 50
	}
	out := make([]domain.ETASnapshot, 0)
	for i := len(r.S.etas) - 1; i >= 0; i-- {
		s := r.S.etas[i]
		if s.TenantID == tenantID && s.RouteID == routeID {
			out = append(out, s)
			if len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}

func (r *RouteRepo) UpsertTrafficHint(_ context.Context, h domain.TrafficHint) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.hints[h.ID] = h
	return nil
}

func (r *RouteRepo) GetTrafficHint(_ context.Context, tenantID, id uuid.UUID) (domain.TrafficHint, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	h, ok := r.S.hints[id]
	if !ok || h.TenantID != tenantID {
		return domain.TrafficHint{}, fmt.Errorf("%w: traffic hint", domain.ErrNotFound)
	}
	return h, nil
}

func (r *RouteRepo) ListActiveTrafficHints(_ context.Context, tenantID uuid.UUID, at time.Time) ([]domain.TrafficHint, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	out := make([]domain.TrafficHint, 0)
	for _, h := range r.S.hints {
		if h.TenantID != tenantID {
			continue
		}
		if !at.Before(h.ValidFrom) && at.Before(h.ValidUntil) {
			out = append(out, h)
		}
	}
	return out, nil
}

var _ ports.RouteRepo = (*RouteRepo)(nil)

// OutboxRepo is an in-memory OutboxRepository.
type OutboxRepo struct{ S *Store }

func (r *OutboxRepo) Enqueue(_ context.Context, m domain.OutboxMessage) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.outbox = append(r.S.outbox, m)
	return nil
}

func (r *OutboxRepo) Update(_ context.Context, m domain.OutboxMessage) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	for i := range r.S.outbox {
		if r.S.outbox[i].ID == m.ID {
			r.S.outbox[i] = m
			return nil
		}
	}
	return fmt.Errorf("%w: outbox", domain.ErrNotFound)
}

func (r *OutboxRepo) ListPending(_ context.Context, limit int) ([]domain.OutboxMessage, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	if limit <= 0 {
		limit = 100
	}
	out := make([]domain.OutboxMessage, 0)
	for _, m := range r.S.outbox {
		if m.Status == domain.OutboxStatusPending {
			out = append(out, m)
			if len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}

var _ ports.OutboxRepository = (*OutboxRepo)(nil)
