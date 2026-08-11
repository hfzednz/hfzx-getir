package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/search-service/internal/app/ports"
	"github.com/nexora/search-service/internal/domain"
)

// DocumentRepo persists product documents.
type DocumentRepo struct{ DB *sql.DB }

func (r *DocumentRepo) Upsert(ctx context.Context, d domain.ProductDocument) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO product_documents (
			tenant_id, product_id, variant_id, sku, title, description, brand_id, brand_name,
			category_ids, category_path, tags, attributes, price_minor, compare_at_minor, discount_pct,
			currency, available, warehouse_ids, city_id, rating_avg, review_count, popularity,
			freshness_score, profit_score, delivery_eta_min, image_ref, version, indexed_at
		) VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26,$27,$28
		)
		ON CONFLICT (tenant_id, product_id) DO UPDATE SET
			variant_id=EXCLUDED.variant_id, sku=EXCLUDED.sku, title=EXCLUDED.title, description=EXCLUDED.description,
			brand_id=EXCLUDED.brand_id, brand_name=EXCLUDED.brand_name, category_ids=EXCLUDED.category_ids,
			category_path=EXCLUDED.category_path, tags=EXCLUDED.tags, attributes=EXCLUDED.attributes,
			price_minor=EXCLUDED.price_minor, compare_at_minor=EXCLUDED.compare_at_minor, discount_pct=EXCLUDED.discount_pct,
			currency=EXCLUDED.currency, available=EXCLUDED.available, warehouse_ids=EXCLUDED.warehouse_ids,
			city_id=EXCLUDED.city_id, rating_avg=EXCLUDED.rating_avg, review_count=EXCLUDED.review_count,
			popularity=EXCLUDED.popularity, freshness_score=EXCLUDED.freshness_score, profit_score=EXCLUDED.profit_score,
			delivery_eta_min=EXCLUDED.delivery_eta_min, image_ref=EXCLUDED.image_ref, version=EXCLUDED.version,
			indexed_at=EXCLUDED.indexed_at`,
		d.TenantID, d.ProductID, nullUUIDValue(d.VariantID), d.SKU, d.Title, d.Description,
		nullUUIDValue(d.BrandID), d.BrandName, UUIDArray(d.CategoryIDs), TextArray(d.CategoryPath), TextArray(d.Tags),
		JSONStringMap(d.Attributes), d.PriceMinor, d.CompareAtMinor, d.DiscountPct, d.Currency, d.Available,
		UUIDArray(d.WarehouseIDs), nullUUIDValue(d.CityID), d.RatingAvg, d.ReviewCount, d.Popularity,
		d.FreshnessScore, d.ProfitScore, d.DeliveryETAMin, d.ImageRef, d.Version, d.IndexedAt.UTC())
	return err
}

func (r *DocumentRepo) Get(ctx context.Context, tenantID, productID uuid.UUID) (domain.ProductDocument, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT tenant_id, product_id, variant_id, sku, title, description, brand_id, brand_name,
			category_ids, category_path, tags, attributes, price_minor, compare_at_minor, discount_pct,
			currency, available, warehouse_ids, city_id, rating_avg, review_count, popularity,
			freshness_score, profit_score, delivery_eta_min, image_ref, version, indexed_at
		FROM product_documents WHERE tenant_id=$1 AND product_id=$2`, tenantID, productID)
	d, err := scanDoc(row)
	if err != nil {
		if isNoRows(err) {
			return domain.ProductDocument{}, domain.ErrNotFound
		}
		return domain.ProductDocument{}, err
	}
	return d, nil
}

