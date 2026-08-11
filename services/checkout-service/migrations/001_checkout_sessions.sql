-- Checkout sessions: orchestration aggregate. No PSP capture, no OMS saga,
-- no inventory ledger — opaque refs + snapshots only. Money = BIGINT minor units.
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TYPE checkout_status AS ENUM (
    'started',
    'validating',
    'ready',
    'blocked',
    'completing',
    'completed',
    'failed',
    'abandoned'
);

CREATE TYPE delivery_option AS ENUM (
    'instant',
    'scheduled',
    'priority',
    'economy',
    'pickup',
    'corporate'
);

CREATE TYPE substitution_policy AS ENUM (
    'allow',
    'ask',
    'refuse'
);

CREATE TABLE checkout_sessions (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id           UUID NOT NULL,
    cart_id             UUID NOT NULL,
    principal_id        UUID NOT NULL,
    status              checkout_status NOT NULL DEFAULT 'started',
    delivery_option     delivery_option,
    address_snapshot    JSONB NOT NULL DEFAULT '{}'::jsonb,
    slot_snapshot       JSONB NOT NULL DEFAULT '{}'::jsonb,
    gift_prefs          JSONB NOT NULL DEFAULT '{}'::jsonb,
    invoice_prefs       JSONB NOT NULL DEFAULT '{}'::jsonb,
    substitutions       substitution_policy NOT NULL DEFAULT 'ask',
    notes               TEXT NOT NULL DEFAULT '',
    tip_minor           BIGINT NOT NULL DEFAULT 0,
    currency            CHAR(3) NOT NULL DEFAULT 'TRY',
    validation_results  JSONB NOT NULL DEFAULT '{}'::jsonb,
    quote_snapshot      JSONB NOT NULL DEFAULT '{}'::jsonb,
    -- Opaque order-service id (no FK across services).
    order_id            TEXT NOT NULL DEFAULT '',
    idempotency_key     TEXT NOT NULL,
    recovery_token      TEXT NOT NULL DEFAULT '',
    city_id             TEXT NOT NULL DEFAULT '',
    coupon_codes        TEXT[] NOT NULL DEFAULT '{}',
    version             BIGINT NOT NULL DEFAULT 1,
    metadata            JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at        TIMESTAMPTZ,
    abandoned_at        TIMESTAMPTZ,
    failed_at           TIMESTAMPTZ,

    CONSTRAINT uq_checkout_sessions_idempotency UNIQUE (tenant_id, idempotency_key),
    CONSTRAINT chk_checkout_currency CHECK (currency ~ '^[A-Z]{3}$'),
    CONSTRAINT chk_checkout_idempotency_key CHECK (idempotency_key <> ''),
    CONSTRAINT chk_checkout_tip_nonneg CHECK (tip_minor >= 0),
    CONSTRAINT chk_checkout_version CHECK (version >= 1)
);

CREATE TABLE checkout_events (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id      UUID NOT NULL REFERENCES checkout_sessions (id) ON DELETE CASCADE,
    tenant_id       UUID NOT NULL,
    type            TEXT NOT NULL,
    payload         JSONB NOT NULL DEFAULT '{}'::jsonb,
    actor_id        UUID,
    actor_type      TEXT NOT NULL DEFAULT '',
    occurred_at     TIMESTAMPTZ NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
