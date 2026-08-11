package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/wallet-service/internal/domain"
)

// TransferInput moves funds between accounts (same or different wallets, same tenant).
type TransferInput struct {
	TenantID         uuid.UUID
	FromWalletID     uuid.UUID
	FromAccountType  domain.AccountType
	ToWalletID       uuid.UUID
	ToAccountType    domain.AccountType
	AmountMinor      int64
	IdempotencyKey   string
	Reference        string
}

// Transfer debits source available and credits destination.
func (d *Deps) Transfer(ctx context.Context, in TransferInput) (domain.Transfer, error) {
	if in.IdempotencyKey == "" {
		return domain.Transfer{}, fmt.Errorf("%w: idempotency_key required", domain.ErrInvalidArgument)
	}
	if in.AmountMinor <= 0 {
		return domain.Transfer{}, fmt.Errorf("%w: amount must be > 0", domain.ErrInvalidArgument)
	}
	if existing, err := d.Wallets.GetTransferByIdempotency(ctx, in.TenantID, in.IdempotencyKey); err == nil {
		return existing, nil
	}
	if in.ToWalletID == uuid.Nil {
		in.ToWalletID = in.FromWalletID
	}

	fromW, err := d.Wallets.GetWallet(ctx, in.TenantID, in.FromWalletID)
	if err != nil {
		return domain.Transfer{}, err
	}
	toW, err := d.Wallets.GetWallet(ctx, in.TenantID, in.ToWalletID)
	if err != nil {
		return domain.Transfer{}, err
	}
	fromAcct, err := d.resolveAccount(ctx, in.TenantID, in.FromWalletID, in.FromAccountType)
	if err != nil {
		return domain.Transfer{}, err
	}
	toType := in.ToAccountType
	if toType == "" {
		toType = domain.AccountCash
	}
	toAcct, err := d.resolveAccount(ctx, in.TenantID, in.ToWalletID, toType)
	if err != nil {
		return domain.Transfer{}, err
	}
	if fromAcct.Currency != toAcct.Currency {
		return domain.Transfer{}, domain.ErrCurrencyMismatch
	}
	if fromAcct.Available() < in.AmountMinor {
		return domain.Transfer{}, fmt.Errorf("%w: available=%d need=%d", domain.ErrOverdraft, fromAcct.Available(), in.AmountMinor)
	}

	now := d.now()
	fromAcct.BalanceMinor -= in.AmountMinor
	fromAcct.UpdatedAt = now
	fromAcct.Version++
	toAcct.BalanceMinor += in.AmountMinor
	toAcct.UpdatedAt = now
	toAcct.Version++
	if err := fromAcct.Validate(); err != nil {
		return domain.Transfer{}, err
	}
	if err := d.Wallets.UpdateAccount(ctx, fromAcct); err != nil {
		return domain.Transfer{}, err
	}
	if err := d.Wallets.UpdateAccount(ctx, toAcct); err != nil {
		return domain.Transfer{}, err
	}

	xfer := domain.Transfer{
		ID: d.newID(), TenantID: in.TenantID,
		FromWalletID: fromW.ID, FromAccountID: fromAcct.ID,
		ToWalletID: toW.ID, ToAccountID: toAcct.ID,
		AmountMinor: in.AmountMinor, Currency: fromAcct.Currency,
		IdempotencyKey: in.IdempotencyKey, Reference: in.Reference, CreatedAt: now,
	}
	if err := d.Wallets.CreateTransfer(ctx, xfer); err != nil {
		return domain.Transfer{}, err
	}
	_ = d.Wallets.CreateEntry(ctx, domain.Entry{
		ID: d.newID(), WalletID: fromW.ID, AccountID: fromAcct.ID, TenantID: in.TenantID,
		Kind: domain.EntryDebit, AmountMinor: in.AmountMinor, Currency: fromAcct.Currency,
		BalanceAfter: fromAcct.BalanceMinor, HeldAfter: fromAcct.HeldMinor,
		Reference: xfer.ID.String(), IdempotencyKey: in.IdempotencyKey + ":from", CreatedAt: now,
	})
	_ = d.Wallets.CreateEntry(ctx, domain.Entry{
		ID: d.newID(), WalletID: toW.ID, AccountID: toAcct.ID, TenantID: in.TenantID,
		Kind: domain.EntryCredit, AmountMinor: in.AmountMinor, Currency: toAcct.Currency,
		BalanceAfter: toAcct.BalanceMinor, HeldAfter: toAcct.HeldMinor,
		Reference: xfer.ID.String(), IdempotencyKey: in.IdempotencyKey + ":to", CreatedAt: now,
	})
	d.emit(ctx, fromW, domain.EventWalletTransfer, map[string]any{
		"transferId": xfer.ID.String(), "amountMinor": in.AmountMinor,
		"toWalletId": toW.ID.String(),
	})
	return xfer, nil
}

