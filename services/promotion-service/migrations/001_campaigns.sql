-- campaigns
CREATE TABLE IF NOT EXISTS campaigns (
    id              UUID PRIMARY KEY,
    tenant_id       UUID NOT NULL,
    name            TEXT NOT NULL,
    description     TEXT NOT NULL DEFAULT '',
    status          TEXT NOT NULL CHECK (status IN ('draft','scheduled','active','paused','expired')),
    starts_at       TIMESTAMPTZ,
    ends_at         TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    version         BIGINT NOT NULL DEFAULT 1,
    CONSTRAINT campaigns_schedule_chk CHECK (ends_at IS NULL OR starts_at IS NULL OR ends_at > starts_at)
);
