-- attachments_meta
CREATE TABLE IF NOT EXISTS attachments_meta (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    owner_type TEXT NOT NULL,
    owner_id UUID NOT NULL,
    file_name TEXT NOT NULL,
    content_type TEXT NOT NULL,
    size_bytes BIGINT NOT NULL DEFAULT 0,
    uri TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);
