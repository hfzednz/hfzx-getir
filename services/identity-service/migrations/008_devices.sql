-- Devices: client device fingerprints bound to principals.
CREATE TABLE devices (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    principal_id    UUID NOT NULL REFERENCES principals (id) ON DELETE CASCADE,
    fingerprint     TEXT NOT NULL,
    platform        TEXT NOT NULL DEFAULT '',
    name            TEXT NOT NULL DEFAULT '',
    trusted         BOOLEAN NOT NULL DEFAULT FALSE,
    trusted_at      TIMESTAMPTZ,
    last_seen_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    revoked_at      TIMESTAMPTZ,

    CONSTRAINT uq_devices_principal_fingerprint UNIQUE (principal_id, fingerprint)
);

COMMENT ON TABLE devices IS 'Known devices for a principal; trust and revoke drive risk decisions.';
COMMENT ON COLUMN devices.fingerprint IS 'Stable client device fingerprint hash.';
COMMENT ON COLUMN devices.trusted IS 'Explicitly trusted device (reduces MFA friction when policy allows).';
