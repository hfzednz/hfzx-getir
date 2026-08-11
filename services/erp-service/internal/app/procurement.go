package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/nexora/erp-service/internal/domain"
)

func (d *Deps) UpsertSupplier(ctx context.Context, s domain.Supplier) (domain.Supplier, error) {
	if s.TenantID == uuid.Nil || s.CompanyID == uuid.Nil || s.Code == "" || s.Name == "" {
		return s, domain.ErrInvalidArgument
	}
	s.Code = domain.NormalizeCode(s.Code)
	if s.ID == uuid.Nil {
		s.ID = d.newID()
	}
	if s.Status == "" {
		s.Status = "active"
	}
	if s.Currency == "" {
		s.Currency = "TRY"
	}
	s.CreatedAt = d.now()
	if d.AI != nil {
		if risk, err := d.AI.SupplierRisk(ctx, s.TenantID, map[string]float64{"new_supplier": 1}); err == nil {
			s.RiskScore = risk
		}
	}
	if err := d.Suppliers.Save(ctx, s); err != nil {
		return s, err
	}
	d.emit(ctx, s.TenantID, s.ID, domain.EventSupplierAdded, map[string]any{"code": s.Code, "risk": s.RiskScore})
	return s, nil
}

func (d *Deps) CreatePR(ctx context.Context, p domain.PurchaseRequest) (domain.PurchaseRequest, error) {
	if p.TenantID == uuid.Nil || p.CompanyID == uuid.Nil || len(p.Lines) == 0 {
		return p, domain.ErrInvalidArgument
	}
	if p.ID == uuid.Nil {
		p.ID = d.newID()
	}
	p.TotalMinor = domain.LinesTotal(p.Lines)
	if p.Currency == "" {
		p.Currency = "TRY"
	}
	p.Status = "submitted"
	p.CreatedAt = d.now()
	if err := d.Procurement.SavePR(ctx, p); err != nil {
		return p, err
	}
	_, _ = d.startApproval(ctx, p.TenantID, p.CompanyID, "purchase", p.ID, p.RequesterID)
	return p, nil
}

func (d *Deps) CreatePO(ctx context.Context, po domain.PurchaseOrder) (domain.PurchaseOrder, error) {
	if po.TenantID == uuid.Nil || po.CompanyID == uuid.Nil || po.SupplierID == uuid.Nil || len(po.Lines) == 0 {
		return po, domain.ErrInvalidArgument
	}
	if _, err := d.Suppliers.Get(ctx, po.TenantID, po.SupplierID); err != nil {
		return po, err
	}
	if po.ID == uuid.Nil {
		po.ID = d.newID()
	}
	po.TotalMinor = domain.LinesTotal(po.Lines)
	if po.Currency == "" {
		po.Currency = "TRY"
	}
	po.Status = "open"
	po.CreatedAt = d.now()
	if err := d.Procurement.SavePO(ctx, po); err != nil {
		return po, err
	}
	d.emit(ctx, po.TenantID, po.ID, domain.EventPurchaseOrderCreated, map[string]any{
		"supplierId": po.SupplierID.String(), "totalMinor": po.TotalMinor,
	})
	return po, nil
}

func (d *Deps) ReceiveGoods(ctx context.Context, g domain.GoodsReceipt, warehouseRef string) (domain.GoodsReceipt, error) {
	po, err := d.Procurement.GetPO(ctx, g.TenantID, g.POID)
	if err != nil {
		return g, err
	}
	if g.ID == uuid.Nil {
		g.ID = d.newID()
	}
	if len(g.Lines) == 0 {
		g.Lines = po.Lines
	}
	g.CompanyID = po.CompanyID
	g.ReceivedAt = d.now()
	if err := d.Procurement.SaveGRN(ctx, g); err != nil {
		return g, err
	}
	if d.Inventory != nil {
		for _, l := range g.Lines {
			_ = d.Inventory.Receive(ctx, g.TenantID, l.SKU, l.Qty, warehouseRef)
		}
	}
	po.Status = "received"
	_ = d.Procurement.SavePO(ctx, po)
	d.emit(ctx, g.TenantID, g.ID, domain.EventGoodsReceived, map[string]any{"poId": g.POID.String()})
	return g, nil
}

