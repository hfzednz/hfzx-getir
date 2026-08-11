package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/catalog-service/internal/app/ports"
	"github.com/nexora/catalog-service/internal/domain"
)

type BundleRepo struct{ DB *sql.DB }

func (r *BundleRepo) Upsert(ctx context.Context, b domain.Bundle) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO bundles (id, product_id, tenant_id, composition, name, metadata, created_at, updated_at)
		VALUES ($1,$2,$3,$4::bundle_composition_type,$5,$6,$7,$8)
		ON CONFLICT (product_id) DO UPDATE SET
			composition=EXCLUDED.composition, name=EXCLUDED.name, metadata=EXCLUDED.metadata,
			updated_at=EXCLUDED.updated_at, id=bundles.id`,
		b.ID, b.ProductID, b.TenantID, string(b.Composition), b.Name, JSONMap(b.Metadata), b.CreatedAt, b.UpdatedAt)
	return err
}

func (r *BundleRepo) GetByProduct(ctx context.Context, tenantID, productID uuid.UUID) (domain.Bundle, error) {
	var b domain.Bundle
	var comp string
	var meta JSONMap
	err := r.DB.QueryRowContext(ctx, `
		SELECT id, product_id, tenant_id, composition::text, name, metadata, created_at, updated_at
		FROM bundles WHERE product_id=$1 AND tenant_id=$2`, productID, tenantID).
		Scan(&b.ID, &b.ProductID, &b.TenantID, &comp, &b.Name, &meta, &b.CreatedAt, &b.UpdatedAt)
	if err != nil {
		return domain.Bundle{}, mapNotFound(err)
	}
	b.Composition = domain.BundleCompositionType(comp)
	b.Metadata = map[string]any(meta)
	return b, nil
}

func (r *BundleRepo) SetItems(ctx context.Context, bundleID uuid.UUID, items []domain.BundleItem) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, `DELETE FROM bundle_items WHERE bundle_id=$1`, bundleID); err != nil {
		return err
	}
	for _, it := range items {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO bundle_items (id, bundle_id, component_variant_id, qty, is_optional, sort_order, metadata, created_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
			it.ID, it.BundleID, it.ComponentVariantID, it.Qty, it.IsOptional, it.SortOrder, JSONMap(it.Metadata), it.CreatedAt); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *BundleRepo) ListItems(ctx context.Context, bundleID uuid.UUID) ([]domain.BundleItem, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, bundle_id, component_variant_id, qty, is_optional, sort_order, metadata, created_at
		FROM bundle_items WHERE bundle_id=$1 ORDER BY sort_order`, bundleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.BundleItem{}
	for rows.Next() {
		var it domain.BundleItem
		var meta JSONMap
		if err := rows.Scan(&it.ID, &it.BundleID, &it.ComponentVariantID, &it.Qty, &it.IsOptional, &it.SortOrder, &meta, &it.CreatedAt); err != nil {
			return nil, err
		}
		it.Metadata = map[string]any(meta)
		out = append(out, it)
	}
	return out, rows.Err()
}

var _ ports.BundleRepository = (*BundleRepo)(nil)

type RelationRepo struct{ DB *sql.DB }

func (r *RelationRepo) Upsert(ctx context.Context, rel domain.ProductRelation) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO product_relations (
			id, tenant_id, source_product_id, target_product_id, type, sort_order, score, metadata, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5::relation_type,$6,$7,$8,$9,$10)
		ON CONFLICT (source_product_id, target_product_id, type) DO UPDATE SET
			sort_order=EXCLUDED.sort_order, score=EXCLUDED.score, metadata=EXCLUDED.metadata,
			updated_at=EXCLUDED.updated_at, id=product_relations.id`,
		rel.ID, rel.TenantID, rel.SourceProductID, rel.TargetProductID, string(rel.Type), rel.SortOrder,
		nullFloat64(rel.Score), JSONMap(rel.Metadata), rel.CreatedAt, rel.UpdatedAt)
	return err
}

func (r *RelationRepo) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	res, err := r.DB.ExecContext(ctx, `DELETE FROM product_relations WHERE id=$1 AND tenant_id=$2`, id, tenantID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *RelationRepo) ListBySource(ctx context.Context, tenantID, sourceID uuid.UUID, typ *domain.RelationType) ([]domain.ProductRelation, error) {
	q := `
		SELECT id, tenant_id, source_product_id, target_product_id, type::text, sort_order, score, metadata, created_at, updated_at
		FROM product_relations WHERE tenant_id=$1 AND source_product_id=$2`
	args := []any{tenantID, sourceID}
	if typ != nil {
		q += ` AND type=$3::relation_type`
		args = append(args, string(*typ))
	}
	q += ` ORDER BY sort_order`
	rows, err := r.DB.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.ProductRelation{}
	for rows.Next() {
		var rel domain.ProductRelation
		var t string
		var score sql.NullFloat64
		var meta JSONMap
		if err := rows.Scan(&rel.ID, &rel.TenantID, &rel.SourceProductID, &rel.TargetProductID, &t, &rel.SortOrder, &score, &meta, &rel.CreatedAt, &rel.UpdatedAt); err != nil {
			return nil, err
		}
		rel.Type = domain.RelationType(t)
		rel.Score = scanNullFloat64(score)
		rel.Metadata = map[string]any(meta)
		out = append(out, rel)
	}
	return out, rows.Err()
}

var _ ports.RelationRepository = (*RelationRepo)(nil)

type VersionRepo struct{ DB *sql.DB }

func (r *VersionRepo) Create(ctx context.Context, v domain.ProductVersion) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO product_versions (id, product_id, tenant_id, version_number, snapshot, status, change_summary, created_by, created_at)
		VALUES ($1,$2,$3,$4,$5,$6::product_status,$7,$8,$9)`,
		v.ID, v.ProductID, v.TenantID, v.VersionNumber, JSONMap(v.Snapshot), string(v.Status), v.ChangeSummary, nullUUID(v.CreatedBy), v.CreatedAt)
	return mapUniqueViolation(err)
}

