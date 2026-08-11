package memory

import (
	"context"
	"sort"

	"github.com/google/uuid"
	"github.com/nexora/settlement-service/internal/app/ports"
	"github.com/nexora/settlement-service/internal/domain"
)

// NewRepos returns in-memory settlement repositories.
func NewRepos(s *Store) (ports.BatchRepository, ports.EventStore, ports.OutboxRepository) {
	return &BatchRepo{S: s}, &EventRepo{S: s}, &OutboxRepo{S: s}
}

func cloneBatch(b domain.SettlementBatch) domain.SettlementBatch {
	cp := b
	if b.Lines != nil {
		cp.Lines = append([]domain.SettlementLine(nil), b.Lines...)
	}
	copyUUID := func(p *uuid.UUID) *uuid.UUID {
		if p == nil {
			return nil
		}
		v := *p
		return &v
	}
	cp.SubmittedBy = copyUUID(b.SubmittedBy)
	cp.ApprovedBy = copyUUID(b.ApprovedBy)
	if b.SubmittedAt != nil {
		t := *b.SubmittedAt
		cp.SubmittedAt = &t
	}
	if b.ApprovedAt != nil {
		t := *b.ApprovedAt
		cp.ApprovedAt = &t
	}
	if b.CompletedAt != nil {
		t := *b.CompletedAt
		cp.CompletedAt = &t
	}
	if b.FailedAt != nil {
		t := *b.FailedAt
		cp.FailedAt = &t
	}
	return cp
}

// BatchRepo is an in-memory BatchRepository.
type BatchRepo struct{ S *Store }

func (r *BatchRepo) Create(_ context.Context, b domain.SettlementBatch) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	if _, ok := r.S.Batches[b.ID]; ok {
		return domain.ErrAlreadyExists
	}
	if b.IdempotencyKey != "" {
		ik := tenantKey(b.TenantID, b.IdempotencyKey)
		if id, ok := r.S.BatchesByIdem[ik]; ok && id != b.ID {
			return domain.ErrAlreadyExists
		}
		r.S.BatchesByIdem[ik] = b.ID
	}
	r.S.Batches[b.ID] = cloneBatch(b)
	return nil
}

func (r *BatchRepo) Update(_ context.Context, b domain.SettlementBatch) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	cur, ok := r.S.Batches[b.ID]
	if !ok || cur.TenantID != b.TenantID {
		return domain.ErrNotFound
	}
	if b.Version < cur.Version {
		return domain.ErrVersionConflict
	}
	r.S.Batches[b.ID] = cloneBatch(b)
	return nil
}

func (r *BatchRepo) GetByID(_ context.Context, tenantID, id uuid.UUID) (domain.SettlementBatch, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	b, ok := r.S.Batches[id]
	if !ok || b.TenantID != tenantID {
		return domain.SettlementBatch{}, domain.ErrNotFound
	}
	return cloneBatch(b), nil
}

func (r *BatchRepo) GetByIdempotencyKey(_ context.Context, tenantID uuid.UUID, key string) (domain.SettlementBatch, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	id, ok := r.S.BatchesByIdem[tenantKey(tenantID, key)]
	if !ok {
		return domain.SettlementBatch{}, domain.ErrNotFound
	}
	b, ok := r.S.Batches[id]
	if !ok || b.TenantID != tenantID {
		return domain.SettlementBatch{}, domain.ErrNotFound
	}
	return cloneBatch(b), nil
}

func (r *BatchRepo) List(_ context.Context, f ports.BatchFilter) ([]domain.SettlementBatch, int, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	all := make([]domain.SettlementBatch, 0)
	for _, b := range r.S.Batches {
		if b.TenantID != f.TenantID {
			continue
		}
		if f.Status != nil && b.Status != *f.Status {
			continue
		}
		all = append(all, cloneBatch(b))
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

func (r *BatchRepo) SavePayout(_ context.Context, p domain.PayoutInstruction) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.Payouts[p.ID] = p
	return nil
}

func (r *BatchRepo) ListPayouts(_ context.Context, tenantID, batchID uuid.UUID) ([]domain.PayoutInstruction, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	out := make([]domain.PayoutInstruction, 0)
	for _, p := range r.S.Payouts {
		if p.TenantID == tenantID && p.BatchID == batchID {
			out = append(out, p)
		}
	}
	return out, nil
}

func (r *BatchRepo) SaveReconciliation(_ context.Context, rec domain.Reconciliation) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.Reconciles[rec.ID] = rec
	return nil
}

func (r *BatchRepo) SaveMismatch(_ context.Context, m domain.Mismatch) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.Mismatches[m.ID] = m
	return nil
}

func (r *BatchRepo) ListMismatches(_ context.Context, tenantID, batchID uuid.UUID) ([]domain.Mismatch, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	out := make([]domain.Mismatch, 0)
	for _, m := range r.S.Mismatches {
		if m.TenantID == tenantID && m.BatchID == batchID {
			out = append(out, m)
		}
	}
	return out, nil
}

// EventRepo is an in-memory EventStore.
type EventRepo struct{ S *Store }

func (r *EventRepo) Append(_ context.Context, e domain.SettlementEvent) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.Events[e.ID] = e
	return nil
}

func (r *EventRepo) ListByBatch(_ context.Context, tenantID, batchID uuid.UUID) ([]domain.SettlementEvent, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	out := make([]domain.SettlementEvent, 0)
	for _, e := range r.S.Events {
		if e.TenantID == tenantID && e.BatchID == batchID {
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
	_ ports.BatchRepository  = (*BatchRepo)(nil)
	_ ports.EventStore       = (*EventRepo)(nil)
	_ ports.OutboxRepository = (*OutboxRepo)(nil)
)