func (d *Deps) CreateAPInvoice(ctx context.Context, inv domain.APInvoice, taxRateBps int64) (domain.APInvoice, error) {
	if inv.TenantID == uuid.Nil || inv.CompanyID == uuid.Nil || inv.SupplierID == uuid.Nil || inv.InvoiceNumber == "" {
		return inv, domain.ErrInvalidArgument
	}
	if inv.ID == uuid.Nil {
		inv.ID = d.newID()
	}
	if inv.Currency == "" {
		inv.Currency = "TRY"
	}
	if inv.TaxMinor == 0 && taxRateBps > 0 {
		inv.TaxMinor = domain.VATAmount(inv.SubtotalMinor, taxRateBps)
	}
	inv.TotalMinor = inv.SubtotalMinor + inv.TaxMinor
	inv.Status = "draft"
	inv.CreatedAt = d.now()

	if inv.POID != nil {
		po, err := d.Procurement.GetPO(ctx, inv.TenantID, *inv.POID)
		if err != nil {
			return inv, err
		}
		grns, _ := d.Procurement.ListGRNByPO(ctx, inv.TenantID, *inv.POID)
		var grnTotal int64
		for _, g := range grns {
			grnTotal += domain.LinesTotal(g.Lines)
		}
		inv.MatchScore = domain.ThreeWayMatchScore(po.TotalMinor, grnTotal, inv.SubtotalMinor)
		if inv.MatchScore >= 0.98 {
			inv.Status = "matched"
		}
	}
	if err := d.AP.Save(ctx, inv); err != nil {
		return inv, err
	}
	return inv, nil
}

func (d *Deps) ApproveAPInvoice(ctx context.Context, tenantID, invoiceID, approverID uuid.UUID) (domain.APInvoice, error) {
	inv, err := d.AP.Get(ctx, tenantID, invoiceID)
	if err != nil {
		return inv, err
	}
	if inv.Status != "matched" && inv.Status != "draft" {
		return inv, domain.ErrIllegalTransition
	}
	now := d.now()
	inv.Status = "approved"
	inv.ApprovedAt = &now
	if err := d.AP.Save(ctx, inv); err != nil {
		return inv, err
	}
	d.emit(ctx, tenantID, inv.ID, domain.EventInvoiceApproved, map[string]any{
		"approverId": approverID.String(), "totalMinor": inv.TotalMinor,
	})
	return inv, nil
}

func (d *Deps) ScheduleAPPayment(ctx context.Context, tenantID, invoiceID uuid.UUID) (domain.APInvoice, string, error) {
	inv, err := d.AP.Get(ctx, tenantID, invoiceID)
	if err != nil {
		return inv, "", err
	}
	if inv.Status != "approved" {
		return inv, "", domain.ErrIllegalTransition
	}
	ref := ""
	if d.Settlement != nil {
		ref, err = d.Settlement.ScheduleSupplierPayment(ctx, tenantID, inv.SupplierID, inv.TotalMinor, inv.Currency, inv.InvoiceNumber)
		if err != nil {
			return inv, "", err
		}
	}
	inv.Status = "scheduled"
	_ = d.AP.Save(ctx, inv)
	d.emit(ctx, tenantID, inv.ID, domain.EventPaymentScheduled, map[string]any{
		"batchRef": ref, "totalMinor": inv.TotalMinor,
	})
	return inv, ref, nil
}

func (d *Deps) UpsertBudget(ctx context.Context, b domain.Budget) (domain.Budget, error) {
	if b.TenantID == uuid.Nil || b.CompanyID == uuid.Nil || b.Label == "" || len(b.Lines) == 0 {
		return b, domain.ErrInvalidArgument
	}
	if b.ID == uuid.Nil {
		b.ID = d.newID()
	}
	if b.Status == "" {
		b.Status = "draft"
	}
	if b.Currency == "" {
		b.Currency = "TRY"
	}
	b.CreatedAt = d.now()
	return b, d.Budgets.Save(ctx, b)
}

func (d *Deps) ApproveBudget(ctx context.Context, tenantID, budgetID, approverID uuid.UUID) (domain.Budget, error) {
	b, err := d.Budgets.Get(ctx, tenantID, budgetID)
	if err != nil {
		return b, err
	}
	if b.Status != "draft" {
		return b, domain.ErrIllegalTransition
	}
	now := d.now()
	b.Status = "approved"
	b.ApprovedAt = &now
	if err := d.Budgets.Save(ctx, b); err != nil {
		return b, err
	}
	d.emit(ctx, tenantID, b.ID, domain.EventBudgetApproved, map[string]any{"approverId": approverID.String()})
	return b, nil
}