func (r *VersionRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (domain.ProductVersion, error) {
	v, err := scanVersion(r.DB.QueryRowContext(ctx, `
		SELECT id, product_id, tenant_id, version_number, snapshot, status::text, change_summary, created_by, created_at
		FROM product_versions WHERE id=$1 AND tenant_id=$2`, id, tenantID))
	if err != nil {
		return domain.ProductVersion{}, mapNotFound(err)
	}
	return v, nil
}

func (r *VersionRepo) GetLatest(ctx context.Context, tenantID, productID uuid.UUID) (domain.ProductVersion, error) {
	v, err := scanVersion(r.DB.QueryRowContext(ctx, `
		SELECT id, product_id, tenant_id, version_number, snapshot, status::text, change_summary, created_by, created_at
		FROM product_versions WHERE tenant_id=$1 AND product_id=$2 ORDER BY version_number DESC LIMIT 1`, tenantID, productID))
	if err != nil {
		return domain.ProductVersion{}, mapNotFound(err)
	}
	return v, nil
}

func (r *VersionRepo) ListByProduct(ctx context.Context, tenantID, productID uuid.UUID, limit, offset int) ([]domain.ProductVersion, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, product_id, tenant_id, version_number, snapshot, status::text, change_summary, created_by, created_at
		FROM product_versions WHERE tenant_id=$1 AND product_id=$2
		ORDER BY version_number DESC LIMIT $3 OFFSET $4`, tenantID, productID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.ProductVersion{}
	for rows.Next() {
		v, err := scanVersion(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func (r *VersionRepo) NextVersionNumber(ctx context.Context, tenantID, productID uuid.UUID) (int, error) {
	var n sql.NullInt64
	err := r.DB.QueryRowContext(ctx, `
		SELECT MAX(version_number) FROM product_versions WHERE tenant_id=$1 AND product_id=$2`, tenantID, productID).Scan(&n)
	if err != nil {
		return 0, err
	}
	if !n.Valid {
		return 1, nil
	}
	return int(n.Int64) + 1, nil
}

type versionScanner interface {
	Scan(dest ...any) error
}

func scanVersion(s versionScanner) (domain.ProductVersion, error) {
	var v domain.ProductVersion
	var status string
	var snap JSONMap
	var createdBy uuid.NullUUID
	err := s.Scan(&v.ID, &v.ProductID, &v.TenantID, &v.VersionNumber, &snap, &status, &v.ChangeSummary, &createdBy, &v.CreatedAt)
	if err != nil {
		return domain.ProductVersion{}, err
	}
	v.Snapshot = map[string]any(snap)
	v.Status = domain.ProductStatus(status)
	v.CreatedBy = scanNullUUID(createdBy)
	return v, nil
}

var _ ports.VersionRepository = (*VersionRepo)(nil)

type WorkflowRepo struct{ DB *sql.DB }

func (r *WorkflowRepo) CreateAction(ctx context.Context, a domain.ApprovalAction) error {
	var from, to any
	if a.FromStatus != nil {
		from = string(*a.FromStatus)
	}
	if a.ToStatus != nil {
		to = string(*a.ToStatus)
	}
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO approval_actions (
			id, product_id, version_id, tenant_id, action, from_status, to_status, actor_id, actor_role, comment, metadata, created_at
		) VALUES ($1,$2,$3,$4,$5::approval_action_type,$6::product_status,$7::product_status,$8,$9,$10,$11,$12)`,
		a.ID, a.ProductID, nullUUID(a.VersionID), a.TenantID, string(a.Action), from, to, a.ActorID, a.ActorRole, a.Comment, JSONMap(a.Metadata), a.CreatedAt)
	return err
}

func (r *WorkflowRepo) ListByProduct(ctx context.Context, tenantID, productID uuid.UUID, limit int) ([]domain.ApprovalAction, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, product_id, version_id, tenant_id, action::text, from_status::text, to_status::text,
			actor_id, actor_role, comment, metadata, created_at
		FROM approval_actions WHERE tenant_id=$1 AND product_id=$2
		ORDER BY created_at DESC LIMIT $3`, tenantID, productID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.ApprovalAction{}
	for rows.Next() {
		var a domain.ApprovalAction
		var action string
		var from, to sql.NullString
		var versionID uuid.NullUUID
		var meta JSONMap
		if err := rows.Scan(&a.ID, &a.ProductID, &versionID, &a.TenantID, &action, &from, &to,
			&a.ActorID, &a.ActorRole, &a.Comment, &meta, &a.CreatedAt); err != nil {
			return nil, err
		}
		a.Action = domain.ApprovalActionType(action)
		a.VersionID = scanNullUUID(versionID)
		if from.Valid {
			s := domain.ProductStatus(from.String)
			a.FromStatus = &s
		}
		if to.Valid {
			s := domain.ProductStatus(to.String)
			a.ToStatus = &s
		}
		a.Metadata = map[string]any(meta)
		out = append(out, a)
	}
	return out, rows.Err()
}

var _ ports.WorkflowRepository = (*WorkflowRepo)(nil)
