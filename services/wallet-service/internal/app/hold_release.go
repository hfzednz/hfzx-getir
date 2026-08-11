package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/wallet-service/internal/domain"
)

// HoldInput reserves available balance.
type HoldInput struct {
	TenantID       uuid.UUID
	WalletID       uuid.UUID
	AccountType    domain.AccountType
	AmountMinor    int64
	IdempotencyKey string
	Reference      string
}

// Hold reserves funds (increases held; available decreases).
func (d *Deps) Hold(ctx context.Context, in HoldInput) (domain.Hold, domain.Account, error) {
	if in.IdempotencyKey == "" {
		return domain.Hold{}, domain.Account{}, fmt.Errorf("%w: idempotency_key required", domain.ErrInvalidArgument)
	}
	if in.AmountMinor <= 0 {
		return domain.Hold{}, domain.Account{}, fmt.Errorf("%w: amount must be > 0", domain.ErrInvalidArgument)
	}
	if existing, err := d.Wallets.GetHoldByIdempotency(ctx, in.TenantID, in.IdempotencyKey); err == nil {
		acct, _ := d.Wallets.GetAccount(ctx, in.TenantID, existing.AccountID)
		return existing, acct, nil
	}

	w, err := d.Wallets.GetWallet(ctx, in.TenantID, in.WalletID)
	if err != nil {
		return domain.Hold{}, domain.Account{}, err
	}
	acct, err := d.resolveAccount(ctx, in.TenantID, in.WalletID, in.AccountType)
	if err != nil {
		return domain.Hold{}, domain.Account{}, err
	}
	if acct.Available() < in.AmountMinor {
		return domain.Hold{}, domain.Account{}, fmt.Errorf("%w: available=%d need=%d", domain.ErrOverdraft, acct.Available(), in.AmountMinor)
	}

	now := d.now()
	acct.HeldMinor += in.AmountMinor
	acct.UpdatedAt = now
	acct.Version++
	if err := acct.Validate(); err != nil {
		return domain.Hold{}, domain.Account{}, err
	}
	if err := d.Wallets.UpdateAccount(ctx, acct); err != nil {
		return domain.Hold{}, domain.Account{}, err
	}
	hold := domain.Hold{
		ID: d.newID(), WalletID: w.ID, AccountID: acct.ID, TenantID: in.TenantID,
		AmountMinor: in.AmountMinor, Currency: acct.Currency, Status: domain.HoldActive,
		Reference: in.Reference, IdempotencyKey: in.IdempotencyKey, CreatedAt: now,
	}
	if err := d.Wallets.CreateHold(ctx, hold); err != nil {
		return domain.Hold{}, domain.Account{}, err
	}
	_ = d.Wallets.CreateEntry(ctx, domain.Entry{
		ID: d.newID(), WalletID: w.ID, AccountID: acct.ID, TenantID: in.TenantID,
		Kind: domain.EntryHold, AmountMinor: in.AmountMinor, Currency: acct.Currency,
		BalanceAfter: acct.BalanceMinor, HeldAfter: acct.HeldMinor,
		Reference: in.Reference, IdempotencyKey: in.IdempotencyKey + ":entry", CreatedAt: now,
	})
	d.emit(ctx, w, domain.EventWalletHeld, map[string]any{
		"holdId": hold.ID.String(), "amountMinor": in.AmountMinor,
	})
	return hold, acct, nil
}

// ReleaseInput releases an active hold.
type ReleaseInput struct {
	TenantID uuid.UUID
	HoldID   uuid.UUID
}

// Release frees held funds back to available.
func (d *Deps) Release(ctx context.Context, in ReleaseInput) (domain.Hold, domain.Account, error) {
	hold, err := d.Wallets.GetHold(ctx, in.TenantID, in.HoldID)
	if err != nil {
		return domain.Hold{}, domain.Account{}, err
	}
	if hold.Status == domain.HoldReleased {
		acct, _ := d.Wallets.GetAccount(ctx, in.TenantID, hold.AccountID)
		return hold, acct, nil
	}
	if hold.Status != domain.HoldActive {
		return domain.Hold{}, domain.Account{}, fmt.Errorf("%w: status %s", domain.ErrHoldReleased, hold.Status)
	}

	acct, err := d.Wallets.GetAccount(ctx, in.TenantID, hold.AccountID)
	if err != nil {
		return domain.Hold{}, domain.Account{}, err
	}
	w, err := d.Wallets.GetWallet(ctx, in.TenantID, hold.WalletID)
	if err != nil {
		return domain.Hold{}, domain.Account{}, err
	}

	now := d.now()
	if acct.HeldMinor < hold.AmountMinor {
		return domain.Hold{}, domain.Account{}, fmt.Errorf("%w: held underflow", domain.ErrInvariant)
	}
	acct.HeldMinor -= hold.AmountMinor
	acct.UpdatedAt = now
	acct.Version++
	if err := d.Wallets.UpdateAccount(ctx, acct); err != nil {
		return domain.Hold{}, domain.Account{}, err
	}
	hold.Status = domain.HoldReleased
	hold.ReleasedAt = &now
	if err := d.Wallets.UpdateHold(ctx, hold); err != nil {
		return domain.Hold{}, domain.Account{}, err
	}
	_ = d.Wallets.CreateEntry(ctx, domain.Entry{
		ID: d.newID(), WalletID: w.ID, AccountID: acct.ID, TenantID: in.TenantID,
		Kind: domain.EntryRelease, AmountMinor: hold.AmountMinor, Currency: acct.Currency,
		BalanceAfter: acct.BalanceMinor, HeldAfter: acct.HeldMinor,
		Reference: hold.ID.String(), IdempotencyKey: "release:" + hold.ID.String(), CreatedAt: now,
	})
	d.emit(ctx, w, domain.EventWalletReleased, map[string]any{"holdId": hold.ID.String()})
	return hold, acct, nil
}
