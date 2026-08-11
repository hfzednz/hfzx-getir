CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE addresses (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id     UUID NOT NULL,
    line1         TEXT NOT NULL DEFAULT '',
    building      TEXT NOT NULL DEFAULT '',
    entrance      TEXT NOT NULL DEFAULT '',
    floor         TEXT NOT NULL DEFAULT '',
    apt           TEXT NOT NULL DEFAULT '',
    landmark      TEXT NOT NULL DEFAULT '',
    place_id      TEXT NOT NULL DEFAULT '',
    lat           DOUBLE PRECISION NOT NULL,
    lng           DOUBLE PRECISION NOT NULL,
    confidence    DOUBLE PRECISION NOT NULL DEFAULT 0,
    risk_score    DOUBLE PRECISION NOT NULL DEFAULT 0,
    components    JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_addresses_lat CHECK (lat >= -90 AND lat <= 90),
    CONSTRAINT chk_addresses_lng CHECK (lng >= -180 AND lng <= 180),
    CONSTRAINT chk_addresses_confidence CHECK (confidence >= 0 AND confidence <= 1),
    CONSTRAINT chk_addresses_risk CHECK (risk_score >= 0 AND risk_score <= 1),
    CONSTRAINT chk_addresses_line_or_place CHECK (line1 <> '' OR place_id <> '')
);

COMMENT ON TABLE addresses IS 'Normalized addresses with place_id, confidence, risk, and components jsonb.';
COMMENT ON COLUMN addresses.lat IS 'WGS84 latitude; prod may add GEOGRAPHY(Point,4326) generated column.';
COMMENT ON COLUMN addresses.lng IS 'WGS84 longitude; prod may add GEOGRAPHY(Point,4326) generated column.';