func (d *Deps) CreateAsset(ctx context.Context, a domain.FixedAsset) (domain.FixedAsset, error) {
	if a.TenantID == uuid.Nil || a.CompanyID == uuid.Nil || a.Code == "" || a.CostMinor <= 0 {
		return a, domain.ErrInvalidArgument
	}
	a.Code = domain.NormalizeCode(a.Code)
	if a.ID == uuid.Nil {
		a.ID = d.newID()
	}
	if a.UsefulLifeMonths <= 0 {
		a.UsefulLifeMonths = 36
	}
	a.Status = "active"
	a.AcquiredAt = d.now()
	if a.Currency == "" {
		a.Currency = "TRY"
	}
	if err := d.Assets.Save(ctx, a); err != nil {
		return a, err
	}
	d.emit(ctx, a.TenantID, a.ID, domain.EventAssetCreated, map[string]any{"code": a.Code, "costMinor": a.CostMinor})
	return a, nil
}

func (d *Deps) RunDepreciation(ctx context.Context, tenantID, companyID, periodID uuid.UUID) (int, error) {
	assets, err := d.Assets.List(ctx, tenantID, companyID)
	if err != nil {
		return 0, err
	}
	n := 0
	for _, a := range assets {
		if a.Status != "active" {
			continue
		}
		monthly := domain.StraightLineDepreciation(a.CostMinor, a.UsefulLifeMonths)
		if monthly <= 0 || a.AccumDepMinor+monthly > a.CostMinor {
			continue
		}
		a.AccumDepMinor += monthly
		_ = d.Assets.Save(ctx, a)
		_, _ = d.PostJournal(ctx, domain.JournalEntry{
			TenantID: tenantID, CompanyID: companyID, PeriodID: periodID,
			Memo: "Depreciation " + a.Code, Currency: a.Currency,
			Lines: []domain.JournalLine{
				{AccountCode: "6100", DebitMinor: monthly, Memo: "dep expense"},
				{AccountCode: "1600", CreditMinor: monthly, Memo: "accum dep"},
			},
			IdempotencyKey: "dep-" + a.ID.String() + "-" + periodID.String(),
		})
		n++
	}
	return n, nil
}

func (d *Deps) CalculateTax(ctx context.Context, t domain.TaxReturn, rateBps int64) (domain.TaxReturn, error) {
	if t.TenantID == uuid.Nil || t.CompanyID == uuid.Nil || t.Kind == "" {
		return t, domain.ErrInvalidArgument
	}
	if t.ID == uuid.Nil {
		t.ID = d.newID()
	}
	t.TaxMinor = domain.VATAmount(t.TaxableMinor, rateBps)
	t.Status = "draft"
	t.CreatedAt = d.now()
	if t.Currency == "" {
		t.Currency = "TRY"
	}
	if err := d.Tax.Save(ctx, t); err != nil {
		return t, err
	}
	d.emit(ctx, t.TenantID, t.ID, domain.EventTaxCalculated, map[string]any{
		"kind": t.Kind, "taxMinor": t.TaxMinor,
	})
	return t, nil
}

func (d *Deps) SubmitExpense(ctx context.Context, e domain.ExpenseReport) (domain.ExpenseReport, error) {
	if e.TenantID == uuid.Nil || e.CompanyID == uuid.Nil || e.EmployeeID == uuid.Nil || len(e.Lines) == 0 {
		return e, domain.ErrInvalidArgument
	}
	if e.ID == uuid.Nil {
		e.ID = d.newID()
	}
	var total int64
	for _, l := range e.Lines {
		total += l.AmountMinor
	}
	e.TotalMinor = total
	e.Status = "submitted"
	e.CreatedAt = d.now()
	if e.Currency == "" {
		e.Currency = "TRY"
	}
	if err := d.Expenses.Save(ctx, e); err != nil {
		return e, err
	}
	_, _ = d.startApproval(ctx, e.TenantID, e.CompanyID, "expense", e.ID, e.EmployeeID)
	d.emit(ctx, e.TenantID, e.ID, domain.EventExpenseSubmitted, map[string]any{"totalMinor": e.TotalMinor})
	return e, nil
}

