-- Transactional outbox for checkout.lifecycle events.
CREATE TYPE outbox_status AS ENUM (
    'pending',
    'published',
    'failed'
);

CREATE TABLE checkout_outbox (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    session_id      UUID NOT NULL REFERENCES checkout_sessions (id) ON DELETE CASCADE,
    topic           TEXT NOT NULL,
    key             TEXT NOT NULL,
    payload         JSONB NOT NULL DEFAULT '{}'::jsonb,
    status          outbox_status NOT NULL DEFAULT 'pending',
    attempts        INT NOT NULL DEFAULT 0,
    last_error      TEXT NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    published_at    TIMESTAMPTZ
);
