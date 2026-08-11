CREATE TABLE IF NOT EXISTS review_votes (
    id         UUID PRIMARY KEY,
    review_id  UUID NOT NULL REFERENCES reviews(id),
    tenant_id  UUID NOT NULL,
    voter_id   UUID NOT NULL,
    helpful    BOOLEAN NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    UNIQUE (review_id, voter_id)
);

CREATE TABLE IF NOT EXISTS review_comments (
    id         UUID PRIMARY KEY,
    review_id  UUID NOT NULL REFERENCES reviews(id),
    tenant_id  UUID NOT NULL,
    author_id  UUID NOT NULL,
    parent_id  UUID,
    body       TEXT NOT NULL,
    status     TEXT NOT NULL DEFAULT 'published',
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS review_reports (
    id          UUID PRIMARY KEY,
    review_id   UUID NOT NULL REFERENCES reviews(id),
    tenant_id   UUID NOT NULL,
    reporter_id UUID NOT NULL,
    reason      TEXT NOT NULL,
    details     TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_review_reports_review ON review_reports (tenant_id, review_id);
