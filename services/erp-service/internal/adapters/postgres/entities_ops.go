package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/erp-service/internal/app/ports"
	"github.com/nexora/erp-service/internal/domain"
)

type ProcurementRepo struct{ DB *sql.DB }

func (r *ProcurementRepo) SavePR(ctx context.Context, p domain.PurchaseRequest) error {
	lines, err := marshalLines(p.Lines)
	if err != nil {
		return err
	}
	_, err = r.DB.ExecContext(ctx, `
		INSERT INTO erp_purchase_requests (
			id, tenant_id, company_id, requester_id, status, currency, total_minor, lines_json, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		ON CONFLICT (id) DO UPDATE SET status=EXCLUDED.status, currency=EXCLUDED.currency,
			total_minor=EXCLUDED.total_minor, lines_json=EXCLUDED.lines_json`,
		p.ID, p.TenantID, p.CompanyID, p.RequesterID, p.Status, p.Currency, p.TotalMinor, lines, p.CreatedAt.UTC())
	return err
}

func (r *ProcurementRepo) GetPR(ctx context.Context, tenantID, id uuid.UUID) (domain.PurchaseRequest, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, company_id, requester_id, status, currency, total_minor, lines_json, created_at
		FROM erp_purchase_requests WHERE id=$1 AND tenant_id=$2`, id, tenantID)
	var p domain.PurchaseRequest
	var raw []byte
	err := row.Scan(&p.ID, &p.TenantID, &p.CompanyID, &p.RequesterID, &p.Status, &p.Currency, &p.TotalMinor, &raw, &p.CreatedAt)
	if err != nil {
		if isNoRows(err) {
			return domain.PurchaseRequest{}, domain.ErrNotFound
		}
		return domain.PurchaseRequest{}, err
	}
	_ = unmarshalLines(raw, &p.Lines)
	p.CreatedAt = p.CreatedAt.UTC()
	return p, nil
}

func (r *ProcurementRepo) SavePO(ctx context.Context, p domain.PurchaseOrder) error {
	lines, err := marshalLines(p.Lines)
	if err != nil {
		return err
	}
	_, err = r.DB.ExecContext(ctx, `
		INSERT INTO erp_purchase_orders (
			id, tenant_id, company_id, supplier_id, pr_id, status, currency, total_minor, lines_json, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		ON CONFLICT (id) DO UPDATE SET status=EXCLUDED.status, currency=EXCLUDED.currency,
			total_minor=EXCLUDED.total_minor, lines_json=EXCLUDED.lines_json, pr_id=EXCLUDED.pr_id`,
		p.ID, p.TenantID, p.CompanyID, p.SupplierID, nullUUID(p.PRID), p.Status, p.Currency, p.TotalMinor, lines, p.CreatedAt.UTC())
	return err
}

func (r *ProcurementRepo) GetPO(ctx context.Context, tenantID, id uuid.UUID) (domain.PurchaseOrder, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, company_id, supplier_id, pr_id, status, currency, total_minor, lines_json, created_at
		FROM erp_purchase_orders WHERE id=$1 AND tenant_id=$2`, id, tenantID)
	var p domain.PurchaseOrder
	var pr uuid.NullUUID
	var raw []byte
	err := row.Scan(&p.ID, &p.TenantID, &p.CompanyID, &p.SupplierID, &pr, &p.Status, &p.Currency, &p.TotalMinor, &raw, &p.CreatedAt)
	if err != nil {
		if isNoRows(err) {
			return domain.PurchaseOrder{}, domain.ErrNotFound
		}
		return domain.PurchaseOrder{}, err
	}
	p.PRID = scanNullUUID(pr)
	_ = unmarshalLines(raw, &p.Lines)
	p.CreatedAt = p.CreatedAt.UTC()
	return p, nil
}

