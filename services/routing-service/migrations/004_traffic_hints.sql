-- Traffic hints: regional traffic multipliers used in ETA.
CREATE TABLE traffic_hints (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id    UUID NOT NULL,
    region_key   TEXT NOT NULL DEFAULT '',
    lat          DOUBLE PRECISION NOT NULL,
    lon          DOUBLE PRECISION NOT NULL,
    radius_m     DOUBLE PRECISION NOT NULL DEFAULT 2000,
    factor       DOUBLE PRECISION NOT NULL,
    valid_from   TIMESTAMPTZ NOT NULL DEFAULT now(),
    valid_until  TIMESTAMPTZ NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_traffic_hints_factor CHECK (factor > 0),
    CONSTRAINT chk_traffic_hints_radius CHECK (radius_m > 0),
    CONSTRAINT chk_traffic_hints_lat CHECK (lat >= -90 AND lat <= 90),
    CONSTRAINT chk_traffic_hints_lon CHECK (lon >= -180 AND lon <= 180),
    CONSTRAINT chk_traffic_hints_window CHECK (valid_until > valid_from)
);

COMMENT ON TABLE traffic_hints IS 'Regional traffic factors applied when origin is within radius.';
