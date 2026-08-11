-- QC inspections on fulfillment units / packages.
CREATE TYPE qc_result AS ENUM (
    'pending',
    'passed',
    'failed',
    'waived'
);

CREATE TABLE qc_inspections (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    warehouse_id    UUID NOT NULL REFERENCES warehouses (id) ON DELETE RESTRICT,
    fulfillment_id  UUID REFERENCES fulfillment_orders (id) ON DELETE SET NULL,
    task_id         UUID REFERENCES tasks (id) ON DELETE SET NULL,
    station_id      UUID REFERENCES stations (id) ON DELETE SET NULL,
    dispatch_unit_id UUID REFERENCES dispatch_units (id) ON DELETE SET NULL,
    inspector_id    UUID,
    result          qc_result NOT NULL DEFAULT 'pending',
    checklist       JSONB NOT NULL DEFAULT '[]'::jsonb,
    notes           TEXT NOT NULL DEFAULT '',
    defects         JSONB NOT NULL DEFAULT '[]'::jsonb,
    inspected_at    TIMESTAMPTZ,
    metadata        JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

COMMENT ON TABLE qc_inspections IS 'Quality control inspections on fulfill units; not finance/refunds.';
COMMENT ON COLUMN qc_inspections.result IS 'pending | passed | failed | waived.';
COMMENT ON COLUMN qc_inspections.inspector_id IS 'Opaque employee UUID.';
