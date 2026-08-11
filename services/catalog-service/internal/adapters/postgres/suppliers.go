package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/catalog-service/internal/app/ports"
	"github.com/nexora/catalog-service/internal/domain"
)

// SupplierRepo persists suppliers and product links.
type SupplierRepo struct{ DB *sql.DB }

func (r *SupplierRepo) Create(ctx context.Context, s domain.Supplier) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO suppliers (
			id, tenant_id, code, name, contact_email, contact_phone, country_code,
			external_ref, metadata, status, created_at, updated_at, deleted_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10::supplier_status,$11,$12,$13)`,
		s.ID, s.TenantID, s.Code, s.Name, s.ContactEmail, s.ContactPhone, s.CountryCode,
		s.ExternalRef, JSONMap(s.Metadata), string(s.Status), s.CreatedAt, s.UpdatedAt, nullTime(s.DeletedAt))
	return mapUniqueViolation(err)
}

func (r *SupplierRepo) Update(ctx context.Context, s domain.Supplier) error {
	res, err := r.DB.ExecContext(ctx, `
		UPDATE suppliers SET name=$3, contact_email=$4, contact_phone=$5, country_code=$6,
			external_ref=$7, metadata=$8, status=$9::supplier_status, updated_at=$10, deleted_at=$11
		WHERE id=$1 AND tenant_id=$2`,
		s.ID, s.TenantID, s.Name, s.ContactEmail, s.ContactPhone, s.CountryCode,
		s.ExternalRef, JSONMap(s.Metadata), string(s.Status), s.UpdatedAt, nullTime(s.DeletedAt))
	if err != nil {
		return mapUniqueViolation(err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *SupplierRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (domain.Supplier, error) {
	return r.scanOne(ctx, `
		SELECT id, tenant_id, code, name, contact_email, contact_phone, country_code,
			external_ref, metadata, status::text, created_at, updated_at, deleted_at
		FROM suppliers WHERE id=$1 AND tenant_id=$2 AND deleted_at IS NULL`, id, tenantID)
}

func (r *SupplierRepo) GetByCode(ctx context.Context, tenantID uuid.UUID, code string) (domain.Supplier, error) {
	return r.scanOne(ctx, `
		SELECT id, tenant_id, code, name, contact_email, contact_phone, country_code,
			external_ref, metadata, status::text, created_at, updated_at, deleted_at
		FROM suppliers WHERE tenant_id=$1 AND code=$2 AND deleted_at IS NULL`, tenantID, code)
}

func (r *SupplierRepo) List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]domain.Supplier, int, error) {
	if limit <= 0 {
		limit = 50
	}
	var total int
	if err := r.DB.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM suppliers WHERE tenant_id=$1 AND deleted_at IS NULL`, tenantID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, code, name, contact_email, contact_phone, country_code,
			external_ref, metadata, status::text, created_at, updated_at, deleted_at
		FROM suppliers WHERE tenant_id=$1 AND deleted_at IS NULL
		ORDER BY name ASC LIMIT $2 OFFSET $3`, tenantID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []domain.Supplier{}
	for rows.Next() {
		s, err := scanSupplier(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, s)
	}
	return out, total, rows.Err()
}

func (r *SupplierRepo) LinkProduct(ctx context.Context, sp domain.SupplierProduct) error {
	var cost any
	if sp.CostHintMinor != nil {
		cost = *sp.CostHintMinor
	}
	var lead any
	if sp.LeadTimeDays != nil {
		lead = *sp.LeadTimeDays
	}
	var moq any
	if sp.MOQ != nil {
		moq = *sp.MOQ
	}
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO supplier_products (
			id, supplier_id, product_id, variant_id, tenant_id, supplier_sku,
			cost_hint_minor, cost_hint_currency, lead_time_days, moq, metadata, is_preferred,
			created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		sp.ID, sp.SupplierID, sp.ProductID, nullUUID(sp.VariantID), sp.TenantID, sp.SupplierSKU,
		cost, sp.CostHintCurrency, lead, moq, JSONMap(sp.Metadata), sp.IsPreferred,
		sp.CreatedAt, sp.UpdatedAt)
	return mapUniqueViolation(err)
}

func (r *SupplierRepo) ListProducts(ctx context.Context, tenantID, supplierID uuid.UUID) ([]domain.SupplierProduct, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, supplier_id, product_id, variant_id, tenant_id, supplier_sku,
			cost_hint_minor, cost_hint_currency, lead_time_days, moq, metadata, is_preferred,
			created_at, updated_at
		FROM supplier_products WHERE tenant_id=$1 AND supplier_id=$2
		ORDER BY created_at DESC`, tenantID, supplierID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.SupplierProduct{}
	for rows.Next() {
		sp, err := scanSupplierProduct(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, sp)
	}
	return out, rows.Err()
}

func (r *SupplierRepo) scanOne(ctx context.Context, q string, args ...any) (domain.Supplier, error) {
	row := r.DB.QueryRowContext(ctx, q, args...)
	s, err := scanSupplier(row)
	if err != nil {
		return domain.Supplier{}, mapNotFound(err)
	}
	return s, nil
}

type supplierScanner interface {
	Scan(dest ...any) error
}

func scanSupplier(row supplierScanner) (domain.Supplier, error) {
	var s domain.Supplier
	var status string
	var meta JSONMap
	var deleted sql.NullTime
	err := row.Scan(
		&s.ID, &s.TenantID, &s.Code, &s.Name, &s.ContactEmail, &s.ContactPhone, &s.CountryCode,
		&s.ExternalRef, &meta, &status, &s.CreatedAt, &s.UpdatedAt, &deleted)
	if err != nil {
		return domain.Supplier{}, err
	}
	s.Status = domain.SupplierStatus(status)
	s.Metadata = map[string]any(meta)
	s.DeletedAt = scanNullTime(deleted)
	return s, nil
}

func scanSupplierProduct(row supplierScanner) (domain.SupplierProduct, error) {
	var sp domain.SupplierProduct
	var variant uuid.NullUUID
	var costHint sql.NullInt64
	var lead sql.NullInt64
	var moq sql.NullInt64
	var meta JSONMap
	err := row.Scan(
		&sp.ID, &sp.SupplierID, &sp.ProductID, &variant, &sp.TenantID, &sp.SupplierSKU,
		&costHint, &sp.CostHintCurrency, &lead, &moq, &meta, &sp.IsPreferred,
		&sp.CreatedAt, &sp.UpdatedAt)
	if err != nil {
		return domain.SupplierProduct{}, err
	}
	sp.VariantID = scanNullUUID(variant)
	if costHint.Valid {
		v := costHint.Int64
		sp.CostHintMinor = &v
	}
	if lead.Valid {
		v := int(lead.Int64)
		sp.LeadTimeDays = &v
	}
	if moq.Valid {
		v := int(moq.Int64)
		sp.MOQ = &v
	}
	sp.Metadata = map[string]any(meta)
	return sp, nil
}

var _ ports.SupplierRepository = (*SupplierRepo)(nil)
