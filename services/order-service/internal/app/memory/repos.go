package memory

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/nexora/order-service/internal/app/ports"
	"github.com/nexora/order-service/internal/domain"
)

// OrderRepo is an in-memory OrderRepository.
type OrderRepo struct{ S *Store }

func (r *OrderRepo) Create(_ context.Context, o domain.Order) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	key := tenantKey(o.TenantID, o.IdempotencyKey)
	if _, ok := r.S.OrdersByIdem[key]; ok {
		return fmt.Errorf("%w: idempotency key", domain.ErrAlreadyExists)
	}
	if _, ok := r.S.Orders[o.ID]; ok {
		return fmt.Errorf("%w: order", domain.ErrAlreadyExists)
	}
	cp := o
	cp.Lines = append([]domain.OrderLine(nil), o.Lines...)
	r.S.Orders[o.ID] = cp
	r.S.OrdersByIdem[key] = o.ID
	return nil
}

func (r *OrderRepo) Update(_ context.Context, o domain.Order) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	cur, ok := r.S.Orders[o.ID]
	if !ok || cur.TenantID != o.TenantID {
		return domain.ErrNotFound
	}
	if o.Version < cur.Version {
		return domain.ErrVersionConflict
	}
	cp := o
	cp.Lines = append([]domain.OrderLine(nil), o.Lines...)
	r.S.Orders[o.ID] = cp
	return nil
}

func (r *OrderRepo) GetByID(_ context.Context, tenantID, id uuid.UUID) (domain.Order, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	o, ok := r.S.Orders[id]
	if !ok || o.TenantID != tenantID {
		return domain.Order{}, domain.ErrNotFound
	}
	cp := o
	cp.Lines = append([]domain.OrderLine(nil), o.Lines...)
	return cp, nil
}

func (r *OrderRepo) GetByIdempotencyKey(_ context.Context, tenantID uuid.UUID, key string) (domain.Order, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	id, ok := r.S.OrdersByIdem[tenantKey(tenantID, key)]
	if !ok {
		return domain.Order{}, domain.ErrNotFound
	}
	o, ok := r.S.Orders[id]
	if !ok || o.TenantID != tenantID {
		return domain.Order{}, domain.ErrNotFound
	}
	cp := o
	cp.Lines = append([]domain.OrderLine(nil), o.Lines...)
	return cp, nil
}

func (r *OrderRepo) List(_ context.Context, f ports.OrderFilter) ([]domain.Order, int, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	out := make([]domain.Order, 0)
	for _, o := range r.S.Orders {
		if o.TenantID != f.TenantID {
			continue
		}
		if f.CustomerID != nil && o.CustomerPrincipalID != *f.CustomerID {
			continue
		}
		if f.Status != nil && o.Status != *f.Status {
			continue
		}
		if f.Query != "" {
			q := strings.ToLower(f.Query)
			if !strings.Contains(strings.ToLower(o.ID.String()), q) &&
				!strings.Contains(strings.ToLower(o.IdempotencyKey), q) &&
				!strings.Contains(strings.ToLower(string(o.Status)), q) {
				continue
			}
		}
		cp := o
		cp.Lines = append([]domain.OrderLine(nil), o.Lines...)
		out = append(out, cp)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	total := len(out)
	if f.Offset > len(out) {
		return nil, total, nil
	}
	end := f.Offset + f.Limit
	if f.Limit <= 0 {
		end = len(out)
	}
	if end > len(out) {
		end = len(out)
	}
	return out[f.Offset:end], total, nil
}

// EventStoreRepo is an in-memory EventStore.
type EventStoreRepo struct{ S *Store }

func (r *EventStoreRepo) Append(_ context.Context, e domain.OrderEvent) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.Events[e.ID] = e
	return nil
}

