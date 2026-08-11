-- Saga instances: place | cancel | refund | return orchestration headers.
CREATE TYPE saga_type AS ENUM (
    'place',
    'cancel',
    'refund',
    'return'
);

CREATE TYPE saga_instance_status AS ENUM (
    'pending',
    'running',
    'compensating',
    'succeeded',
    'failed',
    'compensated'
);

CREATE TABLE saga_instances (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id        UUID NOT NULL REFERENCES orders (id) ON DELETE CASCADE,
    tenant_id       UUID NOT NULL,
    saga_type       saga_type NOT NULL,
    status          saga_instance_status NOT NULL DEFAULT 'pending',
    current_step    TEXT NOT NULL DEFAULT '',
    correlation_id  TEXT NOT NULL DEFAULT '',
    idempotency_key TEXT NOT NULL,
    last_error      TEXT NOT NULL DEFAULT '',
    metadata        JSONB NOT NULL DEFAULT '{}'::jsonb,
    started_at      TIMESTAMPTZ,
    completed_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_saga_instances_idempotency UNIQUE (idempotency_key),
    CONSTRAINT chk_saga_instances_idempotency CHECK (idempotency_key <> '')
);

COMMENT ON TABLE saga_instances IS 'OMS saga orchestration header; no 2PC — compensations via steps.';
COMMENT ON COLUMN saga_instances.saga_type IS 'place | cancel | refund | return.';
COMMENT ON COLUMN saga_instances.current_step IS 'Name of the active/last attempted step.';
COMMENT ON COLUMN saga_instances.correlation_id IS 'Cross-service correlation / trace helper.';
