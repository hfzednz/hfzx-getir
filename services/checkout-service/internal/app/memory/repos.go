package memory

import (
	"context"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/nexora/checkout-service/internal/app/ports"
	"github.com/nexora/checkout-service/internal/domain"
)

// NewRepos returns in-memory checkout repositories.
func NewRepos(s *Store) (ports.CheckoutRepo, ports.EventStore, ports.OutboxRepository) {
	return &SessionRepo{S: s}, &EventRepo{S: s}, &OutboxRepo{S: s}
}

// SessionRepo is an in-memory CheckoutRepo.
type SessionRepo struct{ S *Store }

func (r *SessionRepo) Create(_ context.Context, s domain.Session) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	if _, ok := r.S.Sessions[s.ID]; ok {
		return domain.ErrAlreadyExists
	}
	ik := tenantKey(s.TenantID, s.IdempotencyKey)
	if id, ok := r.S.SessionsByIdem[ik]; ok {
		if id != s.ID {
			return domain.ErrAlreadyExists
		}
	}
	cp := cloneSession(s)
	r.S.Sessions[s.ID] = cp
	r.S.SessionsByIdem[ik] = s.ID
	if s.RecoveryToken != "" {
		r.S.ByRecovery[s.RecoveryToken] = s.ID
	}
	return nil
}

func (r *SessionRepo) Update(_ context.Context, s domain.Session) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	cur, ok := r.S.Sessions[s.ID]
	if !ok || cur.TenantID != s.TenantID {
		return domain.ErrNotFound
	}
	if s.Version < cur.Version {
		return domain.ErrVersionConflict
	}
	// Refresh recovery index if rotated.
	if cur.RecoveryToken != "" && cur.RecoveryToken != s.RecoveryToken {
		delete(r.S.ByRecovery, cur.RecoveryToken)
	}
	if s.RecoveryToken != "" {
		r.S.ByRecovery[s.RecoveryToken] = s.ID
	}
	r.S.Sessions[s.ID] = cloneSession(s)
	return nil
}

func (r *SessionRepo) GetByID(_ context.Context, tenantID, id uuid.UUID) (domain.Session, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	s, ok := r.S.Sessions[id]
	if !ok || s.TenantID != tenantID {
		return domain.Session{}, domain.ErrNotFound
	}
	return cloneSession(s), nil
}

func (r *SessionRepo) GetByIdempotencyKey(_ context.Context, tenantID uuid.UUID, key string) (domain.Session, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	id, ok := r.S.SessionsByIdem[tenantKey(tenantID, key)]
	if !ok {
		return domain.Session{}, domain.ErrNotFound
	}
	s, ok := r.S.Sessions[id]
	if !ok || s.TenantID != tenantID {
		return domain.Session{}, domain.ErrNotFound
	}
	return cloneSession(s), nil
}

func (r *SessionRepo) GetByRecoveryToken(_ context.Context, token string) (domain.Session, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	id, ok := r.S.ByRecovery[token]
	if !ok {
		return domain.Session{}, domain.ErrNotFound
	}
	s, ok := r.S.Sessions[id]
	if !ok {
		return domain.Session{}, domain.ErrNotFound
	}
	return cloneSession(s), nil
}

func (r *SessionRepo) List(_ context.Context, f ports.SessionFilter) ([]domain.Session, int, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	query := strings.ToLower(strings.TrimSpace(f.Query))
	all := make([]domain.Session, 0)
	for _, s := range r.S.Sessions {
		if s.TenantID != f.TenantID {
			continue
		}
		if f.PrincipalID != nil && s.PrincipalID != *f.PrincipalID {
			continue
		}
		if f.Status != nil && s.Status != *f.Status {
			continue
		}
		if query != "" {
			match := strings.Contains(strings.ToLower(s.ID.String()), query) ||
				strings.Contains(strings.ToLower(s.CartID.String()), query) ||
				strings.Contains(strings.ToLower(s.OrderID), query) ||
				strings.Contains(strings.ToLower(string(s.Status)), query)
			if !match {
				continue
			}
		}
		all = append(all, cloneSession(s))
	}
	sort.Slice(all, func(i, j int) bool {
		return all[i].CreatedAt.After(all[j].CreatedAt)
	})
	total := len(all)
	if f.Offset >= len(all) {
		return nil, total, nil
	}
	end := len(all)
	if f.Limit > 0 && f.Offset+f.Limit < end {
		end = f.Offset + f.Limit
	}
	return all[f.Offset:end], total, nil
}

func (r *SessionRepo) CountByStatus(_ context.Context, tenantID uuid.UUID) (map[domain.SessionStatus]int, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	out := map[domain.SessionStatus]int{}
	for _, s := range r.S.Sessions {
		if s.TenantID != tenantID {
			continue
		}
		out[s.Status]++
	}
	return out, nil
}

func cloneSession(s domain.Session) domain.Session {
	cp := s
	if s.CouponCodes != nil {
		cp.CouponCodes = append([]string(nil), s.CouponCodes...)
	}
	if s.Validation.Issues != nil {
		cp.Validation.Issues = append([]domain.ValidationIssue(nil), s.Validation.Issues...)
	}
	if s.Metadata != nil {
		cp.Metadata = map[string]any{}
		for k, v := range s.Metadata {
			cp.Metadata[k] = v
		}
	}
	return cp
}

// EventRepo is an in-memory EventStore.
type EventRepo struct{ S *Store }

func (r *EventRepo) Append(_ context.Context, e domain.SessionEvent) error {
	if err := e.Validate(); err != nil {
		return err
	}
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	if _, ok := r.S.Events[e.ID]; ok {
		return domain.ErrAlreadyExists
	}
	cp := e
	if e.Payload != nil {
		cp.Payload = map[string]any{}
		for k, v := range e.Payload {
			cp.Payload[k] = v
		}
	}
	r.S.Events[e.ID] = cp
	return nil
}

func (r *EventRepo) ListBySession(_ context.Context, tenantID, sessionID uuid.UUID) ([]domain.SessionEvent, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	out := make([]domain.SessionEvent, 0)
	for _, e := range r.S.Events {
		if e.TenantID == tenantID && e.SessionID == sessionID {
			out = append(out, e)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].OccurredAt.Before(out[j].OccurredAt)
	})
	return out, nil
}

// OutboxRepo is an in-memory OutboxRepository.
type OutboxRepo struct{ S *Store }

func (r *OutboxRepo) Enqueue(_ context.Context, m domain.OutboxMessage) error {
	if err := m.Validate(); err != nil {
		return err
	}
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	if _, ok := r.S.Outbox[m.ID]; ok {
		return domain.ErrAlreadyExists
	}
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
	sort.Slice(out, func(i, j int) bool {
		return out[i].CreatedAt.Before(out[j].CreatedAt)
	})
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

// SeedCart inserts a cart for CartClient lookups.
func (s *Store) SeedCart(c ports.CartView) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Carts[c.ID] = c
}

var _ ports.CheckoutRepo = (*SessionRepo)(nil)
var _ ports.EventStore = (*EventRepo)(nil)
var _ ports.OutboxRepository = (*OutboxRepo)(nil)
