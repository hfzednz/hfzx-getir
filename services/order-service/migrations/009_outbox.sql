-- Transactional outbox (AD-11): at-least-once publish to Kafka.
CREATE TABLE outbox (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    order_id        UUID REFERENCES orders (id) ON DELETE SET NULL,
    topic           TEXT NOT NULL,
    key             TEXT NOT NULL DEFAULT '',
    payload         JSONB NOT NULL,
    headers         JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    published_at    TIMESTAMPTZ,
    attempts        INT NOT NULL DEFAULT 0,
    last_error      TEXT NOT NULL DEFAULT '',

    CONSTRAINT chk_outbox_topic CHECK (topic <> ''),
    CONSTRAINT chk_outbox_attempts CHECK (attempts >= 0)
);

COMMENT ON TABLE outbox IS 'Transactional outbox for order.lifecycle and related topics.';
COMMENT ON COLUMN outbox.topic IS 'Kafka topic (e.g. order.lifecycle).';
COMMENT ON COLUMN outbox.key IS 'Partition key (typically order id).';
COMMENT ON COLUMN outbox.payload IS 'Event envelope JSON; camelCase at wire layer.';
COMMENT ON COLUMN outbox.published_at IS 'Null until successfully published; worker sweeps unpublished rows.';
