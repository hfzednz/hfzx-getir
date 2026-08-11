package memory

import (
	"context"

	"github.com/google/uuid"
	"github.com/nexora/cart-service/internal/app/ports"
	"github.com/nexora/cart-service/internal/domain"
)

// NewRepos returns in-memory repositories sharing a store.
func NewRepos(s *Store) (*CartRepo, *EventRepo, *OutboxRepo, *SavedRepo) {
	return &CartRepo{S: s}, &EventRepo{S: s}, &OutboxRepo{S: s}, &SavedRepo{S: s}
}

// CartRepo is an in-memory cart repository.
type CartRepo struct{ S *Store }

func (r *CartRepo) Create(_ context.Context, c domain.Cart) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	if _, ok := r.S.Carts[c.ID]; ok {
		return domain.ErrAlreadyExists
	}
	cp := cloneCart(c)
	r.S.Carts[c.ID] = cp
	r.indexCart(cp)
	return nil
}

func (r *CartRepo) Update(_ context.Context, c domain.Cart) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	old, ok := r.S.Carts[c.ID]
	if !ok || old.TenantID != c.TenantID {
		return domain.ErrNotFound
	}
	r.unindexCart(old)
	cp := cloneCart(c)
	r.S.Carts[c.ID] = cp
	r.indexCart(cp)
	return nil
}

func (r *CartRepo) GetByID(_ context.Context, tenantID, id uuid.UUID) (domain.Cart, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	c, ok := r.S.Carts[id]
	if !ok || c.TenantID != tenantID {
		return domain.Cart{}, domain.ErrNotFound
	}
	return cloneCart(c), nil
}

func (r *CartRepo) GetActiveByGuest(_ context.Context, tenantID uuid.UUID, guestToken string) (domain.Cart, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	id, ok := r.S.CartByGuest[tenantKey(tenantID, guestToken)]
	if !ok {
		return domain.Cart{}, domain.ErrNotFound
	}
	c, ok := r.S.Carts[id]
	if !ok || c.TenantID != tenantID || c.Status != domain.CartStatusActive {
		return domain.Cart{}, domain.ErrNotFound
	}
	return cloneCart(c), nil
}

func (r *CartRepo) GetActiveByPrincipal(_ context.Context, tenantID, principalID uuid.UUID) (domain.Cart, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	id, ok := r.S.CartByPrincipal[tenantKey(tenantID, principalID.String())]
	if !ok {
		return domain.Cart{}, domain.ErrNotFound
	}
	c, ok := r.S.Carts[id]
	if !ok || c.TenantID != tenantID || c.Status != domain.CartStatusActive {
		return domain.Cart{}, domain.ErrNotFound
	}
	return cloneCart(c), nil
}

func (r *CartRepo) indexCart(c domain.Cart) {
	if c.Status != domain.CartStatusActive {
		return
	}
	if c.GuestToken != "" {
		r.S.CartByGuest[tenantKey(c.TenantID, c.GuestToken)] = c.ID
	}
	if c.PrincipalID != nil {
		r.S.CartByPrincipal[tenantKey(c.TenantID, c.PrincipalID.String())] = c.ID
	}
}

func (r *CartRepo) unindexCart(c domain.Cart) {
	if c.GuestToken != "" {
		key := tenantKey(c.TenantID, c.GuestToken)
		if id, ok := r.S.CartByGuest[key]; ok && id == c.ID {
			delete(r.S.CartByGuest, key)
		}
	}
	if c.PrincipalID != nil {
		key := tenantKey(c.TenantID, c.PrincipalID.String())
		if id, ok := r.S.CartByPrincipal[key]; ok && id == c.ID {
			delete(r.S.CartByPrincipal, key)
		}
	}
}

func cloneCart(c domain.Cart) domain.Cart {
	out := c
	if c.Lines != nil {
		out.Lines = append([]domain.CartLine(nil), c.Lines...)
	}
	if c.Coupons != nil {
		out.Coupons = append([]domain.AppliedCoupon(nil), c.Coupons...)
	}
	if c.ReservationRefs != nil {
		out.ReservationRefs = append([]domain.ReservationRef(nil), c.ReservationRefs...)
	}
	if c.Quote != nil {
		q := *c.Quote
		if c.Quote.LineQuotes != nil {
			q.LineQuotes = append([]domain.LineQuote(nil), c.Quote.LineQuotes...)
		}
		out.Quote = &q
	}
	if c.Metadata != nil {
		out.Metadata = make(map[string]any, len(c.Metadata))
		for k, v := range c.Metadata {
			out.Metadata[k] = v
		}
	}
	return out
}

var _ ports.CartRepository = (*CartRepo)(nil)

// EventRepo is an in-memory event store.
type EventRepo struct{ S *Store }

func (r *EventRepo) Append(_ context.Context, e domain.CartEvent) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	if _, ok := r.S.Events[e.ID]; ok {
		return domain.ErrAlreadyExists
	}
	r.S.Events[e.ID] = e
	return nil
}

func (r *EventRepo) ListByCart(_ context.Context, tenantID, cartID uuid.UUID) ([]domain.CartEvent, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	out := make([]domain.CartEvent, 0)
	for _, e := range r.S.Events {
		if e.TenantID == tenantID && e.CartID == cartID {
			out = append(out, e)
		}
	}
	return out, nil
}

var _ ports.EventStore = (*EventRepo)(nil)

// OutboxRepo is an in-memory outbox repository.
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
			if limit > 0 && len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}

var _ ports.OutboxRepository = (*OutboxRepo)(nil)

// SavedRepo is an in-memory saved cart repository.
type SavedRepo struct{ S *Store }

func (r *SavedRepo) Create(_ context.Context, s domain.SavedCart) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	if _, ok := r.S.Saved[s.ID]; ok {
		return domain.ErrAlreadyExists
	}
	r.S.Saved[s.ID] = s
	return nil
}

func (r *SavedRepo) ListByPrincipal(_ context.Context, tenantID, principalID uuid.UUID) ([]domain.SavedCart, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	out := make([]domain.SavedCart, 0)
	for _, s := range r.S.Saved {
		if s.TenantID == tenantID && s.PrincipalID == principalID {
			out = append(out, s)
		}
	}
	return out, nil
}

var _ ports.SavedCartRepository = (*SavedRepo)(nil)
