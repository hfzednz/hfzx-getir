-- Stock movements: append-only ledger of quantity changes.
CREATE TYPE movement_type AS ENUM (
    'receipt',
    'purchase_receipt',
    'sale',
    'courier_pickup',
    'transfer_out',
    'transfer_in',
    'adjust',
    'count',
    'damage',
    'return_in',
    'supplier_return',
    'waste',
    'manual'
);

CREATE TABLE stock_movements (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id           UUID NOT NULL,
    warehouse_id        UUID NOT NULL REFERENCES warehouses (id) ON DELETE RESTRICT,
    balance_id          UUID REFERENCES stock_balances (id) ON DELETE RESTRICT,
    variant_id          UUID NOT NULL,
    sku_code            TEXT NOT NULL DEFAULT '',
    location_id         UUID REFERENCES locations (id) ON DELETE RESTRICT,
    lot_id              UUID REFERENCES stock_lots (id) ON DELETE RESTRICT,
    type                movement_type NOT NULL,
    -- Signed delta applied to on_hand (or related bucket per type).
    qty                 BIGINT NOT NULL CHECK (qty <> 0),
    -- Balance snapshot after this movement.
    on_hand_after       BIGINT NOT NULL CHECK (on_hand_after >= 0),
    reserved_after      BIGINT NOT NULL CHECK (reserved_after >= 0),
    blocked_after       BIGINT NOT NULL CHECK (blocked_after >= 0),
    incoming_after      BIGINT NOT NULL CHECK (incoming_after >= 0),
    in_transit_after    BIGINT NOT NULL CHECK (in_transit_after >= 0),
    idempotency_key     TEXT NOT NULL,
    actor_id            UUID,
    reason              TEXT NOT NULL DEFAULT '',
    external_ref        TEXT NOT NULL DEFAULT '',
    reservation_id      UUID REFERENCES reservations (id) ON DELETE SET NULL,
    transfer_id         UUID,
    metadata            JSONB NOT NULL DEFAULT '{}'::jsonb,
    occurred_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_stock_movements_idempotency UNIQUE (idempotency_key),
    CONSTRAINT chk_stock_movements_sku CHECK (char_length(sku_code) <= 128)
);

COMMENT ON TABLE stock_movements IS 'Append-only quantity ledger; idempotency_key enforces once-only apply.';
COMMENT ON COLUMN stock_movements.qty IS 'Signed quantity delta for the movement type.';
COMMENT ON COLUMN stock_movements.idempotency_key IS 'Caller-supplied unique key; retries must reuse same key.';
COMMENT ON COLUMN stock_movements.variant_id IS 'Opaque catalog variant UUID.';
