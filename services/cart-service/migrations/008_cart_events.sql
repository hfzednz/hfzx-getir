-- Append-only cart lifecycle timeline.
CREATE TABLE cart_events (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cart_id      UUID NOT NULL REFERENCES carts (id) ON DELETE CASCADE,
    tenant_id    UUID NOT NULL,
    type         TEXT NOT NULL,
    payload      JSONB NOT NULL DEFAULT '{}'::jsonb,
    actor_id     UUID,
    actor_type   TEXT NOT NULL DEFAULT '',
    occurred_at  TIMESTAMPTZ NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_cart_events_type CHECK (type <> '')
);

COMMENT ON TABLE cart_events IS 'Append-only cart.lifecycle timeline projection.';