func (d *Deps) startApproval(ctx context.Context, tenantID, companyID uuid.UUID, kind string, subjectID, requesterID uuid.UUID) (domain.ApprovalRequest, error) {
	a := domain.ApprovalRequest{
		ID: d.newID(), TenantID: tenantID, CompanyID: companyID, Kind: kind, SubjectID: subjectID,
		Status: "pending", CreatedAt: d.now(),
		Steps: []domain.ApprovalStep{{Level: 1, ApproverID: requesterID, Status: "pending"}},
	}
	return a, d.Approvals.Save(ctx, a)
}

func (d *Deps) DecideApproval(ctx context.Context, tenantID, approvalID, actorID uuid.UUID, approve bool, note string) (domain.ApprovalRequest, error) {
	a, err := d.Approvals.Get(ctx, tenantID, approvalID)
	if err != nil {
		return a, err
	}
	if a.Status != "pending" {
		return a, domain.ErrIllegalTransition
	}
	now := d.now()
	for i := range a.Steps {
		if a.Steps[i].Status == "pending" {
			if approve {
				a.Steps[i].Status = "approved"
			} else {
				a.Steps[i].Status = "rejected"
			}
			a.Steps[i].ApproverID = actorID
			a.Steps[i].Note = note
			a.Steps[i].DecidedAt = &now
			break
		}
	}
	if approve {
		a.Status = "approved"
	} else {
		a.Status = "rejected"
	}
	a.DecidedAt = &now
	return a, d.Approvals.Save(ctx, a)
}

func (d *Deps) UpsertBank(ctx context.Context, b domain.BankAccount) (domain.BankAccount, error) {
	if b.TenantID == uuid.Nil || b.CompanyID == uuid.Nil || b.Name == "" {
		return b, domain.ErrInvalidArgument
	}
	if b.ID == uuid.Nil {
		b.ID = d.newID()
	}
	if b.Currency == "" {
		b.Currency = "TRY"
	}
	return b, d.Treasury.SaveBank(ctx, b)
}

func (d *Deps) ImportBankTxn(ctx context.Context, t domain.BankTransaction) (domain.BankTransaction, error) {
	if t.TenantID == uuid.Nil || t.BankAccountID == uuid.Nil {
		return t, domain.ErrInvalidArgument
	}
	if t.ID == uuid.Nil {
		t.ID = d.newID()
	}
	if t.BookedAt.IsZero() {
		t.BookedAt = d.now()
	}
	bank, err := d.Treasury.GetBank(ctx, t.TenantID, t.BankAccountID)
	if err != nil {
		return t, err
	}
	bank.BalanceMinor += t.AmountMinor
	_ = d.Treasury.SaveBank(ctx, bank)
	return t, d.Treasury.SaveTxn(ctx, t)
}

func (d *Deps) ReconcileTxn(ctx context.Context, tenantID, bankID, txnID uuid.UUID) (domain.BankTransaction, error) {
	txns, err := d.Treasury.ListTxns(ctx, tenantID, bankID)
	if err != nil {
		return domain.BankTransaction{}, err
	}
	for _, t := range txns {
		if t.ID == txnID {
			t.Reconciled = true
			return t, d.Treasury.SaveTxn(ctx, t)
		}
	}
	return domain.BankTransaction{}, domain.ErrNotFound
}

func (d *Deps) ExportPayroll(ctx context.Context, p domain.PayrollBatch) (domain.PayrollBatch, error) {
	if p.TenantID == uuid.Nil || p.CompanyID == uuid.Nil || p.TotalMinor <= 0 {
		return p, domain.ErrInvalidArgument
	}
	if p.ID == uuid.Nil {
		p.ID = d.newID()
	}
	p.Status = "exported"
	p.ExternalRef = "payroll-" + p.ID.String()
	p.CreatedAt = d.now()
	if p.Currency == "" {
		p.Currency = "TRY"
	}
	return p, d.Payroll.Save(ctx, p)
}

func (d *Deps) CashflowForecast(ctx context.Context, tenantID, companyID uuid.UUID, days int) (map[string]float64, error) {
	if d.AI == nil {
		return map[string]float64{"inflow": 0, "outflow": 0}, nil
	}
	return d.AI.PredictCashflow(ctx, tenantID, companyID, days)
}

func (d *Deps) AdminStats(ctx context.Context, tenantID uuid.UUID) (map[string]any, error) {
	companies, _ := d.Companies.List(ctx, tenantID)
	pending, _ := d.Approvals.ListPending(ctx, tenantID)
	return map[string]any{
		"companies": len(companies), "pendingApprovals": len(pending), "tenantId": tenantID.String(),
	}, nil
}
