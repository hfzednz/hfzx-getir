package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/nexora/supplier-service/internal/app/ports"
	"github.com/nexora/supplier-service/internal/domain"
)

type SupplierRepo struct{ DB *sql.DB }

func (r *SupplierRepo) Save(ctx context.Context, s domain.Supplier) error {
	contacts, err := marshalJSON(s.Contacts)
	if err != nil {
		return err
	}
	_, err = r.DB.ExecContext(ctx, `
		INSERT INTO supplier_masters (
			id, tenant_id, company_id, code, legal_name, display_name, country, tax_id, status,
			partner_kinds, contacts_json, banking_ref, erp_supplier_id, risk_score,
			created_at, updated_at, approved_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)
		ON CONFLICT (tenant_id, code) DO UPDATE SET
			id=EXCLUDED.id, company_id=EXCLUDED.company_id, legal_name=EXCLUDED.legal_name,
			display_name=EXCLUDED.display_name, country=EXCLUDED.country, tax_id=EXCLUDED.tax_id,
			status=EXCLUDED.status, partner_kinds=EXCLUDED.partner_kinds, contacts_json=EXCLUDED.contacts_json,
			banking_ref=EXCLUDED.banking_ref, erp_supplier_id=EXCLUDED.erp_supplier_id,
			risk_score=EXCLUDED.risk_score, updated_at=EXCLUDED.updated_at, approved_at=EXCLUDED.approved_at`,
		s.ID, s.TenantID, s.CompanyID, s.Code, s.LegalName, s.DisplayName, s.Country, s.TaxID, string(s.Status),
		textArray(partnerKindsToStrings(s.PartnerKinds)), contacts, s.BankingRef, s.ErpSupplierID, s.RiskScore,
		s.CreatedAt.UTC(), s.UpdatedAt.UTC(), nullTime(s.ApprovedAt))
	return err
}

func (r *SupplierRepo) Get(ctx context.Context, tenantID, id uuid.UUID) (domain.Supplier, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, company_id, code, legal_name, display_name, country, tax_id, status,
			partner_kinds, contacts_json, banking_ref, erp_supplier_id, risk_score,
			created_at, updated_at, approved_at
		FROM supplier_masters WHERE id=$1 AND tenant_id=$2`, id, tenantID)
	return scanSupplier(row)
}

func (r *SupplierRepo) GetByCode(ctx context.Context, tenantID uuid.UUID, code string) (domain.Supplier, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, company_id, code, legal_name, display_name, country, tax_id, status,
			partner_kinds, contacts_json, banking_ref, erp_supplier_id, risk_score,
			created_at, updated_at, approved_at
		FROM supplier_masters WHERE tenant_id=$1 AND code=$2`, tenantID, code)
	return scanSupplier(row)
}

