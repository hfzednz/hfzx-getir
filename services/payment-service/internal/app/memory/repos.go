package memory

import (
	"context"
	"sort"

	"github.com/google/uuid"
	"github.com/nexora/payment-service/internal/app/ports"
	"github.com/nexora/payment-service/internal/domain"
)

// IntentRepo is an in-memory IntentRepo.
type IntentRepo struct{ S *Store }

// OutboxRepo is an in-memory OutboxRepository.
type OutboxRepo struct{ S *Store }

// NewRepos returns intent + outbox repos sharing a store.
func NewRepos(s *Store) (*IntentRepo, *OutboxRepo) {
	return &IntentRepo{S: s}, &OutboxRepo{S: s}
}

func (r *IntentRepo) CreateIntent(_ context.Context, i domain.PaymentIntent) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	k := idemKey(i.TenantID, i.IdempotencyKey)
	if _, ok := r.S.byIdem[k]; ok {
		return domain.ErrAlreadyExists
	}
	r.S.intents[i.ID] = i
	r.S.byIdem[k] = i.ID
	return nil
}

func (r *IntentRepo) UpdateIntent(_ context.Context, i domain.PaymentIntent) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	if _, ok := r.S.intents[i.ID]; !ok {
		return domain.ErrNotFound
	}
	r.S.intents[i.ID] = i
	return nil
}

func (r *IntentRepo) GetIntent(_ context.Context, tenantID, id uuid.UUID) (domain.PaymentIntent, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	i, ok := r.S.intents[id]
	if !ok || i.TenantID != tenantID {
		return domain.PaymentIntent{}, domain.ErrNotFound
	}
	return i, nil
}

func (r *IntentRepo) GetByIdempotencyKey(_ context.Context, tenantID uuid.UUID, key string) (domain.PaymentIntent, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	id, ok := r.S.byIdem[idemKey(tenantID, key)]
	if !ok {
		return domain.PaymentIntent{}, domain.ErrNotFound
	}
	return r.S.intents[id], nil
}

func (r *IntentRepo) ListIntents(_ context.Context, f ports.IntentFilter) ([]domain.PaymentIntent, int, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	var all []domain.PaymentIntent
	for _, i := range r.S.intents {
		if i.TenantID != f.TenantID {
			continue
		}
		if f.PrincipalID != nil && i.PrincipalID != *f.PrincipalID {
			continue
		}
		if f.Status != nil && i.Status != *f.Status {
			continue
		}
		if f.OrderID != "" && i.OrderID != f.OrderID {
			continue
		}
		all = append(all, i)
	}
	sort.Slice(all, func(a, b int) bool { return all[a].CreatedAt.After(all[b].CreatedAt) })
	total := len(all)
	limit := f.Limit
	if limit <= 0 {
		limit = 50
	}
	off := f.Offset
	if off > total {
		off = total
	}
	end := off + limit
	if end > total {
		end = total
	}
	return all[off:end], total, nil
}

func (r *IntentRepo) CreateAttempt(_ context.Context, a domain.PaymentAttempt) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.attempts[a.IntentID] = append(r.S.attempts[a.IntentID], a)
	return nil
}

func (r *IntentRepo) ListAttempts(_ context.Context, _, intentID uuid.UUID) ([]domain.PaymentAttempt, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	out := append([]domain.PaymentAttempt(nil), r.S.attempts[intentID]...)
	return out, nil
}

func (r *IntentRepo) CreateMethod(_ context.Context, m domain.PaymentMethod) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.methods[m.ID] = m
	return nil
}

func (r *IntentRepo) GetMethod(_ context.Context, tenantID, id uuid.UUID) (domain.PaymentMethod, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	m, ok := r.S.methods[id]
	if !ok || m.TenantID != tenantID {
		return domain.PaymentMethod{}, domain.ErrNotFound
	}
	return m, nil
}

func (r *IntentRepo) ListMethods(_ context.Context, tenantID, principalID uuid.UUID) ([]domain.PaymentMethod, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	var out []domain.PaymentMethod
	for _, m := range r.S.methods {
		if m.TenantID == tenantID && m.PrincipalID == principalID {
			out = append(out, m)
		}
	}
	return out, nil
}

