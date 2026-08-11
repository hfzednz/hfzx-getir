CREATE TABLE IF NOT EXISTS review_revisions (
    id         UUID PRIMARY KEY,
    review_id  UUID NOT NULL REFERENCES reviews(id),
    tenant_id  UUID NOT NULL,
    revision   INT NOT NULL,
    title      TEXT NOT NULL DEFAULT '',
    body       TEXT NOT NULL DEFAULT '',
    locale     TEXT NOT NULL DEFAULT 'tr-TR',
    created_at TIMESTAMPTZ NOT NULL,
    created_by UUID NOT NULL,
    UNIQUE (review_id, revision)
);
