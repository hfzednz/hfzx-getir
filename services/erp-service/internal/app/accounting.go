package app

import (
	"context"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/erp-service/internal/domain"
)

func (d *Deps) UpsertCompany(ctx context.Context, c domain.Company) (domain.Company, error) {
	if c.TenantID == uuid.Nil || c.Code == "" || c.Name == "" || c.Currency == "" {
		return c, domain.ErrInvalidArgument
	}
	c.Code = domain.NormalizeCode(c.Code)
	if c.ID == uuid.Nil {
		c.ID = d.newID()
	}
	c.Active = true
	c.CreatedAt = d.now()
	return c, d.Companies.Save(ctx, c)
}

func (d *Deps) OpenFiscalYear(ctx context.Context, y domain.FiscalYear, months int) (domain.FiscalYear, []domain.AccountingPeriod, error) {
	var periods []domain.AccountingPeriod
	if y.TenantID == uuid.Nil || y.CompanyID == uuid.Nil || y.Label == "" {
		return y, nil, domain.ErrInvalidArgument
	}
	if months <= 0 {
		months = 12
	}
	if y.ID == uuid.Nil {
		y.ID = d.newID()
	}
	if y.StartDate.IsZero() {
		y.StartDate = time.Date(time.Now().Year(), 1, 1, 0, 0, 0, 0, time.UTC)
	}
	if y.EndDate.IsZero() {
		y.EndDate = y.StartDate.AddDate(1, 0, -1)
	}
	if err := d.Periods.SaveYear(ctx, y); err != nil {
		return y, nil, err
	}
	start := y.StartDate
	for i := 0; i < months; i++ {
		next := start.AddDate(0, 1, 0)
		end := next.Add(-time.Nanosecond)
		p := domain.AccountingPeriod{
			ID: d.newID(), TenantID: y.TenantID, CompanyID: y.CompanyID, FiscalYearID: y.ID,
			Label: y.Label + "-" + pad2(i+1), StartDate: start, EndDate: end, Status: "open",
		}
		_ = d.Periods.SavePeriod(ctx, p)
		periods = append(periods, p)
		start = next
	}
	return y, periods, nil
}

func pad2(n int) string {
	if n < 10 {
		return "0" + strconv.Itoa(n)
	}
	return strconv.Itoa(n)
}

func (d *Deps) ClosePeriod(ctx context.Context, tenantID, periodID uuid.UUID) (domain.AccountingPeriod, error) {
	p, err := d.Periods.GetPeriod(ctx, tenantID, periodID)
	if err != nil {
		return p, err
	}
	p.Status = "closed"
	return p, d.Periods.SavePeriod(ctx, p)
}

func (d *Deps) UpsertAccount(ctx context.Context, a domain.ChartAccount) (domain.ChartAccount, error) {
	if a.TenantID == uuid.Nil || a.CompanyID == uuid.Nil || a.Code == "" || !domain.ValidAccountType(a.Type) {
		return a, domain.ErrInvalidArgument
	}
	a.Code = domain.NormalizeCode(a.Code)
	if a.ID == uuid.Nil {
		a.ID = d.newID()
	}
	a.Active = true
	return a, d.Accounts.Save(ctx, a)
}

func (d *Deps) PostJournal(ctx context.Context, j domain.JournalEntry) (domain.JournalEntry, error) {
	if j.TenantID == uuid.Nil || j.CompanyID == uuid.Nil || j.PeriodID == uuid.Nil {
		return j, domain.ErrInvalidArgument
	}
	if j.IdempotencyKey != "" {
		if existing, ok, err := d.Journals.GetByIdempotency(ctx, j.TenantID, j.IdempotencyKey); err != nil {
			return j, err
		} else if ok {
			return existing, nil
		}
	}
	period, err := d.Periods.GetPeriod(ctx, j.TenantID, j.PeriodID)
	if err != nil {
		return j, err
	}
	if period.Status == "closed" {
		return j, domain.ErrPeriodClosed
	}
	if err := j.ValidateBalance(); err != nil {
		return j, err
	}
	for i := range j.Lines {
		j.Lines[i].AccountCode = domain.NormalizeCode(j.Lines[i].AccountCode)
		if _, err := d.Accounts.GetByCode(ctx, j.TenantID, j.CompanyID, j.Lines[i].AccountCode); err != nil {
			return j, domain.ErrInvalidArgument
		}
	}
	if j.ID == uuid.Nil {
		j.ID = d.newID()
	}
	j.Status = "posted"
	now := d.now()
	j.CreatedAt = now
	j.PostedAt = &now
	if j.Currency == "" {
		j.Currency = "TRY"
	}
	if d.Ledger != nil {
		ref, err := d.Ledger.PostJournal(ctx, j.TenantID, j.CompanyID, j.Memo, j.Currency, j.Lines, j.IdempotencyKey)
		if err != nil {
			return j, err
		}
		j.LedgerRef = ref
	}
	if err := d.Journals.Save(ctx, j); err != nil {
		return j, err
	}
	d.emit(ctx, j.TenantID, j.ID, domain.EventJournalCreated, map[string]any{
		"ledgerRef": j.LedgerRef, "companyId": j.CompanyID.String(),
	})
	return j, nil
}
