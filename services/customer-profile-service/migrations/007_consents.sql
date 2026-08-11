-- Consents: channel-level marketing/privacy consent records per profile.
CREATE TYPE consent_channel AS ENUM (
    'email',
    'sms',
    'push',
    'whatsapp',
    'marketing',
    'transactional',
    'privacy',
    'newsletter'
);

CREATE TABLE consents (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    profile_id      UUID NOT NULL REFERENCES customer_profiles (id) ON DELETE CASCADE,
    tenant_id       UUID NOT NULL,
    channel         consent_channel NOT NULL,
    granted         BOOLEAN NOT NULL,
    source          TEXT NOT NULL DEFAULT '',
    recorded_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_consents_profile_channel UNIQUE (profile_id, channel)
);

COMMENT ON TABLE consents IS 'Latest consent state per profile and channel.';
COMMENT ON COLUMN consents.channel IS 'email | sms | push | whatsapp | marketing | transactional | privacy | newsletter.';
COMMENT ON COLUMN consents.source IS 'Capture source: app, web, csr, import, etc.';
COMMENT ON COLUMN consents.recorded_at IS 'When the grant/revoke decision was recorded.';
