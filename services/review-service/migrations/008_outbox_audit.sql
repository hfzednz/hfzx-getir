CREATE TABLE IF NOT EXISTS outbox_messages (
    id            UUID PRIMARY KEY,
    tenant_id     UUID NOT NULL,
    aggregate_id  UUID NOT NULL,
    topic         TEXT NOT NULL,
    key           TEXT NOT NULL,
    payload       JSONB NOT NULL,
    status        TEXT NOT NULL,
    attempts      INT NOT NULL DEFAULT 0,
    last_error    TEXT NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ NOT NULL,
    updated_at    TIMESTAMPTZ NOT NULL,
    published_at  TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_outbox_pending
    ON outbox_messages (status, created_at) WHERE status = 'pending';

CREATE TABLE IF NOT EXISTS audit_logs (
    id         UUID PRIMARY KEY,
    tenant_id  UUID NOT NULL,
    actor_id   UUID,
    action     TEXT NOT NULL,
    entity_type TEXT NOT NULL,
    entity_id  UUID NOT NULL,
    meta       JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL
);
