-- Orders: canonical OMS aggregate header. Opaque refs only — no stock ledger,
-- PSP charges, warehouse tasks, or courier assignment owned here.
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TYPE order_status AS ENUM (
    'draft',
    'pending_payment',
    'payment_processing',
    'payment_failed',
    'inventory_reservation',
    'inventory_failed',
    'warehouse_assigned',
    'picking',
    'packing',
    'ready_for_dispatch',
    'courier_assigned',
    'out_for_delivery',
    'delivered',
    'completed',
    'cancelled',
    'refund_pending',
    'refunded',
    'failed',
    'archived'
);

CREATE TYPE order_type AS ENUM (
    'instant',
    'scheduled',
    'express',
    'pickup',
    'gift',
    'subscription',
    'corporate',
    'multi_warehouse',
    'split',
    'replacement'
);

CREATE TABLE orders (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id               UUID NOT NULL,
    customer_principal_id   UUID NOT NULL,
    status                  order_status NOT NULL DEFAULT 'draft',
    type                    order_type NOT NULL DEFAULT 'instant',
    currency                CHAR(3) NOT NULL,
    -- Money snapshots in integer minor units (never float).
    subtotal_minor          BIGINT NOT NULL DEFAULT 0,
    discount_minor          BIGINT NOT NULL DEFAULT 0,
    tax_minor               BIGINT NOT NULL DEFAULT 0,
    shipping_minor          BIGINT NOT NULL DEFAULT 0,
    tip_minor               BIGINT NOT NULL DEFAULT 0,
    total_minor             BIGINT NOT NULL DEFAULT 0,
    address_snapshot        JSONB NOT NULL DEFAULT '{}'::jsonb,
    notes                   TEXT NOT NULL DEFAULT '',
    gift                    JSONB,
    priority                INT NOT NULL DEFAULT 0,
    warehouse_ids           UUID[] NOT NULL DEFAULT '{}',
    version                 BIGINT NOT NULL DEFAULT 1,
    idempotency_key         TEXT NOT NULL,
    scheduled_at            TIMESTAMPTZ,
    placed_at               TIMESTAMPTZ,
    -- Opaque integration refs (no FK across services).
    payment_intent_ref      TEXT NOT NULL DEFAULT '',
    reservation_ref         TEXT NOT NULL DEFAULT '',
    courier_ref             TEXT NOT NULL DEFAULT '',
    parent_order_id         UUID REFERENCES orders (id) ON DELETE SET NULL,
    metadata                JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    cancelled_at            TIMESTAMPTZ,
    completed_at            TIMESTAMPTZ,
    archived_at             TIMESTAMPTZ,

    CONSTRAINT uq_orders_idempotency_key UNIQUE (idempotency_key),
    CONSTRAINT chk_orders_currency CHECK (currency ~ '^[A-Z]{3}$'),
    CONSTRAINT chk_orders_idempotency_key CHECK (idempotency_key <> ''),
    CONSTRAINT chk_orders_money_nonneg CHECK (
        subtotal_minor >= 0
        AND discount_minor >= 0
        AND tax_minor >= 0
        AND shipping_minor >= 0
        AND tip_minor >= 0
        AND total_minor >= 0
    ),
    CONSTRAINT chk_orders_version CHECK (version >= 1)
);

COMMENT ON TABLE orders IS 'OMS order aggregate header; orchestrates via opaque refs + sagas.';
COMMENT ON COLUMN orders.customer_principal_id IS 'Opaque identity-service principal_id; no FK.';
COMMENT ON COLUMN orders.currency IS 'ISO-4217 currency code; amounts are minor units.';
COMMENT ON COLUMN orders.address_snapshot IS 'Immutable delivery/pickup address snapshot at place time.';
COMMENT ON COLUMN orders.gift IS 'Optional gift wrapping / message payload.';
COMMENT ON COLUMN orders.warehouse_ids IS 'Opaque warehouse-service ids involved in fulfillment splits.';
COMMENT ON COLUMN orders.payment_intent_ref IS 'Opaque payment-service intent/charge id.';
COMMENT ON COLUMN orders.reservation_ref IS 'Opaque inventory-service reservation id.';
COMMENT ON COLUMN orders.courier_ref IS 'Opaque dispatch-service assignment id after handoff.';
COMMENT ON COLUMN orders.idempotency_key IS 'Client/command idempotency key; globally unique.';
COMMENT ON COLUMN orders.version IS 'Optimistic concurrency version.';
