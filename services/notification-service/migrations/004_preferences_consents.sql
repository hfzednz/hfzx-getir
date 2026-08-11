-- preferences and consents
CREATE TABLE IF NOT EXISTS preferences (
    id            UUID PRIMARY KEY,
    tenant_id     UUID NOT NULL,
    principal_id  UUID NOT NULL,
    channel_opt_out JSONB NOT NULL DEFAULT '{}',
    quiet_start   INT NOT NULL DEFAULT 0,
    quiet_end     INT NOT NULL DEFAULT 0,
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, principal_id)
);

CREATE TABLE IF NOT EXISTS consents (
    id            UUID PRIMARY KEY,
    tenant_id     UUID NOT NULL,
    principal_id  UUID NOT NULL,
    purpose       TEXT NOT NULL,
    granted       BOOLEAN NOT NULL,
    source        TEXT NOT NULL DEFAULT '',
    recorded_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);
