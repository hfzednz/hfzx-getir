-- Delivery timeline projections (opaque order_id).
CREATE TYPE timeline_event_type AS ENUM (
    'LocationUpdated', 'Arrived', 'GeofenceEnter', 'GeofenceExit', 'Custom'
);

CREATE TABLE delivery_timelines (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id     UUID NOT NULL,
    order_id      UUID NOT NULL,
    courier_id    UUID,
    type          timeline_event_type NOT NULL,
    lat           DOUBLE PRECISION,
    lon           DOUBLE PRECISION,
    message       TEXT NOT NULL DEFAULT '',
    meta          JSONB NOT NULL DEFAULT '{}'::jsonb,
    occurred_at   TIMESTAMPTZ NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

COMMENT ON TABLE delivery_timelines IS 'Order delivery timeline projection; no OMS ownership.';
