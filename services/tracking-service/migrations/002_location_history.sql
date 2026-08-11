-- Optional capped location history (app layer enforces cap).
CREATE TABLE location_history (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id     UUID NOT NULL,
    courier_id    UUID NOT NULL,
    lat           DOUBLE PRECISION NOT NULL,
    lon           DOUBLE PRECISION NOT NULL,
    accuracy_m    DOUBLE PRECISION NOT NULL DEFAULT 0,
    heading_deg   DOUBLE PRECISION,
    speed_mps     DOUBLE PRECISION,
    recorded_at   TIMESTAMPTZ NOT NULL,
    received_at   TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_hist_lat CHECK (lat >= -90 AND lat <= 90),
    CONSTRAINT chk_hist_lon CHECK (lon >= -180 AND lon <= 180)
);

COMMENT ON TABLE location_history IS 'Capped GPS history; prune by tenant+courier oldest rows.';
