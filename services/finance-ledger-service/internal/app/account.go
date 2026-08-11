package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/nexora/finance-ledger-service/internal/domain"
)

// EnsureAccountInput creates or returns an existing account by code.
type EnsureAccountInput struct {
	TenantID uuid.UUID
	Code     string
	Name     string
	Type     domain.AccountType
	Currency string
}

// EnsureAccount upserts a chart-of-accounts entry by tenant+code (idempotent).
func (d *Deps) EnsureAccount(ctx context.Context, in EnsureAccountInput) (domain.Account, error) {
	if in.TenantID == uuid.Nil {
		return domain.Account{}, fmt.Errorf("%w: tenant_id", domain.ErrInvalidArgument)
	}
	code := strings.TrimSpace(in.Code)
	if code == "" {
		return domain.Account{}, fmt.Errorf("%w: code", domain.ErrInvalidArgument)
	}
	if existing, err := d.Accounts.GetByCode(ctx, in.TenantID, code); err == nil {
		return existing, nil
	} else if err != domain.ErrNotFound {
		return domain.Account{}, err
	}
	now := d.now()
	acc := domain.Account{
		ID:        d.newID(),
		TenantID:  in.TenantID,
		Code:      code,
		Name:      strings.TrimSpace(in.Name),
		Type:      in.Type,
		Currency:  strings.ToUpper(strings.TrimSpace(in.Currency)),
		Active:    true,
		CreatedAt: now,
		UpdatedAt: now,
		Version:   1,
	}
	if acc.Name == "" {
		acc.Name = code
	}
	if err := acc.Validate(); err != nil {
		return domain.Account{}, err
	}
	if err := d.Accounts.Create(ctx, acc); err != nil {
		return domain.Account{}, err
	}
	_ = d.appendEvent(ctx, acc.ID, acc.TenantID, domain.EventAccountEnsured, map[string]any{
		"code": acc.Code, "type": string(acc.Type), "currency": acc.Currency,
	})
	return acc, nil
}

// GetBalanceInput fetches account balance from posted lines.
type GetBalanceInput struct {
	TenantID  uuid.UUID
	AccountID uuid.UUID
}

// AccountBalance is the balance projection.
type AccountBalance struct {
	AccountID     uuid.UUID
	Currency      string
	BalanceMinor  int64
}

// GetBalance returns debit−credit net for a posted-lines account.
func (d *Deps) GetBalance(ctx context.Context, in GetBalanceInput) (AccountBalance, error) {
	if in.TenantID == uuid.Nil || in.AccountID == uuid.Nil {
		return AccountBalance{}, fmt.Errorf("%w: tenant_id and account_id required", domain.ErrInvalidArgument)
	}
	acc, err := d.Accounts.GetByID(ctx, in.TenantID, in.AccountID)
	if err != nil {
		return AccountBalance{}, err
	}
	bal, err := d.Journals.BalanceMinor(ctx, in.TenantID, in.AccountID)
	if err != nil {
		return AccountBalance{}, err
	}
	return AccountBalance{AccountID: acc.ID, Currency: acc.Currency, BalanceMinor: bal}, nil
}
