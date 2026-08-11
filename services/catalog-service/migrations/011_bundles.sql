-- Bundle / kit composition: component variants and quantities (composition only).
CREATE TYPE bundle_composition_type AS ENUM (
    'static',
    'dynamic'
);

CREATE TABLE bundles (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id      UUID NOT NULL REFERENCES products (id) ON DELETE CASCADE,
    tenant_id       UUID NOT NULL,
    composition     bundle_composition_type NOT NULL DEFAULT 'static',
    name            TEXT NOT NULL DEFAULT '',
    metadata        JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_bundles_product UNIQUE (product_id)
);

CREATE TABLE bundle_items (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bundle_id               UUID NOT NULL REFERENCES bundles (id) ON DELETE CASCADE,
    component_variant_id    UUID NOT NULL REFERENCES variants (id) ON DELETE RESTRICT,
    qty                     INT NOT NULL DEFAULT 1 CHECK (qty > 0),
    is_optional             BOOLEAN NOT NULL DEFAULT FALSE,
    sort_order              INT NOT NULL DEFAULT 0,
    metadata                JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_bundle_items_component UNIQUE (bundle_id, component_variant_id)
);

COMMENT ON TABLE bundles IS 'Bundle/kit header for products with kind bundle|kit|pack.';
COMMENT ON COLUMN bundles.composition IS 'static | dynamic; dynamic rules may live in metadata.';
COMMENT ON TABLE bundle_items IS 'Component variants and qty; qty is composition count, NOT warehouse stock.';
COMMENT ON COLUMN bundle_items.qty IS 'Number of component units in the bundle (BOM), not on-hand inventory.';
