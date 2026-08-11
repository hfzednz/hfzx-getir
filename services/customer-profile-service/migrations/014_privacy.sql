-- Privacy requests: GDPR-style export and delete workflows.
CREATE TYPE privacy_request_kind AS ENUM (
    'export',
    'delete'
);

CREATE TYPE privacy_request_status AS ENUM (
    'pending',
    'processing',
    'completed',
    'failed',
    'cancelled'
);

CREATE TABLE privacy_requests (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    profile_id      UUID NOT NULL REFERENCES customer_profiles (id) ON DELETE CASCADE,
    tenant_id       UUID NOT NULL,
    kind            privacy_request_kind NOT NULL,
    status          privacy_request_status NOT NULL DEFAULT 'pending',
    payload_ref     TEXT NOT NULL DEFAULT '',
    requested_by    UUID,
    reason          TEXT NOT NULL DEFAULT '',
    error_message   TEXT NOT NULL DEFAULT '',
    completed_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

COMMENT ON TABLE privacy_requests IS 'Privacy export/delete requests and their processing status.';
COMMENT ON COLUMN privacy_requests.kind IS 'export | delete.';
COMMENT ON COLUMN privacy_requests.payload_ref IS 'Storage key / URI for export archive or deletion evidence.';
COMMENT ON COLUMN privacy_requests.requested_by IS 'Requester principal_id (self or CSR).';
