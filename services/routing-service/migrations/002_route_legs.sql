-- Route legs: segments between consecutive waypoints.
CREATE TABLE route_legs (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id         UUID NOT NULL,
    route_id          UUID NOT NULL REFERENCES routes (id) ON DELETE CASCADE,
    from_sequence     INT NOT NULL,
    to_sequence       INT NOT NULL,
    distance_meters   DOUBLE PRECISION NOT NULL DEFAULT 0,
    duration_seconds  DOUBLE PRECISION NOT NULL DEFAULT 0,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_route_legs_seq CHECK (to_sequence > from_sequence),
    CONSTRAINT chk_route_legs_distance CHECK (distance_meters >= 0),
    CONSTRAINT chk_route_legs_duration CHECK (duration_seconds >= 0)
);

COMMENT ON TABLE route_legs IS 'Per-leg distance/duration for a route plan.';