func (r *ProcurementRepo) SaveGRN(ctx context.Context, g domain.GoodsReceipt) error {
	lines, err := marshalLines(g.Lines)
	if err != nil {
		return err
	}
	_, err = r.DB.ExecContext(ctx, `
		INSERT INTO erp_goods_receipts (
			id, tenant_id, company_id, po_id, lines_json, received_at, created_by
		) VALUES ($1,$2,$3,$4,$5,$6,$7)
		ON CONFLICT (id) DO UPDATE SET lines_json=EXCLUDED.lines_json, received_at=EXCLUDED.received_at`,
		g.ID, g.TenantID, g.CompanyID, g.POID, lines, g.ReceivedAt.UTC(), nullUUIDValue(g.CreatedBy))
	return err
}

func (r *ProcurementRepo) ListGRNByPO(ctx context.Context, tenantID, poID uuid.UUID) ([]domain.GoodsReceipt, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, company_id, po_id, lines_json, received_at, COALESCE(created_by,'00000000-0000-0000-0000-000000000000'::uuid)
		FROM erp_goods_receipts WHERE tenant_id=$1 AND po_id=$2 ORDER BY received_at ASC`, tenantID, poID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.GoodsReceipt{}
	for rows.Next() {
		var g domain.GoodsReceipt
		var raw []byte
		if err := rows.Scan(&g.ID, &g.TenantID, &g.CompanyID, &g.POID, &raw, &g.ReceivedAt, &g.CreatedBy); err != nil {
			return nil, err
		}
		_ = unmarshalLines(raw, &g.Lines)
		g.ReceivedAt = g.ReceivedAt.UTC()
		out = append(out, g)
	}
	return out, rows.Err()
}

type APRepo struct{ DB *sql.DB }

func (r *APRepo) Save(ctx context.Context, inv domain.APInvoice) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO erp_ap_invoices (
			id, tenant_id, company_id, supplier_id, po_id, invoice_number, currency,
			subtotal_minor, tax_minor, total_minor, status, match_score, due_date, created_at, approved_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
		ON CONFLICT (id) DO UPDATE SET status=EXCLUDED.status, match_score=EXCLUDED.match_score,
			subtotal_minor=EXCLUDED.subtotal_minor, tax_minor=EXCLUDED.tax_minor, total_minor=EXCLUDED.total_minor,
			approved_at=EXCLUDED.approved_at`,
		inv.ID, inv.TenantID, inv.CompanyID, inv.SupplierID, nullUUID(inv.POID), inv.InvoiceNumber, inv.Currency,
		inv.SubtotalMinor, inv.TaxMinor, inv.TotalMinor, inv.Status, inv.MatchScore, nullDate(inv.DueDate),
		inv.CreatedAt.UTC(), nullTime(inv.ApprovedAt))
	return err
}

