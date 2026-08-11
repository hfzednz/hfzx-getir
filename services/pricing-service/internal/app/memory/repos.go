package memory

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/pricing-service/internal/app/ports"
	"github.com/nexora/pricing-service/internal/domain"
)

// Repos bundles in-memory port implementations.
type Repos struct {
	Prices   *PriceRepo
	Taxes    *TaxRepo
	Dynamics *DynamicRepo
	Audits   *QuoteAuditRepo
	Outbox   *OutboxRepo
}

// NewRepos wires all memory repositories to a shared store.
func NewRepos(s *Store) *Repos {
	return &Repos{
		Prices:   &PriceRepo{S: s},
		Taxes:    &TaxRepo{S: s},
		Dynamics: &DynamicRepo{S: s},
		Audits:   &QuoteAuditRepo{S: s},
		Outbox:   &OutboxRepo{S: s},
	}
}

// PriceRepo is an in-memory PriceRepo.
type PriceRepo struct{ S *Store }

func (r *PriceRepo) UpsertBook(_ context.Context, b domain.PriceBook) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.books[b.ID] = b
	return nil
}

func (r *PriceRepo) GetBook(_ context.Context, tenantID, bookID uuid.UUID) (domain.PriceBook, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	b, ok := r.S.books[bookID]
	if !ok || b.TenantID != tenantID {
		return domain.PriceBook{}, fmt.Errorf("%w: price book", domain.ErrNotFound)
	}
	return b, nil
}

func (r *PriceRepo) ListBooks(_ context.Context, tenantID uuid.UUID) ([]domain.PriceBook, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	out := make([]domain.PriceBook, 0)
	for _, b := range r.S.books {
		if b.TenantID == tenantID {
			out = append(out, b)
		}
	}
	return out, nil
}

func (r *PriceRepo) UpsertEntry(_ context.Context, e domain.PriceEntry) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.entries[e.ID] = e
	return nil
}

func (r *PriceRepo) GetEntry(_ context.Context, tenantID, entryID uuid.UUID) (domain.PriceEntry, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	e, ok := r.S.entries[entryID]
	if !ok || e.TenantID != tenantID {
		return domain.PriceEntry{}, fmt.Errorf("%w: price entry", domain.ErrNotFound)
	}
	return e, nil
}

func (r *PriceRepo) ListEntries(_ context.Context, tenantID uuid.UUID, bookID, variantID *uuid.UUID) ([]domain.PriceEntry, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	out := make([]domain.PriceEntry, 0)
	for _, e := range r.S.entries {
		if e.TenantID != tenantID {
			continue
		}
		if bookID != nil && e.PriceBookID != *bookID {
			continue
		}
		if variantID != nil && e.VariantID != *variantID {
			continue
		}
		out = append(out, e)
	}
	return out, nil
}

func (r *PriceRepo) ListEntriesForVariant(_ context.Context, tenantID, variantID uuid.UUID) ([]domain.PriceEntry, error) {
	return r.ListEntries(context.Background(), tenantID, nil, &variantID)
}

var _ ports.PriceRepo = (*PriceRepo)(nil)

// TaxRepo is an in-memory TaxRepo.
type TaxRepo struct{ S *Store }

func (r *TaxRepo) Upsert(_ context.Context, t domain.TaxRule) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.taxes[t.ID] = t
	return nil
}

func (r *TaxRepo) Get(_ context.Context, tenantID, id uuid.UUID) (domain.TaxRule, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	t, ok := r.S.taxes[id]
	if !ok || t.TenantID != tenantID {
		return domain.TaxRule{}, fmt.Errorf("%w: tax rule", domain.ErrNotFound)
	}
	return t, nil
}

func (r *TaxRepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.TaxRule, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	out := make([]domain.TaxRule, 0)
	for _, t := range r.S.taxes {
		if t.TenantID == tenantID {
			out = append(out, t)
		}
	}
	return out, nil
}

var _ ports.TaxRepo = (*TaxRepo)(nil)

// DynamicRepo is an in-memory DynamicRepo.
type DynamicRepo struct{ S *Store }

func (r *DynamicRepo) Upsert(_ context.Context, d domain.DynamicRule) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.dynamics[d.ID] = d
	return nil
}

func (r *DynamicRepo) Get(_ context.Context, tenantID, id uuid.UUID) (domain.DynamicRule, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	d, ok := r.S.dynamics[id]
	if !ok || d.TenantID != tenantID {
		return domain.DynamicRule{}, fmt.Errorf("%w: dynamic rule", domain.ErrNotFound)
	}
	return d, nil
}

func (r *DynamicRepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.DynamicRule, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	out := make([]domain.DynamicRule, 0)
	for _, d := range r.S.dynamics {
		if d.TenantID == tenantID {
			out = append(out, d)
		}
	}
	return out, nil
}

func (r *DynamicRepo) ListActive(ctx context.Context, tenantID uuid.UUID) ([]domain.DynamicRule, error) {
	all, err := r.List(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	out := make([]domain.DynamicRule, 0)
	for _, d := range all {
		if d.Active {
			out = append(out, d)
		}
	}
	return out, nil
}

var _ ports.DynamicRepo = (*DynamicRepo)(nil)

// QuoteAuditRepo is an in-memory QuoteAuditRepo.
type QuoteAuditRepo struct{ S *Store }

func (r *QuoteAuditRepo) Create(_ context.Context, a domain.QuoteAudit) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.audits = append(r.S.audits, a)
	return nil
}

func (r *QuoteAuditRepo) Get(_ context.Context, tenantID, id uuid.UUID) (domain.QuoteAudit, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	for _, a := range r.S.audits {
		if a.ID == id && a.TenantID == tenantID {
			return a, nil
		}
	}
	return domain.QuoteAudit{}, fmt.Errorf("%w: quote audit", domain.ErrNotFound)
}

func (r *QuoteAuditRepo) List(_ context.Context, tenantID uuid.UUID, limit int) ([]domain.QuoteAudit, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	out := make([]domain.QuoteAudit, 0)
	for i := len(r.S.audits) - 1; i >= 0 && len(out) < limit; i-- {
		if r.S.audits[i].TenantID == tenantID {
			out = append(out, r.S.audits[i])
		}
	}
	return out, nil
}

var _ ports.QuoteAuditRepo = (*QuoteAuditRepo)(nil)

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
