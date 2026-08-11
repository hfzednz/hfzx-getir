-- Geofence zones: polygon vertices and/or radius geometry.
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TYPE zone_kind AS ENUM ('delivery', 'restricted', 'warehouse');

CREATE TABLE zones (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL,
    name        TEXT NOT NULL,
    city        TEXT NOT NULL DEFAULT '',
    kind        zone_kind NOT NULL,
    vertices    JSONB NOT NULL DEFAULT '[]'::jsonb,
    geojson     JSONB,
    center_lat  DOUBLE PRECISION,
    center_lng  DOUBLE PRECISION,
    radius_m    DOUBLE PRECISION,
    active      BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_zones_radius CHECK (radius_m IS NULL OR radius_m > 0),
    CONSTRAINT chk_zones_geometry CHECK (
        jsonb_array_length(vertices) >= 3
        OR (center_lat IS NOT NULL AND center_lng IS NOT NULL AND radius_m IS NOT NULL AND radius_m > 0)
    )
);

COMMENT ON TABLE zones IS 'Delivery / restricted / warehouse polygons or radius zones.';
