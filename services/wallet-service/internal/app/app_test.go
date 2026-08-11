package app_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/wallet-service/internal/app"
	"github.com/nexora/wallet-service/internal/app/memory"
	"github.com/nexora/wallet-service/internal/domain"
)

type testEnv struct {
	Deps      *app.Deps
	Tenant    uuid.UUID
	Principal uuid.UUID
}

func testDeps(t *testing.T) *testEnv {
	t.Helper()
	store := memory.NewStore()
	wallets, outbox := memory.NewRepos(store)
	deps := &app.Deps{
		Wallets: wallets, Outbox: outbox,
		Publisher: &memory.EventPublisher{S: store},
		Ledger:    &memory.LedgerClient{},
		Clock:     &memory.Clock{T: time.Date(2026, 8, 6, 12, 0, 0, 0, time.UTC)},
		IDs:       memory.IDGen{},
	}
	return &testEnv{
		Deps:      deps,
		Tenant:    uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		Principal: uuid.MustParse("22222222-2222-2222-2222-222222222222"),
	}
}

func TestCreditDebit(t *testing.T) {
	env := testDeps(t)
	ctx := context.Background()
	view, err := env.Deps.GetOrCreate(ctx, app.GetOrCreateInput{
		TenantID: env.Tenant, PrincipalID: env.Principal, Currency: "TRY",
	})
	if err != nil {
		t.Fatalf("getorcreate: %v", err)
	}
	acct, _, err := env.Deps.Credit(ctx, app.MoneyInput{
		TenantID: env.Tenant, WalletID: view.Wallet.ID, AccountType: domain.AccountCash,
		AmountMinor: 10000, IdempotencyKey: "c1",
	})
	if err != nil {
		t.Fatalf("credit: %v", err)
	}
	if acct.BalanceMinor != 10000 || acct.Available() != 10000 {
		t.Fatalf("after credit: %+v", acct)
	}
	acct, _, err = env.Deps.Debit(ctx, app.MoneyInput{
		TenantID: env.Tenant, WalletID: view.Wallet.ID, AccountType: domain.AccountCash,
		AmountMinor: 3000, IdempotencyKey: "d1",
	})
	if err != nil {
		t.Fatalf("debit: %v", err)
	}
	if acct.BalanceMinor != 7000 {
		t.Fatalf("balance=%d", acct.BalanceMinor)
	}
}

func TestHoldBlocksDebit(t *testing.T) {
	env := testDeps(t)
	ctx := context.Background()
	view, _ := env.Deps.GetOrCreate(ctx, app.GetOrCreateInput{
		TenantID: env.Tenant, PrincipalID: env.Principal, Currency: "TRY",
	})
	_, _, _ = env.Deps.Credit(ctx, app.MoneyInput{
		TenantID: env.Tenant, WalletID: view.Wallet.ID, AmountMinor: 5000, IdempotencyKey: "c-hold",
	})
	hold, acct, err := env.Deps.Hold(ctx, app.HoldInput{
		TenantID: env.Tenant, WalletID: view.Wallet.ID, AmountMinor: 4000, IdempotencyKey: "h1",
	})
	if err != nil {
		t.Fatalf("hold: %v", err)
	}
	if acct.Available() != 1000 || hold.Status != domain.HoldActive {
		t.Fatalf("available=%d hold=%+v", acct.Available(), hold)
	}
	_, _, err = env.Deps.Debit(ctx, app.MoneyInput{
		TenantID: env.Tenant, WalletID: view.Wallet.ID, AmountMinor: 2000, IdempotencyKey: "d-blocked",
	})
	if !errors.Is(err, domain.ErrOverdraft) {
		t.Fatalf("expected overdraft, got %v", err)
	}
}

func TestRelease(t *testing.T) {
	env := testDeps(t)
	ctx := context.Background()
	view, _ := env.Deps.GetOrCreate(ctx, app.GetOrCreateInput{
		TenantID: env.Tenant, PrincipalID: env.Principal, Currency: "TRY",
	})
	_, _, _ = env.Deps.Credit(ctx, app.MoneyInput{
		TenantID: env.Tenant, WalletID: view.Wallet.ID, AmountMinor: 5000, IdempotencyKey: "c-rel",
	})
	hold, _, _ := env.Deps.Hold(ctx, app.HoldInput{
		TenantID: env.Tenant, WalletID: view.Wallet.ID, AmountMinor: 3000, IdempotencyKey: "h-rel",
	})
	_, acct, err := env.Deps.Release(ctx, app.ReleaseInput{TenantID: env.Tenant, HoldID: hold.ID})
	if err != nil {
		t.Fatalf("release: %v", err)
	}
	if acct.HeldMinor != 0 || acct.Available() != 5000 {
		t.Fatalf("after release: avail=%d held=%d", acct.Available(), acct.HeldMinor)
	}
}

func TestTransfer(t *testing.T) {
	env := testDeps(t)
	ctx := context.Background()
	view, _ := env.Deps.GetOrCreate(ctx, app.GetOrCreateInput{
		TenantID: env.Tenant, PrincipalID: env.Principal, Currency: "TRY",
	})
	_, _, _ = env.Deps.Credit(ctx, app.MoneyInput{
		TenantID: env.Tenant, WalletID: view.Wallet.ID, AccountType: domain.AccountCash,
		AmountMinor: 8000, IdempotencyKey: "c-xfer",
	})
	_, err := env.Deps.Transfer(ctx, app.TransferInput{
		TenantID: env.Tenant, FromWalletID: view.Wallet.ID, FromAccountType: domain.AccountCash,
		ToWalletID: view.Wallet.ID, ToAccountType: domain.AccountPromo,
		AmountMinor: 2000, IdempotencyKey: "xfer-1",
	})
	if err != nil {
		t.Fatalf("transfer: %v", err)
	}
	cash, _ := env.Deps.Wallets.GetAccountByType(ctx, env.Tenant, view.Wallet.ID, domain.AccountCash)
	promo, _ := env.Deps.Wallets.GetAccountByType(ctx, env.Tenant, view.Wallet.ID, domain.AccountPromo)
	if cash.BalanceMinor != 6000 || promo.BalanceMinor != 2000 {
		t.Fatalf("cash=%d promo=%d", cash.BalanceMinor, promo.BalanceMinor)
	}
}

func TestOverdraftFails(t *testing.T) {
	env := testDeps(t)
	ctx := context.Background()
	view, _ := env.Deps.GetOrCreate(ctx, app.GetOrCreateInput{
		TenantID: env.Tenant, PrincipalID: env.Principal, Currency: "TRY",
	})
	_, _, _ = env.Deps.Credit(ctx, app.MoneyInput{
		TenantID: env.Tenant, WalletID: view.Wallet.ID, AmountMinor: 1000, IdempotencyKey: "c-od",
	})
	_, _, err := env.Deps.Debit(ctx, app.MoneyInput{
		TenantID: env.Tenant, WalletID: view.Wallet.ID, AmountMinor: 1001, IdempotencyKey: "d-od",
	})
	if !errors.Is(err, domain.ErrOverdraft) {
		t.Fatalf("expected overdraft, got %v", err)
	}
}
