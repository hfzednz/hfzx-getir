-- Addresses: delivery locations for a customer profile.
CREATE TYPE address_label AS ENUM (
    'home',
    'work',
    'vacation',
    'custom'
);

CREATE TABLE addresses (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    profile_id          UUID NOT NULL REFERENCES customer_profiles (id) ON DELETE CASCADE,
    tenant_id           UUID NOT NULL,
    label               address_label NOT NULL DEFAULT 'home',
    custom_label        TEXT NOT NULL DEFAULT '',
    line1               TEXT NOT NULL DEFAULT '',
    building            TEXT NOT NULL DEFAULT '',
    apartment           TEXT NOT NULL DEFAULT '',
    entrance            TEXT NOT NULL DEFAULT '',
    floor               TEXT NOT NULL DEFAULT '',
    door                TEXT NOT NULL DEFAULT '',
    notes               TEXT NOT NULL DEFAULT '',
    lat                 DOUBLE PRECISION NOT NULL,
    lng                 DOUBLE PRECISION NOT NULL,
    city_id             UUID,
    zone_validated_at   TIMESTAMPTZ,
    is_default          BOOLEAN NOT NULL DEFAULT FALSE,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at          TIMESTAMPTZ,

    CONSTRAINT chk_addresses_lat CHECK (lat >= -90 AND lat <= 90),
    CONSTRAINT chk_addresses_lng CHECK (lng >= -180 AND lng <= 180),
    CONSTRAINT chk_addresses_custom_label CHECK (
        label <> 'custom' OR length(trim(custom_label)) > 0
    )
);

COMMENT ON TABLE addresses IS 'Customer delivery addresses; zone validation via geofence port.';
COMMENT ON COLUMN addresses.lat IS 'WGS84 latitude; required for delivery.';
COMMENT ON COLUMN addresses.lng IS 'WGS84 longitude; required for delivery.';
COMMENT ON COLUMN addresses.zone_validated_at IS 'Last successful geofence/zone validation timestamp.';
COMMENT ON COLUMN addresses.city_id IS 'Optional city catalog id from geo/catalog services.';
