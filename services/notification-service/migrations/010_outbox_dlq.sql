-- outbox and dlq
CREATE TABLE IF NOT EXISTS outbox (
    id           UUID PRIMARY KEY,
    tenant_id    UUID NOT NULL,
    message_id   UUID NOT NULL,
    topic        TEXT NOT NULL,
    key          TEXT NOT NULL,
    payload      JSONB NOT NULL DEFAULT '{}',
    status       TEXT NOT NULL DEFAULT 'pending',
    attempts     INT NOT NULL DEFAULT 0,
    last_error   TEXT NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    published_at TIMESTAMPTZ NULL
);

CREATE TABLE IF NOT EXISTS dlq (
    id          UUID PRIMARY KEY,
    tenant_id   UUID NOT NULL,
    message_id  UUID NOT NULL,
    reason      TEXT NOT NULL,
    payload     JSONB NOT NULL DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
