-- Active pick session bound to a pick task.
CREATE TABLE pick_sessions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id         UUID NOT NULL REFERENCES tasks (id) ON DELETE CASCADE,
    warehouse_id    UUID NOT NULL REFERENCES warehouses (id) ON DELETE RESTRICT,
    fulfillment_id  UUID REFERENCES fulfillment_orders (id) ON DELETE SET NULL,
    strategy        pick_strategy NOT NULL DEFAULT 'single',
    -- Ordered pick route: [{line_id, location_code, seq, ...}]
    route           JSONB NOT NULL DEFAULT '[]'::jsonb,
    started_at      TIMESTAMPTZ,
    completed_at    TIMESTAMPTZ,
    metadata        JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_pick_sessions_task UNIQUE (task_id)
);

COMMENT ON TABLE pick_sessions IS 'Pick run for a task; route is AI/optimizer ordered line hints.';
COMMENT ON COLUMN pick_sessions.route IS 'JSON array of ordered pick steps with opaque location_code hints.';
COMMENT ON COLUMN pick_sessions.strategy IS 'single | batch | wave | zone | cluster | priority | express.';
