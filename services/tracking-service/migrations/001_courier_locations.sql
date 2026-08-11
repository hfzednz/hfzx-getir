-- Latest courier locations (one row per tenant+courier).
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE courier_locations (
    tenant_id     UUID NOT NULL,
    courier_id    UUID NOT NULL,
    lat           DOUBLE PRECISION NOT NULL,
    lon           DOUBLE PRECISION NOT NULL,
    accuracy_m    DOUBLE PRECISION NOT NULL DEFAULT 0,
    heading_deg   DOUBLE PRECISION,
    speed_mps     DOUBLE PRECISION,
    recorded_at   TIMESTAMPTZ NOT NULL,
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),

    PRIMARY KEY (tenant_id, courier_id),
    CONSTRAINT chk_courier_loc_lat CHECK (lat >= -90 AND lat <= 90),
    CONSTRAINT chk_courier_loc_lon CHECK (lon >= -180 AND lon <= 180)
);

COMMENT ON TABLE courier_locations IS 'Latest live courier GPS; opaque courier_id only.';
