-- inbox items
CREATE TABLE IF NOT EXISTS inbox_items (
    id            UUID PRIMARY KEY,
    tenant_id     UUID NOT NULL,
    principal_id  UUID NOT NULL,
    message_id    UUID NOT NULL,
    title         TEXT NOT NULL DEFAULT '',
    body          TEXT NOT NULL DEFAULT '',
    read          BOOLEAN NOT NULL DEFAULT false,
    archived      BOOLEAN NOT NULL DEFAULT false,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    read_at       TIMESTAMPTZ NULL
);
