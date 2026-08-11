CREATE TABLE geocode_cache (
    query_hash    TEXT PRIMARY KEY,
    result        JSONB NOT NULL DEFAULT '{}'::jsonb,
    expires_at    TIMESTAMPTZ NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

COMMENT ON TABLE geocode_cache IS 'Geocode/reverse/autocomplete query hash → result jsonb with TTL.';
