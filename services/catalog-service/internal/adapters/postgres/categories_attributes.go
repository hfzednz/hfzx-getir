package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/catalog-service/internal/app/ports"
	"github.com/nexora/catalog-service/internal/domain"
)

type CategoryRepo struct{ DB *sql.DB }

func (r *CategoryRepo) Create(ctx context.Context, c domain.Category) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO categories (
			id, tenant_id, parent_id, name, slug, kind, path, depth, sort_order, description,
			image_url, metadata, is_active, created_at, updated_at, deleted_at
		) VALUES ($1,$2,$3,$4,$5,$6::category_kind,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)`,
		c.ID, c.TenantID, nullUUID(c.ParentID), c.Name, c.Slug, string(c.Kind), c.Path, c.Depth, c.SortOrder, c.Description,
		c.ImageURL, JSONMap(c.Metadata), c.IsActive, c.CreatedAt, c.UpdatedAt, nullTime(c.DeletedAt))
	return mapUniqueViolation(err)
}

func (r *CategoryRepo) Update(ctx context.Context, c domain.Category) error {
	res, err := r.DB.ExecContext(ctx, `
		UPDATE categories SET parent_id=$2, name=$3, slug=$4, kind=$5::category_kind, path=$6, depth=$7, sort_order=$8,
			description=$9, image_url=$10, metadata=$11, is_active=$12, updated_at=$13, deleted_at=$14
		WHERE id=$1 AND tenant_id=$15`,
		c.ID, nullUUID(c.ParentID), c.Name, c.Slug, string(c.Kind), c.Path, c.Depth, c.SortOrder,
		c.Description, c.ImageURL, JSONMap(c.Metadata), c.IsActive, c.UpdatedAt, nullTime(c.DeletedAt), c.TenantID)
	if err != nil {
		return mapUniqueViolation(err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *CategoryRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (domain.Category, error) {
	c, err := scanCategory(r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, parent_id, name, slug, kind::text, path, depth, sort_order, description,
			image_url, metadata, is_active, created_at, updated_at, deleted_at
		FROM categories WHERE id=$1 AND tenant_id=$2`, id, tenantID))
	if err != nil {
		return domain.Category{}, mapNotFound(err)
	}
	return c, nil
}

