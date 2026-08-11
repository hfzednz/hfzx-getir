package memory

import (
	"context"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/nexora/finance-ledger-service/internal/app/ports"
	"github.com/nexora/finance-ledger-service/internal/domain"
)

// NewRepos returns in-memory ledger repositories.
func NewRepos(s *Store) (
	ports.AccountRepository,
	ports.JournalRepository,
	ports.InvoiceRepository,
	ports.TaxRuleRepository,
	ports.EventStore,
	ports.OutboxRepository,
) {
	return &AccountRepo{S: s}, &JournalRepo{S: s}, &InvoiceRepo{S: s}, &TaxRuleRepo{S: s}, &EventRepo{S: s}, &OutboxRepo{S: s}
}

// AccountRepo is an in-memory AccountRepository.
type AccountRepo struct{ S *Store }

func (r *AccountRepo) Create(_ context.Context, a domain.Account) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	if _, ok := r.S.Accounts[a.ID]; ok {
		return domain.ErrAlreadyExists
	}
	ck := tenantKey(a.TenantID, strings.ToUpper(a.Code))
	if _, ok := r.S.AccountsByCode[ck]; ok {
		return domain.ErrAlreadyExists
	}
	cp := a
	r.S.Accounts[a.ID] = cp
	r.S.AccountsByCode[ck] = a.ID
	return nil
}

func (r *AccountRepo) Update(_ context.Context, a domain.Account) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	cur, ok := r.S.Accounts[a.ID]
	if !ok || cur.TenantID != a.TenantID {
		return domain.ErrNotFound
	}
	if a.Version < cur.Version {
		return domain.ErrVersionConflict
	}
	r.S.Accounts[a.ID] = a
	return nil
}

func (r *AccountRepo) GetByID(_ context.Context, tenantID, id uuid.UUID) (domain.Account, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	a, ok := r.S.Accounts[id]
	if !ok || a.TenantID != tenantID {
		return domain.Account{}, domain.ErrNotFound
	}
	return a, nil
}

func (r *AccountRepo) GetByCode(_ context.Context, tenantID uuid.UUID, code string) (domain.Account, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	id, ok := r.S.AccountsByCode[tenantKey(tenantID, strings.ToUpper(strings.TrimSpace(code)))]
	if !ok {
		return domain.Account{}, domain.ErrNotFound
	}
	a, ok := r.S.Accounts[id]
	if !ok || a.TenantID != tenantID {
		return domain.Account{}, domain.ErrNotFound
	}
	return a, nil
}

// JournalRepo is an in-memory JournalRepository.
type JournalRepo struct{ S *Store }

func cloneJournal(j domain.Journal) domain.Journal {
	cp := j
	if j.Lines != nil {
		cp.Lines = append([]domain.JournalLine(nil), j.Lines...)
	}
	if j.PostedAt != nil {
		t := *j.PostedAt
		cp.PostedAt = &t
	}
	return cp
}

func (r *JournalRepo) Create(_ context.Context, j domain.Journal) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	if _, ok := r.S.Journals[j.ID]; ok {
		return domain.ErrAlreadyExists
	}
	if j.IdempotencyKey != "" {
		ik := tenantKey(j.TenantID, j.IdempotencyKey)
		if id, ok := r.S.JournalsByIdem[ik]; ok && id != j.ID {
			return domain.ErrAlreadyExists
		}
		r.S.JournalsByIdem[ik] = j.ID
	}
	r.S.Journals[j.ID] = cloneJournal(j)
	return nil
}

func (r *JournalRepo) Update(_ context.Context, j domain.Journal) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	cur, ok := r.S.Journals[j.ID]
	if !ok || cur.TenantID != j.TenantID {
		return domain.ErrNotFound
	}
	if cur.Status == domain.JournalStatusPosted {
		return domain.ErrJournalImmutable
	}
	if j.Version < cur.Version {
		return domain.ErrVersionConflict
	}
	r.S.Journals[j.ID] = cloneJournal(j)
	return nil
}

func (r *JournalRepo) GetByID(_ context.Context, tenantID, id uuid.UUID) (domain.Journal, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	j, ok := r.S.Journals[id]
	if !ok || j.TenantID != tenantID {
		return domain.Journal{}, domain.ErrNotFound
	}
	return cloneJournal(j), nil
}

func (r *JournalRepo) GetByIdempotencyKey(_ context.Context, tenantID uuid.UUID, key string) (domain.Journal, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	id, ok := r.S.JournalsByIdem[tenantKey(tenantID, key)]
	if !ok {
		return domain.Journal{}, domain.ErrNotFound
	}
	j, ok := r.S.Journals[id]
	if !ok || j.TenantID != tenantID {
		return domain.Journal{}, domain.ErrNotFound
	}
	return cloneJournal(j), nil
}

