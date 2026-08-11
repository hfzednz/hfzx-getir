CREATE TABLE provider_configs (
    tenant_id     UUID PRIMARY KEY,
    google_enabled  BOOLEAN NOT NULL DEFAULT FALSE,
    apple_enabled   BOOLEAN NOT NULL DEFAULT FALSE,
    mapbox_enabled  BOOLEAN NOT NULL DEFAULT FALSE,
    osm_enabled     BOOLEAN NOT NULL DEFAULT TRUE,
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

COMMENT ON TABLE provider_configs IS 'Client map SDK provider enable flags (google|apple|mapbox|osm); no tile serving.';
