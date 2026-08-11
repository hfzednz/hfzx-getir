-- devices (push tokens)
CREATE TABLE IF NOT EXISTS devices (
    id            UUID PRIMARY KEY,
    tenant_id     UUID NOT NULL,
    principal_id  UUID NOT NULL,
    platform      TEXT NOT NULL,
    token         TEXT NOT NULL,
    locale        TEXT NOT NULL DEFAULT 'en',
    active        BOOLEAN NOT NULL DEFAULT true,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS ux_devices_tenant_token
    ON devices (tenant_id, token);