func (r *SupplierRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.Supplier, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, company_id, code, legal_name, display_name, country, tax_id, status,
			partner_kinds, contacts_json, banking_ref, erp_supplier_id, risk_score,
			created_at, updated_at, approved_at
		FROM supplier_masters WHERE tenant_id=$1 ORDER BY code ASC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Supplier{}
	for rows.Next() {
		s, err := scanSupplier(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

type scannable interface{ Scan(dest ...any) error }

func scanSupplier(row scannable) (domain.Supplier, error) {
	var s domain.Supplier
	var status string
	var kinds []string
	var contactsRaw []byte
	var approved sql.NullTime
	err := row.Scan(
		&s.ID, &s.TenantID, &s.CompanyID, &s.Code, &s.LegalName, &s.DisplayName, &s.Country, &s.TaxID, &status,
		pq.Array(&kinds), &contactsRaw, &s.BankingRef, &s.ErpSupplierID, &s.RiskScore,
		&s.CreatedAt, &s.UpdatedAt, &approved)
	if err != nil {
		if isNoRows(err) {
			return domain.Supplier{}, domain.ErrNotFound
		}
		return domain.Supplier{}, err
	}
	s.Status = domain.SupplierStatus(status)
	s.PartnerKinds = stringsToPartnerKinds(kinds)
	_ = unmarshalJSON(contactsRaw, &s.Contacts)
	s.ApprovedAt = scanNullTime(approved)
	s.CreatedAt = s.CreatedAt.UTC()
	s.UpdatedAt = s.UpdatedAt.UTC()
	return s, nil
}

type DocumentRepo struct{ DB *sql.DB }

func (r *DocumentRepo) Save(ctx context.Context, d domain.SupplierDocument) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO supplier_documents (
			id, tenant_id, supplier_id, kind, name, uri, expires_at, verified, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		ON CONFLICT (id) DO UPDATE SET name=EXCLUDED.name, uri=EXCLUDED.uri, expires_at=EXCLUDED.expires_at, verified=EXCLUDED.verified`,
		d.ID, d.TenantID, d.SupplierID, string(d.Kind), d.Name, d.URI, nullTime(d.ExpiresAt), d.Verified, d.CreatedAt.UTC())
	return err
}

func (r *DocumentRepo) ListBySupplier(ctx context.Context, tenantID, supplierID uuid.UUID) ([]domain.SupplierDocument, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, supplier_id, kind, name, uri, expires_at, verified, created_at
		FROM supplier_documents WHERE tenant_id=$1 AND supplier_id=$2 ORDER BY created_at DESC`, tenantID, supplierID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.SupplierDocument{}
	for rows.Next() {
		var d domain.SupplierDocument
		var kind string
		var expires sql.NullTime
		if err := rows.Scan(&d.ID, &d.TenantID, &d.SupplierID, &kind, &d.Name, &d.URI, &expires, &d.Verified, &d.CreatedAt); err != nil {
			return nil, err
		}
		d.Kind = domain.DocumentKind(kind)
		d.ExpiresAt = scanNullTime(expires)
		d.CreatedAt = d.CreatedAt.UTC()
		out = append(out, d)
	}
	return out, rows.Err()
}

type CertRepo struct{ DB *sql.DB }

func (r *CertRepo) Save(ctx context.Context, c domain.Certification) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO supplier_certifications (
			id, tenant_id, supplier_id, name, issuer, valid_until, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7)
		ON CONFLICT (id) DO UPDATE SET name=EXCLUDED.name, issuer=EXCLUDED.issuer, valid_until=EXCLUDED.valid_until`,
		c.ID, c.TenantID, c.SupplierID, c.Name, c.Issuer, nullTime(c.ValidUntil), c.CreatedAt.UTC())
	return err
}

func (r *CertRepo) ListBySupplier(ctx context.Context, tenantID, supplierID uuid.UUID) ([]domain.Certification, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, supplier_id, name, issuer, valid_until, created_at
		FROM supplier_certifications WHERE tenant_id=$1 AND supplier_id=$2 ORDER BY created_at DESC`, tenantID, supplierID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Certification{}
	for rows.Next() {
		var c domain.Certification
		var until sql.NullTime
		if err := rows.Scan(&c.ID, &c.TenantID, &c.SupplierID, &c.Name, &c.Issuer, &until, &c.CreatedAt); err != nil {
			return nil, err
		}
		c.ValidUntil = scanNullTime(until)
		c.CreatedAt = c.CreatedAt.UTC()
		out = append(out, c)
	}
	return out, rows.Err()
}

type ContractRepo struct{ DB *sql.DB }

func (r *ContractRepo) Save(ctx context.Context, c domain.Contract) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO supplier_contracts (
			id, tenant_id, supplier_id, title, version, status, starts_at, ends_at, terms_uri, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		ON CONFLICT (id) DO UPDATE SET title=EXCLUDED.title, version=EXCLUDED.version, status=EXCLUDED.status,
			starts_at=EXCLUDED.starts_at, ends_at=EXCLUDED.ends_at, terms_uri=EXCLUDED.terms_uri, updated_at=EXCLUDED.updated_at`,
		c.ID, c.TenantID, c.SupplierID, c.Title, c.Version, string(c.Status), c.StartsAt.UTC(), c.EndsAt.UTC(),
		c.TermsURI, c.CreatedAt.UTC(), c.UpdatedAt.UTC())
	return err
}

func (r *ContractRepo) Get(ctx context.Context, tenantID, id uuid.UUID) (domain.Contract, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, supplier_id, title, version, status, starts_at, ends_at, terms_uri, created_at, updated_at
		FROM supplier_contracts WHERE id=$1 AND tenant_id=$2`, id, tenantID)
	var c domain.Contract
	var status string
	err := row.Scan(&c.ID, &c.TenantID, &c.SupplierID, &c.Title, &c.Version, &status, &c.StartsAt, &c.EndsAt, &c.TermsURI, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		if isNoRows(err) {
			return domain.Contract{}, domain.ErrNotFound
		}
		return domain.Contract{}, err
	}
	c.Status = domain.ContractStatus(status)
	c.StartsAt = c.StartsAt.UTC()
	c.EndsAt = c.EndsAt.UTC()
	c.CreatedAt = c.CreatedAt.UTC()
	c.UpdatedAt = c.UpdatedAt.UTC()
	return c, nil
}

func (r *ContractRepo) ListBySupplier(ctx context.Context, tenantID, supplierID uuid.UUID) ([]domain.Contract, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, supplier_id, title, version, status, starts_at, ends_at, terms_uri, created_at, updated_at
		FROM supplier_contracts WHERE tenant_id=$1 AND supplier_id=$2 ORDER BY updated_at DESC`, tenantID, supplierID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Contract{}
	for rows.Next() {
		var c domain.Contract
		var status string
		if err := rows.Scan(&c.ID, &c.TenantID, &c.SupplierID, &c.Title, &c.Version, &status, &c.StartsAt, &c.EndsAt, &c.TermsURI, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		c.Status = domain.ContractStatus(status)
		c.StartsAt = c.StartsAt.UTC()
		c.EndsAt = c.EndsAt.UTC()
		c.CreatedAt = c.CreatedAt.UTC()
		c.UpdatedAt = c.UpdatedAt.UTC()
		out = append(out, c)
	}
	return out, rows.Err()
}

type RFQRepo struct{ DB *sql.DB }

func (r *RFQRepo) Save(ctx context.Context, rfq domain.RFQ) error {
	lines, err := marshalJSON(rfq.Lines)
	if err != nil {
		return err
	}
	_, err = r.DB.ExecContext(ctx, `
		INSERT INTO supplier_rfqs (
			id, tenant_id, company_id, number, title, status, lines_json, due_at, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		ON CONFLICT (tenant_id, number) DO UPDATE SET
			id=EXCLUDED.id, title=EXCLUDED.title, status=EXCLUDED.status, lines_json=EXCLUDED.lines_json, due_at=EXCLUDED.due_at`,
		rfq.ID, rfq.TenantID, rfq.CompanyID, rfq.Number, rfq.Title, string(rfq.Status), lines, rfq.DueAt.UTC(), rfq.CreatedAt.UTC())
	return err
}

func (r *RFQRepo) Get(ctx context.Context, tenantID, id uuid.UUID) (domain.RFQ, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, company_id, number, title, status, lines_json, due_at, created_at
		FROM supplier_rfqs WHERE id=$1 AND tenant_id=$2`, id, tenantID)
	var rfq domain.RFQ
	var status string
	var raw []byte
	err := row.Scan(&rfq.ID, &rfq.TenantID, &rfq.CompanyID, &rfq.Number, &rfq.Title, &status, &raw, &rfq.DueAt, &rfq.CreatedAt)
	if err != nil {
		if isNoRows(err) {
			return domain.RFQ{}, domain.ErrNotFound
		}
		return domain.RFQ{}, err
	}
	rfq.Status = domain.RFQStatus(status)
	_ = unmarshalJSON(raw, &rfq.Lines)
	rfq.DueAt = rfq.DueAt.UTC()
	rfq.CreatedAt = rfq.CreatedAt.UTC()
	return rfq, nil
}

func (r *RFQRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.RFQ, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, company_id, number, title, status, lines_json, due_at, created_at
		FROM supplier_rfqs WHERE tenant_id=$1 ORDER BY created_at DESC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.RFQ{}
	for rows.Next() {
		var rfq domain.RFQ
		var status string
		var raw []byte
		if err := rows.Scan(&rfq.ID, &rfq.TenantID, &rfq.CompanyID, &rfq.Number, &rfq.Title, &status, &raw, &rfq.DueAt, &rfq.CreatedAt); err != nil {
			return nil, err
		}
		rfq.Status = domain.RFQStatus(status)
		_ = unmarshalJSON(raw, &rfq.Lines)
		rfq.DueAt = rfq.DueAt.UTC()
		rfq.CreatedAt = rfq.CreatedAt.UTC()
		out = append(out, rfq)
	}
	return out, rows.Err()
}

type QuotationRepo struct{ DB *sql.DB }

func (r *QuotationRepo) Save(ctx context.Context, q domain.Quotation) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO supplier_quotations (
			id, tenant_id, rfq_id, supplier_id, currency, total_minor, lead_time_days, status, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		ON CONFLICT (id) DO UPDATE SET status=EXCLUDED.status, total_minor=EXCLUDED.total_minor, lead_time_days=EXCLUDED.lead_time_days`,
		q.ID, q.TenantID, q.RFQID, q.SupplierID, q.Currency, q.TotalMinor, q.LeadTimeDays, q.Status, q.CreatedAt.UTC())
	return err
}

func (r *QuotationRepo) Get(ctx context.Context, tenantID, id uuid.UUID) (domain.Quotation, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, rfq_id, supplier_id, currency, total_minor, lead_time_days, status, created_at
		FROM supplier_quotations WHERE id=$1 AND tenant_id=$2`, id, tenantID)
	var q domain.Quotation
	err := row.Scan(&q.ID, &q.TenantID, &q.RFQID, &q.SupplierID, &q.Currency, &q.TotalMinor, &q.LeadTimeDays, &q.Status, &q.CreatedAt)
	if err != nil {
		if isNoRows(err) {
			return domain.Quotation{}, domain.ErrNotFound
		}
		return domain.Quotation{}, err
	}
	q.CreatedAt = q.CreatedAt.UTC()
	return q, nil
}

func (r *QuotationRepo) ListByRFQ(ctx context.Context, tenantID, rfqID uuid.UUID) ([]domain.Quotation, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, rfq_id, supplier_id, currency, total_minor, lead_time_days, status, created_at
		FROM supplier_quotations WHERE tenant_id=$1 AND rfq_id=$2 ORDER BY created_at DESC`, tenantID, rfqID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Quotation{}
	for rows.Next() {
		var q domain.Quotation
		if err := rows.Scan(&q.ID, &q.TenantID, &q.RFQID, &q.SupplierID, &q.Currency, &q.TotalMinor, &q.LeadTimeDays, &q.Status, &q.CreatedAt); err != nil {
			return nil, err
		}
		q.CreatedAt = q.CreatedAt.UTC()
		out = append(out, q)
	}
	return out, rows.Err()
}

var (
	_ ports.SupplierRepo  = (*SupplierRepo)(nil)
	_ ports.DocumentRepo  = (*DocumentRepo)(nil)
	_ ports.CertRepo      = (*CertRepo)(nil)
	_ ports.ContractRepo  = (*ContractRepo)(nil)
	_ ports.RFQRepo       = (*RFQRepo)(nil)
	_ ports.QuotationRepo = (*QuotationRepo)(nil)
)
