-- Inbound returns to inventory (customer / courier / warehouse / supplier).
CREATE TYPE return_source AS ENUM (
    'customer',
    'courier',
    'warehouse',
    'supplier'
);

CREATE TYPE return_disposition AS ENUM (
    'restock',
    'quarantine',
    'waste'
);

CREATE TYPE return_status AS ENUM (
    'draft',
    'received',
    'disposed',
    'cancelled'
);

CREATE TABLE inventory_returns (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    warehouse_id    UUID NOT NULL REFERENCES warehouses (id) ON DELETE RESTRICT,
    source          return_source NOT NULL,
    disposition     return_disposition NOT NULL DEFAULT 'restock',
    status          return_status NOT NULL DEFAULT 'draft',
    -- Opaque upstream refs (order/shipment/RMA) — no order aggregate here.
    external_ref    TEXT NOT NULL DEFAULT '',
    actor_id        UUID,
    reason          TEXT NOT NULL DEFAULT '',
    metadata        JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    received_at     TIMESTAMPTZ,
    disposed_at     TIMESTAMPTZ
);

CREATE TABLE inventory_return_lines (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    return_id       UUID NOT NULL REFERENCES inventory_returns (id) ON DELETE CASCADE,
    variant_id      UUID NOT NULL,
    sku_code        TEXT NOT NULL DEFAULT '',
    lot_id          UUID REFERENCES stock_lots (id) ON DELETE RESTRICT,
    location_id     UUID REFERENCES locations (id) ON DELETE RESTRICT,
    qty             BIGINT NOT NULL CHECK (qty > 0),
    disposition     return_disposition,
    condition_notes TEXT NOT NULL DEFAULT '',
    metadata        JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_return_lines_sku CHECK (char_length(sku_code) <= 128)
);

COMMENT ON TABLE inventory_returns IS 'Return-to-stock headers; disposition drives restock/quarantine/waste.';
COMMENT ON TABLE inventory_return_lines IS 'Return line quantities; variant_id is opaque catalog ref.';
COMMENT ON COLUMN inventory_returns.source IS 'customer | courier | warehouse | supplier.';
COMMENT ON COLUMN inventory_returns.disposition IS 'restock | quarantine | waste.';
COMMENT ON COLUMN inventory_returns.external_ref IS 'Opaque order/shipment/RMA id; not an FK.';
