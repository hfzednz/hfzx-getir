package postgres

import (
	"context"
	"database/sql"
	"strings"

	"github.com/nexora/platform-ops-service/internal/app/ports"
	"github.com/nexora/platform-ops-service/internal/domain"
)

var _ ports.Registry = (*Registry)(nil)

// Registry persists the super-admin tenant/company directory in PostgreSQL.
type Registry struct{ DB *sql.DB }

func (r *Registry) ListTenants(ctx context.Context) ([]domain.PlatformTenant, []domain.DualControlProposal, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, slug, name, company_id, company_name, isolation_mode, status, region, description, created_at
		FROM platform_tenants ORDER BY created_at DESC`)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	tenants := []domain.PlatformTenant{}
	for rows.Next() {
		t, err := scanTenant(rows)
		if err != nil {
			return nil, nil, err
		}
		tenants = append(tenants, t)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}
	props, err := r.listProposals(ctx)
	if err != nil {
		return nil, nil, err
	}
	return tenants, props, nil
}

func (r *Registry) GetTenant(ctx context.Context, id string) (domain.PlatformTenant, error) {
	return scanTenant(r.DB.QueryRowContext(ctx, `
		SELECT id, slug, name, company_id, company_name, isolation_mode, status, region, description, created_at
		FROM platform_tenants WHERE id=$1`, id))
}

func (r *Registry) SaveTenant(ctx context.Context, t domain.PlatformTenant) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if t.Slug != "" {
		var existing string
		err := tx.QueryRowContext(ctx, `SELECT id FROM platform_tenants WHERE slug=$1 AND id<>$2`, t.Slug, t.ID).Scan(&existing)
		if err == nil {
			return domain.ErrConflict
		}
		if err != nil && err != sql.ErrNoRows {
			return err
		}
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO platform_tenants (
			id, slug, name, company_id, company_name, isolation_mode, status, region, description, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		ON CONFLICT (id) DO UPDATE SET
			slug=EXCLUDED.slug, name=EXCLUDED.name, company_id=EXCLUDED.company_id,
			company_name=EXCLUDED.company_name, isolation_mode=EXCLUDED.isolation_mode,
			status=EXCLUDED.status, region=EXCLUDED.region, description=EXCLUDED.description`,
		t.ID, t.Slug, t.Name, t.CompanyID, t.CompanyName, t.IsolationMode, t.Status, t.Region, t.Description, t.CreatedAt.UTC())
	if err != nil {
		return mapUniqueViolation(err)
	}
	if t.CompanyID != "" {
		var n int
		if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM platform_tenants WHERE company_id=$1`, t.CompanyID).Scan(&n); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `UPDATE platform_companies SET tenant_count=$2 WHERE id=$1`, t.CompanyID, n); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *Registry) ListCompanies(ctx context.Context) ([]domain.PlatformCompany, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, legal_name, trade_name, country_code, status, tenant_count, primary_currency,
			industry, tax_id, vat_number, billing_email, registered_addr, default_locale, time_zone, primary_color, created_at
		FROM platform_companies ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.PlatformCompany{}
	for rows.Next() {
		c, err := scanCompany(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *Registry) GetCompany(ctx context.Context, id string) (domain.PlatformCompany, error) {
	return scanCompany(r.DB.QueryRowContext(ctx, `
		SELECT id, legal_name, trade_name, country_code, status, tenant_count, primary_currency,
			industry, tax_id, vat_number, billing_email, registered_addr, default_locale, time_zone, primary_color, created_at
		FROM platform_companies WHERE id=$1`, id))
}

func (r *Registry) SaveCompany(ctx context.Context, c domain.PlatformCompany) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO platform_companies (
			id, legal_name, trade_name, country_code, status, tenant_count, primary_currency,
			industry, tax_id, vat_number, billing_email, registered_addr, default_locale, time_zone, primary_color, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)
		ON CONFLICT (id) DO UPDATE SET
			legal_name=EXCLUDED.legal_name, trade_name=EXCLUDED.trade_name, country_code=EXCLUDED.country_code,
			status=EXCLUDED.status, tenant_count=EXCLUDED.tenant_count, primary_currency=EXCLUDED.primary_currency,
			industry=EXCLUDED.industry, tax_id=EXCLUDED.tax_id, vat_number=EXCLUDED.vat_number,
			billing_email=EXCLUDED.billing_email, registered_addr=EXCLUDED.registered_addr,
			default_locale=EXCLUDED.default_locale, time_zone=EXCLUDED.time_zone, primary_color=EXCLUDED.primary_color`,
		c.ID, c.LegalName, c.TradeName, c.CountryCode, c.Status, c.TenantCount, c.PrimaryCurrency,
		c.Industry, c.TaxID, c.VATNumber, c.BillingEmail, c.RegisteredAddr, c.DefaultLocale, c.TimeZone, c.PrimaryColor, c.CreatedAt.UTC())
	return mapUniqueViolation(err)
}

