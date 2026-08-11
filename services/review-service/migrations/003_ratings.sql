CREATE TABLE IF NOT EXISTS ratings (
    id          UUID PRIMARY KEY,
    tenant_id   UUID NOT NULL,
    author_id   UUID NOT NULL,
    target_type TEXT NOT NULL,
    target_id   UUID NOT NULL,
    review_id   UUID REFERENCES reviews(id),
    scheme      TEXT NOT NULL,
    value       DOUBLE PRECISION NOT NULL,
    stars       DOUBLE PRECISION NOT NULL,
    verified    BOOLEAN NOT NULL DEFAULT FALSE,
    weight      DOUBLE PRECISION NOT NULL DEFAULT 1,
    created_at  TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_ratings_target
    ON ratings (tenant_id, target_type, target_id, created_at DESC);

CREATE TABLE IF NOT EXISTS rating_aggregates (
    tenant_id       UUID NOT NULL,
    target_type     TEXT NOT NULL,
    target_id       UUID NOT NULL,
    scheme          TEXT NOT NULL DEFAULT 'stars_5',
    count           INT NOT NULL DEFAULT 0,
    sum_stars       DOUBLE PRECISION NOT NULL DEFAULT 0,
    avg_stars       DOUBLE PRECISION NOT NULL DEFAULT 0,
    bayesian_avg    DOUBLE PRECISION NOT NULL DEFAULT 0,
    time_decay_avg  DOUBLE PRECISION NOT NULL DEFAULT 0,
    verified_count  INT NOT NULL DEFAULT 0,
    verified_avg    DOUBLE PRECISION NOT NULL DEFAULT 0,
    updated_at      TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (tenant_id, target_type, target_id, scheme)
);
