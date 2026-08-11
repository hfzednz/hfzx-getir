CREATE TYPE poi_kind AS ENUM ('warehouse', 'pickup', 'partner', 'store', 'dropoff');

CREATE TABLE pois (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id     UUID NOT NULL,
    kind          poi_kind NOT NULL,
    ref_id        TEXT NOT NULL,
    name          TEXT NOT NULL DEFAULT '',
    lat           DOUBLE PRECISION NOT NULL,
    lng           DOUBLE PRECISION NOT NULL,
    meta          JSONB NOT NULL DEFAULT '{}'::jsonb,
    active        BOOLEAN NOT NULL DEFAULT TRUE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_pois_lat CHECK (lat >= -90 AND lat <= 90),
    CONSTRAINT chk_pois_lng CHECK (lng >= -180 AND lng <= 180),
    CONSTRAINT chk_pois_ref CHECK (ref_id <> '')
);

COMMENT ON TABLE pois IS 'Spatial POI index; ref_id is opaque (warehouse/pickup/partner/store/dropoff).';