func (r *DocumentRepo) List(ctx context.Context, tenantID uuid.UUID, limit int) ([]domain.ProductDocument, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := r.DB.QueryContext(ctx, `
		SELECT tenant_id, product_id, variant_id, sku, title, description, brand_id, brand_name,
			category_ids, category_path, tags, attributes, price_minor, compare_at_minor, discount_pct,
			currency, available, warehouse_ids, city_id, rating_avg, review_count, popularity,
			freshness_score, profit_score, delivery_eta_min, image_ref, version, indexed_at
		FROM product_documents WHERE tenant_id=$1 ORDER BY indexed_at DESC LIMIT $2`, tenantID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.ProductDocument{}
	for rows.Next() {
		d, err := scanDoc(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func (r *DocumentRepo) Delete(ctx context.Context, tenantID, productID uuid.UUID) error {
	_, err := r.DB.ExecContext(ctx, `DELETE FROM product_documents WHERE tenant_id=$1 AND product_id=$2`, tenantID, productID)
	return err
}

type scannable interface{ Scan(dest ...any) error }

func scanDoc(row scannable) (domain.ProductDocument, error) {
	var d domain.ProductDocument
	var variant, brand, city uuid.NullUUID
	var cats, wh UUIDArray
	var path, tags TextArray
	var attrs JSONStringMap
	err := row.Scan(
		&d.TenantID, &d.ProductID, &variant, &d.SKU, &d.Title, &d.Description, &brand, &d.BrandName,
		&cats, &path, &tags, &attrs, &d.PriceMinor, &d.CompareAtMinor, &d.DiscountPct,
		&d.Currency, &d.Available, &wh, &city, &d.RatingAvg, &d.ReviewCount, &d.Popularity,
		&d.FreshnessScore, &d.ProfitScore, &d.DeliveryETAMin, &d.ImageRef, &d.Version, &d.IndexedAt)
	if err != nil {
		return domain.ProductDocument{}, err
	}
	d.VariantID = scanUUIDOrNil(variant)
	d.BrandID = scanUUIDOrNil(brand)
	d.CityID = scanUUIDOrNil(city)
	d.CategoryIDs = []uuid.UUID(cats)
	d.WarehouseIDs = []uuid.UUID(wh)
	d.CategoryPath = []string(path)
	d.Tags = []string(tags)
	d.Attributes = map[string]string(attrs)
	d.IndexedAt = d.IndexedAt.UTC()
	return d, nil
}

// SynonymRepo persists synonym rules.
type SynonymRepo struct{ DB *sql.DB }

func (r *SynonymRepo) Save(ctx context.Context, s domain.SynonymRule) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO synonym_rules (id, tenant_id, locale, term, synonyms, active, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		ON CONFLICT (id) DO UPDATE SET
			locale=EXCLUDED.locale, term=EXCLUDED.term, synonyms=EXCLUDED.synonyms,
			active=EXCLUDED.active, updated_at=EXCLUDED.updated_at`,
		s.ID, s.TenantID, s.Locale, s.Term, TextArray(s.Synonyms), s.Active, s.UpdatedAt.UTC())
	return mapUniqueViolation(err)
}

func (r *SynonymRepo) List(ctx context.Context, tenantID uuid.UUID, locale string) ([]domain.SynonymRule, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, locale, term, synonyms, active, updated_at
		FROM synonym_rules WHERE tenant_id=$1 AND ($2='' OR locale='' OR locale=$2)
		ORDER BY term`, tenantID, locale)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.SynonymRule{}
	for rows.Next() {
		var s domain.SynonymRule
		var syn TextArray
		if err := rows.Scan(&s.ID, &s.TenantID, &s.Locale, &s.Term, &syn, &s.Active, &s.UpdatedAt); err != nil {
			return nil, err
		}
		s.Synonyms = []string(syn)
		s.UpdatedAt = s.UpdatedAt.UTC()
		out = append(out, s)
	}
	return out, rows.Err()
}

// BoostRepo persists boost rules.
type BoostRepo struct{ DB *sql.DB }

func (r *BoostRepo) Save(ctx context.Context, b domain.BoostRule) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO boost_rules (
			id, tenant_id, name, kind, product_ids, category_id, brand_id, weight, priority,
			active, starts_at, ends_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
		ON CONFLICT (id) DO UPDATE SET
			name=EXCLUDED.name, kind=EXCLUDED.kind, product_ids=EXCLUDED.product_ids,
			category_id=EXCLUDED.category_id, brand_id=EXCLUDED.brand_id, weight=EXCLUDED.weight,
			priority=EXCLUDED.priority, active=EXCLUDED.active, starts_at=EXCLUDED.starts_at,
			ends_at=EXCLUDED.ends_at, updated_at=EXCLUDED.updated_at`,
		b.ID, b.TenantID, b.Name, b.Kind, UUIDArray(b.ProductIDs), nullUUID(b.CategoryID), nullUUID(b.BrandID),
		b.Weight, b.Priority, b.Active, nullTime(b.StartsAt), nullTime(b.EndsAt), b.UpdatedAt.UTC())
	return mapUniqueViolation(err)
}

func (r *BoostRepo) ListActive(ctx context.Context, tenantID uuid.UUID, now time.Time) ([]domain.BoostRule, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, name, kind, product_ids, category_id, brand_id, weight, priority,
			active, starts_at, ends_at, updated_at
		FROM boost_rules
		WHERE tenant_id=$1 AND active=TRUE
			AND (starts_at IS NULL OR starts_at <= $2)
			AND (ends_at IS NULL OR ends_at >= $2)
		ORDER BY priority DESC, name`, tenantID, now.UTC())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.BoostRule{}
	for rows.Next() {
		var b domain.BoostRule
		var pids UUIDArray
		var cat, brand uuid.NullUUID
		var starts, ends sql.NullTime
		if err := rows.Scan(&b.ID, &b.TenantID, &b.Name, &b.Kind, &pids, &cat, &brand, &b.Weight, &b.Priority,
			&b.Active, &starts, &ends, &b.UpdatedAt); err != nil {
			return nil, err
		}
		b.ProductIDs = []uuid.UUID(pids)
		b.CategoryID = scanNullUUID(cat)
		b.BrandID = scanNullUUID(brand)
		b.StartsAt = scanNullTime(starts)
		b.EndsAt = scanNullTime(ends)
		b.UpdatedAt = b.UpdatedAt.UTC()
		out = append(out, b)
	}
	return out, rows.Err()
}

// IndexJobRepo persists index jobs.
type IndexJobRepo struct{ DB *sql.DB }

func (r *IndexJobRepo) Save(ctx context.Context, j domain.IndexJob) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO index_jobs (id, tenant_id, mode, status, docs_total, docs_done, error, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		ON CONFLICT (id) DO UPDATE SET
			mode=EXCLUDED.mode, status=EXCLUDED.status, docs_total=EXCLUDED.docs_total,
			docs_done=EXCLUDED.docs_done, error=EXCLUDED.error, updated_at=EXCLUDED.updated_at`,
		j.ID, j.TenantID, j.Mode, j.Status, j.DocsTotal, j.DocsDone, j.Error, j.CreatedAt.UTC(), j.UpdatedAt.UTC())
	return mapUniqueViolation(err)
}

func (r *IndexJobRepo) Get(ctx context.Context, tenantID, id uuid.UUID) (domain.IndexJob, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, mode, status, docs_total, docs_done, error, created_at, updated_at
		FROM index_jobs WHERE tenant_id=$1 AND id=$2`, tenantID, id)
	var j domain.IndexJob
	err := row.Scan(&j.ID, &j.TenantID, &j.Mode, &j.Status, &j.DocsTotal, &j.DocsDone, &j.Error, &j.CreatedAt, &j.UpdatedAt)
	if err != nil {
		if isNoRows(err) {
			return domain.IndexJob{}, domain.ErrNotFound
		}
		return domain.IndexJob{}, err
	}
	j.CreatedAt = j.CreatedAt.UTC()
	j.UpdatedAt = j.UpdatedAt.UTC()
	return j, nil
}

func (r *IndexJobRepo) List(ctx context.Context, tenantID uuid.UUID, limit int) ([]domain.IndexJob, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, mode, status, docs_total, docs_done, error, created_at, updated_at
		FROM index_jobs WHERE tenant_id=$1 ORDER BY created_at DESC LIMIT $2`, tenantID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.IndexJob{}
	for rows.Next() {
		var j domain.IndexJob
		if err := rows.Scan(&j.ID, &j.TenantID, &j.Mode, &j.Status, &j.DocsTotal, &j.DocsDone, &j.Error, &j.CreatedAt, &j.UpdatedAt); err != nil {
			return nil, err
		}
		j.CreatedAt = j.CreatedAt.UTC()
		j.UpdatedAt = j.UpdatedAt.UTC()
		out = append(out, j)
	}
	return out, rows.Err()
}

// TrendRepo persists trend entries.
type TrendRepo struct{ DB *sql.DB }

func (r *TrendRepo) Save(ctx context.Context, t domain.TrendEntry) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO trend_entries (tenant_id, kind, key, entity_id, score, region, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		ON CONFLICT (tenant_id, kind, key) DO UPDATE SET
			entity_id=EXCLUDED.entity_id, score=EXCLUDED.score, region=EXCLUDED.region, updated_at=EXCLUDED.updated_at`,
		t.TenantID, t.Kind, t.Key, nullUUID(t.EntityID), t.Score, t.Region, t.UpdatedAt.UTC())
	return err
}