func (r *CategoryRepo) GetBySlug(ctx context.Context, tenantID uuid.UUID, slug string) (domain.Category, error) {
	c, err := scanCategory(r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, parent_id, name, slug, kind::text, path, depth, sort_order, description,
			image_url, metadata, is_active, created_at, updated_at, deleted_at
		FROM categories WHERE tenant_id=$1 AND slug=$2`, tenantID, slug))
	if err != nil {
		return domain.Category{}, mapNotFound(err)
	}
	return c, nil
}

func (r *CategoryRepo) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]domain.Category, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, parent_id, name, slug, kind::text, path, depth, sort_order, description,
			image_url, metadata, is_active, created_at, updated_at, deleted_at
		FROM categories WHERE tenant_id=$1 ORDER BY depth ASC, sort_order ASC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Category{}
	for rows.Next() {
		c, err := scanCategory(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *CategoryRepo) ListChildren(ctx context.Context, tenantID, parentID uuid.UUID) ([]domain.Category, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, parent_id, name, slug, kind::text, path, depth, sort_order, description,
			image_url, metadata, is_active, created_at, updated_at, deleted_at
		FROM categories WHERE tenant_id=$1 AND parent_id=$2 ORDER BY sort_order ASC`, tenantID, parentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Category{}
	for rows.Next() {
		c, err := scanCategory(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *CategoryRepo) AssignProduct(ctx context.Context, pc domain.ProductCategory) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO product_categories (product_id, category_id, is_primary, sort_order, assigned_at)
		VALUES ($1,$2,$3,$4,$5)
		ON CONFLICT (product_id, category_id) DO UPDATE SET
			is_primary=EXCLUDED.is_primary, sort_order=EXCLUDED.sort_order, assigned_at=EXCLUDED.assigned_at`,
		pc.ProductID, pc.CategoryID, pc.IsPrimary, pc.SortOrder, pc.AssignedAt)
	return err
}

func (r *CategoryRepo) RemoveProduct(ctx context.Context, productID, categoryID uuid.UUID) error {
	_, err := r.DB.ExecContext(ctx, `DELETE FROM product_categories WHERE product_id=$1 AND category_id=$2`, productID, categoryID)
	return err
}

func (r *CategoryRepo) ListProductCategories(ctx context.Context, productID uuid.UUID) ([]domain.ProductCategory, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT product_id, category_id, is_primary, sort_order, assigned_at
		FROM product_categories WHERE product_id=$1 ORDER BY sort_order`, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.ProductCategory{}
	for rows.Next() {
		var pc domain.ProductCategory
		if err := rows.Scan(&pc.ProductID, &pc.CategoryID, &pc.IsPrimary, &pc.SortOrder, &pc.AssignedAt); err != nil {
			return nil, err
		}
		out = append(out, pc)
	}
	return out, rows.Err()
}

func (r *CategoryRepo) ListProductsInCategory(ctx context.Context, tenantID, categoryID uuid.UUID, limit, offset int) ([]uuid.UUID, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.DB.QueryContext(ctx, `
		SELECT pc.product_id FROM product_categories pc
		JOIN products p ON p.id = pc.product_id
		WHERE pc.category_id=$1 AND p.tenant_id=$2
		ORDER BY pc.sort_order LIMIT $3 OFFSET $4`, categoryID, tenantID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []uuid.UUID{}
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

type categoryScanner interface {
	Scan(dest ...any) error
}

func scanCategory(s categoryScanner) (domain.Category, error) {
	var c domain.Category
	var parent uuid.NullUUID
	var kind string
	var meta JSONMap
	var deleted sql.NullTime
	err := s.Scan(&c.ID, &c.TenantID, &parent, &c.Name, &c.Slug, &kind, &c.Path, &c.Depth, &c.SortOrder, &c.Description,
		&c.ImageURL, &meta, &c.IsActive, &c.CreatedAt, &c.UpdatedAt, &deleted)
	if err != nil {
		return domain.Category{}, err
	}
	c.ParentID = scanNullUUID(parent)
	c.Kind = domain.CategoryKind(kind)
	c.Metadata = map[string]any(meta)
	c.DeletedAt = scanNullTime(deleted)
	return c, nil
}

var _ ports.CategoryRepository = (*CategoryRepo)(nil)

type AttributeRepo struct{ DB *sql.DB }

func (r *AttributeRepo) CreateDef(ctx context.Context, d domain.AttributeDef) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO attribute_defs (
			id, tenant_id, code, name, description, type, schema, is_required, is_filterable,
			is_variant_axis, sort_order, created_at, updated_at, deleted_at
		) VALUES ($1,$2,$3,$4,$5,$6::attribute_type,$7,$8,$9,$10,$11,$12,$13,$14)`,
		d.ID, d.TenantID, d.Code, d.Name, d.Description, string(d.Type), JSONMap(d.Schema), d.IsRequired, d.IsFilterable,
		d.IsVariantAxis, d.SortOrder, d.CreatedAt, d.UpdatedAt, nullTime(d.DeletedAt))
	return mapUniqueViolation(err)
}

func (r *AttributeRepo) UpdateDef(ctx context.Context, d domain.AttributeDef) error {
	res, err := r.DB.ExecContext(ctx, `
		UPDATE attribute_defs SET name=$2, description=$3, type=$4::attribute_type, schema=$5, is_required=$6,
			is_filterable=$7, is_variant_axis=$8, sort_order=$9, updated_at=$10, deleted_at=$11
		WHERE id=$1 AND tenant_id=$12`,
		d.ID, d.Name, d.Description, string(d.Type), JSONMap(d.Schema), d.IsRequired,
		d.IsFilterable, d.IsVariantAxis, d.SortOrder, d.UpdatedAt, nullTime(d.DeletedAt), d.TenantID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *AttributeRepo) GetDefByID(ctx context.Context, tenantID, id uuid.UUID) (domain.AttributeDef, error) {
	d, err := scanAttrDef(r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, code, name, description, type::text, schema, is_required, is_filterable,
			is_variant_axis, sort_order, created_at, updated_at, deleted_at
		FROM attribute_defs WHERE id=$1 AND tenant_id=$2`, id, tenantID))
	if err != nil {
		return domain.AttributeDef{}, mapNotFound(err)
	}
	return d, nil
}

func (r *AttributeRepo) GetDefByCode(ctx context.Context, tenantID uuid.UUID, code string) (domain.AttributeDef, error) {
	d, err := scanAttrDef(r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, code, name, description, type::text, schema, is_required, is_filterable,
			is_variant_axis, sort_order, created_at, updated_at, deleted_at
		FROM attribute_defs WHERE tenant_id=$1 AND code=$2`, tenantID, code))
	if err != nil {
		return domain.AttributeDef{}, mapNotFound(err)
	}
	return d, nil
}

func (r *AttributeRepo) ListDefs(ctx context.Context, tenantID uuid.UUID) ([]domain.AttributeDef, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, code, name, description, type::text, schema, is_required, is_filterable,
			is_variant_axis, sort_order, created_at, updated_at, deleted_at
		FROM attribute_defs WHERE tenant_id=$1 ORDER BY sort_order, code`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.AttributeDef{}
	for rows.Next() {
		d, err := scanAttrDef(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func (r *AttributeRepo) UpsertProductAttribute(ctx context.Context, a domain.ProductAttribute) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO product_attributes (id, product_id, attribute_def_id, tenant_id, value, locale, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (product_id, attribute_def_id, locale) DO UPDATE SET
			value=EXCLUDED.value, updated_at=EXCLUDED.updated_at, id=product_attributes.id`,
		a.ID, a.ProductID, a.AttributeDefID, a.TenantID, JSONMap(a.Value), a.Locale, a.CreatedAt, a.UpdatedAt)
	return err
}

func (r *AttributeRepo) ListProductAttributes(ctx context.Context, tenantID, productID uuid.UUID) ([]domain.ProductAttribute, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, product_id, attribute_def_id, tenant_id, value, locale, created_at, updated_at
		FROM product_attributes WHERE tenant_id=$1 AND product_id=$2`, tenantID, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.ProductAttribute{}
	for rows.Next() {
		var a domain.ProductAttribute
		var val JSONMap
		if err := rows.Scan(&a.ID, &a.ProductID, &a.AttributeDefID, &a.TenantID, &val, &a.Locale, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		a.Value = map[string]any(val)
		out = append(out, a)
	}
	return out, rows.Err()
}

func (r *AttributeRepo) DeleteProductAttribute(ctx context.Context, tenantID, id uuid.UUID) error {
	res, err := r.DB.ExecContext(ctx, `DELETE FROM product_attributes WHERE id=$1 AND tenant_id=$2`, id, tenantID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

type attrDefScanner interface {
	Scan(dest ...any) error
}

func scanAttrDef(s attrDefScanner) (domain.AttributeDef, error) {
	var d domain.AttributeDef
	var typ string
	var schema JSONMap
	var deleted sql.NullTime
	err := s.Scan(&d.ID, &d.TenantID, &d.Code, &d.Name, &d.Description, &typ, &schema, &d.IsRequired, &d.IsFilterable,
		&d.IsVariantAxis, &d.SortOrder, &d.CreatedAt, &d.UpdatedAt, &deleted)
	if err != nil {
		return domain.AttributeDef{}, err
	}
	d.Type = domain.AttributeType(typ)
	d.Schema = map[string]any(schema)
	d.DeletedAt = scanNullTime(deleted)
	return d, nil
}

var _ ports.AttributeRepository = (*AttributeRepo)(nil)
