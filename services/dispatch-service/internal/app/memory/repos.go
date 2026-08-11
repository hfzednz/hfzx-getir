package memory

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/dispatch-service/internal/app/ports"
	"github.com/nexora/dispatch-service/internal/domain"
)

// Repos holds all memory repository implementations.
type Repos struct {
	Dispatches *DispatchRepo
	Couriers   *CourierPool
	Vehicles   *VehicleRepo
	Outbox     *OutboxRepo
}

// NewRepos wires repos against a shared store.
func NewRepos(s *Store) *Repos {
	return &Repos{
		Dispatches: &DispatchRepo{S: s},
		Couriers:   &CourierPool{S: s},
		Vehicles:   &VehicleRepo{S: s},
		Outbox:     &OutboxRepo{S: s},
	}
}

// DispatchRepo is an in-memory DispatchRepo.
type DispatchRepo struct{ S *Store }

func (r *DispatchRepo) Create(_ context.Context, d domain.Dispatch) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	if _, ok := r.S.dispatches[d.ID]; ok {
		return fmt.Errorf("%w: dispatch %s", domain.ErrAlreadyExists, d.ID)
	}
	r.S.dispatches[d.ID] = d
	return nil
}

func (r *DispatchRepo) Update(_ context.Context, d domain.Dispatch) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	cur, ok := r.S.dispatches[d.ID]
	if !ok || cur.TenantID != d.TenantID {
		return fmt.Errorf("%w: dispatch %s", domain.ErrNotFound, d.ID)
	}
	r.S.dispatches[d.ID] = d
	return nil
}

func (r *DispatchRepo) Get(_ context.Context, tenantID, id uuid.UUID) (domain.Dispatch, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	d, ok := r.S.dispatches[id]
	if !ok || d.TenantID != tenantID {
		return domain.Dispatch{}, fmt.Errorf("%w: dispatch %s", domain.ErrNotFound, id)
	}
	return d, nil
}

func (r *DispatchRepo) List(_ context.Context, tenantID uuid.UUID, status domain.DispatchStatus, limit, offset int) ([]domain.Dispatch, int, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	all := make([]domain.Dispatch, 0)
	for _, d := range r.S.dispatches {
		if d.TenantID != tenantID {
			continue
		}
		if status != "" && d.Status != status {
			continue
		}
		all = append(all, d)
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

func (r *DispatchRepo) AppendEvent(_ context.Context, e domain.DispatchEvent) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.events = append(r.S.events, e)
	return nil
}

func (r *DispatchRepo) ListEvents(_ context.Context, tenantID, dispatchID uuid.UUID) ([]domain.DispatchEvent, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	out := make([]domain.DispatchEvent, 0)
	for _, e := range r.S.events {
		if e.TenantID == tenantID && e.DispatchID == dispatchID {
			out = append(out, e)
		}
	}
	return out, nil
}

func (r *DispatchRepo) AppendAttempt(_ context.Context, a domain.AssignmentAttempt) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.attempts = append(r.S.attempts, a)
	return nil
}

func (r *DispatchRepo) CreateBatch(_ context.Context, b domain.Batch) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.batches[b.ID] = b
	return nil
}

func (r *DispatchRepo) GetBatch(_ context.Context, tenantID, id uuid.UUID) (domain.Batch, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	b, ok := r.S.batches[id]
	if !ok || b.TenantID != tenantID {
		return domain.Batch{}, fmt.Errorf("%w: batch %s", domain.ErrNotFound, id)
	}
	return b, nil
}

// CourierPool is an in-memory CourierPool.
type CourierPool struct{ S *Store }

func (r *CourierPool) Upsert(_ context.Context, c domain.CourierSnapshot) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.couriers[courierKey(c.TenantID, c.CourierPrincipalID)] = c
	return nil
}

func (r *CourierPool) Get(_ context.Context, tenantID, courierPrincipalID uuid.UUID) (domain.CourierSnapshot, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	c, ok := r.S.couriers[courierKey(tenantID, courierPrincipalID)]
	if !ok {
		return domain.CourierSnapshot{}, fmt.Errorf("%w: courier %s", domain.ErrNotFound, courierPrincipalID)
	}
	return c, nil
}

func (r *CourierPool) ListAvailable(_ context.Context, tenantID uuid.UUID) ([]domain.CourierSnapshot, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	out := make([]domain.CourierSnapshot, 0)
	for _, c := range r.S.couriers {
		if c.TenantID == tenantID {
			out = append(out, c)
		}
	}
	return out, nil
}

func (r *CourierPool) AdjustLoad(_ context.Context, tenantID, courierPrincipalID uuid.UUID, delta int) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	key := courierKey(tenantID, courierPrincipalID)
	c, ok := r.S.couriers[key]
	if !ok {
		return fmt.Errorf("%w: courier %s", domain.ErrNotFound, courierPrincipalID)
	}
	c.CurrentLoad += delta
	if c.CurrentLoad < 0 {
		c.CurrentLoad = 0
	}
	r.S.couriers[key] = c
	return nil
}

// VehicleRepo is an in-memory VehicleRepo.
type VehicleRepo struct{ S *Store }

func (r *VehicleRepo) Upsert(_ context.Context, v domain.Vehicle) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.vehicles[v.ID] = v
	return nil
}

func (r *VehicleRepo) Get(_ context.Context, tenantID, id uuid.UUID) (domain.Vehicle, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	v, ok := r.S.vehicles[id]
	if !ok || v.TenantID != tenantID {
		return domain.Vehicle{}, fmt.Errorf("%w: vehicle %s", domain.ErrNotFound, id)
	}
	return v, nil
}

func (r *VehicleRepo) List(_ context.Context, tenantID uuid.UUID, limit, offset int) ([]domain.Vehicle, int, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	all := make([]domain.Vehicle, 0)
	for _, v := range r.S.vehicles {
		if v.TenantID == tenantID {
			all = append(all, v)
		}
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
	_ ports.DispatchRepo     = (*DispatchRepo)(nil)
	_ ports.CourierPool      = (*CourierPool)(nil)
	_ ports.VehicleRepo      = (*VehicleRepo)(nil)
	_ ports.OutboxRepository = (*OutboxRepo)(nil)
	_ ports.EventPublisher   = (*EventPublisher)(nil)
)
