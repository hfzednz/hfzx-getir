-- Pack session at a station for a pack task.
CREATE TYPE pack_session_status AS ENUM (
    'queued',
    'claimed',
    'verified',
    'sealed',
    'labeled',
    'completed',
    'cancelled'
);

CREATE TABLE pack_sessions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    task_id         UUID NOT NULL REFERENCES tasks (id) ON DELETE CASCADE,
    warehouse_id    UUID NOT NULL REFERENCES warehouses (id) ON DELETE RESTRICT,
    station_id      UUID REFERENCES stations (id) ON DELETE RESTRICT,
    fulfillment_id  UUID REFERENCES fulfillment_orders (id) ON DELETE SET NULL,
    packer_id       UUID,
    status          pack_session_status NOT NULL DEFAULT 'queued',
    expected_weight_g BIGINT NOT NULL DEFAULT 0,
    weight_tolerance  BIGINT NOT NULL DEFAULT 50,
    actual_weight_g   BIGINT,
    weight_g        INT CHECK (weight_g IS NULL OR weight_g >= 0),
    length_mm       INT CHECK (length_mm IS NULL OR length_mm >= 0),
    width_mm        INT CHECK (width_mm IS NULL OR width_mm >= 0),
    height_mm       INT CHECK (height_mm IS NULL OR height_mm >= 0),
    materials       JSONB NOT NULL DEFAULT '[]'::jsonb,
    cold_chain      BOOLEAN NOT NULL DEFAULT FALSE,
    fragile         BOOLEAN NOT NULL DEFAULT FALSE,
    hazard          BOOLEAN NOT NULL DEFAULT FALSE,
    sealed_at       TIMESTAMPTZ,
    labeled_at      TIMESTAMPTZ,
    label_id        UUID,
    label_payload   JSONB NOT NULL DEFAULT '{}'::jsonb,
    started_at      TIMESTAMPTZ,
    completed_at    TIMESTAMPTZ,
    metadata        JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_pack_sessions_task UNIQUE (task_id)
);

COMMENT ON TABLE pack_sessions IS 'Pack run: materials, weight/dim checks, seal, label payload.';
COMMENT ON COLUMN pack_sessions.materials IS 'JSON array of packing materials used.';
COMMENT ON COLUMN pack_sessions.label_payload IS 'Label print intent / payload snapshot after seal.';