func (r *APRepo) Get(ctx context.Context, tenantID, id uuid.UUID) (domain.APInvoice, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, company_id, supplier_id, po_id, invoice_number, currency,
			subtotal_minor, tax_minor, total_minor, status, match_score, due_date, created_at, approved_at
		FROM erp_ap_invoices WHERE id=$1 AND tenant_id=$2`, id, tenantID)
	var inv domain.APInvoice
	var po uuid.NullUUID
	var due, approved sql.NullTime
	err := row.Scan(&inv.ID, &inv.TenantID, &inv.CompanyID, &inv.SupplierID, &po, &inv.InvoiceNumber, &inv.Currency,
		&inv.SubtotalMinor, &inv.TaxMinor, &inv.TotalMinor, &inv.Status, &inv.MatchScore, &due, &inv.CreatedAt, &approved)
	if err != nil {
		if isNoRows(err) {
			return domain.APInvoice{}, domain.ErrNotFound
		}
		return domain.APInvoice{}, err
	}
	inv.POID = scanNullUUID(po)
	inv.DueDate = scanDate(due)
	inv.ApprovedAt = scanNullTime(approved)
	inv.CreatedAt = inv.CreatedAt.UTC()
	return inv, nil
}

func (r *APRepo) List(ctx context.Context, tenantID, companyID uuid.UUID, status string) ([]domain.APInvoice, error) {
	q := `
		SELECT id, tenant_id, company_id, supplier_id, po_id, invoice_number, currency,
			subtotal_minor, tax_minor, total_minor, status, match_score, due_date, created_at, approved_at
		FROM erp_ap_invoices WHERE tenant_id=$1 AND company_id=$2`
	args := []any{tenantID, companyID}
	if status != "" {
		q += ` AND status=$3`
		args = append(args, status)
	}
	q += ` ORDER BY created_at DESC`
	rows, err := r.DB.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.APInvoice{}
	for rows.Next() {
		var inv domain.APInvoice
		var po uuid.NullUUID
		var due, approved sql.NullTime
		if err := rows.Scan(&inv.ID, &inv.TenantID, &inv.CompanyID, &inv.SupplierID, &po, &inv.InvoiceNumber, &inv.Currency,
			&inv.SubtotalMinor, &inv.TaxMinor, &inv.TotalMinor, &inv.Status, &inv.MatchScore, &due, &inv.CreatedAt, &approved); err != nil {
			return nil, err
		}
		inv.POID = scanNullUUID(po)
		inv.DueDate = scanDate(due)
		inv.ApprovedAt = scanNullTime(approved)
		inv.CreatedAt = inv.CreatedAt.UTC()
		out = append(out, inv)
	}
	return out, rows.Err()
}

type ARRepo struct{ DB *sql.DB }

func (r *ARRepo) Save(ctx context.Context, inv domain.ARInvoice) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO erp_ar_invoices (
			id, tenant_id, company_id, customer_ref, invoice_number, currency, total_minor, status, due_date, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		ON CONFLICT (id) DO UPDATE SET status=EXCLUDED.status, total_minor=EXCLUDED.total_minor, due_date=EXCLUDED.due_date`,
		inv.ID, inv.TenantID, inv.CompanyID, inv.CustomerRef, inv.InvoiceNumber, inv.Currency, inv.TotalMinor,
		inv.Status, nullDate(inv.DueDate), inv.CreatedAt.UTC())
	return err
}

func (r *ARRepo) Get(ctx context.Context, tenantID, id uuid.UUID) (domain.ARInvoice, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, company_id, customer_ref, invoice_number, currency, total_minor, status, due_date, created_at
		FROM erp_ar_invoices WHERE id=$1 AND tenant_id=$2`, id, tenantID)
	var inv domain.ARInvoice
	var due sql.NullTime
	err := row.Scan(&inv.ID, &inv.TenantID, &inv.CompanyID, &inv.CustomerRef, &inv.InvoiceNumber, &inv.Currency,
		&inv.TotalMinor, &inv.Status, &due, &inv.CreatedAt)
	if err != nil {
		if isNoRows(err) {
			return domain.ARInvoice{}, domain.ErrNotFound
		}
		return domain.ARInvoice{}, err
	}
	inv.DueDate = scanDate(due)
	inv.CreatedAt = inv.CreatedAt.UTC()
	return inv, nil
}

type TreasuryRepo struct{ DB *sql.DB }

func (r *TreasuryRepo) SaveBank(ctx context.Context, b domain.BankAccount) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO erp_bank_accounts (id, tenant_id, company_id, name, iban, currency, balance_minor)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		ON CONFLICT (id) DO UPDATE SET name=EXCLUDED.name, iban=EXCLUDED.iban, balance_minor=EXCLUDED.balance_minor`,
		b.ID, b.TenantID, b.CompanyID, b.Name, b.IBAN, b.Currency, b.BalanceMinor)
	return err
}

