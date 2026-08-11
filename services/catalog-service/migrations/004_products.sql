-- Master products: catalog truth only — no stock qty, no sell prices, no orders.
CREATE TYPE product_kind AS ENUM (
    'standard',
    'bundle',
    'kit',
    'pack',
    'subscription',
    'digital',
    'gift',
    'seasonal',
    'limited'
);

CREATE TYPE product_status AS ENUM (
    'draft',
    'pending_review',
    'approved',
    'published',
    'hidden',
    'archived',
    'deleted',
    'scheduled'
);

CREATE TABLE products (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    brand_id        UUID REFERENCES brands (id) ON DELETE SET NULL,
    kind            product_kind NOT NULL DEFAULT 'standard',
    status          product_status NOT NULL DEFAULT 'draft',
    slug            TEXT NOT NULL,
    sku_code        TEXT NOT NULL DEFAULT '',
    external_ref    TEXT NOT NULL DEFAULT '',
    gtin_base       TEXT NOT NULL DEFAULT '',
    manufacturer_sku TEXT NOT NULL DEFAULT '',
    metadata        JSONB NOT NULL DEFAULT '{}'::jsonb,
    scheduled_at    TIMESTAMPTZ,
    published_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_products_tenant_slug UNIQUE (tenant_id, slug),
    CONSTRAINT chk_products_slug CHECK (slug ~ '^[a-z0-9]+(?:-[a-z0-9]+)*$'),
    CONSTRAINT chk_products_scheduled CHECK (
        status <> 'scheduled' OR scheduled_at IS NOT NULL
    )
);

-- Wire product_categories.product_id FK deferred from 002.
ALTER TABLE product_categories
    ADD CONSTRAINT fk_product_categories_product
    FOREIGN KEY (product_id) REFERENCES products (id) ON DELETE CASCADE;

COMMENT ON TABLE products IS 'Master product aggregate; inventory qty and sell prices are out of scope.';
COMMENT ON COLUMN products.kind IS 'standard | bundle | kit | pack | subscription | digital | gift | seasonal | limited.';
COMMENT ON COLUMN products.status IS 'draft → pending_review → approved → published; also hidden|archived|deleted|scheduled.';
COMMENT ON COLUMN products.external_ref IS 'External ERP/PIM product identifier.';
COMMENT ON COLUMN products.scheduled_at IS 'When status=scheduled, publish time; pricing/stock still external.';
