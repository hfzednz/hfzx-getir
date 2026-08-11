package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/erp-service/internal/app/ports"
	"github.com/nexora/erp-service/internal/domain"
)

type scannable interface{ Scan(dest ...any) error }

type CompanyRepo struct{ DB *sql.DB }

func (r *CompanyRepo) Save(ctx context.Context, c domain.Company) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO erp_companies (id, tenant_id, code, name, country, currency, active, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (tenant_id, code) DO UPDATE SET id=EXCLUDED.id, name=EXCLUDED.name, country=EXCLUDED.country,
			currency=EXCLUDED.currency, active=EXCLUDED.active`,
		c.ID, c.TenantID, c.Code, c.Name, c.Country, c.Currency, c.Active, c.CreatedAt.UTC())
	return err
}

func (r *CompanyRepo) Get(ctx context.Context, tenantID, id uuid.UUID) (domain.Company, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, code, name, country, currency, active, created_at
		FROM erp_companies WHERE id=$1 AND tenant_id=$2`, id, tenantID)
	var c domain.Company
	err := row.Scan(&c.ID, &c.TenantID, &c.Code, &c.Name, &c.Country, &c.Currency, &c.Active, &c.CreatedAt)
	if err != nil {
		if isNoRows(err) {
			return domain.Company{}, domain.ErrNotFound
		}
		return domain.Company{}, err
	}
	c.CreatedAt = c.CreatedAt.UTC()
	return c, nil
}

func (r *CompanyRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.Company, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, code, name, country, currency, active, created_at
		FROM erp_companies WHERE tenant_id=$1 ORDER BY code ASC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Company{}
	for rows.Next() {
		var c domain.Company
		if err := rows.Scan(&c.ID, &c.TenantID, &c.Code, &c.Name, &c.Country, &c.Currency, &c.Active, &c.CreatedAt); err != nil {
			return nil, err
		}
		c.CreatedAt = c.CreatedAt.UTC()
		out = append(out, c)
	}
	return out, rows.Err()
}

type PeriodRepo struct{ DB *sql.DB }

func (r *PeriodRepo) SaveYear(ctx context.Context, y domain.FiscalYear) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO erp_fiscal_years (id, tenant_id, company_id, label, start_date, end_date, closed)
		VALUES ($1,$2,$3,$4,$5::date,$6::date,$7)
		ON CONFLICT (id) DO UPDATE SET label=EXCLUDED.label, start_date=EXCLUDED.start_date,
			end_date=EXCLUDED.end_date, closed=EXCLUDED.closed`,
		y.ID, y.TenantID, y.CompanyID, y.Label, dateOnly(y.StartDate), dateOnly(y.EndDate), y.Closed)
	return err
}

func (r *PeriodRepo) SavePeriod(ctx context.Context, p domain.AccountingPeriod) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO erp_accounting_periods (id, tenant_id, company_id, fiscal_year_id, label, start_date, end_date, status)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (id) DO UPDATE SET label=EXCLUDED.label, start_date=EXCLUDED.start_date,
			end_date=EXCLUDED.end_date, status=EXCLUDED.status`,
		p.ID, p.TenantID, p.CompanyID, p.FiscalYearID, p.Label, p.StartDate.UTC(), p.EndDate.UTC(), p.Status)
	return err
}

func (r *PeriodRepo) GetPeriod(ctx context.Context, tenantID, id uuid.UUID) (domain.AccountingPeriod, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, company_id, fiscal_year_id, label, start_date, end_date, status
		FROM erp_accounting_periods WHERE id=$1 AND tenant_id=$2`, id, tenantID)
	var p domain.AccountingPeriod
	err := row.Scan(&p.ID, &p.TenantID, &p.CompanyID, &p.FiscalYearID, &p.Label, &p.StartDate, &p.EndDate, &p.Status)
	if err != nil {
		if isNoRows(err) {
			return domain.AccountingPeriod{}, domain.ErrNotFound
		}
		return domain.AccountingPeriod{}, err
	}
	p.StartDate = p.StartDate.UTC()
	p.EndDate = p.EndDate.UTC()
	return p, nil
}

func (r *PeriodRepo) ListPeriods(ctx context.Context, tenantID, companyID uuid.UUID) ([]domain.AccountingPeriod, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, company_id, fiscal_year_id, label, start_date, end_date, status
		FROM erp_accounting_periods WHERE tenant_id=$1 AND company_id=$2 ORDER BY start_date ASC`, tenantID, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.AccountingPeriod{}
	for rows.Next() {
		var p domain.AccountingPeriod
		if err := rows.Scan(&p.ID, &p.TenantID, &p.CompanyID, &p.FiscalYearID, &p.Label, &p.StartDate, &p.EndDate, &p.Status); err != nil {
			return nil, err
		}
		p.StartDate = p.StartDate.UTC()
		p.EndDate = p.EndDate.UTC()
		out = append(out, p)
	}
	return out, rows.Err()
}

