-- Segments: local dynamic/behavior segment definitions and membership cache.
CREATE TYPE segment_kind AS ENUM (
    'dynamic',
    'behavior',
    'location',
    'revenue',
    'retention',
    'ai'
);

CREATE TABLE segments (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    name            TEXT NOT NULL,
    kind            segment_kind NOT NULL DEFAULT 'dynamic',
    description     TEXT NOT NULL DEFAULT '',
    rules           JSONB NOT NULL DEFAULT '{}'::jsonb,
    active          BOOLEAN NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_segments_tenant_name UNIQUE (tenant_id, name)
);

CREATE TABLE segment_members (
    segment_id      UUID NOT NULL REFERENCES segments (id) ON DELETE CASCADE,
    profile_id      UUID NOT NULL REFERENCES customer_profiles (id) ON DELETE CASCADE,
    joined_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at      TIMESTAMPTZ,
    source          TEXT NOT NULL DEFAULT 'rules',

    PRIMARY KEY (segment_id, profile_id)
);

COMMENT ON TABLE segments IS 'Local segment definitions; heavy ML batch lives in segmentation-service.';
COMMENT ON COLUMN segments.kind IS 'dynamic | behavior | location | revenue | retention | ai.';
COMMENT ON COLUMN segments.rules IS 'Evaluation rules payload for dynamic membership.';
COMMENT ON TABLE segment_members IS 'Membership cache / assignments for segments.';