func (r *JournalRepo) List(_ context.Context, f ports.JournalFilter) ([]domain.Journal, int, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	all := make([]domain.Journal, 0)
	for _, j := range r.S.Journals {
		if j.TenantID != f.TenantID {
			continue
		}
		if f.Status != nil && j.Status != *f.Status {
			continue
		}
		all = append(all, cloneJournal(j))
	}
	sort.Slice(all, func(i, k int) bool {
		return all[i].CreatedAt.After(all[k].CreatedAt)
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

func (r *JournalRepo) BalanceMinor(_ context.Context, tenantID, accountID uuid.UUID) (int64, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	var bal int64
	for _, j := range r.S.Journals {
		if j.TenantID != tenantID || j.Status != domain.JournalStatusPosted {
			continue
		}
		for _, line := range j.Lines {
			if line.AccountID != accountID {
				continue
			}
			bal += line.DebitMinor - line.CreditMinor
		}
	}
	return bal, nil
}

// InvoiceRepo is an in-memory InvoiceRepository.
type InvoiceRepo struct{ S *Store }

func cloneInvoice(inv domain.Invoice) domain.Invoice {
	cp := inv
	if inv.Lines != nil {
		cp.Lines = append([]domain.InvoiceLine(nil), inv.Lines...)
	}
	if inv.IssuedAt != nil {
		t := *inv.IssuedAt
		cp.IssuedAt = &t
	}
	return cp
}

func (r *InvoiceRepo) Create(_ context.Context, inv domain.Invoice) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	if _, ok := r.S.Invoices[inv.ID]; ok {
		return domain.ErrAlreadyExists
	}
	if inv.IdempotencyKey != "" {
		ik := tenantKey(inv.TenantID, inv.IdempotencyKey)
		if id, ok := r.S.InvoicesByIdem[ik]; ok && id != inv.ID {
			return domain.ErrAlreadyExists
		}
		r.S.InvoicesByIdem[ik] = inv.ID
	}
	r.S.Invoices[inv.ID] = cloneInvoice(inv)
	return nil
}

func (r *InvoiceRepo) Update(_ context.Context, inv domain.Invoice) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	cur, ok := r.S.Invoices[inv.ID]
	if !ok || cur.TenantID != inv.TenantID {
		return domain.ErrNotFound
	}
	if inv.Version < cur.Version {
		return domain.ErrVersionConflict
	}
	r.S.Invoices[inv.ID] = cloneInvoice(inv)
	return nil
}

func (r *InvoiceRepo) GetByID(_ context.Context, tenantID, id uuid.UUID) (domain.Invoice, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	inv, ok := r.S.Invoices[id]
	if !ok || inv.TenantID != tenantID {
		return domain.Invoice{}, domain.ErrNotFound
	}
	return cloneInvoice(inv), nil
}

func (r *InvoiceRepo) GetByIdempotencyKey(_ context.Context, tenantID uuid.UUID, key string) (domain.Invoice, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	id, ok := r.S.InvoicesByIdem[tenantKey(tenantID, key)]
	if !ok {
		return domain.Invoice{}, domain.ErrNotFound
	}
	inv, ok := r.S.Invoices[id]
	if !ok || inv.TenantID != tenantID {
		return domain.Invoice{}, domain.ErrNotFound
	}
	return cloneInvoice(inv), nil
}

func (r *InvoiceRepo) CreateCreditNote(_ context.Context, cn domain.CreditNote) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	if _, ok := r.S.CreditNotes[cn.ID]; ok {
		return domain.ErrAlreadyExists
	}
	if cn.IdempotencyKey != "" {
		ik := tenantKey(cn.TenantID, cn.IdempotencyKey)
		if id, ok := r.S.CreditByIdem[ik]; ok && id != cn.ID {
			return domain.ErrAlreadyExists
		}
		r.S.CreditByIdem[ik] = cn.ID
	}
	r.S.CreditNotes[cn.ID] = cn
	return nil
}

func (r *InvoiceRepo) GetCreditNoteByIdempotencyKey(_ context.Context, tenantID uuid.UUID, key string) (domain.CreditNote, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	id, ok := r.S.CreditByIdem[tenantKey(tenantID, key)]
	if !ok {
		return domain.CreditNote{}, domain.ErrNotFound
	}
	cn, ok := r.S.CreditNotes[id]
	if !ok || cn.TenantID != tenantID {
		return domain.CreditNote{}, domain.ErrNotFound
	}
	return cn, nil
}

// TaxRuleRepo is an in-memory TaxRuleRepository.
type TaxRuleRepo struct{ S *Store }

func (r *TaxRuleRepo) Upsert(_ context.Context, rule domain.TaxRule) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.TaxRules[tenantKey(rule.TenantID, strings.ToUpper(rule.Code))] = rule
	return nil
}

func (r *TaxRuleRepo) GetByCode(_ context.Context, tenantID uuid.UUID, code string) (domain.TaxRule, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	rule, ok := r.S.TaxRules[tenantKey(tenantID, strings.ToUpper(strings.TrimSpace(code)))]
	if !ok {
		return domain.TaxRule{}, domain.ErrNotFound
	}
	return rule, nil
}

// EventRepo is an in-memory EventStore.
type EventRepo struct{ S *Store }

func (r *EventRepo) Append(_ context.Context, e domain.LedgerEvent) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.Events[e.ID] = e
	return nil
}

func (r *EventRepo) ListByEntity(_ context.Context, tenantID, entityID uuid.UUID) ([]domain.LedgerEvent, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	out := make([]domain.LedgerEvent, 0)
	for _, e := range r.S.Events {
		if e.TenantID == tenantID && e.EntityID == entityID {
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
	sort.Slice(out, func(i, j int) bool {
		return out[i].CreatedAt.Before(out[j].CreatedAt)
	})
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

var (
	_ ports.AccountRepository  = (*AccountRepo)(nil)
	_ ports.JournalRepository  = (*JournalRepo)(nil)
	_ ports.InvoiceRepository  = (*InvoiceRepo)(nil)
	_ ports.TaxRuleRepository  = (*TaxRuleRepo)(nil)
	_ ports.EventStore         = (*EventRepo)(nil)
	_ ports.OutboxRepository   = (*OutboxRepo)(nil)
)
