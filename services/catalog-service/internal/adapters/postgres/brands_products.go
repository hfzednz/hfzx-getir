package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/catalog-service/internal/app/ports"
	"github.com/nexora/catalog-service/internal/domain"
)

type BrandRepo struct{ DB *sql.DB }

func (r *BrandRepo) Create(ctx context.Context, b domain.Brand) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO brands (
			id, tenant_id, name, slug, description, logo_url, website_url, country_code,
			external_ref, metadata, status, created_at, updated_at, deleted_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11::brand_status,$12,$13,$14)`,
		b.ID, b.TenantID, b.Name, b.Slug, b.Description, b.LogoURL, b.WebsiteURL, b.CountryCode,
		b.ExternalRef, JSONMap(b.Metadata), string(b.Status), b.CreatedAt, b.UpdatedAt, nullTime(b.DeletedAt))
	return mapUniqueViolation(err)
}

func (r *BrandRepo) Update(ctx context.Context, b domain.Brand) error {
	res, err := r.DB.ExecContext(ctx, `
		UPDATE brands SET name=$2, slug=$3, description=$4, logo_url=$5, website_url=$6, country_code=$7,
			external_ref=$8, metadata=$9, status=$10::brand_status, updated_at=$11, deleted_at=$12
		WHERE id=$1 AND tenant_id=$13`,
		b.ID, b.Name, b.Slug, b.Description, b.LogoURL, b.WebsiteURL, b.CountryCode,
		b.ExternalRef, JSONMap(b.Metadata), string(b.Status), b.UpdatedAt, nullTime(b.DeletedAt), b.TenantID)
	if err != nil {
		return mapUniqueViolation(err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *BrandRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (domain.Brand, error) {
	return r.scanOne(ctx, `
		SELECT id, tenant_id, name, slug, description, logo_url, website_url, country_code,
			external_ref, metadata, status::text, created_at, updated_at, deleted_at
		FROM brands WHERE id=$1 AND tenant_id=$2`, id, tenantID)
}

func (r *BrandRepo) GetBySlug(ctx context.Context, tenantID uuid.UUID, slug string) (domain.Brand, error) {
	return r.scanOne(ctx, `
		SELECT id, tenant_id, name, slug, description, logo_url, website_url, country_code,
			external_ref, metadata, status::text, created_at, updated_at, deleted_at
		FROM brands WHERE tenant_id=$1 AND slug=$2`, tenantID, slug)
}

func (r *BrandRepo) List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]domain.Brand, int, error) {
	if limit <= 0 {
		limit = 50
	}
	var total int
	if err := r.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM brands WHERE tenant_id=$1`, tenantID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, name, slug, description, logo_url, website_url, country_code,
			external_ref, metadata, status::text, created_at, updated_at, deleted_at
		FROM brands WHERE tenant_id=$1 ORDER BY name ASC LIMIT $2 OFFSET $3`, tenantID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []domain.Brand{}
	for rows.Next() {
		b, err := scanBrand(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, b)
	}
	return out, total, rows.Err()
}

func (r *BrandRepo) scanOne(ctx context.Context, q string, args ...any) (domain.Brand, error) {
	row := r.DB.QueryRowContext(ctx, q, args...)
	b, err := scanBrand(row)
	if err != nil {
		return domain.Brand{}, mapNotFound(err)
	}
	return b, nil
}

type brandScanner interface {
	Scan(dest ...any) error
}

func scanBrand(s brandScanner) (domain.Brand, error) {
	var b domain.Brand
	var status string
	var meta JSONMap
	var deleted sql.NullTime
	err := s.Scan(&b.ID, &b.TenantID, &b.Name, &b.Slug, &b.Description, &b.LogoURL, &b.WebsiteURL, &b.CountryCode,
		&b.ExternalRef, &meta, &status, &b.CreatedAt, &b.UpdatedAt, &deleted)
	if err != nil {
		return domain.Brand{}, err
	}
	b.Metadata = map[string]any(meta)
	b.Status = domain.BrandStatus(status)
	b.DeletedAt = scanNullTime(deleted)
	return b, nil
}

var _ ports.BrandRepository = (*BrandRepo)(nil)

type ProductRepo struct{ DB *sql.DB }

func (r *ProductRepo) Create(ctx context.Context, p domain.Product) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO products (
			id, tenant_id, brand_id, kind, status, slug, sku_code, external_ref, gtin_base,
			manufacturer_sku, metadata, scheduled_at, published_at, created_at, updated_at, deleted_at
		) VALUES ($1,$2,$3,$4::product_kind,$5::product_status,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)`,
		p.ID, p.TenantID, nullUUID(p.BrandID), string(p.Kind), string(p.Status), p.Slug, p.SKUCode, p.ExternalRef, p.GTINBase,
		p.ManufacturerSKU, JSONMap(p.Metadata), nullTime(p.ScheduledAt), nullTime(p.PublishedAt), p.CreatedAt, p.UpdatedAt, nullTime(p.DeletedAt))
	return mapUniqueViolation(err)
}