type AccountRepo struct{ DB *sql.DB }

func (r *AccountRepo) Save(ctx context.Context, a domain.ChartAccount) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO erp_chart_accounts (id, tenant_id, company_id, code, name, account_type, parent_id, active)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (tenant_id, company_id, code) DO UPDATE SET id=EXCLUDED.id, name=EXCLUDED.name,
			account_type=EXCLUDED.account_type, parent_id=EXCLUDED.parent_id, active=EXCLUDED.active`,
		a.ID, a.TenantID, a.CompanyID, a.Code, a.Name, a.Type, nullUUID(a.ParentID), a.Active)
	return err
}

func (r *AccountRepo) GetByCode(ctx context.Context, tenantID, companyID uuid.UUID, code string) (domain.ChartAccount, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, company_id, code, name, account_type, parent_id, active
		FROM erp_chart_accounts WHERE tenant_id=$1 AND company_id=$2 AND code=$3`, tenantID, companyID, domain.NormalizeCode(code))
	var a domain.ChartAccount
	var parent uuid.NullUUID
	err := row.Scan(&a.ID, &a.TenantID, &a.CompanyID, &a.Code, &a.Name, &a.Type, &parent, &a.Active)
	if err != nil {
		if isNoRows(err) {
			return domain.ChartAccount{}, domain.ErrNotFound
		}
		return domain.ChartAccount{}, err
	}
	a.ParentID = scanNullUUID(parent)
	return a, nil
}

func (r *AccountRepo) List(ctx context.Context, tenantID, companyID uuid.UUID) ([]domain.ChartAccount, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, company_id, code, name, account_type, parent_id, active
		FROM erp_chart_accounts WHERE tenant_id=$1 AND company_id=$2 ORDER BY code ASC`, tenantID, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.ChartAccount{}
	for rows.Next() {
		var a domain.ChartAccount
		var parent uuid.NullUUID
		if err := rows.Scan(&a.ID, &a.TenantID, &a.CompanyID, &a.Code, &a.Name, &a.Type, &parent, &a.Active); err != nil {
			return nil, err
		}
		a.ParentID = scanNullUUID(parent)
		out = append(out, a)
	}
	return out, rows.Err()
}

type JournalRepo struct{ DB *sql.DB }

func (r *JournalRepo) Save(ctx context.Context, j domain.JournalEntry) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	_, err = tx.ExecContext(ctx, `
		INSERT INTO erp_journal_entries (
			id, tenant_id, company_id, period_id, memo, currency, status, ledger_ref, idempotency_key, created_by, created_at, posted_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		ON CONFLICT (id) DO UPDATE SET memo=EXCLUDED.memo, currency=EXCLUDED.currency, status=EXCLUDED.status,
			ledger_ref=EXCLUDED.ledger_ref, posted_at=EXCLUDED.posted_at`,
		j.ID, j.TenantID, j.CompanyID, j.PeriodID, j.Memo, j.Currency, j.Status, j.LedgerRef,
		nullString(j.IdempotencyKey), nullUUIDValue(j.CreatedBy), j.CreatedAt.UTC(), nullTime(j.PostedAt))
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM erp_journal_lines WHERE journal_id=$1`, j.ID); err != nil {
		return err
	}
	for _, line := range j.Lines {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO erp_journal_lines (journal_id, account_code, cost_center, debit_minor, credit_minor, memo)
			VALUES ($1,$2,$3,$4,$5,$6)`,
			j.ID, line.AccountCode, line.CostCenter, line.DebitMinor, line.CreditMinor, line.Memo); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *JournalRepo) Get(ctx context.Context, tenantID, id uuid.UUID) (domain.JournalEntry, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, company_id, period_id, memo, currency, status, ledger_ref, COALESCE(idempotency_key,''),
			COALESCE(created_by,'00000000-0000-0000-0000-000000000000'::uuid), created_at, posted_at
		FROM erp_journal_entries WHERE id=$1 AND tenant_id=$2`, id, tenantID)
	j, err := scanJournalHeader(row)
	if err != nil {
		return domain.JournalEntry{}, err
	}
	lines, err := r.loadLines(ctx, j.ID)
	if err != nil {
		return domain.JournalEntry{}, err
	}
	j.Lines = lines
	return j, nil
}

