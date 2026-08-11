-- Merge jobs: duplicate detection and profile merge workflows.
CREATE TYPE merge_job_status AS ENUM (
    'pending',
    'running',
    'completed',
    'failed',
    'cancelled'
);

CREATE TABLE merge_jobs (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id           UUID NOT NULL,
    source_profile_id   UUID NOT NULL REFERENCES customer_profiles (id),
    target_profile_id   UUID NOT NULL REFERENCES customer_profiles (id),
    status              merge_job_status NOT NULL DEFAULT 'pending',
    detection_score     DOUBLE PRECISION,
    detection_reason    JSONB NOT NULL DEFAULT '{}'::jsonb,
    result              JSONB NOT NULL DEFAULT '{}'::jsonb,
    requested_by        UUID,
    error_message       TEXT NOT NULL DEFAULT '',
    started_at          TIMESTAMPTZ,
    completed_at        TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_merge_jobs_distinct_profiles CHECK (source_profile_id <> target_profile_id)
);

COMMENT ON TABLE merge_jobs IS 'Duplicate detection / profile merge jobs; source is absorbed into target.';
COMMENT ON COLUMN merge_jobs.source_profile_id IS 'Profile being merged away (becomes status=merged).';
COMMENT ON COLUMN merge_jobs.target_profile_id IS 'Surviving profile.';
COMMENT ON COLUMN merge_jobs.detection_score IS 'Duplicate confidence score in [0,1] when auto-detected.';
COMMENT ON COLUMN merge_jobs.detection_reason IS 'Signals that triggered duplicate detection.';
