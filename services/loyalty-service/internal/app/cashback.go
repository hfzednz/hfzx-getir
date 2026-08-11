package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/loyalty-service/internal/app/ports"
	"github.com/nexora/loyalty-service/internal/domain"
)

// GrantCashbackInput creates a pending cashback grant.
type GrantCashbackInput struct {
	TenantID       uuid.UUID
	PrincipalID    uuid.UUID
	AmountMinor    int64
	Currency       string
	AccountType    string // cashback | promo
	OrderID        *uuid.UUID
	IdempotencyKey string
}

// GrantCashback records a pending grant (wallet not yet credited).
func (d *Deps) GrantCashback(ctx context.Context, in GrantCashbackInput) (domain.CashbackGrant, error) {
	if in.AmountMinor <= 0 {
		return domain.CashbackGrant{}, fmt.Errorf("%w: amount_minor must be > 0", domain.ErrInvalidArgument)
	}
	if in.Currency == "" {
		in.Currency = "TRY"
	}
	if in.AccountType == "" {
		in.AccountType = "cashback"
	}
	if in.IdempotencyKey == "" {
		return domain.CashbackGrant{}, fmt.Errorf("%w: idempotency_key required", domain.ErrInvalidArgument)
	}
	if existing, err := d.Cashbacks.GetGrantByIdempotency(ctx, in.TenantID, in.IdempotencyKey); err == nil {
		return existing, nil
	}

	acct, err := d.EnsureAccount(ctx, EnsureAccountInput{TenantID: in.TenantID, PrincipalID: in.PrincipalID})
	if err != nil {
		return domain.CashbackGrant{}, err
	}
	now := d.now()
	g := domain.CashbackGrant{
		ID: d.newID(), TenantID: in.TenantID, AccountID: acct.ID, PrincipalID: in.PrincipalID,
		AmountMinor: in.AmountMinor, Currency: in.Currency, AccountType: in.AccountType,
		Status: domain.CashbackPending, OrderID: in.OrderID, IdempotencyKey: in.IdempotencyKey,
		CreatedAt: now, UpdatedAt: now,
	}
	if err := g.Validate(); err != nil {
		return domain.CashbackGrant{}, err
	}
	if err := d.Cashbacks.CreateGrant(ctx, g); err != nil {
		return domain.CashbackGrant{}, err
	}
	return g, nil
}

// ConfirmCashbackInput confirms a grant to wallet.
type ConfirmCashbackInput struct {
	TenantID uuid.UUID
	GrantID  uuid.UUID
}

// ConfirmCashbackToWallet calls WalletClient.Credit and marks grant issued/failed.
func (d *Deps) ConfirmCashbackToWallet(ctx context.Context, in ConfirmCashbackInput) (domain.CashbackGrant, error) {
	g, err := d.Cashbacks.GetGrant(ctx, in.TenantID, in.GrantID)
	if err != nil {
		return domain.CashbackGrant{}, err
	}
	if g.Status == domain.CashbackIssued {
		return g, nil
	}
	if g.Status != domain.CashbackPending {
		return domain.CashbackGrant{}, fmt.Errorf("%w: status=%s", domain.ErrCashbackState, g.Status)
	}
	if d.Wallet == nil {
		return domain.CashbackGrant{}, fmt.Errorf("%w: wallet client required", domain.ErrInvariant)
	}

	res, err := d.Wallet.Credit(ctx, ports.WalletCreditRequest{
		TenantID: g.TenantID, PrincipalID: g.PrincipalID,
		AmountMinor: g.AmountMinor, Currency: g.Currency, AccountType: g.AccountType,
		IdempotencyKey: "loyalty-cashback:" + g.ID.String(),
		Reference:      "cashback:" + g.ID.String(),
	})
	now := d.now()
	if err != nil {
		g.Status = domain.CashbackFailed
		g.FailureReason = err.Error()
		g.UpdatedAt = now
		_ = d.Cashbacks.UpdateGrant(ctx, g)
		return g, err
	}
	g.Status = domain.CashbackIssued
	g.WalletRef = res.EntryID
	if g.WalletRef == "" {
		g.WalletRef = res.WalletID
	}
	g.UpdatedAt = now
	if err := d.Cashbacks.UpdateGrant(ctx, g); err != nil {
		return domain.CashbackGrant{}, err
	}
	acct, _ := d.Accounts.GetAccount(ctx, g.TenantID, g.AccountID)
	d.emit(ctx, acct, domain.EventCashbackIssued, map[string]any{
		"grantId": g.ID.String(), "amountMinor": g.AmountMinor, "currency": g.Currency,
	})
	return g, nil
}

// ConfirmToWallet is an alias matching the use-case name.
func (d *Deps) ConfirmToWallet(ctx context.Context, in ConfirmCashbackInput) (domain.CashbackGrant, error) {
	return d.ConfirmCashbackToWallet(ctx, in)
}
