CREATE TABLE IF NOT EXISTS trust_scores (
    tenant_id          UUID NOT NULL,
    reviewer_id        UUID NOT NULL,
    score              DOUBLE PRECISION NOT NULL DEFAULT 40,
    verified_purchases INT NOT NULL DEFAULT 0,
    published_reviews  INT NOT NULL DEFAULT 0,
    rejected_reviews   INT NOT NULL DEFAULT 0,
    helpful_received   INT NOT NULL DEFAULT 0,
    badges             TEXT[] NOT NULL DEFAULT '{}',
    updated_at         TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (tenant_id, reviewer_id)
);

CREATE TABLE IF NOT EXISTS reputation_scores (
    tenant_id    UUID NOT NULL,
    target_type  TEXT NOT NULL,
    target_id    UUID NOT NULL,
    score        DOUBLE PRECISION NOT NULL DEFAULT 50,
    tier         TEXT NOT NULL DEFAULT 'fair',
    review_count INT NOT NULL DEFAULT 0,
    updated_at   TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (tenant_id, target_type, target_id)
);
