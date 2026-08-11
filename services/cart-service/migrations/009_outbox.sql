-- Transactional outbox for cart.lifecycle Kafka publish.
CREATE TYPE outbox_status AS ENUM (
    'pending',
    'published',
    'failed'
);

CREATE TABLE outbox (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id     UUID NOT NULL,
    cart_id       UUID NOT NULL,
    topic         TEXT NOT NULL,
    key           TEXT NOT NULL DEFAULT '',
    payload       JSONB NOT NULL DEFAULT '{}'::jsonb,
    status        outbox_status NOT NULL DEFAULT 'pending',
    attempts      INT NOT NULL DEFAULT 0,
    last_error    TEXT NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    published_at  TIMESTAMPTZ,

    CONSTRAINT chk_outbox_topic CHECK (topic <> '')
);

COMMENT ON TABLE outbox IS 'Transactional outbox; cart_id is opaque correlation (no cross-service FK).';
