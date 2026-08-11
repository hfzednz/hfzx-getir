-- schedules
CREATE TABLE IF NOT EXISTS schedules (
    id               UUID PRIMARY KEY,
    tenant_id        UUID NOT NULL,
    principal_id     UUID NOT NULL,
    channel          TEXT NOT NULL,
    priority         TEXT NOT NULL,
    template_key     TEXT NOT NULL DEFAULT '',
    locale           TEXT NOT NULL DEFAULT 'en',
    recipient        TEXT NOT NULL DEFAULT '',
    subject          TEXT NOT NULL DEFAULT '',
    body             TEXT NOT NULL DEFAULT '',
    vars_json        JSONB NOT NULL DEFAULT '{}',
    idempotency_key  TEXT NOT NULL DEFAULT '',
    send_at          TIMESTAMPTZ NOT NULL,
    status           TEXT NOT NULL DEFAULT 'pending',
    message_id       UUID NULL,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);