func (r *TreasuryRepo) GetBank(ctx context.Context, tenantID, id uuid.UUID) (domain.BankAccount, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, company_id, name, iban, currency, balance_minor
		FROM erp_bank_accounts WHERE id=$1 AND tenant_id=$2`, id, tenantID)
	var b domain.BankAccount
	err := row.Scan(&b.ID, &b.TenantID, &b.CompanyID, &b.Name, &b.IBAN, &b.Currency, &b.BalanceMinor)
	if err != nil {
		if isNoRows(err) {
			return domain.BankAccount{}, domain.ErrNotFound
		}
		return domain.BankAccount{}, err
	}
	return b, nil
}

func (r *TreasuryRepo) ListBanks(ctx context.Context, tenantID, companyID uuid.UUID) ([]domain.BankAccount, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, company_id, name, iban, currency, balance_minor
		FROM erp_bank_accounts WHERE tenant_id=$1 AND company_id=$2 ORDER BY name ASC`, tenantID, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.BankAccount{}
	for rows.Next() {
		var b domain.BankAccount
		if err := rows.Scan(&b.ID, &b.TenantID, &b.CompanyID, &b.Name, &b.IBAN, &b.Currency, &b.BalanceMinor); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

func (r *TreasuryRepo) SaveTxn(ctx context.Context, t domain.BankTransaction) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO erp_bank_transactions (
			id, tenant_id, bank_account_id, external_ref, amount_minor, currency, booked_at, reconciled, memo
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		ON CONFLICT (id) DO UPDATE SET reconciled=EXCLUDED.reconciled, memo=EXCLUDED.memo`,
		t.ID, t.TenantID, t.BankAccountID, t.ExternalRef, t.AmountMinor, t.Currency, t.BookedAt.UTC(), t.Reconciled, t.Memo)
	return err
}

func (r *TreasuryRepo) ListTxns(ctx context.Context, tenantID, bankID uuid.UUID) ([]domain.BankTransaction, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, bank_account_id, external_ref, amount_minor, currency, booked_at, reconciled, memo
		FROM erp_bank_transactions WHERE tenant_id=$1 AND bank_account_id=$2 ORDER BY booked_at DESC`, tenantID, bankID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.BankTransaction{}
	for rows.Next() {
		var t domain.BankTransaction
		if err := rows.Scan(&t.ID, &t.TenantID, &t.BankAccountID, &t.ExternalRef, &t.AmountMinor, &t.Currency, &t.BookedAt, &t.Reconciled, &t.Memo); err != nil {
			return nil, err
		}
		t.BookedAt = t.BookedAt.UTC()
		out = append(out, t)
	}
	return out, rows.Err()
}

type BudgetRepo struct{ DB *sql.DB }

func (r *BudgetRepo) Save(ctx context.Context, b domain.Budget) error {
	lines, err := marshalLines(b.Lines)
	if err != nil {
		return err
	}
	_, err = r.DB.ExecContext(ctx, `
		INSERT INTO erp_budgets (
			id, tenant_id, company_id, label, period, year, status, currency, lines_json, created_at, approved_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		ON CONFLICT (id) DO UPDATE SET status=EXCLUDED.status, lines_json=EXCLUDED.lines_json, approved_at=EXCLUDED.approved_at`,
		b.ID, b.TenantID, b.CompanyID, b.Label, b.Period, b.Year, b.Status, b.Currency, lines, b.CreatedAt.UTC(), nullTime(b.ApprovedAt))
	return err
}

func (r *BudgetRepo) Get(ctx context.Context, tenantID, id uuid.UUID) (domain.Budget, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, company_id, label, period, year, status, currency, lines_json, created_at, approved_at
		FROM erp_budgets WHERE id=$1 AND tenant_id=$2`, id, tenantID)
	var b domain.Budget
	var raw []byte
	var approved sql.NullTime
	err := row.Scan(&b.ID, &b.TenantID, &b.CompanyID, &b.Label, &b.Period, &b.Year, &b.Status, &b.Currency, &raw, &b.CreatedAt, &approved)
	if err != nil {
		if isNoRows(err) {
			return domain.Budget{}, domain.ErrNotFound
		}
		return domain.Budget{}, err
	}
	_ = unmarshalLines(raw, &b.Lines)
	b.ApprovedAt = scanNullTime(approved)
	b.CreatedAt = b.CreatedAt.UTC()
	return b, nil
}

func (r *BudgetRepo) List(ctx context.Context, tenantID, companyID uuid.UUID) ([]domain.Budget, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, company_id, label, period, year, status, currency, lines_json, created_at, approved_at
		FROM erp_budgets WHERE tenant_id=$1 AND company_id=$2 ORDER BY year DESC, label ASC`, tenantID, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Budget{}
	for rows.Next() {
		var b domain.Budget
		var raw []byte
		var approved sql.NullTime
		if err := rows.Scan(&b.ID, &b.TenantID, &b.CompanyID, &b.Label, &b.Period, &b.Year, &b.Status, &b.Currency, &raw, &b.CreatedAt, &approved); err != nil {
			return nil, err
		}
		_ = unmarshalLines(raw, &b.Lines)
		b.ApprovedAt = scanNullTime(approved)
		b.CreatedAt = b.CreatedAt.UTC()
		out = append(out, b)
	}
	return out, rows.Err()
}

type AssetRepo struct{ DB *sql.DB }

func (r *AssetRepo) Save(ctx context.Context, a domain.FixedAsset) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO erp_fixed_assets (
			id, tenant_id, company_id, code, name, cost_minor, currency, useful_life_months, accum_dep_minor, status, acquired_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		ON CONFLICT (tenant_id, company_id, code) DO UPDATE SET id=EXCLUDED.id, name=EXCLUDED.name,
			cost_minor=EXCLUDED.cost_minor, accum_dep_minor=EXCLUDED.accum_dep_minor, status=EXCLUDED.status`,
		a.ID, a.TenantID, a.CompanyID, a.Code, a.Name, a.CostMinor, a.Currency, a.UsefulLifeMonths,
		a.AccumDepMinor, a.Status, a.AcquiredAt.UTC())
	return err
}

func (r *AssetRepo) Get(ctx context.Context, tenantID, id uuid.UUID) (domain.FixedAsset, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, company_id, code, name, cost_minor, currency, useful_life_months, accum_dep_minor, status, acquired_at
		FROM erp_fixed_assets WHERE id=$1 AND tenant_id=$2`, id, tenantID)
	var a domain.FixedAsset
	err := row.Scan(&a.ID, &a.TenantID, &a.CompanyID, &a.Code, &a.Name, &a.CostMinor, &a.Currency,
		&a.UsefulLifeMonths, &a.AccumDepMinor, &a.Status, &a.AcquiredAt)
	if err != nil {
		if isNoRows(err) {
			return domain.FixedAsset{}, domain.ErrNotFound
		}
		return domain.FixedAsset{}, err
	}
	a.AcquiredAt = a.AcquiredAt.UTC()
	return a, nil
}

func (r *AssetRepo) List(ctx context.Context, tenantID, companyID uuid.UUID) ([]domain.FixedAsset, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, company_id, code, name, cost_minor, currency, useful_life_months, accum_dep_minor, status, acquired_at
		FROM erp_fixed_assets WHERE tenant_id=$1 AND company_id=$2 ORDER BY code ASC`, tenantID, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.FixedAsset{}
	for rows.Next() {
		var a domain.FixedAsset
		if err := rows.Scan(&a.ID, &a.TenantID, &a.CompanyID, &a.Code, &a.Name, &a.CostMinor, &a.Currency,
			&a.UsefulLifeMonths, &a.AccumDepMinor, &a.Status, &a.AcquiredAt); err != nil {
			return nil, err
		}
		a.AcquiredAt = a.AcquiredAt.UTC()
		out = append(out, a)
	}
	return out, rows.Err()
}

type TaxRepo struct{ DB *sql.DB }

func (r *TaxRepo) Save(ctx context.Context, t domain.TaxReturn) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO erp_tax_returns (
			id, tenant_id, company_id, kind, period_label, currency, taxable_minor, tax_minor, status, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		ON CONFLICT (id) DO UPDATE SET status=EXCLUDED.status, taxable_minor=EXCLUDED.taxable_minor, tax_minor=EXCLUDED.tax_minor`,
		t.ID, t.TenantID, t.CompanyID, t.Kind, t.PeriodLabel, t.Currency, t.TaxableMinor, t.TaxMinor, t.Status, t.CreatedAt.UTC())
	return err
}