func (r *IntentRepo) CreateRefund(_ context.Context, rf domain.Refund) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	k := idemKey(rf.TenantID, rf.IdempotencyKey)
	if _, ok := r.S.refundIdem[k]; ok {
		return domain.ErrAlreadyExists
	}
	r.S.refunds[rf.ID] = rf
	r.S.refundIdem[k] = rf.ID
	return nil
}

func (r *IntentRepo) UpdateRefund(_ context.Context, rf domain.Refund) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	if _, ok := r.S.refunds[rf.ID]; !ok {
		return domain.ErrNotFound
	}
	r.S.refunds[rf.ID] = rf
	return nil
}

func (r *IntentRepo) GetRefundByIdempotency(_ context.Context, tenantID uuid.UUID, key string) (domain.Refund, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	id, ok := r.S.refundIdem[idemKey(tenantID, key)]
	if !ok {
		return domain.Refund{}, domain.ErrNotFound
	}
	return r.S.refunds[id], nil
}

func (r *IntentRepo) ListRefunds(_ context.Context, _, intentID uuid.UUID) ([]domain.Refund, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	var out []domain.Refund
	for _, rf := range r.S.refunds {
		if rf.IntentID == intentID {
			out = append(out, rf)
		}
	}
	return out, nil
}

func (r *IntentRepo) CreateChargeback(_ context.Context, c domain.Chargeback) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.chargebacks = append(r.S.chargebacks, c)
	return nil
}

func (r *IntentRepo) ListChargebacks(_ context.Context, tenantID uuid.UUID, intentID *uuid.UUID) ([]domain.Chargeback, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	var out []domain.Chargeback
	for _, c := range r.S.chargebacks {
		if c.TenantID != tenantID {
			continue
		}
		if intentID != nil && c.IntentID != *intentID {
			continue
		}
		out = append(out, c)
	}
	return out, nil
}

func (r *IntentRepo) UpsertRoute(_ context.Context, route domain.ProviderRoute) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	for i, existing := range r.S.routes {
		if existing.TenantID == route.TenantID && existing.MethodType == route.MethodType && existing.Currency == route.Currency {
			route.ID = existing.ID
			route.CreatedAt = existing.CreatedAt
			r.S.routes[i] = route
			return nil
		}
	}
	r.S.routes = append(r.S.routes, route)
	return nil
}

func (r *IntentRepo) ListRoutes(_ context.Context, tenantID uuid.UUID, method domain.PaymentMethodType, currency string) ([]domain.ProviderRoute, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	var out []domain.ProviderRoute
	for _, route := range r.S.routes {
		if route.TenantID != tenantID || !route.Active {
			continue
		}
		if method != "" && route.MethodType != method {
			continue
		}
		if currency != "" && route.Currency != "" && route.Currency != currency {
			continue
		}
		out = append(out, route)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Priority < out[j].Priority })
	return out, nil
}

func (r *IntentRepo) CreateFraudScore(_ context.Context, f domain.FraudScore) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.fraud = append(r.S.fraud, f)
	return nil
}

func (r *IntentRepo) CreateAudit(_ context.Context, a domain.AuditEntry) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.audits = append(r.S.audits, a)
	return nil
}

func (o *OutboxRepo) Enqueue(_ context.Context, m domain.OutboxMessage) error {
	o.S.mu.Lock()
	defer o.S.mu.Unlock()
	o.S.outbox = append(o.S.outbox, m)
	return nil
}

func (o *OutboxRepo) Update(_ context.Context, m domain.OutboxMessage) error {
	o.S.mu.Lock()
	defer o.S.mu.Unlock()
	for i, existing := range o.S.outbox {
		if existing.ID == m.ID {
			o.S.outbox[i] = m
			return nil
		}
	}
	return domain.ErrNotFound
}

func (o *OutboxRepo) ListPending(_ context.Context, limit int) ([]domain.OutboxMessage, error) {
	o.S.mu.RLock()
	defer o.S.mu.RUnlock()
	var out []domain.OutboxMessage
	for _, m := range o.S.outbox {
		if m.Status == domain.OutboxStatusPending {
			out = append(out, m)
			if limit > 0 && len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}

var _ ports.IntentRepo = (*IntentRepo)(nil)
var _ ports.OutboxRepository = (*OutboxRepo)(nil)