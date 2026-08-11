-- Media references: binaries owned by media-service; catalog stores association + CDN URL.
CREATE TYPE media_kind AS ENUM (
    'image',
    'video',
    '360',
    'ar',
    'pdf',
    'audio'
);

CREATE TABLE product_media (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id      UUID NOT NULL REFERENCES products (id) ON DELETE CASCADE,
    variant_id      UUID REFERENCES variants (id) ON DELETE CASCADE,
    tenant_id       UUID NOT NULL,
    media_asset_id  UUID NOT NULL,
    kind            media_kind NOT NULL DEFAULT 'image',
    sort_order      INT NOT NULL DEFAULT 0,
    cdn_url         TEXT NOT NULL DEFAULT '',
    alt_text        TEXT NOT NULL DEFAULT '',
    locale          TEXT NOT NULL DEFAULT '',
    is_primary      BOOLEAN NOT NULL DEFAULT FALSE,
    metadata        JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_product_media_asset UNIQUE (product_id, media_asset_id, locale)
);

COMMENT ON TABLE product_media IS 'Product↔media associations; upload policy owned by media-service.';
COMMENT ON COLUMN product_media.media_asset_id IS 'UUID from media-service; no binary storage here.';
COMMENT ON COLUMN product_media.kind IS 'image | video | 360 | ar | pdf | audio.';
COMMENT ON COLUMN product_media.cdn_url IS 'Resolved CDN URL snapshot for read paths.';
COMMENT ON COLUMN product_media.locale IS 'Empty = all locales; otherwise BCP-47 tag.';
