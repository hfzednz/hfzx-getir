-- Profile activity log (profile-side audit of mutations / admin views).
CREATE TABLE profile_activity (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    profile_id      UUID NOT NULL REFERENCES customer_profiles (id) ON DELETE CASCADE,
    tenant_id       UUID NOT NULL,
    actor_id        UUID,
    action          TEXT NOT NULL,
    resource_type   TEXT NOT NULL DEFAULT '',
    resource_id     UUID,
    payload         JSONB NOT NULL DEFAULT '{}'::jsonb,
    ip              INET,
    user_agent      TEXT NOT NULL DEFAULT '',
    occurred_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

COMMENT ON TABLE profile_activity IS 'Append-only activity log for profile mutations and sensitive admin reads.';
COMMENT ON COLUMN profile_activity.action IS 'e.g. profile.updated, address.added, consent.changed, admin.viewed.';
COMMENT ON COLUMN profile_activity.actor_id IS 'Actor principal_id; null for system/async jobs.';
