-- usage counters, simulations, outbox
CREATE TABLE IF NOT EXISTS usage_counters (
    id             UUID PRIMARY KEY,
    tenant_id      UUID NOT NULL,
    promotion_id   UUID NOT NULL REFERENCES promotions(id),
    scope          TEXT NOT NULL CHECK (scope IN ('global','user','order','device')),
    scope_key      TEXT NOT NULL,
    count          INT NOT NULL DEFAULT 0,
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, promotion_id, scope, scope_key)
);

CREATE TABLE IF NOT EXISTS simulations (
    id               UUID PRIMARY KEY,
    tenant_id        UUID NOT NULL,
    request_payload  JSONB NOT NULL DEFAULT '{}',
    result_payload   JSONB NOT NULL DEFAULT '{}',
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS outbox (
    id             UUID PRIMARY KEY,
    tenant_id      UUID NOT NULL,
    aggregate_id   UUID NOT NULL,
    topic          TEXT NOT NULL,
    key            TEXT NOT NULL DEFAULT '',
    payload        JSONB NOT NULL DEFAULT '{}',
    status         TEXT NOT NULL CHECK (status IN ('pending','published','failed')),
    attempts       INT NOT NULL DEFAULT 0,
    last_error     TEXT NOT NULL DEFAULT '',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    published_at   TIMESTAMPTZ
);
