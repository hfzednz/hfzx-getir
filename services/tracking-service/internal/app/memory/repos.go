package memory

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/tracking-service/internal/app/ports"
	"github.com/nexora/tracking-service/internal/domain"
)

// Repos bundles in-memory port implementations.
type Repos struct {
	Locations *LocationRepo
	Timelines *TimelineRepo
	Outbox    *OutboxRepo
}

// NewRepos wires all memory repositories to a shared store.
func NewRepos(s *Store) *Repos {
	return &Repos{
		Locations: &LocationRepo{S: s},
		Timelines: &TimelineRepo{S: s},
		Outbox:    &OutboxRepo{S: s},
	}
}

// LocationRepo is an in-memory LocationRepo.
type LocationRepo struct{ S *Store }

func (r *LocationRepo) UpsertLatest(_ context.Context, loc domain.CourierLocation) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.latest[courierKey{Tenant: loc.TenantID, Courier: loc.CourierID}] = loc
	return nil
}

func (r *LocationRepo) GetLatest(_ context.Context, tenantID, courierID uuid.UUID) (domain.CourierLocation, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	loc, ok := r.S.latest[courierKey{Tenant: tenantID, Courier: courierID}]
	if !ok {
		return domain.CourierLocation{}, fmt.Errorf("%w: courier location", domain.ErrNotFound)
	}
	return loc, nil
}

func (r *LocationRepo) AppendHistory(_ context.Context, entry domain.LocationHistoryEntry, cap int) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	if cap <= 0 {
		cap = domain.DefaultHistoryCap
	}
	key := courierKey{Tenant: entry.TenantID, Courier: entry.CourierID}
	hist := append(r.S.history[key], entry)
	if len(hist) > cap {
		hist = hist[len(hist)-cap:]
	}
	r.S.history[key] = hist
	return nil
}

func (r *LocationRepo) ListHistory(_ context.Context, tenantID, courierID uuid.UUID, limit int) ([]domain.LocationHistoryEntry, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	hist := r.S.history[courierKey{Tenant: tenantID, Courier: courierID}]
	if limit <= 0 || limit > len(hist) {
		limit = len(hist)
	}
	out := make([]domain.LocationHistoryEntry, limit)
	copy(out, hist[len(hist)-limit:])
	return out, nil
}

func (r *LocationRepo) ListNearby(_ context.Context, tenantID uuid.UUID, lat, lon, radiusM float64, limit int) ([]domain.CourierLocation, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	if limit <= 0 {
		limit = 50
	}
	out := make([]domain.CourierLocation, 0)
	for _, loc := range r.S.latest {
		if loc.TenantID != tenantID {
			continue
		}
		if domain.HaversineMeters(lat, lon, loc.Lat, loc.Lon) <= radiusM {
			out = append(out, loc)
			if len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}

var _ ports.LocationRepo = (*LocationRepo)(nil)

// TimelineRepo is an in-memory TimelineRepo.
type TimelineRepo struct{ S *Store }

func (r *TimelineRepo) Append(_ context.Context, e domain.TimelineEvent) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.timelines = append(r.S.timelines, e)
	return nil
}

func (r *TimelineRepo) ListByOrder(_ context.Context, tenantID, orderID uuid.UUID, limit int) ([]domain.TimelineEvent, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	if limit <= 0 {
		limit = 100
	}
	out := make([]domain.TimelineEvent, 0)
	for _, e := range r.S.timelines {
		if e.TenantID == tenantID && e.OrderID == orderID {
			out = append(out, e)
			if len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}

func (r *TimelineRepo) SaveGeofenceEvent(_ context.Context, e domain.GeofenceEvent) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.geofence = append(r.S.geofence, e)
	return nil
}

func (r *TimelineRepo) ListGeofenceEvents(_ context.Context, tenantID uuid.UUID, courierID *uuid.UUID, limit int) ([]domain.GeofenceEvent, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	if limit <= 0 {
		limit = 100
	}
	out := make([]domain.GeofenceEvent, 0)
	for _, e := range r.S.geofence {
		if e.TenantID != tenantID {
			continue
		}
		if courierID != nil && e.CourierID != *courierID {
			continue
		}
		out = append(out, e)
		if len(out) >= limit {
			break
		}
	}
	return out, nil
}

var _ ports.TimelineRepo = (*TimelineRepo)(nil)

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