func (r *Registry) DeleteCompany(ctx context.Context, id string) error {
	res, err := r.DB.ExecContext(ctx, `DELETE FROM platform_companies WHERE id=$1`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *Registry) SaveProposal(ctx context.Context, p domain.DualControlProposal) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO platform_dual_control (
			id, action, tenant_id, tenant_name, requester_id, requester_email, status, reason, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		ON CONFLICT (id) DO UPDATE SET
			action=EXCLUDED.action, tenant_id=EXCLUDED.tenant_id, tenant_name=EXCLUDED.tenant_name,
			requester_id=EXCLUDED.requester_id, requester_email=EXCLUDED.requester_email,
			status=EXCLUDED.status, reason=EXCLUDED.reason`,
		p.ID, p.Action, p.TenantID, p.TenantName, p.RequesterID, p.RequesterEmail, p.Status, p.Reason, p.CreatedAt.UTC())
	return mapUniqueViolation(err)
}

func (r *Registry) GetProposal(ctx context.Context, id string) (domain.DualControlProposal, error) {
	return scanProposal(r.DB.QueryRowContext(ctx, `
		SELECT id, action, tenant_id, tenant_name, requester_id, requester_email, status, reason, created_at
		FROM platform_dual_control WHERE id=$1`, id))
}

func (r *Registry) listProposals(ctx context.Context) ([]domain.DualControlProposal, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, action, tenant_id, tenant_name, requester_id, requester_email, status, reason, created_at
		FROM platform_dual_control ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.DualControlProposal{}
	for rows.Next() {
		p, err := scanProposal(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *Registry) AppendAudit(ctx context.Context, e domain.PlatformAuditEntry) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO platform_registry_audit (
			id, actor_id, actor_email, action, resource, resource_id, occurred_at, loc, device, ip, session_id,
			old_value, new_value, severity, sealed
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`,
		e.ID, e.ActorID, e.ActorEmail, e.Action, e.Resource, e.ResourceID, e.When.UTC(), e.Where, e.Device, e.IP, e.SessionID,
		nullString(e.OldValue), nullString(e.NewValue), e.Severity, e.Sealed)
	return err
}

func (r *Registry) ListAudit(ctx context.Context, q string) ([]domain.PlatformAuditEntry, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, actor_id, actor_email, action, resource, resource_id, occurred_at, loc, device, ip, session_id,
			old_value, new_value, severity, sealed
		FROM platform_registry_audit ORDER BY occurred_at DESC LIMIT 500`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	q = strings.ToLower(strings.TrimSpace(q))
	out := []domain.PlatformAuditEntry{}
	for rows.Next() {
		e, err := scanAudit(rows)
		if err != nil {
			return nil, err
		}
		if q == "" || strings.Contains(strings.ToLower(e.ActorEmail+e.Action+e.Resource+e.IP+e.SessionID), q) {
			out = append(out, e)
		}
	}
	return out, rows.Err()
}

func (r *Registry) ListPeople(ctx context.Context) ([]domain.PlatformPerson, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, name, email, kind, org_unit_id, org_unit_name, status FROM platform_people ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.PlatformPerson{}
	for rows.Next() {
		var p domain.PlatformPerson
		if err := rows.Scan(&p.ID, &p.Name, &p.Email, &p.Kind, &p.OrgUnitID, &p.OrgUnitName, &p.Status); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *Registry) SavePerson(ctx context.Context, p domain.PlatformPerson) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO platform_people (id, name, email, kind, org_unit_id, org_unit_name, status)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		ON CONFLICT (id) DO UPDATE SET
			name=EXCLUDED.name, email=EXCLUDED.email, kind=EXCLUDED.kind,
			org_unit_id=EXCLUDED.org_unit_id, org_unit_name=EXCLUDED.org_unit_name, status=EXCLUDED.status`,
		p.ID, p.Name, p.Email, p.Kind, p.OrgUnitID, p.OrgUnitName, p.Status)
	return err
}

type scanner interface {
	Scan(dest ...any) error
}

func scanTenant(row scanner) (domain.PlatformTenant, error) {
	var t domain.PlatformTenant
	err := row.Scan(&t.ID, &t.Slug, &t.Name, &t.CompanyID, &t.CompanyName, &t.IsolationMode, &t.Status, &t.Region, &t.Description, &t.CreatedAt)
	if isNoRows(err) {
		return domain.PlatformTenant{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.PlatformTenant{}, err
	}
	t.CreatedAt = t.CreatedAt.UTC()
	return t, nil
}

func scanCompany(row scanner) (domain.PlatformCompany, error) {
	var c domain.PlatformCompany
	err := row.Scan(&c.ID, &c.LegalName, &c.TradeName, &c.CountryCode, &c.Status, &c.TenantCount, &c.PrimaryCurrency,
		&c.Industry, &c.TaxID, &c.VATNumber, &c.BillingEmail, &c.RegisteredAddr, &c.DefaultLocale, &c.TimeZone, &c.PrimaryColor, &c.CreatedAt)
	if isNoRows(err) {
		return domain.PlatformCompany{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.PlatformCompany{}, err
	}
	c.CreatedAt = c.CreatedAt.UTC()
	return c, nil
}

func scanProposal(row scanner) (domain.DualControlProposal, error) {
	var p domain.DualControlProposal
	err := row.Scan(&p.ID, &p.Action, &p.TenantID, &p.TenantName, &p.RequesterID, &p.RequesterEmail, &p.Status, &p.Reason, &p.CreatedAt)
	if isNoRows(err) {
		return domain.DualControlProposal{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.DualControlProposal{}, err
	}
	p.CreatedAt = p.CreatedAt.UTC()
	return p, nil
}

func scanAudit(row scanner) (domain.PlatformAuditEntry, error) {
	var e domain.PlatformAuditEntry
	var old, neu sql.NullString
	err := row.Scan(&e.ID, &e.ActorID, &e.ActorEmail, &e.Action, &e.Resource, &e.ResourceID, &e.When, &e.Where, &e.Device, &e.IP, &e.SessionID,
		&old, &neu, &e.Severity, &e.Sealed)
	if err != nil {
		return domain.PlatformAuditEntry{}, err
	}
	e.When = e.When.UTC()
	if old.Valid {
		v := old.String
		e.OldValue = &v
	}
	if neu.Valid {
		v := neu.String
		e.NewValue = &v
	}
	return e, nil
}

func nullString(v *string) any {
	if v == nil {
		return nil
	}
	return *v
}