func (r *JournalRepo) GetByIdempotency(ctx context.Context, tenantID uuid.UUID, key string) (domain.JournalEntry, bool, error) {
	if key == "" {
		return domain.JournalEntry{}, false, nil
	}
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, company_id, period_id, memo, currency, status, ledger_ref, COALESCE(idempotency_key,''),
			COALESCE(created_by,'00000000-0000-0000-0000-000000000000'::uuid), created_at, posted_at
		FROM erp_journal_entries WHERE tenant_id=$1 AND idempotency_key=$2`, tenantID, key)
	j, err := scanJournalHeader(row)
	if err != nil {
		if err == domain.ErrNotFound {
			return domain.JournalEntry{}, false, nil
		}
		return domain.JournalEntry{}, false, err
	}
	lines, err := r.loadLines(ctx, j.ID)
	if err != nil {
		return domain.JournalEntry{}, false, err
	}
	j.Lines = lines
	return j, true, nil
}

func (r *JournalRepo) loadLines(ctx context.Context, journalID uuid.UUID) ([]domain.JournalLine, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT account_code, cost_center, debit_minor, credit_minor, memo
		FROM erp_journal_lines WHERE journal_id=$1 ORDER BY id ASC`, journalID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.JournalLine{}
	for rows.Next() {
		var l domain.JournalLine
		if err := rows.Scan(&l.AccountCode, &l.CostCenter, &l.DebitMinor, &l.CreditMinor, &l.Memo); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

func scanJournalHeader(row scannable) (domain.JournalEntry, error) {
	var j domain.JournalEntry
	var posted sql.NullTime
	err := row.Scan(&j.ID, &j.TenantID, &j.CompanyID, &j.PeriodID, &j.Memo, &j.Currency, &j.Status, &j.LedgerRef,
		&j.IdempotencyKey, &j.CreatedBy, &j.CreatedAt, &posted)
	if err != nil {
		if isNoRows(err) {
			return domain.JournalEntry{}, domain.ErrNotFound
		}
		return domain.JournalEntry{}, err
	}
	j.PostedAt = scanNullTime(posted)
	j.CreatedAt = j.CreatedAt.UTC()
	return j, nil
}

type SupplierRepo struct{ DB *sql.DB }

func (r *SupplierRepo) Save(ctx context.Context, s domain.Supplier) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO erp_suppliers (id, tenant_id, company_id, code, name, tax_id, country, currency, risk_score, status, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		ON CONFLICT (tenant_id, company_id, code) DO UPDATE SET id=EXCLUDED.id, name=EXCLUDED.name, tax_id=EXCLUDED.tax_id,
			country=EXCLUDED.country, currency=EXCLUDED.currency, risk_score=EXCLUDED.risk_score, status=EXCLUDED.status`,
		s.ID, s.TenantID, s.CompanyID, s.Code, s.Name, s.TaxID, s.Country, s.Currency, s.RiskScore, s.Status, s.CreatedAt.UTC())
	return err
}

func (r *SupplierRepo) Get(ctx context.Context, tenantID, id uuid.UUID) (domain.Supplier, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, company_id, code, name, tax_id, country, currency, risk_score, status, created_at
		FROM erp_suppliers WHERE id=$1 AND tenant_id=$2`, id, tenantID)
	var s domain.Supplier
	err := row.Scan(&s.ID, &s.TenantID, &s.CompanyID, &s.Code, &s.Name, &s.TaxID, &s.Country, &s.Currency, &s.RiskScore, &s.Status, &s.CreatedAt)
	if err != nil {
		if isNoRows(err) {
			return domain.Supplier{}, domain.ErrNotFound
		}
		return domain.Supplier{}, err
	}
	s.CreatedAt = s.CreatedAt.UTC()
	return s, nil
}

func (r *SupplierRepo) List(ctx context.Context, tenantID, companyID uuid.UUID) ([]domain.Supplier, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, company_id, code, name, tax_id, country, currency, risk_score, status, created_at
		FROM erp_suppliers WHERE tenant_id=$1 AND company_id=$2 ORDER BY code ASC`, tenantID, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Supplier{}
	for rows.Next() {
		var s domain.Supplier
		if err := rows.Scan(&s.ID, &s.TenantID, &s.CompanyID, &s.Code, &s.Name, &s.TaxID, &s.Country, &s.Currency, &s.RiskScore, &s.Status, &s.CreatedAt); err != nil {
			return nil, err
		}
		s.CreatedAt = s.CreatedAt.UTC()
		out = append(out, s)
	}
	return out, rows.Err()
}

var (
	_ ports.CompanyRepo  = (*CompanyRepo)(nil)
	_ ports.PeriodRepo   = (*PeriodRepo)(nil)
	_ ports.AccountRepo  = (*AccountRepo)(nil)
	_ ports.JournalRepo  = (*JournalRepo)(nil)
	_ ports.SupplierRepo = (*SupplierRepo)(nil)
)
