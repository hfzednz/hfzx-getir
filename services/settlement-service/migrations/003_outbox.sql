-- settlement-service: transactional outbox + timeline
CREATE TABLE IF NOT EXISTS settlement_outbox (
    id            UUID PRIMARY KEY,
    tenant_id     UUID NOT NULL,
    batch_id      UUID NOT NULL,
    topic         TEXT NOT NULL,
    message_key   TEXT NOT NULL DEFAULT '',
    payload       JSONB NOT NULL DEFAULT '{}'::jsonb,
    status        TEXT NOT NULL CHECK (status IN ('pending','published','failed')),
    attempts      INT NOT NULL DEFAULT 0,
    last_error    TEXT NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    published_at  TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS settlement_events (
    id            UUID PRIMARY KEY,
    batch_id      UUID NOT NULL,
    tenant_id     UUID NOT NULL,
    event_type    TEXT NOT NULL,
    payload       JSONB NOT NULL DEFAULT '{}'::jsonb,
    actor_id      UUID,
    actor_type    TEXT NOT NULL DEFAULT '',
    occurred_at   TIMESTAMPTZ NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
