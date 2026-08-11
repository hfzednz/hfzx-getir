package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/catalog-service/internal/app/ports"
	"github.com/nexora/catalog-service/internal/domain"
)

type LocaleRepo struct{ DB *sql.DB }

func (r *LocaleRepo) Upsert(ctx context.Context, l domain.ProductLocale) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO product_locales (
			id, product_id, tenant_id, lang, title, subtitle, description, short_description,
			specs, usage_text, warnings, ingredients, allergens, nutrition, storage, origin,
			metadata, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19)
		ON CONFLICT (product_id, lang) DO UPDATE SET
			title=EXCLUDED.title, subtitle=EXCLUDED.subtitle, description=EXCLUDED.description,
			short_description=EXCLUDED.short_description, specs=EXCLUDED.specs, usage_text=EXCLUDED.usage_text,
			warnings=EXCLUDED.warnings, ingredients=EXCLUDED.ingredients, allergens=EXCLUDED.allergens,
			nutrition=EXCLUDED.nutrition, storage=EXCLUDED.storage, origin=EXCLUDED.origin,
			metadata=EXCLUDED.metadata, updated_at=EXCLUDED.updated_at, id=product_locales.id`,
		l.ID, l.ProductID, l.TenantID, l.Lang, l.Title, l.Subtitle, l.Description, l.ShortDescription,
		l.Specs, l.Usage, l.Warnings, l.Ingredients, l.Allergens, l.Nutrition, l.Storage, l.Origin,
		JSONMap(l.Metadata), l.CreatedAt, l.UpdatedAt)
	return err
}

func (r *LocaleRepo) Get(ctx context.Context, tenantID, productID uuid.UUID, lang string) (domain.ProductLocale, error) {
	l, err := scanLocale(r.DB.QueryRowContext(ctx, `
		SELECT id, product_id, tenant_id, lang, title, subtitle, description, short_description,
			specs, usage_text, warnings, ingredients, allergens, nutrition, storage, origin,
			metadata, created_at, updated_at
		FROM product_locales WHERE tenant_id=$1 AND product_id=$2 AND lang=$3`, tenantID, productID, lang))
	if err != nil {
		return domain.ProductLocale{}, mapNotFound(err)
	}
	return l, nil
}

func (r *LocaleRepo) ListByProduct(ctx context.Context, tenantID, productID uuid.UUID) ([]domain.ProductLocale, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, product_id, tenant_id, lang, title, subtitle, description, short_description,
			specs, usage_text, warnings, ingredients, allergens, nutrition, storage, origin,
			metadata, created_at, updated_at
		FROM product_locales WHERE tenant_id=$1 AND product_id=$2`, tenantID, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.ProductLocale{}
	for rows.Next() {
		l, err := scanLocale(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

func (r *LocaleRepo) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	res, err := r.DB.ExecContext(ctx, `DELETE FROM product_locales WHERE id=$1 AND tenant_id=$2`, id, tenantID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

type localeScanner interface {
	Scan(dest ...any) error
}

func scanLocale(s localeScanner) (domain.ProductLocale, error) {
	var l domain.ProductLocale
	var meta JSONMap
	err := s.Scan(&l.ID, &l.ProductID, &l.TenantID, &l.Lang, &l.Title, &l.Subtitle, &l.Description, &l.ShortDescription,
		&l.Specs, &l.Usage, &l.Warnings, &l.Ingredients, &l.Allergens, &l.Nutrition, &l.Storage, &l.Origin,
		&meta, &l.CreatedAt, &l.UpdatedAt)
	if err != nil {
		return domain.ProductLocale{}, err
	}
	l.Metadata = map[string]any(meta)
	return l, nil
}

var _ ports.LocaleRepository = (*LocaleRepo)(nil)

type SEORepo struct{ DB *sql.DB }

func (r *SEORepo) Upsert(ctx context.Context, s domain.SEO) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO seo (
			id, tenant_id, entity_type, entity_id, lang, slug, meta_title, meta_description, meta_keywords,
			canonical_url, og_title, og_description, og_image_url, twitter_card, twitter_title,
			twitter_description, twitter_image_url, jsonld, robots, created_at, updated_at
		) VALUES ($1,$2,$3::seo_entity_type,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21)
		ON CONFLICT (tenant_id, entity_type, entity_id, lang) DO UPDATE SET
			slug=EXCLUDED.slug, meta_title=EXCLUDED.meta_title, meta_description=EXCLUDED.meta_description,
			meta_keywords=EXCLUDED.meta_keywords, canonical_url=EXCLUDED.canonical_url, og_title=EXCLUDED.og_title,
			og_description=EXCLUDED.og_description, og_image_url=EXCLUDED.og_image_url, twitter_card=EXCLUDED.twitter_card,
			twitter_title=EXCLUDED.twitter_title, twitter_description=EXCLUDED.twitter_description,
			twitter_image_url=EXCLUDED.twitter_image_url, jsonld=EXCLUDED.jsonld, robots=EXCLUDED.robots,
			updated_at=EXCLUDED.updated_at, id=seo.id`,
		s.ID, s.TenantID, string(s.EntityType), s.EntityID, s.Lang, s.Slug, s.MetaTitle, s.MetaDescription, s.MetaKeywords,
		s.CanonicalURL, s.OGTitle, s.OGDescription, s.OGImageURL, s.TwitterCard, s.TwitterTitle,
		s.TwitterDescription, s.TwitterImageURL, JSONMap(s.JSONLD), s.Robots, s.CreatedAt, s.UpdatedAt)
	return err
}

func (r *SEORepo) Get(ctx context.Context, tenantID uuid.UUID, entityType domain.SEOEntityType, entityID uuid.UUID, lang string) (domain.SEO, error) {
	s, err := scanSEO(r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, entity_type::text, entity_id, lang, slug, meta_title, meta_description, meta_keywords,
			canonical_url, og_title, og_description, og_image_url, twitter_card, twitter_title,
			twitter_description, twitter_image_url, jsonld, robots, created_at, updated_at
		FROM seo WHERE tenant_id=$1 AND entity_type=$2::seo_entity_type AND entity_id=$3 AND lang=$4`,
		tenantID, string(entityType), entityID, lang))
	if err != nil {
		return domain.SEO{}, mapNotFound(err)
	}
	return s, nil
}

