-- Fulfillment projection of an opaque order — no order aggregate owned here.
CREATE TYPE fulfillment_status AS ENUM (
    'received',
    'reserved',
    'pick_queued',
    'picking',
    'picked',
    'pack_queued',
    'packing',
    'packed',
    'dispatch_queued',
    'dispatched',
    'cancelled',
    'failed'
);

CREATE TYPE pick_strategy AS ENUM (
    'single',
    'batch',
    'wave',
    'zone',
    'cluster',
    'priority',
    'express'
);

CREATE TABLE fulfillment_orders (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    -- Opaque order-service id (UUID form when available).
    order_id        UUID,
    -- Opaque order-service wire id (string); unique per tenant.
    external_order_id TEXT NOT NULL DEFAULT '',
    warehouse_id    UUID NOT NULL REFERENCES warehouses (id) ON DELETE RESTRICT,
    -- Opaque inventory reservation id when reserved via inventory port.
    reservation_id  UUID,
    status          fulfillment_status NOT NULL DEFAULT 'received',
    priority        INT NOT NULL DEFAULT 0,
    strategy        pick_strategy NOT NULL DEFAULT 'single',
    vip             BOOLEAN NOT NULL DEFAULT FALSE,
    express         BOOLEAN NOT NULL DEFAULT FALSE,
    sla_deadline    TIMESTAMPTZ,
    courier_ref     TEXT NOT NULL DEFAULT '',
    metadata        JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    cancelled_at    TIMESTAMPTZ,
    failed_at       TIMESTAMPTZ,
    completed_at    TIMESTAMPTZ,

    CONSTRAINT uq_fulfillment_orders_order_id UNIQUE (order_id),
    CONSTRAINT uq_fulfillment_orders_external UNIQUE (tenant_id, external_order_id),
    CONSTRAINT chk_fulfillment_orders_order_ref CHECK (
        order_id IS NOT NULL OR external_order_id <> ''
    )
);

COMMENT ON TABLE fulfillment_orders IS 'WH fulfillment projection; order_id/reservation_id are opaque external refs.';
COMMENT ON COLUMN fulfillment_orders.order_id IS 'Opaque order-service UUID — not an FK; no order aggregate here.';
COMMENT ON COLUMN fulfillment_orders.external_order_id IS 'Opaque order wire id (string); unique per tenant.';
COMMENT ON COLUMN fulfillment_orders.reservation_id IS 'Opaque inventory reservation UUID from inventory-service port.';
COMMENT ON COLUMN fulfillment_orders.status IS 'received→reserved→pick_queued→picking→picked→pack_queued→packing→packed→dispatch_queued→dispatched | cancelled|failed.';
COMMENT ON COLUMN fulfillment_orders.strategy IS 'single | batch | wave | zone | cluster | priority | express.';
