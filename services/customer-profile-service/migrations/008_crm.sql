-- CRM notes and timeline events (profile-side; ticket lifecycle owned by crm-service).
CREATE TABLE crm_notes (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    profile_id      UUID NOT NULL REFERENCES customer_profiles (id) ON DELETE CASCADE,
    tenant_id       UUID NOT NULL,
    author_id       UUID NOT NULL,
    body            TEXT NOT NULL,
    pinned          BOOLEAN NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ
);

CREATE TABLE timeline_events (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    profile_id      UUID NOT NULL REFERENCES customer_profiles (id) ON DELETE CASCADE,
    tenant_id       UUID NOT NULL,
    type            TEXT NOT NULL,
    payload         JSONB NOT NULL DEFAULT '{}'::jsonb,
    actor_id        UUID,
    occurred_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

COMMENT ON TABLE crm_notes IS 'Internal CSR/agent notes on a customer profile.';
COMMENT ON COLUMN crm_notes.author_id IS 'Actor principal_id who wrote the note.';
COMMENT ON TABLE timeline_events IS 'Append-only profile timeline (orders linked, status changes, CSR actions).';
COMMENT ON COLUMN timeline_events.type IS 'Event type discriminator, e.g. note_added, address_changed, consent_changed.';
COMMENT ON COLUMN timeline_events.payload IS 'Opaque typed payload as JSON.';
COMMENT ON COLUMN timeline_events.actor_id IS 'Optional actor principal_id (system when null).';
