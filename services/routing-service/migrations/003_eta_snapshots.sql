-- ETA snapshots: point-in-time ETA estimates for audit / recalc history.
CREATE TABLE eta_snapshots (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id         UUID NOT NULL,
    route_id          UUID NOT NULL REFERENCES routes (id) ON DELETE CASCADE,
    distance_meters   DOUBLE PRECISION NOT NULL DEFAULT 0,
    duration_seconds  DOUBLE PRECISION NOT NULL DEFAULT 0,
    eta_at            TIMESTAMPTZ NOT NULL,
    traffic_factor    DOUBLE PRECISION NOT NULL DEFAULT 1,
    weather_factor    DOUBLE PRECISION NOT NULL DEFAULT 1,
    speed_mps         DOUBLE PRECISION NOT NULL DEFAULT 8.333333333333334,
    reason            TEXT NOT NULL DEFAULT '',
    captured_at       TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_eta_distance CHECK (distance_meters >= 0),
    CONSTRAINT chk_eta_duration CHECK (duration_seconds >= 0)
);

COMMENT ON TABLE eta_snapshots IS 'Historical ETA captures; Event ETAUpdated → routing.eta.';
