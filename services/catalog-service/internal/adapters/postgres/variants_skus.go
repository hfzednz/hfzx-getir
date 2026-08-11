package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/catalog-service/internal/app/ports"
	"github.com/nexora/catalog-service/internal/domain"
)

type VariantRepo struct{ DB *sql.DB }

func (r *VariantRepo) Create(ctx context.Context, v domain.Variant) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO variants (
			id, product_id, tenant_id, sku_code, name, option_values, status, sort_order,
			barcode_hint, metadata, created_at, updated_at, deleted_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7::variant_status,$8,$9,$10,$11,$12,$13)`,
		v.ID, v.ProductID, v.TenantID, v.SKUCode, v.Name, JSONMap(v.OptionValues), string(v.Status), v.SortOrder,
		v.BarcodeHint, JSONMap(v.Metadata), v.CreatedAt, v.UpdatedAt, nullTime(v.DeletedAt))
	return mapUniqueViolation(err)
}

func (r *VariantRepo) Update(ctx context.Context, v domain.Variant) error {
	res, err := r.DB.ExecContext(ctx, `
		UPDATE variants SET sku_code=$2, name=$3, option_values=$4, status=$5::variant_status, sort_order=$6,
			barcode_hint=$7, metadata=$8, updated_at=$9, deleted_at=$10
		WHERE id=$1 AND tenant_id=$11`,
		v.ID, v.SKUCode, v.Name, JSONMap(v.OptionValues), string(v.Status), v.SortOrder,
		v.BarcodeHint, JSONMap(v.Metadata), v.UpdatedAt, nullTime(v.DeletedAt), v.TenantID)
	if err != nil {
		return mapUniqueViolation(err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *VariantRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (domain.Variant, error) {
	v, err := scanVariant(r.DB.QueryRowContext(ctx, `
		SELECT id, product_id, tenant_id, sku_code, name, option_values, status::text, sort_order,
			barcode_hint, metadata, created_at, updated_at, deleted_at
		FROM variants WHERE id=$1 AND tenant_id=$2`, id, tenantID))
	if err != nil {
		return domain.Variant{}, mapNotFound(err)
	}
	return v, nil
}

func (r *VariantRepo) ListByProduct(ctx context.Context, tenantID, productID uuid.UUID) ([]domain.Variant, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, product_id, tenant_id, sku_code, name, option_values, status::text, sort_order,
			barcode_hint, metadata, created_at, updated_at, deleted_at
		FROM variants WHERE tenant_id=$1 AND product_id=$2 ORDER BY sort_order ASC`, tenantID, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Variant{}
	for rows.Next() {
		v, err := scanVariant(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func (r *VariantRepo) Delete(ctx context.Context, tenantID, id uuid.UUID, at time.Time) error {
	v, err := r.GetByID(ctx, tenantID, id)
	if err != nil {
		return err
	}
	v.Status = domain.VariantStatusDeleted
	v.DeletedAt = &at
	v.UpdatedAt = at
	return r.Update(ctx, v)
}

type variantScanner interface {
	Scan(dest ...any) error
}

func scanVariant(s variantScanner) (domain.Variant, error) {
	var v domain.Variant
	var status string
	var opts, meta JSONMap
	var deleted sql.NullTime
	err := s.Scan(&v.ID, &v.ProductID, &v.TenantID, &v.SKUCode, &v.Name, &opts, &status, &v.SortOrder,
		&v.BarcodeHint, &meta, &v.CreatedAt, &v.UpdatedAt, &deleted)
	if err != nil {
		return domain.Variant{}, err
	}
	v.OptionValues = map[string]any(opts)
	v.Status = domain.VariantStatus(status)
	v.Metadata = map[string]any(meta)
	v.DeletedAt = scanNullTime(deleted)
	return v, nil
}

var _ ports.VariantRepository = (*VariantRepo)(nil)

type SKURepo struct{ DB *sql.DB }

func (r *SKURepo) Create(ctx context.Context, s domain.SKUIdentifier) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO sku_identifiers (id, variant_id, tenant_id, type, value, is_primary, metadata, created_at, updated_at)
		VALUES ($1,$2,$3,$4::sku_identifier_type,$5,$6,$7,$8,$9)`,
		s.ID, s.VariantID, s.TenantID, string(s.Type), s.Value, s.IsPrimary, JSONMap(s.Metadata), s.CreatedAt, s.UpdatedAt)
	return mapUniqueViolation(err)
}

func (r *SKURepo) Update(ctx context.Context, s domain.SKUIdentifier) error {
	res, err := r.DB.ExecContext(ctx, `
		UPDATE sku_identifiers SET type=$2::sku_identifier_type, value=$3, is_primary=$4, metadata=$5, updated_at=$6
		WHERE id=$1 AND tenant_id=$7`,
		s.ID, string(s.Type), s.Value, s.IsPrimary, JSONMap(s.Metadata), s.UpdatedAt, s.TenantID)
	if err != nil {
		return mapUniqueViolation(err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *SKURepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (domain.SKUIdentifier, error) {
	s, err := scanSKU(r.DB.QueryRowContext(ctx, `
		SELECT id, variant_id, tenant_id, type::text, value, is_primary, metadata, created_at, updated_at
		FROM sku_identifiers WHERE id=$1 AND tenant_id=$2`, id, tenantID))
	if err != nil {
		return domain.SKUIdentifier{}, mapNotFound(err)
	}
	return s, nil
}

func (r *SKURepo) ListByVariant(ctx context.Context, tenantID, variantID uuid.UUID) ([]domain.SKUIdentifier, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, variant_id, tenant_id, type::text, value, is_primary, metadata, created_at, updated_at
		FROM sku_identifiers WHERE tenant_id=$1 AND variant_id=$2 ORDER BY created_at`, tenantID, variantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.SKUIdentifier{}
	for rows.Next() {
		s, err := scanSKU(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *SKURepo) FindByValue(ctx context.Context, tenantID uuid.UUID, typ domain.SKUIdentifierType, value string) (domain.SKUIdentifier, error) {
	s, err := scanSKU(r.DB.QueryRowContext(ctx, `
		SELECT id, variant_id, tenant_id, type::text, value, is_primary, metadata, created_at, updated_at
		FROM sku_identifiers WHERE tenant_id=$1 AND type=$2::sku_identifier_type AND value=$3`, tenantID, string(typ), value))
	if err != nil {
		return domain.SKUIdentifier{}, mapNotFound(err)
	}
	return s, nil
}

func (r *SKURepo) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	res, err := r.DB.ExecContext(ctx, `DELETE FROM sku_identifiers WHERE id=$1 AND tenant_id=$2`, id, tenantID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

type skuScanner interface {
	Scan(dest ...any) error
}

func scanSKU(s skuScanner) (domain.SKUIdentifier, error) {
	var out domain.SKUIdentifier
	var typ string
	var meta JSONMap
	err := s.Scan(&out.ID, &out.VariantID, &out.TenantID, &typ, &out.Value, &out.IsPrimary, &meta, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		return domain.SKUIdentifier{}, err
	}
	out.Type = domain.SKUIdentifierType(typ)
	out.Metadata = map[string]any(meta)
	return out, nil
}

var _ ports.SKUIdentifierRepository = (*SKURepo)(nil)
