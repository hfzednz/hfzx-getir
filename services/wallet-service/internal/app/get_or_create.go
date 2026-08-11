package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/wallet-service/internal/app/ports"
	"github.com/nexora/wallet-service/internal/domain"
)

// GetOrCreateInput finds or creates a wallet for a principal.
type GetOrCreateInput struct {
	TenantID    uuid.UUID
	PrincipalID uuid.UUID
	Currency    string
}

// GetOrCreate returns an existing wallet or creates one with all account types.
func (d *Deps) GetOrCreate(ctx context.Context, in GetOrCreateInput) (ports.WalletView, error) {
	if in.TenantID == uuid.Nil || in.PrincipalID == uuid.Nil {
		return ports.WalletView{}, fmt.Errorf("%w: tenant and principal required", domain.ErrInvalidArgument)
	}
	currency := in.Currency
	if currency == "" {
		currency = "TRY"
	}
	if _, err := domain.NewMoney(0, currency); err != nil {
		return ports.WalletView{}, err
	}

	if w, err := d.Wallets.GetWalletByPrincipal(ctx, in.TenantID, in.PrincipalID); err == nil {
		accts, _ := d.Wallets.ListAccounts(ctx, in.TenantID, w.ID)
		return ports.WalletView{Wallet: w, Accounts: accts}, nil
	}

	now := d.now()
	w := domain.Wallet{
		ID: d.newID(), TenantID: in.TenantID, PrincipalID: in.PrincipalID,
		Currency: currency, Active: true, Version: 1, CreatedAt: now, UpdatedAt: now,
	}
	var accounts []domain.Account
	for _, t := range domain.AllAccountTypes() {
		accounts = append(accounts, domain.Account{
			ID: d.newID(), WalletID: w.ID, TenantID: in.TenantID,
			AccountType: t, BalanceMinor: 0, HeldMinor: 0, Currency: currency,
			Version: 1, UpdatedAt: now,
		})
	}
	if err := d.Wallets.CreateWallet(ctx, w, accounts); err != nil {
		if existing, e2 := d.Wallets.GetWalletByPrincipal(ctx, in.TenantID, in.PrincipalID); e2 == nil {
			accts, _ := d.Wallets.ListAccounts(ctx, in.TenantID, existing.ID)
			return ports.WalletView{Wallet: existing, Accounts: accts}, nil
		}
		return ports.WalletView{}, err
	}
	return ports.WalletView{Wallet: w, Accounts: accounts}, nil
}

// GetWallet loads wallet + accounts by id.
func (d *Deps) GetWallet(ctx context.Context, tenantID, walletID uuid.UUID) (ports.WalletView, error) {
	w, err := d.Wallets.GetWallet(ctx, tenantID, walletID)
	if err != nil {
		return ports.WalletView{}, err
	}
	accts, err := d.Wallets.ListAccounts(ctx, tenantID, walletID)
	if err != nil {
		return ports.WalletView{}, err
	}
	return ports.WalletView{Wallet: w, Accounts: accts}, nil
}

func (d *Deps) resolveAccount(ctx context.Context, tenantID, walletID uuid.UUID, t domain.AccountType) (domain.Account, error) {
	if t == "" {
		t = domain.AccountCash
	}
	if !t.Valid() {
		return domain.Account{}, fmt.Errorf("%w: account type %q", domain.ErrInvalidArgument, t)
	}
	return d.Wallets.GetAccountByType(ctx, tenantID, walletID, t)
}
