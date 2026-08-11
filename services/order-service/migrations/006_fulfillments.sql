-- Fulfillment split units: OMS projection of warehouse work (opaque refs only).
CREATE TYPE fulfillment_unit_status AS ENUM (
    'pending',
    'assigned',
    'picking',
    'packing',
    'ready',
    'dispatched',
    'delivered',
    'cancelled',
    'failed'
);

CREATE TABLE fulfillments (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id            UUID NOT NULL REFERENCES orders (id) ON DELETE CASCADE,
    tenant_id           UUID NOT NULL,
    warehouse_id        UUID NOT NULL,
    status              fulfillment_unit_status NOT NULL DEFAULT 'pending',
    -- Opaque inventory-service reservation id for this split unit.
    reservation_id      TEXT NOT NULL DEFAULT '',
    -- Opaque warehouse-service fulfillment order id.
    fulfillment_ref     TEXT NOT NULL DEFAULT '',
    priority            INT NOT NULL DEFAULT 0,
    metadata            JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    cancelled_at        TIMESTAMPTZ,
    completed_at        TIMESTAMPTZ
);

COMMENT ON TABLE fulfillments IS 'Split fulfillment units per warehouse; status is OMS projection only.';
COMMENT ON COLUMN fulfillments.warehouse_id IS 'Opaque warehouse-service warehouse id.';
COMMENT ON COLUMN fulfillments.reservation_id IS 'Opaque inventory reservation id (soft/hard).';
COMMENT ON COLUMN fulfillments.fulfillment_ref IS 'Opaque warehouse-service fulfillment order id.';
COMMENT ON COLUMN fulfillments.status IS 'Projected WH pipeline status — no pick/pack tasks owned here.';
