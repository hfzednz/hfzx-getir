-- provider routes
CREATE TABLE IF NOT EXISTS provider_routes (
    id          UUID PRIMARY KEY,
    tenant_id   UUID NOT NULL,
    channel     TEXT NOT NULL,
    provider    TEXT NOT NULL,
    priority    INT NOT NULL DEFAULT 0,
    enabled     BOOLEAN NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS ux_provider_routes_tenant_channel_provider
    ON provider_routes (tenant_id, channel, provider);
