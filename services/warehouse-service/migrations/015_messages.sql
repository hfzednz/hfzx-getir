-- Internal warehouse announcements and operational alerts.
CREATE TYPE message_kind AS ENUM (
    'announcement',
    'alert',
    'sla_warning',
    'broadcast',
    'direct'
);

CREATE TYPE message_severity AS ENUM (
    'info',
    'warning',
    'critical'
);

CREATE TABLE messages (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    warehouse_id    UUID REFERENCES warehouses (id) ON DELETE CASCADE,
    kind            message_kind NOT NULL DEFAULT 'announcement',
    severity        message_severity NOT NULL DEFAULT 'info',
    title           TEXT NOT NULL,
    body            TEXT NOT NULL DEFAULT '',
    audience        JSONB NOT NULL DEFAULT '{}'::jsonb,
    author_id       UUID,
    expires_at      TIMESTAMPTZ,
    acknowledged    BOOLEAN NOT NULL DEFAULT FALSE,
    metadata        JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_messages_title CHECK (title <> '')
);

COMMENT ON TABLE messages IS 'Internal WH announcements/alerts for mobile/ops boards.';
COMMENT ON COLUMN messages.audience IS 'Target roles/stations/employees filter JSON.';
COMMENT ON COLUMN messages.warehouse_id IS 'Null = tenant-wide broadcast.';