// HistoryInput lists wallet entries.
type HistoryInput struct {
	TenantID uuid.UUID
	WalletID uuid.UUID
	Limit    int
	Offset   int
}

// History returns paginated ledger entries.
func (d *Deps) History(ctx context.Context, in HistoryInput) ([]domain.Entry, int, error) {
	limit := in.Limit
	if limit <= 0 {
		limit = 50
	}
	return d.Wallets.ListEntries(ctx, in.TenantID, in.WalletID, limit, in.Offset)
}

// AdminAdjustInput is an audited balance adjustment (can go either direction).
type AdminAdjustInput struct {
	TenantID       uuid.UUID
	WalletID       uuid.UUID
	AccountType    domain.AccountType
	AmountMinor    int64 // positive = credit, negative = debit
	IdempotencyKey string
	Reason         string
	ActorID        *uuid.UUID
}

// AdminAdjust applies an audited admin balance change.
func (d *Deps) AdminAdjust(ctx context.Context, in AdminAdjustInput) (domain.Account, domain.Entry, error) {
	if in.IdempotencyKey == "" {
		return domain.Account{}, domain.Entry{}, fmt.Errorf("%w: idempotency_key required", domain.ErrInvalidArgument)
	}
	if in.AmountMinor == 0 {
		return domain.Account{}, domain.Entry{}, fmt.Errorf("%w: amount required", domain.ErrInvalidArgument)
	}
	if in.Reason == "" {
		return domain.Account{}, domain.Entry{}, fmt.Errorf("%w: reason required", domain.ErrInvalidArgument)
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
	abs := in.AmountMinor
	if abs < 0 {
		abs = -abs
	}
	if in.AmountMinor > 0 {
		acct.BalanceMinor += in.AmountMinor
	} else {
		if acct.Available() < abs {
			return domain.Account{}, domain.Entry{}, fmt.Errorf("%w: available=%d need=%d", domain.ErrOverdraft, acct.Available(), abs)
		}
		acct.BalanceMinor -= abs
	}
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
		Kind: domain.EntryAdjust, AmountMinor: abs, Currency: acct.Currency,
		BalanceAfter: acct.BalanceMinor, HeldAfter: acct.HeldMinor,
		Reference: in.Reason, IdempotencyKey: in.IdempotencyKey,
		Metadata: map[string]any{"direction": sign(in.AmountMinor), "reason": in.Reason},
		CreatedAt: now,
	}
	if err := d.Wallets.CreateEntry(ctx, entry); err != nil {
		return domain.Account{}, domain.Entry{}, err
	}
	_ = d.Wallets.CreateAudit(ctx, domain.AuditEntry{
		ID: d.newID(), TenantID: in.TenantID, WalletID: w.ID,
		Action: "admin_adjust", ActorID: in.ActorID,
		AmountMinor: in.AmountMinor, Currency: acct.Currency,
		Detail: map[string]any{"reason": in.Reason, "entryId": entry.ID.String()},
		CreatedAt: now,
	})
	d.emit(ctx, w, domain.EventWalletAdjusted, map[string]any{
		"amountMinor": in.AmountMinor, "reason": in.Reason, "entryId": entry.ID.String(),
	})
	return acct, entry, nil
}

func sign(n int64) string {
	if n < 0 {
		return "debit"
	}
	return "credit"
}
