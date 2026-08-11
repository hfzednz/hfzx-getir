-- Append-only order timeline for CQRS read/timeline APIs.
CREATE TABLE order_events (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id        UUID NOT NULL REFERENCES orders (id) ON DELETE CASCADE,
    tenant_id       UUID NOT NULL,
    type            TEXT NOT NULL,
    payload         JSONB NOT NULL DEFAULT '{}'::jsonb,
    actor_id        UUID,
    actor_type      TEXT NOT NULL DEFAULT '',
    occurred_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_order_events_type CHECK (type <> '')
);

COMMENT ON TABLE order_events IS 'Append-only OMS timeline; never updated/deleted in app path.';
COMMENT ON COLUMN order_events.type IS 'Domain event type (e.g. OrderCreated, Delivered).';
COMMENT ON COLUMN order_events.payload IS 'Event payload snapshot (camelCase at wire layer).';
COMMENT ON COLUMN order_events.actor_id IS 'Opaque principal/system actor that caused the event.';
