-- search documents metadata (OpenSearch is primary retrieval store)
CREATE TABLE IF NOT EXISTS product_documents (
    tenant_id        UUID NOT NULL,
    product_id       UUID NOT NULL,
    variant_id       UUID,
    sku              TEXT NOT NULL DEFAULT '',
    title            TEXT NOT NULL,
    description      TEXT NOT NULL DEFAULT '',
    brand_id         UUID,
    brand_name       TEXT NOT NULL DEFAULT '',
    category_ids     UUID[] NOT NULL DEFAULT '{}',
    category_path    TEXT[] NOT NULL DEFAULT '{}',
    tags             TEXT[] NOT NULL DEFAULT '{}',
    attributes       JSONB NOT NULL DEFAULT '{}',
    price_minor      BIGINT NOT NULL DEFAULT 0,
    compare_at_minor BIGINT NOT NULL DEFAULT 0,
    discount_pct     DOUBLE PRECISION NOT NULL DEFAULT 0,
    currency         TEXT NOT NULL DEFAULT 'TRY',
    available        BOOLEAN NOT NULL DEFAULT TRUE,
    warehouse_ids    UUID[] NOT NULL DEFAULT '{}',
    city_id          UUID,
    rating_avg       DOUBLE PRECISION NOT NULL DEFAULT 0,
    review_count     INT NOT NULL DEFAULT 0,
    popularity       DOUBLE PRECISION NOT NULL DEFAULT 0,
    freshness_score  DOUBLE PRECISION NOT NULL DEFAULT 0,
    profit_score     DOUBLE PRECISION NOT NULL DEFAULT 0,
    delivery_eta_min INT NOT NULL DEFAULT 0,
    image_ref        TEXT NOT NULL DEFAULT '',
    version          BIGINT NOT NULL DEFAULT 1,
    indexed_at       TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (tenant_id, product_id)
);

CREATE INDEX IF NOT EXISTS idx_docs_brand ON product_documents (tenant_id, brand_id);
CREATE INDEX IF NOT EXISTS idx_docs_available ON product_documents (tenant_id, available);