func (r *TaxRepo) List(ctx context.Context, tenantID, companyID uuid.UUID) ([]domain.TaxReturn, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, company_id, kind, period_label, currency, taxable_minor, tax_minor, status, created_at
		FROM erp_tax_returns WHERE tenant_id=$1 AND company_id=$2 ORDER BY created_at DESC`, tenantID, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.TaxReturn{}
	for rows.Next() {
		var t domain.TaxReturn
		if err := rows.Scan(&t.ID, &t.TenantID, &t.CompanyID, &t.Kind, &t.PeriodLabel, &t.Currency, &t.TaxableMinor, &t.TaxMinor, &t.Status, &t.CreatedAt); err != nil {
			return nil, err
		}
		t.CreatedAt = t.CreatedAt.UTC()
		out = append(out, t)
	}
	return out, rows.Err()
}

type ExpenseRepo struct{ DB *sql.DB }

func (r *ExpenseRepo) Save(ctx context.Context, e domain.ExpenseReport) error {
	lines, err := marshalLines(e.Lines)
	if err != nil {
		return err
	}
	_, err = r.DB.ExecContext(ctx, `
		INSERT INTO erp_expense_reports (
			id, tenant_id, company_id, employee_id, currency, total_minor, status, lines_json, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		ON CONFLICT (id) DO UPDATE SET status=EXCLUDED.status, total_minor=EXCLUDED.total_minor, lines_json=EXCLUDED.lines_json`,
		e.ID, e.TenantID, e.CompanyID, e.EmployeeID, e.Currency, e.TotalMinor, e.Status, lines, e.CreatedAt.UTC())
	return err
}

func (r *ExpenseRepo) Get(ctx context.Context, tenantID, id uuid.UUID) (domain.ExpenseReport, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, company_id, employee_id, currency, total_minor, status, lines_json, created_at
		FROM erp_expense_reports WHERE id=$1 AND tenant_id=$2`, id, tenantID)
	var e domain.ExpenseReport
	var raw []byte
	err := row.Scan(&e.ID, &e.TenantID, &e.CompanyID, &e.EmployeeID, &e.Currency, &e.TotalMinor, &e.Status, &raw, &e.CreatedAt)
	if err != nil {
		if isNoRows(err) {
			return domain.ExpenseReport{}, domain.ErrNotFound
		}
		return domain.ExpenseReport{}, err
	}
	_ = unmarshalLines(raw, &e.Lines)
	e.CreatedAt = e.CreatedAt.UTC()
	return e, nil
}

