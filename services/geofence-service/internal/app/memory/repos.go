package memory

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/geofence-service/internal/app/ports"
	"github.com/nexora/geofence-service/internal/domain"
)

// Repos holds all memory repository implementations.
type Repos struct {
	Zones  *ZoneRepo
	Outbox *OutboxRepo
}

// NewRepos wires repos against a shared store.
func NewRepos(s *Store) *Repos {
	return &Repos{
		Zones:  &ZoneRepo{S: s},
		Outbox: &OutboxRepo{S: s},
	}
}

// ZoneRepo is an in-memory ZoneRepo.
type ZoneRepo struct{ S *Store }

func (r *ZoneRepo) Create(_ context.Context, z domain.Zone) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	if _, ok := r.S.zones[z.ID]; ok {
		return fmt.Errorf("%w: zone %s", domain.ErrAlreadyExists, z.ID)
	}
	r.S.zones[z.ID] = z
	return nil
}

func (r *ZoneRepo) Update(_ context.Context, z domain.Zone) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	cur, ok := r.S.zones[z.ID]
	if !ok || cur.TenantID != z.TenantID {
		return fmt.Errorf("%w: zone %s", domain.ErrNotFound, z.ID)
	}
	r.S.zones[z.ID] = z
	return nil
}

func (r *ZoneRepo) Delete(_ context.Context, tenantID, id uuid.UUID) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	cur, ok := r.S.zones[id]
	if !ok || cur.TenantID != tenantID {
		return fmt.Errorf("%w: zone %s", domain.ErrNotFound, id)
	}
	delete(r.S.zones, id)
	return nil
}

func (r *ZoneRepo) Get(_ context.Context, tenantID, id uuid.UUID) (domain.Zone, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	z, ok := r.S.zones[id]
	if !ok || z.TenantID != tenantID {
		return domain.Zone{}, fmt.Errorf("%w: zone %s", domain.ErrNotFound, id)
	}
	return z, nil
}

func (r *ZoneRepo) List(_ context.Context, tenantID uuid.UUID, city string, kind domain.ZoneKind, limit, offset int) ([]domain.Zone, int, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	all := make([]domain.Zone, 0)
	for _, z := range r.S.zones {
		if z.TenantID != tenantID {
			continue
		}
		if city != "" && z.City != city {
			continue
		}
		if kind != "" && z.Kind != kind {
			continue
		}
		all = append(all, z)
	}
	total := len(all)
	if offset > total {
		offset = total
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return all[offset:end], total, nil
}

func (r *ZoneRepo) ListActive(_ context.Context, tenantID uuid.UUID, city string) ([]domain.Zone, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	out := make([]domain.Zone, 0)
	for _, z := range r.S.zones {
		if z.TenantID != tenantID || !z.Active {
			continue
		}
		if city != "" && z.City != "" && z.City != city {
			continue
		}
		out = append(out, z)
	}
	return out, nil
}

// OutboxRepo is an in-memory OutboxRepository.
type OutboxRepo struct{ S *Store }

func (r *OutboxRepo) Enqueue(_ context.Context, msg domain.OutboxMessage) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.outbox = append(r.S.outbox, msg)
	return nil
}

func (r *OutboxRepo) Update(_ context.Context, msg domain.OutboxMessage) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	for i := range r.S.outbox {
		if r.S.outbox[i].ID == msg.ID {
			r.S.outbox[i] = msg
			return nil
		}
	}
	return fmt.Errorf("%w: outbox %s", domain.ErrNotFound, msg.ID)
}

func (r *OutboxRepo) ListPending(_ context.Context, limit int) ([]domain.OutboxMessage, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	out := make([]domain.OutboxMessage, 0)
	for _, m := range r.S.outbox {
		if m.Status == domain.OutboxStatusPending {
			out = append(out, m)
			if limit > 0 && len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}

var (
	_ ports.ZoneRepo          = (*ZoneRepo)(nil)
	_ ports.OutboxRepository  = (*OutboxRepo)(nil)
	_ ports.EventPublisher    = (*EventPublisher)(nil)
	_ ports.Clock             = (*Clock)(nil)
	_ ports.IDGen             = (IDGen)(IDGen{})
)
