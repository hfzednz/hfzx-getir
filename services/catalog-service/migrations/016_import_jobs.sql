-- Bulk import / export job tracking for catalog data.
CREATE TYPE import_job_kind AS ENUM (
    'import',
    'export'
);

CREATE TYPE import_job_status AS ENUM (
    'pending',
    'validating',
    'running',
    'completed',
    'failed',
    'cancelled',
    'partial'
);

CREATE TABLE import_jobs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    kind            import_job_kind NOT NULL DEFAULT 'import',
    status          import_job_status NOT NULL DEFAULT 'pending',
    source_format   TEXT NOT NULL DEFAULT 'csv',
    source_uri      TEXT NOT NULL DEFAULT '',
    result_uri      TEXT NOT NULL DEFAULT '',
    total_rows      INT NOT NULL DEFAULT 0 CHECK (total_rows >= 0),
    processed_rows  INT NOT NULL DEFAULT 0 CHECK (processed_rows >= 0),
    success_rows    INT NOT NULL DEFAULT 0 CHECK (success_rows >= 0),
    error_rows      INT NOT NULL DEFAULT 0 CHECK (error_rows >= 0),
    errors          JSONB NOT NULL DEFAULT '[]'::jsonb,
    options         JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_by      UUID,
    started_at      TIMESTAMPTZ,
    finished_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

COMMENT ON TABLE import_jobs IS 'Async catalog import/export jobs; payloads in object storage via source_uri/result_uri.';
COMMENT ON COLUMN import_jobs.errors IS 'Sample / aggregated row-level error records.';
COMMENT ON COLUMN import_jobs.options IS 'Job options: dry_run, upsert mode, locale, mapping, etc.';
