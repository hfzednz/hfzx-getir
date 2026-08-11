CREATE TABLE IF NOT EXISTS review_media (
    id            UUID PRIMARY KEY,
    review_id     UUID NOT NULL REFERENCES reviews(id),
    tenant_id     UUID NOT NULL,
    media_ref     TEXT NOT NULL,
    kind          TEXT NOT NULL,
    mime_type     TEXT NOT NULL DEFAULT '',
    width         INT NOT NULL DEFAULT 0,
    height        INT NOT NULL DEFAULT 0,
    duration_ms   INT NOT NULL DEFAULT 0,
    verified      BOOLEAN NOT NULL DEFAULT FALSE,
    moderation_ok BOOLEAN NOT NULL DEFAULT FALSE,
    created_at    TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_review_media_review ON review_media (review_id);