func (r *EventStoreRepo) ListByOrder(_ context.Context, tenantID, orderID uuid.UUID) ([]domain.OrderEvent, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	out := make([]domain.OrderEvent, 0)
	for _, e := range r.S.Events {
		if e.TenantID == tenantID && e.OrderID == orderID {
			out = append(out, e)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].OccurredAt.Before(out[j].OccurredAt) })
	return out, nil
}

// SagaRepo is an in-memory SagaRepository.
type SagaRepo struct{ S *Store }

func (r *SagaRepo) Create(_ context.Context, s domain.SagaInstance) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	key := tenantKey(s.TenantID, s.IdempotencyKey)
	if _, ok := r.S.SagasByIdem[key]; ok {
		return fmt.Errorf("%w: saga idempotency", domain.ErrAlreadyExists)
	}
	cp := s
	cp.Steps = append([]domain.SagaStep(nil), s.Steps...)
	r.S.Sagas[s.ID] = cp
	r.S.SagasByIdem[key] = s.ID
	return nil
}

func (r *SagaRepo) Update(_ context.Context, s domain.SagaInstance) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	if _, ok := r.S.Sagas[s.ID]; !ok {
		return domain.ErrNotFound
	}
	cp := s
	cp.Steps = append([]domain.SagaStep(nil), s.Steps...)
	r.S.Sagas[s.ID] = cp
	return nil
}

func (r *SagaRepo) GetByID(_ context.Context, tenantID, id uuid.UUID) (domain.SagaInstance, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	s, ok := r.S.Sagas[id]
	if !ok || s.TenantID != tenantID {
		return domain.SagaInstance{}, domain.ErrNotFound
	}
	cp := s
	cp.Steps = append([]domain.SagaStep(nil), s.Steps...)
	return cp, nil
}

func (r *SagaRepo) GetByIdempotencyKey(_ context.Context, tenantID uuid.UUID, key string) (domain.SagaInstance, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	id, ok := r.S.SagasByIdem[tenantKey(tenantID, key)]
	if !ok {
		return domain.SagaInstance{}, domain.ErrNotFound
	}
	s, ok := r.S.Sagas[id]
	if !ok || s.TenantID != tenantID {
		return domain.SagaInstance{}, domain.ErrNotFound
	}
	cp := s
	cp.Steps = append([]domain.SagaStep(nil), s.Steps...)
	return cp, nil
}

func (r *SagaRepo) ListByOrder(_ context.Context, tenantID, orderID uuid.UUID) ([]domain.SagaInstance, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	out := make([]domain.SagaInstance, 0)
	for _, s := range r.S.Sagas {
		if s.TenantID == tenantID && s.OrderID == orderID {
			cp := s
			cp.Steps = append([]domain.SagaStep(nil), s.Steps...)
			out = append(out, cp)
		}
	}
	return out, nil
}

// OutboxRepo is an in-memory OutboxRepository.
type OutboxRepo struct{ S *Store }

func (r *OutboxRepo) Enqueue(_ context.Context, m domain.OutboxMessage) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.Outbox[m.ID] = m
	return nil
}

func (r *OutboxRepo) Update(_ context.Context, m domain.OutboxMessage) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	if _, ok := r.S.Outbox[m.ID]; !ok {
		return domain.ErrNotFound
	}
	r.S.Outbox[m.ID] = m
	return nil
}