func (r *SEORepo) ListByEntity(ctx context.Context, tenantID uuid.UUID, entityType domain.SEOEntityType, entityID uuid.UUID) ([]domain.SEO, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, entity_type::text, entity_id, lang, slug, meta_title, meta_description, meta_keywords,
			canonical_url, og_title, og_description, og_image_url, twitter_card, twitter_title,
			twitter_description, twitter_image_url, jsonld, robots, created_at, updated_at
		FROM seo WHERE tenant_id=$1 AND entity_type=$2::seo_entity_type AND entity_id=$3`,
		tenantID, string(entityType), entityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.SEO{}
	for rows.Next() {
		s, err := scanSEO(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

type seoScanner interface {
	Scan(dest ...any) error
}

func scanSEO(s seoScanner) (domain.SEO, error) {
	var out domain.SEO
	var et string
	var jsonld JSONMap
	err := s.Scan(&out.ID, &out.TenantID, &et, &out.EntityID, &out.Lang, &out.Slug, &out.MetaTitle, &out.MetaDescription, &out.MetaKeywords,
		&out.CanonicalURL, &out.OGTitle, &out.OGDescription, &out.OGImageURL, &out.TwitterCard, &out.TwitterTitle,
		&out.TwitterDescription, &out.TwitterImageURL, &jsonld, &out.Robots, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		return domain.SEO{}, err
	}
	out.EntityType = domain.SEOEntityType(et)
	out.JSONLD = map[string]any(jsonld)
	return out, nil
}

var _ ports.SEORepository = (*SEORepo)(nil)

type MediaRepo struct{ DB *sql.DB }

func (r *MediaRepo) Create(ctx context.Context, m domain.ProductMedia) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO product_media (
			id, product_id, variant_id, tenant_id, media_asset_id, kind, sort_order, cdn_url,
			alt_text, locale, is_primary, metadata, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6::media_kind,$7,$8,$9,$10,$11,$12,$13,$14)`,
		m.ID, m.ProductID, nullUUID(m.VariantID), m.TenantID, m.MediaAssetID, string(m.Kind), m.SortOrder, m.CDNURL,
		m.AltText, m.Locale, m.IsPrimary, JSONMap(m.Metadata), m.CreatedAt, m.UpdatedAt)
	return mapUniqueViolation(err)
}

func (r *MediaRepo) Update(ctx context.Context, m domain.ProductMedia) error {
	res, err := r.DB.ExecContext(ctx, `
		UPDATE product_media SET variant_id=$2, media_asset_id=$3, kind=$4::media_kind, sort_order=$5, cdn_url=$6,
			alt_text=$7, locale=$8, is_primary=$9, metadata=$10, updated_at=$11
		WHERE id=$1 AND tenant_id=$12`,
		m.ID, nullUUID(m.VariantID), m.MediaAssetID, string(m.Kind), m.SortOrder, m.CDNURL,
		m.AltText, m.Locale, m.IsPrimary, JSONMap(m.Metadata), m.UpdatedAt, m.TenantID)
	if err != nil {
		return mapUniqueViolation(err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *MediaRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (domain.ProductMedia, error) {
	m, err := scanMedia(r.DB.QueryRowContext(ctx, `
		SELECT id, product_id, variant_id, tenant_id, media_asset_id, kind::text, sort_order, cdn_url,
			alt_text, locale, is_primary, metadata, created_at, updated_at
		FROM product_media WHERE id=$1 AND tenant_id=$2`, id, tenantID))
	if err != nil {
		return domain.ProductMedia{}, mapNotFound(err)
	}
	return m, nil
}

func (r *MediaRepo) ListByProduct(ctx context.Context, tenantID, productID uuid.UUID) ([]domain.ProductMedia, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, product_id, variant_id, tenant_id, media_asset_id, kind::text, sort_order, cdn_url,
			alt_text, locale, is_primary, metadata, created_at, updated_at
		FROM product_media WHERE tenant_id=$1 AND product_id=$2 ORDER BY sort_order`, tenantID, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.ProductMedia{}
	for rows.Next() {
		m, err := scanMedia(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (r *MediaRepo) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	res, err := r.DB.ExecContext(ctx, `DELETE FROM product_media WHERE id=$1 AND tenant_id=$2`, id, tenantID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

type mediaScanner interface {
	Scan(dest ...any) error
}

func scanMedia(s mediaScanner) (domain.ProductMedia, error) {
	var m domain.ProductMedia
	var variant uuid.NullUUID
	var kind string
	var meta JSONMap
	err := s.Scan(&m.ID, &m.ProductID, &variant, &m.TenantID, &m.MediaAssetID, &kind, &m.SortOrder, &m.CDNURL,
		&m.AltText, &m.Locale, &m.IsPrimary, &meta, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		return domain.ProductMedia{}, err
	}
	m.VariantID = scanNullUUID(variant)
	m.Kind = domain.MediaKind(kind)
	m.Metadata = map[string]any(meta)
	return m, nil
}

var _ ports.MediaRepository = (*MediaRepo)(nil)
