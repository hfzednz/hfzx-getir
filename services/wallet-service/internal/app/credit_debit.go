package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/wallet-service/internal/domain"
)

// MoneyInput is shared for credit/debit/hold/adjust.
type MoneyInput struct {
	TenantID       uuid.UUID
	WalletID       uuid.UUID
	AccountType    domain.AccountType
	AmountMinor    int64
	IdempotencyKey string
	Reference      string
	Metadata       map[string]any
}

// Credit adds funds to an account.
func (d *Deps) Credit(ctx context.Context, in MoneyInput) (domain.Account, domain.Entry, error) {
	if in.IdempotencyKey == "" {
		return domain.Account{}, domain.Entry{}, fmt.Errorf("%w: idempotency_key required", domain.ErrInvalidArgument)
	}
	if in.AmountMinor <= 0 {
		return domain.Account{}, domain.Entry{}, fmt.Errorf("%w: amount must be > 0", domain.ErrInvalidArgument)
	}
	if existing, err := d.Wallets.GetEntryByIdempotency(ctx, in.TenantID, in.IdempotencyKey); err == nil {
		acct, _ := d.Wallets.GetAccount(ctx, in.TenantID, existing.AccountID)
		return acct, existing, nil
	}

	w, err := d.Wallets.GetWallet(ctx, in.TenantID, in.WalletID)
	if err != nil {
		return domain.Account{}, domain.Entry{}, err
	}
	acct, err := d.resolveAccount(ctx, in.TenantID, in.WalletID, in.AccountType)
	if err != nil {
		return domain.Account{}, domain.Entry{}, err
	}

	now := d.now()
	acct.BalanceMinor += in.AmountMinor
	acct.UpdatedAt = now
	acct.Version++
	if err := acct.Validate(); err != nil {
		return domain.Account{}, domain.Entry{}, err
	}
	if err := d.Wallets.UpdateAccount(ctx, acct); err != nil {
		return domain.Account{}, domain.Entry{}, err
	}
	entry := domain.Entry{
		ID: d.newID(), WalletID: w.ID, AccountID: acct.ID, TenantID: in.TenantID,
		Kind: domain.EntryCredit, AmountMinor: in.AmountMinor, Currency: acct.Currency,
		BalanceAfter: acct.BalanceMinor, HeldAfter: acct.HeldMinor,
		Reference: in.Reference, IdempotencyKey: in.IdempotencyKey,
		Metadata: in.Metadata, CreatedAt: now,
	}
	if err := d.Wallets.CreateEntry(ctx, entry); err != nil {
		return domain.Account{}, domain.Entry{}, err
	}
	d.emit(ctx, w, domain.EventWalletCredited, map[string]any{
		"accountType": acct.AccountType, "amountMinor": in.AmountMinor, "entryId": entry.ID.String(),
	})
	d.postLedger(ctx, in.TenantID, in.IdempotencyKey, w.ID.String(), in.AmountMinor, acct.Currency)
	return acct, entry, nil
}

// Debit removes available funds (fails if available < amount).
func (d *Deps) Debit(ctx context.Context, in MoneyInput) (domain.Account, domain.Entry, error) {
	if in.IdempotencyKey == "" {
		return domain.Account{}, domain.Entry{}, fmt.Errorf("%w: idempotency_key required", domain.ErrInvalidArgument)
	}
	if in.AmountMinor <= 0 {
		return domain.Account{}, domain.Entry{}, fmt.Errorf("%w: amount must be > 0", domain.ErrInvalidArgument)
	}
	if existing, err := d.Wallets.GetEntryByIdempotency(ctx, in.TenantID, in.IdempotencyKey); err == nil {
		acct, _ := d.Wallets.GetAccount(ctx, in.TenantID, existing.AccountID)
		return acct, existing, nil
	}

	w, err := d.Wallets.GetWallet(ctx, in.TenantID, in.WalletID)
	if err != nil {
		return domain.Account{}, domain.Entry{}, err
	}
	acct, err := d.resolveAccount(ctx, in.TenantID, in.WalletID, in.AccountType)
	if err != nil {
		return domain.Account{}, domain.Entry{}, err
	}
	if acct.Available() < in.AmountMinor {
		return domain.Account{}, domain.Entry{}, fmt.Errorf("%w: available=%d need=%d", domain.ErrOverdraft, acct.Available(), in.AmountMinor)
	}

	now := d.now()
	acct.BalanceMinor -= in.AmountMinor
	acct.UpdatedAt = now
	acct.Version++
	if err := acct.Validate(); err != nil {
		return domain.Account{}, domain.Entry{}, err
	}
	if err := d.Wallets.UpdateAccount(ctx, acct); err != nil {
		return domain.Account{}, domain.Entry{}, err
	}
	entry := domain.Entry{
		ID: d.newID(), WalletID: w.ID, AccountID: acct.ID, TenantID: in.TenantID,
		Kind: domain.EntryDebit, AmountMinor: in.AmountMinor, Currency: acct.Currency,
		BalanceAfter: acct.BalanceMinor, HeldAfter: acct.HeldMinor,
		Reference: in.Reference, IdempotencyKey: in.IdempotencyKey,
		Metadata: in.Metadata, CreatedAt: now,
	}
	if err := d.Wallets.CreateEntry(ctx, entry); err != nil {
		return domain.Account{}, domain.Entry{}, err
	}
	d.emit(ctx, w, domain.EventWalletDebited, map[string]any{
		"accountType": acct.AccountType, "amountMinor": in.AmountMinor, "entryId": entry.ID.String(),
	})
	d.postLedger(ctx, in.TenantID, in.IdempotencyKey, w.ID.String(), in.AmountMinor, acct.Currency)
	return acct, entry, nil
}
