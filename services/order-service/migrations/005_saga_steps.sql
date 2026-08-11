-- Saga steps: per-step progress with retries and compensation status.
CREATE TYPE saga_step_status AS ENUM (
    'pending',
    'succeeded',
    'failed',
    'compensated'
);

CREATE TABLE saga_steps (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    saga_id         UUID NOT NULL REFERENCES saga_instances (id) ON DELETE CASCADE,
    order_id        UUID NOT NULL REFERENCES orders (id) ON DELETE CASCADE,
    tenant_id       UUID NOT NULL,
    name            TEXT NOT NULL,
    status          saga_step_status NOT NULL DEFAULT 'pending',
    attempt         INT NOT NULL DEFAULT 0,
    last_error      TEXT NOT NULL DEFAULT '',
    idempotency_key TEXT NOT NULL,
    compensation_of TEXT NOT NULL DEFAULT '',
    payload         JSONB NOT NULL DEFAULT '{}'::jsonb,
    started_at      TIMESTAMPTZ,
    completed_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_saga_steps_idempotency UNIQUE (idempotency_key),
    CONSTRAINT chk_saga_steps_name CHECK (name <> ''),
    CONSTRAINT chk_saga_steps_idempotency CHECK (idempotency_key <> ''),
    CONSTRAINT chk_saga_steps_attempt CHECK (attempt >= 0)
);

COMMENT ON TABLE saga_steps IS 'Individual saga steps + compensations; timeouts/retries/DLQ driven here.';
COMMENT ON COLUMN saga_steps.name IS 'Validate|SoftReserve|AuthorizePayment|ConfirmHard|StartFulfillment|RequestDispatch|Complete or ReleaseReserve|VoidOrRefundPayment.';
COMMENT ON COLUMN saga_steps.compensation_of IS 'When set, this step compensates the named forward step.';
COMMENT ON COLUMN saga_steps.idempotency_key IS 'Idempotent port invocation key for retries.';
