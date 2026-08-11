-- Routes: ordered multi-stop plans owned by routing-service.
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TYPE route_status AS ENUM ('draft', 'optimized', 'active', 'completed', 'cancelled');
CREATE TYPE waypoint_kind AS ENUM ('warehouse', 'stop', 'courier');

CREATE TABLE routes (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id         UUID NOT NULL,
    dispatch_id       UUID,
    courier_id        UUID,
    warehouse_id      UUID,
    status            route_status NOT NULL DEFAULT 'draft',
    distance_meters   DOUBLE PRECISION NOT NULL DEFAULT 0,
    duration_seconds  DOUBLE PRECISION NOT NULL DEFAULT 0,
    eta_at            TIMESTAMPTZ,
    traffic_factor    DOUBLE PRECISION NOT NULL DEFAULT 1,
    weather_factor    DOUBLE PRECISION NOT NULL DEFAULT 1,
    speed_mps         DOUBLE PRECISION NOT NULL DEFAULT 8.333333333333334,
    waypoints         JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_routes_distance CHECK (distance_meters >= 0),
    CONSTRAINT chk_routes_duration CHECK (duration_seconds >= 0),
    CONSTRAINT chk_routes_traffic CHECK (traffic_factor >= 0),
    CONSTRAINT chk_routes_weather CHECK (weather_factor >= 0),
    CONSTRAINT chk_routes_speed CHECK (speed_mps > 0)
);

COMMENT ON TABLE routes IS 'Route plans; opaque dispatch/courier/warehouse/order ids only.';
COMMENT ON COLUMN routes.waypoints IS 'Ordered waypoints JSON [{id,sequence,kind,lat,lon,orderId,label,etaAt}].';
