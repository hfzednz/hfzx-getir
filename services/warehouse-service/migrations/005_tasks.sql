-- Work tasks: pick/pack/dispatch/qc and ancillary work.
CREATE TYPE task_type AS ENUM (
    'pick',
    'pack',
    'dispatch',
    'qc',
    'replenish',
    'clean',
    'maintenance',
    'emergency'
);

CREATE TYPE task_status AS ENUM (
    'queued',
    'claimed',
    'in_progress',
    'completed',
    'cancelled',
    'escalated'
);

CREATE TABLE tasks (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id           UUID NOT NULL,
    warehouse_id        UUID NOT NULL REFERENCES warehouses (id) ON DELETE RESTRICT,
    fulfillment_id      UUID REFERENCES fulfillment_orders (id) ON DELETE SET NULL,
    station_id          UUID REFERENCES stations (id) ON DELETE SET NULL,
    type                task_type NOT NULL,
    status              task_status NOT NULL DEFAULT 'queued',
    assignee_id         UUID,
    priority            INT NOT NULL DEFAULT 0,
    wave_id             UUID,
    batch_id            UUID,
    sla_deadline        TIMESTAMPTZ,
    claimed_at          TIMESTAMPTZ,
    started_at          TIMESTAMPTZ,
    completed_at        TIMESTAMPTZ,
    cancelled_at        TIMESTAMPTZ,
    escalated_at        TIMESTAMPTZ,
    escalation_note     TEXT NOT NULL DEFAULT '',
    history             JSONB NOT NULL DEFAULT '[]'::jsonb,
    metadata            JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

COMMENT ON TABLE tasks IS 'Warehouse work queue; claim only allowed from queued.';
COMMENT ON COLUMN tasks.type IS 'pick | pack | dispatch | qc | replenish | clean | maintenance | emergency.';
COMMENT ON COLUMN tasks.status IS 'queued → claimed → in_progress → completed | cancelled | escalated.';
COMMENT ON COLUMN tasks.assignee_id IS 'Opaque employee/principal UUID.';
COMMENT ON COLUMN tasks.wave_id IS 'Opaque wave grouping id when strategy=wave.';
COMMENT ON COLUMN tasks.batch_id IS 'Opaque batch grouping id when strategy=batch.';
