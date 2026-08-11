-- Transactional outbox + admin audit for wallet money mutations.
CREATE TYPE outbox_status AS ENUM (
    'pending',
    'published',
    'failed'
);

CREATE TABLE wallet_outbox (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    wallet_id       UUID NOT NULL REFERENCES wallets (id) ON DELETE CASCADE,
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

CREATE TABLE wallet_audit (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    wallet_id       UUID NOT NULL,
    action          TEXT NOT NULL,
    actor_id        UUID,
    amount_minor    BIGINT NOT NULL DEFAULT 0,
    currency        CHAR(3) NOT NULL DEFAULT '',
    detail          JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
