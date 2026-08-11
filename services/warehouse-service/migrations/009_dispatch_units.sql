-- Dispatch units ready for courier handoff.
CREATE TYPE dispatch_unit_status AS ENUM (
    'queued',
    'verified',
    'handed_off',
    'failed_pickup'
);

CREATE TABLE dispatch_units (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    warehouse_id    UUID NOT NULL REFERENCES warehouses (id) ON DELETE RESTRICT,
    fulfillment_id  UUID NOT NULL REFERENCES fulfillment_orders (id) ON DELETE RESTRICT,
    task_id         UUID REFERENCES tasks (id) ON DELETE SET NULL,
    pack_session_id UUID REFERENCES pack_sessions (id) ON DELETE SET NULL,
    station_id      UUID REFERENCES stations (id) ON DELETE SET NULL,
    label_id        UUID,
    package_code    TEXT NOT NULL DEFAULT '',
    tracking_code   TEXT NOT NULL DEFAULT '',
    -- Opaque courier/dispatch-service assignment reference.
    courier_ref     TEXT NOT NULL DEFAULT '',
    status          dispatch_unit_status NOT NULL DEFAULT 'queued',
    verified_at     TIMESTAMPTZ,
    handed_off_at   TIMESTAMPTZ,
    failed_at       TIMESTAMPTZ,
    metadata        JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_dispatch_units_package_code UNIQUE (warehouse_id, package_code),
    CONSTRAINT chk_dispatch_units_code CHECK (package_code <> '' OR tracking_code <> '')
);

COMMENT ON TABLE dispatch_units IS 'Packed unit in dispatch queue; courier_ref is opaque from dispatch-service.';
COMMENT ON COLUMN dispatch_units.package_code IS 'Physical package identifier for verify/handoff QR.';
COMMENT ON COLUMN dispatch_units.tracking_code IS 'Shipping/label tracking code.';
COMMENT ON COLUMN dispatch_units.courier_ref IS 'Opaque courier assignment ref; not an FK.';
COMMENT ON COLUMN dispatch_units.status IS 'queued | verified | handed_off | failed_pickup.';
