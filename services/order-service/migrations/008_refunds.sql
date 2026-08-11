-- Refund requests: amount in minor units; payment_refund_ref is opaque PSP/wallet ref.
CREATE TYPE refund_method AS ENUM (
    'wallet',
    'card',
    'store_credit'
);

CREATE TYPE refund_status AS ENUM (
    'pending',
    'authorized',
    'processing',
    'succeeded',
    'failed',
    'cancelled'
);

CREATE TABLE refunds (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id            UUID NOT NULL REFERENCES orders (id) ON DELETE CASCADE,
    tenant_id           UUID NOT NULL,
    return_id           UUID REFERENCES returns (id) ON DELETE SET NULL,
    amount_minor        BIGINT NOT NULL,
    currency            CHAR(3) NOT NULL,
    method              refund_method NOT NULL,
    status              refund_status NOT NULL DEFAULT 'pending',
    reason              TEXT NOT NULL DEFAULT '',
    -- Opaque payment-service refund id.
    payment_refund_ref  TEXT NOT NULL DEFAULT '',
    actor_id            UUID,
    metadata            JSONB NOT NULL DEFAULT '{}'::jsonb,
    requested_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at        TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_refunds_amount CHECK (amount_minor > 0),
    CONSTRAINT chk_refunds_currency CHECK (currency ~ '^[A-Z]{3}$')
);

-- Optional FK from returns.refund_id now that refunds exists.
ALTER TABLE returns
    ADD CONSTRAINT fk_returns_refund
    FOREIGN KEY (refund_id) REFERENCES refunds (id) ON DELETE SET NULL;

COMMENT ON TABLE refunds IS 'OMS refund requests; PSP/wallet execution via payment-service port.';
COMMENT ON COLUMN refunds.amount_minor IS 'Refund amount in integer minor units (never float).';
COMMENT ON COLUMN refunds.method IS 'wallet | card | store_credit.';
COMMENT ON COLUMN refunds.payment_refund_ref IS 'Opaque payment-service refund reference.';
