-- SEO metadata for products (and optionally categories via entity_type).
CREATE TYPE seo_entity_type AS ENUM (
    'product',
    'category',
    'brand',
    'collection'
);

CREATE TABLE seo (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    entity_type     seo_entity_type NOT NULL DEFAULT 'product',
    entity_id       UUID NOT NULL,
    lang            TEXT NOT NULL DEFAULT 'en',
    slug            TEXT NOT NULL DEFAULT '',
    meta_title      TEXT NOT NULL DEFAULT '',
    meta_description TEXT NOT NULL DEFAULT '',
    meta_keywords   TEXT NOT NULL DEFAULT '',
    canonical_url   TEXT NOT NULL DEFAULT '',
    og_title        TEXT NOT NULL DEFAULT '',
    og_description  TEXT NOT NULL DEFAULT '',
    og_image_url    TEXT NOT NULL DEFAULT '',
    twitter_card    TEXT NOT NULL DEFAULT '',
    twitter_title   TEXT NOT NULL DEFAULT '',
    twitter_description TEXT NOT NULL DEFAULT '',
    twitter_image_url TEXT NOT NULL DEFAULT '',
    jsonld          JSONB NOT NULL DEFAULT '{}'::jsonb,
    robots          TEXT NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_seo_entity_lang UNIQUE (tenant_id, entity_type, entity_id, lang),
    CONSTRAINT chk_seo_slug CHECK (slug = '' OR slug ~ '^[a-z0-9]+(?:-[a-z0-9]+)*$')
);

COMMENT ON TABLE seo IS 'SEO / OpenGraph / Twitter / JSON-LD for catalog entities.';
COMMENT ON COLUMN seo.jsonld IS 'Structured data JSON-LD payload (Product, BreadcrumbList, etc.).';
COMMENT ON COLUMN seo.canonical_url IS 'Canonical URL; empty means derive from slug/channel.';
