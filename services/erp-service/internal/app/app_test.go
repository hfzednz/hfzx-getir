package app_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/nexora/erp-service/internal/app"
	"github.com/nexora/erp-service/internal/app/memory"
	"github.com/nexora/erp-service/internal/domain"
)

func testDeps() *app.Deps {
	s := memory.NewStore()
	r := memory.NewRepos(s)
	return &app.Deps{
		Companies: r.Companies, Periods: r.Periods, Accounts: r.Accounts,
		Journals: r.Journals, Suppliers: r.Suppliers, Procurement: r.Procurement,
		AP: r.AP, AR: r.AR, Treasury: r.Treasury, Budgets: r.Budgets,
		Assets: r.Assets, Tax: r.Tax, Expenses: r.Expenses,
		Approvals: r.Approvals, Payroll: r.Payroll, Outbox: r.Outbox,
		Ledger: r.Ledger, Inventory: r.Inventory, Settlement: r.Settlement, AI: r.AI,
		Clock: app.SystemClock{}, IDs: app.UUIDGen{},
	}
}

func TestJournalAndProcurementFlow(t *testing.T) {
	ctx := context.Background()
	d := testDeps()
	tid := uuid.New()
	company, err := d.UpsertCompany(ctx, domain.Company{TenantID: tid, Code: "nx", Name: "Nexora TR", Currency: "TRY"})
	if err != nil {
		t.Fatal(err)
	}
	fy, periods, err := d.OpenFiscalYear(ctx, domain.FiscalYear{TenantID: tid, CompanyID: company.ID, Label: "2026"}, 12)
	if err != nil || fy.ID == uuid.Nil || len(periods) != 12 {
		t.Fatalf("fy %v periods %d err %v", fy.ID, len(periods), err)
	}
	for _, code := range []struct {
		c, n, typ string
	}{
		{"1000", "Cash", domain.AccountAsset},
		{"2000", "AP", domain.AccountLiability},
		{"1600", "Accum Dep", domain.AccountAsset},
		{"6100", "Dep Exp", domain.AccountExpense},
	} {
		if _, err := d.UpsertAccount(ctx, domain.ChartAccount{
			TenantID: tid, CompanyID: company.ID, Code: code.c, Name: code.n, Type: code.typ,
		}); err != nil {
			t.Fatal(err)
		}
	}
	j, err := d.PostJournal(ctx, domain.JournalEntry{
		TenantID: tid, CompanyID: company.ID, PeriodID: periods[0].ID, Currency: "TRY",
		Memo: "opening", IdempotencyKey: "j1",
		Lines: []domain.JournalLine{
			{AccountCode: "1000", DebitMinor: 5000},
			{AccountCode: "2000", CreditMinor: 5000},
		},
	})
	if err != nil || j.Status != "posted" || j.LedgerRef == "" {
		t.Fatalf("journal %+v err %v", j, err)
	}
	j2, err := d.PostJournal(ctx, domain.JournalEntry{
		TenantID: tid, CompanyID: company.ID, PeriodID: periods[0].ID, Currency: "TRY",
		IdempotencyKey: "j1",
		Lines: []domain.JournalLine{
			{AccountCode: "1000", DebitMinor: 1},
			{AccountCode: "2000", CreditMinor: 1},
		},
	})
	if err != nil || j2.ID != j.ID {
		t.Fatalf("idempotency failed %+v", j2)
	}

	sup, err := d.UpsertSupplier(ctx, domain.Supplier{
		TenantID: tid, CompanyID: company.ID, Code: "acme", Name: "ACME", Currency: "TRY",
	})
	if err != nil {
		t.Fatal(err)
	}
	lines := []domain.ProcurementLine{{SKU: "SKU1", Qty: 2, UnitMinor: 1000}}
	po, err := d.CreatePO(ctx, domain.PurchaseOrder{
		TenantID: tid, CompanyID: company.ID, SupplierID: sup.ID, Lines: lines, Currency: "TRY",
	})
	if err != nil || po.TotalMinor != 2000 {
		t.Fatalf("po %+v err %v", po, err)
	}
	grn, err := d.ReceiveGoods(ctx, domain.GoodsReceipt{TenantID: tid, POID: po.ID}, "wh-1")
	if err != nil || grn.ID == uuid.Nil {
		t.Fatal(err)
	}
	poID := po.ID
	inv, err := d.CreateAPInvoice(ctx, domain.APInvoice{
		TenantID: tid, CompanyID: company.ID, SupplierID: sup.ID, POID: &poID,
		InvoiceNumber: "INV-1", SubtotalMinor: 2000, Currency: "TRY",
	}, 2000)
	if err != nil || inv.Status != "matched" || inv.TaxMinor != 400 {
		t.Fatalf("ap %+v err %v", inv, err)
	}
	inv, err = d.ApproveAPInvoice(ctx, tid, inv.ID, uuid.New())
	if err != nil || inv.Status != "approved" {
		t.Fatal(err)
	}
	inv, ref, err := d.ScheduleAPPayment(ctx, tid, inv.ID)
	if err != nil || inv.Status != "scheduled" || ref == "" {
		t.Fatalf("schedule %v %s %v", inv.Status, ref, err)
	}

	asset, err := d.CreateAsset(ctx, domain.FixedAsset{
		TenantID: tid, CompanyID: company.ID, Code: "LAP1", Name: "Laptop",
		CostMinor: 36000, UsefulLifeMonths: 36, Currency: "TRY",
	})
	if err != nil {
		t.Fatal(err)
	}
	n, err := d.RunDepreciation(ctx, tid, company.ID, periods[0].ID)
	if err != nil || n != 1 {
		t.Fatalf("dep n=%d err=%v", n, err)
	}
	got, _ := d.Assets.Get(ctx, tid, asset.ID)
	if got.AccumDepMinor != 1000 {
		t.Fatalf("accum %d", got.AccumDepMinor)
	}

	b, err := d.UpsertBudget(ctx, domain.Budget{
		TenantID: tid, CompanyID: company.ID, Label: "FY26", Period: "annual", Year: 2026,
		Lines: []domain.BudgetLine{{AccountCode: "6100", AmountMinor: 12000}},
	})
	if err != nil {
		t.Fatal(err)
	}
	b, err = d.ApproveBudget(ctx, tid, b.ID, uuid.New())
	if err != nil || b.Status != "approved" {
		t.Fatal(err)
	}
	tax, err := d.CalculateTax(ctx, domain.TaxReturn{
		TenantID: tid, CompanyID: company.ID, Kind: "vat", TaxableMinor: 10000, Currency: "TRY",
	}, 2000)
	if err != nil || tax.TaxMinor != 2000 {
		t.Fatal(err)
	}
}