func (r *TrendRepo) List(ctx context.Context, tenantID uuid.UUID, kind string, limit int) ([]domain.TrendEntry, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.DB.QueryContext(ctx, `
		SELECT tenant_id, kind, key, entity_id, score, region, updated_at
		FROM trend_entries WHERE tenant_id=$1 AND ($2='' OR kind=$2)
		ORDER BY score DESC LIMIT $3`, tenantID, kind, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.TrendEntry{}
	for rows.Next() {
		var t domain.TrendEntry
		var entity uuid.NullUUID
		if err := rows.Scan(&t.TenantID, &t.Kind, &t.Key, &entity, &t.Score, &t.Region, &t.UpdatedAt); err != nil {
			return nil, err
		}
		t.EntityID = scanNullUUID(entity)
		t.UpdatedAt = t.UpdatedAt.UTC()
		out = append(out, t)
	}
	return out, rows.Err()
}

func (r *TrendRepo) Bump(ctx context.Context, tenantID uuid.UUID, kind, key string, entityID *uuid.UUID, delta float64, now time.Time) error {
	if key == "" {
		return nil
	}
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO trend_entries (tenant_id, kind, key, entity_id, score, region, updated_at)
		VALUES ($1,$2,$3,$4,$5,'',$6)
		ON CONFLICT (tenant_id, kind, key) DO UPDATE SET
			entity_id=COALESCE(EXCLUDED.entity_id, trend_entries.entity_id),
			score=trend_entries.score + EXCLUDED.score,
			updated_at=EXCLUDED.updated_at`,
		tenantID, kind, key, nullUUID(entityID), delta, now.UTC())
	return err
}

// SuggestRepo persists suggest candidates.
type SuggestRepo struct{ DB *sql.DB }

func (r *SuggestRepo) Upsert(ctx context.Context, tenantID uuid.UUID, c domain.SuggestCandidate) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO suggest_candidates (tenant_id, text, product_id, category_id, weight)
		VALUES ($1,$2,$3,$4,$5)
		ON CONFLICT (tenant_id, text) DO UPDATE SET
			product_id=EXCLUDED.product_id, category_id=EXCLUDED.category_id, weight=EXCLUDED.weight`,
		tenantID, c.Text, nullUUID(c.ProductID), nullUUID(c.CategoryID), c.Weight)
	return err
}

