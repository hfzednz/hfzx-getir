package memory

import (
	"context"

	"github.com/google/uuid"
	"github.com/nexora/wallet-service/internal/app/ports"
	"github.com/nexora/wallet-service/internal/domain"
)

// WalletRepo is an in-memory WalletRepo.
type WalletRepo struct{ S *Store }

// OutboxRepo is an in-memory OutboxRepository.
type OutboxRepo struct{ S *Store }

// NewRepos returns wallet + outbox repos sharing a store.
func NewRepos(s *Store) (*WalletRepo, *OutboxRepo) {
	return &WalletRepo{S: s}, &OutboxRepo{S: s}
}

func (r *WalletRepo) CreateWallet(_ context.Context, w domain.Wallet, accounts []domain.Account) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	k := prinKey(w.TenantID, w.PrincipalID)
	if _, ok := r.S.byPrin[k]; ok {
		return domain.ErrAlreadyExists
	}
	r.S.wallets[w.ID] = w
	r.S.byPrin[k] = w.ID
	for _, a := range accounts {
		r.S.accounts[a.ID] = a
		r.S.acctByType[acctKey(a.TenantID, a.WalletID, a.AccountType)] = a.ID
	}
	return nil
}

func (r *WalletRepo) GetWallet(_ context.Context, tenantID, walletID uuid.UUID) (domain.Wallet, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	w, ok := r.S.wallets[walletID]
	if !ok || w.TenantID != tenantID {
		return domain.Wallet{}, domain.ErrNotFound
	}
	return w, nil
}

func (r *WalletRepo) GetWalletByPrincipal(_ context.Context, tenantID, principalID uuid.UUID) (domain.Wallet, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	id, ok := r.S.byPrin[prinKey(tenantID, principalID)]
	if !ok {
		return domain.Wallet{}, domain.ErrNotFound
	}
	return r.S.wallets[id], nil
}

func (r *WalletRepo) UpdateWallet(_ context.Context, w domain.Wallet) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	if _, ok := r.S.wallets[w.ID]; !ok {
		return domain.ErrNotFound
	}
	r.S.wallets[w.ID] = w
	return nil
}

func (r *WalletRepo) GetAccount(_ context.Context, tenantID, accountID uuid.UUID) (domain.Account, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	a, ok := r.S.accounts[accountID]
	if !ok || a.TenantID != tenantID {
		return domain.Account{}, domain.ErrNotFound
	}
	return a, nil
}

func (r *WalletRepo) GetAccountByType(_ context.Context, tenantID, walletID uuid.UUID, t domain.AccountType) (domain.Account, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	id, ok := r.S.acctByType[acctKey(tenantID, walletID, t)]
	if !ok {
		return domain.Account{}, domain.ErrNotFound
	}
	return r.S.accounts[id], nil
}

func (r *WalletRepo) ListAccounts(_ context.Context, tenantID, walletID uuid.UUID) ([]domain.Account, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	var out []domain.Account
	for _, a := range r.S.accounts {
		if a.TenantID == tenantID && a.WalletID == walletID {
			out = append(out, a)
		}
	}
	return out, nil
}

func (r *WalletRepo) UpdateAccount(_ context.Context, a domain.Account) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	if _, ok := r.S.accounts[a.ID]; !ok {
		return domain.ErrNotFound
	}
	r.S.accounts[a.ID] = a
	return nil
}

func (r *WalletRepo) CreateEntry(_ context.Context, e domain.Entry) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	if e.IdempotencyKey != "" {
		k := idemKey(e.TenantID, e.IdempotencyKey)
		if _, ok := r.S.entryIdem[k]; ok {
			return domain.ErrAlreadyExists
		}
		r.S.entryIdem[k] = e.ID
	}
	r.S.entries = append(r.S.entries, e)
	return nil
}

func (r *WalletRepo) GetEntryByIdempotency(_ context.Context, tenantID uuid.UUID, key string) (domain.Entry, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	id, ok := r.S.entryIdem[idemKey(tenantID, key)]
	if !ok {
		return domain.Entry{}, domain.ErrNotFound
	}
	for _, e := range r.S.entries {
		if e.ID == id {
			return e, nil
		}
	}
	return domain.Entry{}, domain.ErrNotFound
}

func (r *WalletRepo) ListEntries(_ context.Context, tenantID, walletID uuid.UUID, limit, offset int) ([]domain.Entry, int, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	var all []domain.Entry
	for i := len(r.S.entries) - 1; i >= 0; i-- {
		e := r.S.entries[i]
		if e.TenantID == tenantID && e.WalletID == walletID {
			all = append(all, e)
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

func (r *WalletRepo) CreateHold(_ context.Context, h domain.Hold) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	k := idemKey(h.TenantID, h.IdempotencyKey)
	if _, ok := r.S.holdIdem[k]; ok {
		return domain.ErrAlreadyExists
	}
	r.S.holds[h.ID] = h
	r.S.holdIdem[k] = h.ID
	return nil
}

func (r *WalletRepo) GetHold(_ context.Context, tenantID, holdID uuid.UUID) (domain.Hold, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	h, ok := r.S.holds[holdID]
	if !ok || h.TenantID != tenantID {
		return domain.Hold{}, domain.ErrHoldNotFound
	}
	return h, nil
}

func (r *WalletRepo) GetHoldByIdempotency(_ context.Context, tenantID uuid.UUID, key string) (domain.Hold, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	id, ok := r.S.holdIdem[idemKey(tenantID, key)]
	if !ok {
		return domain.Hold{}, domain.ErrNotFound
	}
	return r.S.holds[id], nil
}

func (r *WalletRepo) UpdateHold(_ context.Context, h domain.Hold) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	if _, ok := r.S.holds[h.ID]; !ok {
		return domain.ErrHoldNotFound
	}
	r.S.holds[h.ID] = h
	return nil
}

func (r *WalletRepo) CreateTransfer(_ context.Context, t domain.Transfer) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	k := idemKey(t.TenantID, t.IdempotencyKey)
	if _, ok := r.S.xferIdem[k]; ok {
		return domain.ErrAlreadyExists
	}
	r.S.transfers[t.ID] = t
	r.S.xferIdem[k] = t.ID
	return nil
}

func (r *WalletRepo) GetTransferByIdempotency(_ context.Context, tenantID uuid.UUID, key string) (domain.Transfer, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	id, ok := r.S.xferIdem[idemKey(tenantID, key)]
	if !ok {
		return domain.Transfer{}, domain.ErrNotFound
	}
	return r.S.transfers[id], nil
}

func (r *WalletRepo) UpsertLimit(_ context.Context, l domain.Limit) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	k := l.WalletID.String() + "|" + l.LimitType + "|" + l.WindowKey
	r.S.limits[k] = l
	return nil
}

func (r *WalletRepo) GetLimit(_ context.Context, tenantID, walletID uuid.UUID, limitType, windowKey string) (domain.Limit, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	k := walletID.String() + "|" + limitType + "|" + windowKey
	l, ok := r.S.limits[k]
	if !ok || l.TenantID != tenantID {
		return domain.Limit{}, domain.ErrNotFound
	}
	return l, nil
}

func (r *WalletRepo) CreateAudit(_ context.Context, a domain.AuditEntry) error {
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

var _ ports.WalletRepo = (*WalletRepo)(nil)
var _ ports.OutboxRepository = (*OutboxRepo)(nil)
