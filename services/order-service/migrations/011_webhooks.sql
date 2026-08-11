-- Optional outbound webhook endpoint registrations (stub for partner callbacks).
CREATE TABLE webhooks (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    url             TEXT NOT NULL,
    secret          TEXT NOT NULL DEFAULT '',
    events          TEXT[] NOT NULL DEFAULT '{}',
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    description     TEXT NOT NULL DEFAULT '',
    metadata        JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    disabled_at     TIMESTAMPTZ,

    CONSTRAINT chk_webhooks_url CHECK (url ~ '^https?://')
);

COMMENT ON TABLE webhooks IS 'Optional outbound webhook endpoints; delivery is best-effort via outbox/worker.';
COMMENT ON COLUMN webhooks.events IS 'Subscribed event type names (empty = all order.lifecycle events).';
COMMENT ON COLUMN webhooks.secret IS 'HMAC signing secret; never log in plain text.';
