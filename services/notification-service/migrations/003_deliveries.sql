-- deliveries / attempts
CREATE TABLE IF NOT EXISTS deliveries (
    id            UUID PRIMARY KEY,
    tenant_id     UUID NOT NULL,
    message_id    UUID NOT NULL REFERENCES messages(id),
    attempt_no    INT NOT NULL,
    provider      TEXT NOT NULL,
    status        TEXT NOT NULL,
    provider_ref  TEXT NOT NULL DEFAULT '',
    error         TEXT NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
