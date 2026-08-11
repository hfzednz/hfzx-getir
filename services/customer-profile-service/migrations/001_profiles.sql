-- Customer profiles: CRM/profile attributes keyed by identity-service principal_id.
-- No credentials, sessions, or IAM live here.
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TYPE profile_status AS ENUM (
    'active',
    'merged',
    'deleted'
);

CREATE TYPE gender_kind AS ENUM (
    'unspecified',
    'female',
    'male',
    'non_binary',
    'other',
    'prefer_not_to_say'
);

CREATE TABLE customer_profiles (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    principal_id    UUID NOT NULL,
    tenant_id       UUID NOT NULL,
    display_name    TEXT NOT NULL DEFAULT '',
    full_name       TEXT NOT NULL DEFAULT '',
    nickname        TEXT NOT NULL DEFAULT '',
    avatar_url      TEXT NOT NULL DEFAULT '',
    gender          gender_kind NOT NULL DEFAULT 'unspecified',
    birthday        DATE,
    language        TEXT NOT NULL DEFAULT '',
    country_code    CHAR(2) NOT NULL DEFAULT '',
    city            TEXT NOT NULL DEFAULT '',
    timezone        TEXT NOT NULL DEFAULT '',
    occupation      TEXT NOT NULL DEFAULT '',
    family_size     SMALLINT NOT NULL DEFAULT 0 CHECK (family_size >= 0),
    dietary         JSONB NOT NULL DEFAULT '{}'::jsonb,
    accessibility   JSONB NOT NULL DEFAULT '{}'::jsonb,
    status          profile_status NOT NULL DEFAULT 'active',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT uq_customer_profiles_principal UNIQUE (principal_id)
);

COMMENT ON TABLE customer_profiles IS 'Customer profile & CRM attributes keyed by identity principal_id.';
COMMENT ON COLUMN customer_profiles.principal_id IS 'UUID from identity-service; no FK across services.';
COMMENT ON COLUMN customer_profiles.dietary IS 'Dietary preferences/restrictions as JSON object.';
COMMENT ON COLUMN customer_profiles.accessibility IS 'Accessibility needs as JSON object.';
COMMENT ON COLUMN customer_profiles.status IS 'Lifecycle: active | merged | deleted.';
COMMENT ON COLUMN customer_profiles.deleted_at IS 'Soft-delete timestamp; status should be deleted when set.';
