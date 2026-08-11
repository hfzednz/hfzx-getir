-- Inter/intra warehouse transfers with line items.
CREATE TYPE transfer_status AS ENUM (
    'draft',
    'pending_approval',
    'approved',
    'in_transit',
    'completed',
    'cancelled'
);

CREATE TABLE transfers (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id           UUID NOT NULL,
    code                TEXT NOT NULL DEFAULT '',
    from_warehouse_id   UUID NOT NULL REFERENCES warehouses (id) ON DELETE RESTRICT,
    to_warehouse_id     UUID NOT NULL REFERENCES warehouses (id) ON DELETE RESTRICT,
    from_location_id    UUID REFERENCES locations (id) ON DELETE RESTRICT,
    to_location_id      UUID REFERENCES locations (id) ON DELETE RESTRICT,
    status              transfer_status NOT NULL DEFAULT 'draft',
    requested_by        UUID,
    approved_by         UUID,
    reason              TEXT NOT NULL DEFAULT '',
    metadata            JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    approved_at         TIMESTAMPTZ,
    shipped_at          TIMESTAMPTZ,
    completed_at        TIMESTAMPTZ,
    cancelled_at        TIMESTAMPTZ,

    CONSTRAINT chk_transfers_distinct_sites CHECK (
        from_warehouse_id <> to_warehouse_id
        OR from_location_id IS DISTINCT FROM to_location_id
    )
);

CREATE TABLE transfer_lines (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transfer_id     UUID NOT NULL REFERENCES transfers (id) ON DELETE CASCADE,
    variant_id      UUID NOT NULL,
    sku_code        TEXT NOT NULL DEFAULT '',
    lot_id          UUID REFERENCES stock_lots (id) ON DELETE RESTRICT,
    qty_requested   BIGINT NOT NULL CHECK (qty_requested > 0),
    qty_shipped     BIGINT NOT NULL DEFAULT 0 CHECK (qty_shipped >= 0),
    qty_received    BIGINT NOT NULL DEFAULT 0 CHECK (qty_received >= 0),
    metadata        JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_transfer_lines_shipped CHECK (qty_shipped <= qty_requested),
    CONSTRAINT chk_transfer_lines_received CHECK (qty_received <= qty_shipped),
    CONSTRAINT chk_transfer_lines_sku CHECK (char_length(sku_code) <= 128)
);

-- Late FK from movements.transfer_id once transfers exist.
ALTER TABLE stock_movements
    ADD CONSTRAINT fk_stock_movements_transfer
    FOREIGN KEY (transfer_id) REFERENCES transfers (id) ON DELETE SET NULL;

COMMENT ON TABLE transfers IS 'Stock transfer header between warehouses/locations.';
COMMENT ON TABLE transfer_lines IS 'Transfer line quantities; variant_id is opaque catalog ref.';
COMMENT ON COLUMN transfers.status IS 'draft | pending_approval | approved | in_transit | completed | cancelled.';