type ApprovalRepo struct{ DB *sql.DB }

func (r *ApprovalRepo) Save(ctx context.Context, a domain.ApprovalRequest) error {
	steps, err := marshalLines(a.Steps)
	if err != nil {
		return err
	}
	_, err = r.DB.ExecContext(ctx, `
		INSERT INTO erp_approvals (
			id, tenant_id, company_id, kind, subject_id, status, steps_json, created_at, decided_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		ON CONFLICT (id) DO UPDATE SET status=EXCLUDED.status, steps_json=EXCLUDED.steps_json, decided_at=EXCLUDED.decided_at`,
		a.ID, a.TenantID, a.CompanyID, a.Kind, a.SubjectID, a.Status, steps, a.CreatedAt.UTC(), nullTime(a.DecidedAt))
	return err
}

func (r *ApprovalRepo) Get(ctx context.Context, tenantID, id uuid.UUID) (domain.ApprovalRequest, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, company_id, kind, subject_id, status, steps_json, created_at, decided_at
		FROM erp_approvals WHERE id=$1 AND tenant_id=$2`, id, tenantID)
	var a domain.ApprovalRequest
	var raw []byte
	var decided sql.NullTime
	err := row.Scan(&a.ID, &a.TenantID, &a.CompanyID, &a.Kind, &a.SubjectID, &a.Status, &raw, &a.CreatedAt, &decided)
	if err != nil {
		if isNoRows(err) {
			return domain.ApprovalRequest{}, domain.ErrNotFound
		}
		return domain.ApprovalRequest{}, err
	}
	_ = unmarshalLines(raw, &a.Steps)
	a.DecidedAt = scanNullTime(decided)
	a.CreatedAt = a.CreatedAt.UTC()
	return a, nil
}

func (r *ApprovalRepo) ListPending(ctx context.Context, tenantID uuid.UUID) ([]domain.ApprovalRequest, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, company_id, kind, subject_id, status, steps_json, created_at, decided_at
		FROM erp_approvals WHERE tenant_id=$1 AND status='pending' ORDER BY created_at ASC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.ApprovalRequest{}
	for rows.Next() {
		var a domain.ApprovalRequest
		var raw []byte
		var decided sql.NullTime
		if err := rows.Scan(&a.ID, &a.TenantID, &a.CompanyID, &a.Kind, &a.SubjectID, &a.Status, &raw, &a.CreatedAt, &decided); err != nil {
			return nil, err
		}
		_ = unmarshalLines(raw, &a.Steps)
		a.DecidedAt = scanNullTime(decided)
		a.CreatedAt = a.CreatedAt.UTC()
		out = append(out, a)
	}
	return out, rows.Err()
}

type PayrollRepo struct{ DB *sql.DB }

func (r *PayrollRepo) Save(ctx context.Context, p domain.PayrollBatch) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO erp_payroll_batches (
			id, tenant_id, company_id, label, currency, total_minor, status, external_ref, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		ON CONFLICT (id) DO UPDATE SET status=EXCLUDED.status, total_minor=EXCLUDED.total_minor, external_ref=EXCLUDED.external_ref`,
		p.ID, p.TenantID, p.CompanyID, p.Label, p.Currency, p.TotalMinor, p.Status, p.ExternalRef, p.CreatedAt.UTC())
	return err
}

