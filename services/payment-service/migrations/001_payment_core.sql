-- Payment core tables. Money = BIGINT minor units. Tokens only (never PAN).
-- Opaque order_id refs only — no order/cart/inventory engines.
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TYPE payment_intent_status AS ENUM (
    'initiated',
    'authorized',
    'captured',
    'voided',
    'failed',
    'refunded'
);

CREATE TYPE payment_method_type AS ENUM (
    'card',
    'wallet',
    'apple_pay',
    'google_pay'
);

CREATE TYPE attempt_kind AS ENUM (
    'authorize',
    'capture',
    'void',
    'refund'
);

CREATE TYPE attempt_status AS ENUM (
    'pending',
    'success',
    'failed'
);

CREATE TYPE refund_status AS ENUM (
    'requested',
    'completed',
    'failed'
);

CREATE TYPE chargeback_status AS ENUM (
    'opened',
    'won',
    'lost',
    'canceled'
);

CREATE TYPE fraud_decision AS ENUM (
    'allow',
    'challenge',
    'block'
);

CREATE TABLE payment_intents (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id            UUID NOT NULL,
    principal_id         UUID NOT NULL,
    order_id             TEXT NOT NULL DEFAULT '',
    status               payment_intent_status NOT NULL DEFAULT 'initiated',
    amount_minor         BIGINT NOT NULL,
    captured_minor       BIGINT NOT NULL DEFAULT 0,
    refunded_minor       BIGINT NOT NULL DEFAULT 0,
    currency             CHAR(3) NOT NULL,
    method_type          payment_method_type NOT NULL DEFAULT 'card',
    payment_method_id    UUID,
    provider             TEXT NOT NULL DEFAULT '',
    provider_intent_ref  TEXT NOT NULL DEFAULT '',
    idempotency_key      TEXT NOT NULL,
    fraud_score          INT NOT NULL DEFAULT 0,
    fraud_decision       TEXT NOT NULL DEFAULT '',
    failure_reason       TEXT NOT NULL DEFAULT '',
    metadata             JSONB NOT NULL DEFAULT '{}'::jsonb,
    version              BIGINT NOT NULL DEFAULT 1,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    authorized_at        TIMESTAMPTZ,
    captured_at          TIMESTAMPTZ,
    voided_at            TIMESTAMPTZ,
    failed_at            TIMESTAMPTZ,

    CONSTRAINT uq_payment_intents_idempotency UNIQUE (tenant_id, idempotency_key),
    CONSTRAINT chk_payment_currency CHECK (currency ~ '^[A-Z]{3}$'),
    CONSTRAINT chk_payment_amounts CHECK (
        amount_minor > 0
        AND captured_minor >= 0
        AND refunded_minor >= 0
        AND captured_minor <= amount_minor
        AND refunded_minor <= captured_minor
    ),
    CONSTRAINT chk_payment_idempotency_key CHECK (idempotency_key <> ''),
    CONSTRAINT chk_payment_version CHECK (version >= 1)
);

CREATE TABLE payment_attempts (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    intent_id        UUID NOT NULL REFERENCES payment_intents (id) ON DELETE CASCADE,
    tenant_id        UUID NOT NULL,
    kind             attempt_kind NOT NULL,
    status           attempt_status NOT NULL DEFAULT 'pending',
    provider         TEXT NOT NULL DEFAULT '',
    provider_ref     TEXT NOT NULL DEFAULT '',
    amount_minor     BIGINT NOT NULL DEFAULT 0,
    currency         CHAR(3) NOT NULL DEFAULT 'TRY',
    error_code       TEXT NOT NULL DEFAULT '',
    error_message    TEXT NOT NULL DEFAULT '',
    idempotency_key  TEXT NOT NULL DEFAULT '',
    is_failover      BOOLEAN NOT NULL DEFAULT false,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Tokenized payment methods only — PAN never persisted.
CREATE TABLE payment_methods (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id      UUID NOT NULL,
    principal_id   UUID NOT NULL,
    method_type    payment_method_type NOT NULL,
    token          TEXT NOT NULL,
    last4          CHAR(4) NOT NULL DEFAULT '',
    brand          TEXT NOT NULL DEFAULT '',
    exp_month      INT NOT NULL DEFAULT 0,
    exp_year       INT NOT NULL DEFAULT 0,
    provider       TEXT NOT NULL DEFAULT '',
    active         BOOLEAN NOT NULL DEFAULT true,
    metadata       JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_payment_methods_token CHECK (token <> '')
);

ALTER TABLE payment_intents
    ADD CONSTRAINT fk_payment_intents_method
    FOREIGN KEY (payment_method_id) REFERENCES payment_methods (id);

CREATE TABLE refunds (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    intent_id        UUID NOT NULL REFERENCES payment_intents (id) ON DELETE CASCADE,
    tenant_id        UUID NOT NULL,
    amount_minor     BIGINT NOT NULL,
    currency         CHAR(3) NOT NULL,
    status           refund_status NOT NULL DEFAULT 'requested',
    provider         TEXT NOT NULL DEFAULT '',
    provider_ref     TEXT NOT NULL DEFAULT '',
    reason           TEXT NOT NULL DEFAULT '',
    idempotency_key  TEXT NOT NULL,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at     TIMESTAMPTZ,

    CONSTRAINT uq_refunds_idempotency UNIQUE (tenant_id, idempotency_key),
    CONSTRAINT chk_refunds_amount CHECK (amount_minor > 0),
    CONSTRAINT chk_refunds_currency CHECK (currency ~ '^[A-Z]{3}$')
);

CREATE TABLE chargebacks (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    intent_id      UUID NOT NULL REFERENCES payment_intents (id) ON DELETE CASCADE,
    tenant_id      UUID NOT NULL,
    amount_minor   BIGINT NOT NULL,
    currency       CHAR(3) NOT NULL,
    status         chargeback_status NOT NULL DEFAULT 'opened',
    provider       TEXT NOT NULL DEFAULT '',
    provider_ref   TEXT NOT NULL DEFAULT '',
    reason_code    TEXT NOT NULL DEFAULT '',
    reason         TEXT NOT NULL DEFAULT '',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_chargebacks_amount CHECK (amount_minor > 0)
);

CREATE TABLE provider_routes (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id    UUID NOT NULL,
    method_type  payment_method_type NOT NULL,
    currency     CHAR(3) NOT NULL DEFAULT '',
    providers    TEXT[] NOT NULL,
    active       BOOLEAN NOT NULL DEFAULT true,
    priority     INT NOT NULL DEFAULT 100,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_provider_routes_providers CHECK (cardinality(providers) > 0)
);

CREATE TABLE fraud_scores (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    intent_id   UUID NOT NULL REFERENCES payment_intents (id) ON DELETE CASCADE,
    tenant_id   UUID NOT NULL,
    score       INT NOT NULL,
    decision    fraud_decision NOT NULL,
    reasons     TEXT[] NOT NULL DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_fraud_score_range CHECK (score >= 0 AND score <= 100)
);
