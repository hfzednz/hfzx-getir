-- Stock balances: quantity ledger per (warehouse, variant, optional location).
-- variant_id / sku_code are opaque catalog references — no product content here.
CREATE TABLE stock_balances (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    warehouse_id    UUID NOT NULL REFERENCES warehouses (id) ON DELETE RESTRICT,
    variant_id      UUID NOT NULL,
    sku_code        TEXT NOT NULL DEFAULT '',
    location_id     UUID REFERENCES locations (id) ON DELETE RESTRICT,
    on_hand         BIGINT NOT NULL DEFAULT 0 CHECK (on_hand >= 0),
    reserved        BIGINT NOT NULL DEFAULT 0 CHECK (reserved >= 0),
    blocked         BIGINT NOT NULL DEFAULT 0 CHECK (blocked >= 0),
    incoming        BIGINT NOT NULL DEFAULT 0 CHECK (incoming >= 0),
    in_transit      BIGINT NOT NULL DEFAULT 0 CHECK (in_transit >= 0),
    -- Optimistic concurrency token; bump on every mutate.
    version         BIGINT NOT NULL DEFAULT 1 CHECK (version >= 1),
    safety_min      BIGINT NOT NULL DEFAULT 0 CHECK (safety_min >= 0),
    reorder_point   BIGINT NOT NULL DEFAULT 0 CHECK (reorder_point >= 0),
    max_stock       BIGINT,
    metadata        JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_stock_balances_available CHECK (on_hand >= reserved + blocked),
    CONSTRAINT chk_stock_balances_max_stock CHECK (max_stock IS NULL OR max_stock >= 0),
    CONSTRAINT chk_stock_balances_sku CHECK (char_length(sku_code) <= 128)
);

-- UNIQUE (warehouse_id, variant_id, COALESCE(location_id, nil-uuid))
CREATE UNIQUE INDEX uq_stock_balances_wh_variant_loc
    ON stock_balances (
        warehouse_id,
        variant_id,
        (COALESCE(location_id, '00000000-0000-0000-0000-000000000000'::uuid))
    );

COMMENT ON TABLE stock_balances IS 'Quantity SoT per warehouse/variant/(location); available = on_hand - reserved - blocked.';
COMMENT ON COLUMN stock_balances.variant_id IS 'Opaque catalog variant UUID; no product master content stored.';
COMMENT ON COLUMN stock_balances.sku_code IS 'Opaque catalog SKU string denormalized for lookup only.';
COMMENT ON COLUMN stock_balances.version IS 'Optimistic lock; must bump on every balance mutation.';
COMMENT ON COLUMN stock_balances.location_id IS 'Null = warehouse-level aggregate balance.';