func (r *PayrollRepo) List(ctx context.Context, tenantID, companyID uuid.UUID) ([]domain.PayrollBatch, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, company_id, label, currency, total_minor, status, external_ref, created_at
		FROM erp_payroll_batches WHERE tenant_id=$1 AND company_id=$2 ORDER BY created_at DESC`, tenantID, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.PayrollBatch{}
	for rows.Next() {
		var p domain.PayrollBatch
		if err := rows.Scan(&p.ID, &p.TenantID, &p.CompanyID, &p.Label, &p.Currency, &p.TotalMinor, &p.Status, &p.ExternalRef, &p.CreatedAt); err != nil {
			return nil, err
		}
		p.CreatedAt = p.CreatedAt.UTC()
		out = append(out, p)
	}
	return out, rows.Err()
}

var (
	_ ports.ProcurementRepo = (*ProcurementRepo)(nil)
	_ ports.APRepo          = (*APRepo)(nil)
	_ ports.ARRepo          = (*ARRepo)(nil)
	_ ports.TreasuryRepo    = (*TreasuryRepo)(nil)
	_ ports.BudgetRepo      = (*BudgetRepo)(nil)
	_ ports.AssetRepo       = (*AssetRepo)(nil)
	_ ports.TaxRepo         = (*TaxRepo)(nil)
	_ ports.ExpenseRepo     = (*ExpenseRepo)(nil)
	_ ports.ApprovalRepo    = (*ApprovalRepo)(nil)
	_ ports.PayrollRepo     = (*PayrollRepo)(nil)
)
