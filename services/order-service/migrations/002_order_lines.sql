-- Order lines: priced line snapshots. Variant/SKU/warehouse are opaque refs.
CREATE TABLE order_lines (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id            UUID NOT NULL REFERENCES orders (id) ON DELETE CASCADE,
    tenant_id           UUID NOT NULL,
    variant_id          UUID NOT NULL,
    sku_code            TEXT NOT NULL DEFAULT '',
    title_snapshot      TEXT NOT NULL DEFAULT '',
    qty                 INT NOT NULL,
    unit_price_minor    BIGINT NOT NULL DEFAULT 0,
    discounts_minor     BIGINT NOT NULL DEFAULT 0,
    tax_minor           BIGINT NOT NULL DEFAULT 0,
    line_total_minor    BIGINT NOT NULL DEFAULT 0,
    warehouse_id        UUID,
    sort_order          INT NOT NULL DEFAULT 0,
    metadata            JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_order_lines_qty CHECK (qty > 0),
    CONSTRAINT chk_order_lines_money_nonneg CHECK (
        unit_price_minor >= 0
        AND discounts_minor >= 0
        AND tax_minor >= 0
        AND line_total_minor >= 0
    )
);

COMMENT ON TABLE order_lines IS 'Order line snapshots; pricing frozen at place; no catalog ownership.';
COMMENT ON COLUMN order_lines.variant_id IS 'Opaque catalog-service variant id.';
COMMENT ON COLUMN order_lines.sku_code IS 'SKU code snapshot for display/ops.';
COMMENT ON COLUMN order_lines.title_snapshot IS 'Product/variant title at place time.';
COMMENT ON COLUMN order_lines.warehouse_id IS 'Nullable opaque warehouse for split/multi-warehouse lines.';
COMMENT ON COLUMN order_lines.unit_price_minor IS 'Unit sell price in minor units.';
COMMENT ON COLUMN order_lines.discounts_minor IS 'Line-level discount in minor units.';
