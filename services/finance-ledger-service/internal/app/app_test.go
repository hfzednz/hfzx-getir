package app_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/finance-ledger-service/internal/app"
	"github.com/nexora/finance-ledger-service/internal/app/memory"
	"github.com/nexora/finance-ledger-service/internal/domain"
)

func testDeps(t *testing.T) (*app.Deps, *memory.Store) {
	t.Helper()
	store := memory.NewStore()
	accounts, journals, invoices, taxRules, events, outbox := memory.NewRepos(store)
	deps := &app.Deps{
		Accounts:  accounts,
		Journals:  journals,
		Invoices:  invoices,
		TaxRules:  taxRules,
		Events:    events,
		Outbox:    outbox,
		Publisher: &memory.EventPublisher{S: store},
		Clock:     &memory.Clock{T: time.Date(2026, 8, 6, 12, 0, 0, 0, time.UTC)},
		IDs:       memory.IDGen{},
	}
	return deps, store
}

func TestPostJournalBalancedOK(t *testing.T) {
	deps, _ := testDeps(t)
	ctx := context.Background()
	tenant := uuid.New()

	cash, err := deps.EnsureAccount(ctx, app.EnsureAccountInput{
		TenantID: tenant, Code: "1000", Name: "Cash", Type: domain.AccountTypeAsset, Currency: "TRY",
	})
	if err != nil {
		t.Fatal(err)
	}
	rev, err := deps.EnsureAccount(ctx, app.EnsureAccountInput{
		TenantID: tenant, Code: "4000", Name: "Revenue", Type: domain.AccountTypeRevenue, Currency: "TRY",
	})
	if err != nil {
		t.Fatal(err)
	}

	j, err := deps.PostJournal(ctx, app.PostJournalInput{
		TenantID: tenant, Currency: "TRY", Reference: "pay-opaque-1", IdempotencyKey: "idem-1",
		Lines: []app.JournalLineInput{
			{AccountID: cash.ID, DebitMinor: 10000},
			{AccountID: rev.ID, CreditMinor: 10000},
		},
	})
	if err != nil {
		t.Fatalf("expected balanced post ok: %v", err)
	}
	if j.Status != domain.JournalStatusPosted {
		t.Fatalf("status=%s", j.Status)
	}
	if j.DebitTotal() != 10000 || j.CreditTotal() != 10000 {
		t.Fatalf("totals debit=%d credit=%d", j.DebitTotal(), j.CreditTotal())
	}
}

func TestPostJournalUnbalancedRejected(t *testing.T) {
	deps, _ := testDeps(t)
	ctx := context.Background()
	tenant := uuid.New()

	cash, _ := deps.EnsureAccount(ctx, app.EnsureAccountInput{
		TenantID: tenant, Code: "1000", Name: "Cash", Type: domain.AccountTypeAsset, Currency: "TRY",
	})
	rev, _ := deps.EnsureAccount(ctx, app.EnsureAccountInput{
		TenantID: tenant, Code: "4000", Name: "Revenue", Type: domain.AccountTypeRevenue, Currency: "TRY",
	})

	_, err := deps.PostJournal(ctx, app.PostJournalInput{
		TenantID: tenant, Currency: "TRY",
		Lines: []app.JournalLineInput{
			{AccountID: cash.ID, DebitMinor: 10000},
			{AccountID: rev.ID, CreditMinor: 9000},
		},
	})
	if !errors.Is(err, domain.ErrUnbalancedJournal) {
		t.Fatalf("expected ErrUnbalancedJournal, got %v", err)
	}
}

func TestCreateInvoice(t *testing.T) {
	deps, _ := testDeps(t)
	ctx := context.Background()
	tenant := uuid.New()

	_, err := deps.UpsertTaxRule(ctx, app.UpsertTaxRuleInput{
		TenantID: tenant, Code: "KDV18", Name: "VAT 18%", RateBps: 1800, Currency: "TRY", Active: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	inv, err := deps.CreateInvoice(ctx, app.CreateInvoiceInput{
		TenantID: tenant, Currency: "TRY", CounterpartyRef: "merchant-opaque", Issue: true,
		IdempotencyKey: "inv-1", DefaultTaxCode: "KDV18",
		Lines: []app.InvoiceLineInput{
			{Description: "Platform fee", Qty: 1, UnitMinor: 10000},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if inv.Status != domain.InvoiceStatusIssued {
		t.Fatalf("status=%s", inv.Status)
	}
	if inv.SubtotalMinor != 10000 {
		t.Fatalf("subtotal=%d", inv.SubtotalMinor)
	}
	if inv.TaxMinor != 1800 {
		t.Fatalf("tax=%d want 1800", inv.TaxMinor)
	}
	if inv.TotalMinor != 11800 {
		t.Fatalf("total=%d", inv.TotalMinor)
	}
}

func TestAccountBalanceFromLines(t *testing.T) {
	deps, _ := testDeps(t)
	ctx := context.Background()
	tenant := uuid.New()

	cash, _ := deps.EnsureAccount(ctx, app.EnsureAccountInput{
		TenantID: tenant, Code: "1000", Name: "Cash", Type: domain.AccountTypeAsset, Currency: "TRY",
	})
	rev, _ := deps.EnsureAccount(ctx, app.EnsureAccountInput{
		TenantID: tenant, Code: "4000", Name: "Revenue", Type: domain.AccountTypeRevenue, Currency: "TRY",
	})
	exp, _ := deps.EnsureAccount(ctx, app.EnsureAccountInput{
		TenantID: tenant, Code: "5000", Name: "Expense", Type: domain.AccountTypeExpense, Currency: "TRY",
	})

	_, err := deps.PostJournal(ctx, app.PostJournalInput{
		TenantID: tenant, Currency: "TRY",
		Lines: []app.JournalLineInput{
			{AccountID: cash.ID, DebitMinor: 5000},
			{AccountID: rev.ID, CreditMinor: 5000},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = deps.PostJournal(ctx, app.PostJournalInput{
		TenantID: tenant, Currency: "TRY",
		Lines: []app.JournalLineInput{
			{AccountID: exp.ID, DebitMinor: 2000},
			{AccountID: cash.ID, CreditMinor: 2000},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	bal, err := deps.GetBalance(ctx, app.GetBalanceInput{TenantID: tenant, AccountID: cash.ID})
	if err != nil {
		t.Fatal(err)
	}
	if bal.BalanceMinor != 3000 {
		t.Fatalf("cash balance=%d want 3000", bal.BalanceMinor)
	}
}