func (r *ProductRepo) Update(ctx context.Context, p domain.Product) error {
	res, err := r.DB.ExecContext(ctx, `
		UPDATE products SET brand_id=$2, kind=$3::product_kind, status=$4::product_status, slug=$5, sku_code=$6,
			external_ref=$7, gtin_base=$8, manufacturer_sku=$9, metadata=$10, scheduled_at=$11, published_at=$12,
			updated_at=$13, deleted_at=$14
		WHERE id=$1 AND tenant_id=$15`,
		p.ID, nullUUID(p.BrandID), string(p.Kind), string(p.Status), p.Slug, p.SKUCode,
		p.ExternalRef, p.GTINBase, p.ManufacturerSKU, JSONMap(p.Metadata), nullTime(p.ScheduledAt), nullTime(p.PublishedAt),
		p.UpdatedAt, nullTime(p.DeletedAt), p.TenantID)
	if err != nil {
		return mapUniqueViolation(err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *ProductRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (domain.Product, error) {
	return r.scanOne(ctx, `
		SELECT id, tenant_id, brand_id, kind::text, status::text, slug, sku_code, external_ref, gtin_base,
			manufacturer_sku, metadata, scheduled_at, published_at, created_at, updated_at, deleted_at
		FROM products WHERE id=$1 AND tenant_id=$2`, id, tenantID)
}

func (r *ProductRepo) GetBySlug(ctx context.Context, tenantID uuid.UUID, slug string) (domain.Product, error) {
	return r.scanOne(ctx, `
		SELECT id, tenant_id, brand_id, kind::text, status::text, slug, sku_code, external_ref, gtin_base,
			manufacturer_sku, metadata, scheduled_at, published_at, created_at, updated_at, deleted_at
		FROM products WHERE tenant_id=$1 AND slug=$2`, tenantID, slug)
}

func (r *ProductRepo) List(ctx context.Context, f ports.ProductFilter) ([]domain.Product, int, error) {
	limit := f.Limit
	if limit <= 0 {
		limit = 50
	}
	args := []any{f.TenantID}
	where := `tenant_id=$1`
	if f.Status != nil {
		args = append(args, string(*f.Status))
		where += ` AND status=$` + itoa(len(args)) + `::product_status`
	}
	if f.BrandID != nil {
		args = append(args, *f.BrandID)
		where += ` AND brand_id=$` + itoa(len(args))
	}
	if f.Query != "" {
		args = append(args, "%"+f.Query+"%")
		n := itoa(len(args))
		where += ` AND (slug ILIKE $` + n + ` OR sku_code ILIKE $` + n + `)`
	}
	var total int
	if err := r.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM products WHERE `+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	args = append(args, limit, f.Offset)
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, brand_id, kind::text, status::text, slug, sku_code, external_ref, gtin_base,
			manufacturer_sku, metadata, scheduled_at, published_at, created_at, updated_at, deleted_at
		FROM products WHERE `+where+` ORDER BY created_at ASC LIMIT $`+itoa(len(args)-1)+` OFFSET $`+itoa(len(args)), args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []domain.Product{}
	for rows.Next() {
		p, err := scanProduct(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, p)
	}
	return out, total, rows.Err()
}

func (r *ProductRepo) Delete(ctx context.Context, tenantID, id uuid.UUID, at time.Time) error {
	p, err := r.GetByID(ctx, tenantID, id)
	if err != nil {
		return err
	}
	p.Status = domain.ProductStatusDeleted
	p.DeletedAt = &at
	p.UpdatedAt = at
	return r.Update(ctx, p)
}

func (r *ProductRepo) scanOne(ctx context.Context, q string, args ...any) (domain.Product, error) {
	p, err := scanProduct(r.DB.QueryRowContext(ctx, q, args...))
	if err != nil {
		return domain.Product{}, mapNotFound(err)
	}
	return p, nil
}

type productScanner interface {
	Scan(dest ...any) error
}

func scanProduct(s productScanner) (domain.Product, error) {
	var p domain.Product
	var brand uuid.NullUUID
	var kind, status string
	var meta JSONMap
	var scheduled, published, deleted sql.NullTime
	err := s.Scan(&p.ID, &p.TenantID, &brand, &kind, &status, &p.Slug, &p.SKUCode, &p.ExternalRef, &p.GTINBase,
		&p.ManufacturerSKU, &meta, &scheduled, &published, &p.CreatedAt, &p.UpdatedAt, &deleted)
	if err != nil {
		return domain.Product{}, err
	}
	p.BrandID = scanNullUUID(brand)
	p.Kind = domain.ProductKind(kind)
	p.Status = domain.ProductStatus(status)
	p.Metadata = map[string]any(meta)
	p.ScheduledAt = scanNullTime(scheduled)
	p.PublishedAt = scanNullTime(published)
	p.DeletedAt = scanNullTime(deleted)
	return p, nil
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}

var _ ports.ProductRepository = (*ProductRepo)(nil)