func (r *SuggestRepo) Suggest(ctx context.Context, tenantID uuid.UUID, prefix string, limit int) ([]domain.SuggestCandidate, error) {
	if limit <= 0 {
		limit = 10
	}
	rows, err := r.DB.QueryContext(ctx, `
		SELECT text, product_id, category_id, weight
		FROM suggest_candidates
		WHERE tenant_id=$1 AND ($2='' OR lower(text) LIKE lower($2) || '%' OR lower(text) LIKE '%' || lower($2) || '%')
		ORDER BY weight DESC LIMIT $3`, tenantID, prefix, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.SuggestCandidate{}
	for rows.Next() {
		var c domain.SuggestCandidate
		var pid, cid uuid.NullUUID
		if err := rows.Scan(&c.Text, &pid, &cid, &c.Weight); err != nil {
			return nil, err
		}
		c.ProductID = scanNullUUID(pid)
		c.CategoryID = scanNullUUID(cid)
		out = append(out, c)
	}
	return out, rows.Err()
}

var (
	_ ports.DocumentRepo = (*DocumentRepo)(nil)
	_ ports.SynonymRepo  = (*SynonymRepo)(nil)
	_ ports.BoostRepo    = (*BoostRepo)(nil)
	_ ports.IndexJobRepo = (*IndexJobRepo)(nil)
	_ ports.TrendRepo    = (*TrendRepo)(nil)
	_ ports.SuggestRepo  = (*SuggestRepo)(nil)
)
