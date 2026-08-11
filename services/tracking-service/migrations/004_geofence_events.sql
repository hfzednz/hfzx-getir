-- Geofence enter/exit events recorded by tracking-service.
CREATE TYPE geofence_event_kind AS ENUM ('enter', 'exit');

CREATE TABLE geofence_events (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id     UUID NOT NULL,
    courier_id    UUID NOT NULL,
    order_id      UUID,
    zone_id       TEXT NOT NULL,
    kind          geofence_event_kind NOT NULL,
    lat           DOUBLE PRECISION NOT NULL,
    lon           DOUBLE PRECISION NOT NULL,
    occurred_at   TIMESTAMPTZ NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

COMMENT ON TABLE geofence_events IS 'Zone enter/exit; zone definitions owned by geofence-service.';