func (r *OutboxRepo) ListPending(_ context.Context, limit int) ([]domain.OutboxMessage, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	out := make([]domain.OutboxMessage, 0)
	for _, m := range r.S.Outbox {
		if m.Status == domain.OutboxStatusPending {
			out = append(out, m)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

// FulfillmentRepo is an in-memory FulfillmentRepository.
type FulfillmentRepo struct{ S *Store }

func (r *FulfillmentRepo) Create(_ context.Context, f domain.Fulfillment) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.Fulfillments[f.ID] = f
	return nil
}

func (r *FulfillmentRepo) Update(_ context.Context, f domain.Fulfillment) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	if _, ok := r.S.Fulfillments[f.ID]; !ok {
		return domain.ErrNotFound
	}
	r.S.Fulfillments[f.ID] = f
	return nil
}

func (r *FulfillmentRepo) GetByID(_ context.Context, tenantID, id uuid.UUID) (domain.Fulfillment, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	f, ok := r.S.Fulfillments[id]
	if !ok || f.TenantID != tenantID {
		return domain.Fulfillment{}, domain.ErrNotFound
	}
	return f, nil
}

func (r *FulfillmentRepo) ListByOrder(_ context.Context, tenantID, orderID uuid.UUID) ([]domain.Fulfillment, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	out := make([]domain.Fulfillment, 0)
	for _, f := range r.S.Fulfillments {
		if f.TenantID == tenantID && f.OrderID == orderID {
			out = append(out, f)
		}
	}
	return out, nil
}

// ReturnRepo is an in-memory ReturnRepository.
type ReturnRepo struct{ S *Store }

func (r *ReturnRepo) Create(_ context.Context, ret domain.Return) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	cp := ret
	cp.Lines = append([]domain.ReturnLine(nil), ret.Lines...)
	r.S.Returns[ret.ID] = cp
	return nil
}

func (r *ReturnRepo) Update(_ context.Context, ret domain.Return) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	if _, ok := r.S.Returns[ret.ID]; !ok {
		return domain.ErrNotFound
	}
	cp := ret
	cp.Lines = append([]domain.ReturnLine(nil), ret.Lines...)
	r.S.Returns[ret.ID] = cp
	return nil
}

func (r *ReturnRepo) GetByID(_ context.Context, tenantID, id uuid.UUID) (domain.Return, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	ret, ok := r.S.Returns[id]
	if !ok || ret.TenantID != tenantID {
		return domain.Return{}, domain.ErrNotFound
	}
	cp := ret
	cp.Lines = append([]domain.ReturnLine(nil), ret.Lines...)
	return cp, nil
}

func (r *ReturnRepo) ListByOrder(_ context.Context, tenantID, orderID uuid.UUID) ([]domain.Return, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	out := make([]domain.Return, 0)
	for _, ret := range r.S.Returns {
		if ret.TenantID == tenantID && ret.OrderID == orderID {
			cp := ret
			cp.Lines = append([]domain.ReturnLine(nil), ret.Lines...)
			out = append(out, cp)
		}
	}
	return out, nil
}

// RefundRepo is an in-memory RefundRepository.
type RefundRepo struct{ S *Store }

func (r *RefundRepo) Create(_ context.Context, ref domain.Refund) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.Refunds[ref.ID] = ref
	return nil
}

func (r *RefundRepo) Update(_ context.Context, ref domain.Refund) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	if _, ok := r.S.Refunds[ref.ID]; !ok {
		return domain.ErrNotFound
	}
	r.S.Refunds[ref.ID] = ref
	return nil
}

func (r *RefundRepo) GetByID(_ context.Context, tenantID, id uuid.UUID) (domain.Refund, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	ref, ok := r.S.Refunds[id]
	if !ok || ref.TenantID != tenantID {
		return domain.Refund{}, domain.ErrNotFound
	}
	return ref, nil
}

func (r *RefundRepo) ListByOrder(_ context.Context, tenantID, orderID uuid.UUID) ([]domain.Refund, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	out := make([]domain.Refund, 0)
	for _, ref := range r.S.Refunds {
		if ref.TenantID == tenantID && ref.OrderID == orderID {
			out = append(out, ref)
		}
	}
	return out, nil
}

// NewRepos wires memory repositories.
func NewRepos(s *Store) (
	*OrderRepo, *EventStoreRepo, *SagaRepo, *OutboxRepo,
	*FulfillmentRepo, *ReturnRepo, *RefundRepo,
) {
	return &OrderRepo{S: s}, &EventStoreRepo{S: s}, &SagaRepo{S: s}, &OutboxRepo{S: s},
		&FulfillmentRepo{S: s}, &ReturnRepo{S: s}, &RefundRepo{S: s}
}
