-- Product version snapshots and approval workflow audit trail.
CREATE TYPE approval_action_type AS ENUM (
    'submit',
    'approve',
    'reject',
    'request_changes',
    'publish',
    'unpublish',
    'schedule',
    'rollback',
    'archive'
);

CREATE TABLE product_versions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id      UUID NOT NULL REFERENCES products (id) ON DELETE CASCADE,
    tenant_id       UUID NOT NULL,
    version_number  INT NOT NULL CHECK (version_number > 0),
    -- Full product aggregate snapshot at this version.
    snapshot        JSONB NOT NULL DEFAULT '{}'::jsonb,
    status          product_status NOT NULL DEFAULT 'draft',
    change_summary  TEXT NOT NULL DEFAULT '',
    created_by      UUID,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_product_versions UNIQUE (product_id, version_number)
);

CREATE TABLE approval_actions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id      UUID NOT NULL REFERENCES products (id) ON DELETE CASCADE,
    version_id      UUID REFERENCES product_versions (id) ON DELETE SET NULL,
    tenant_id       UUID NOT NULL,
    action          approval_action_type NOT NULL,
    from_status     product_status,
    to_status       product_status,
    actor_id        UUID NOT NULL,
    actor_role      TEXT NOT NULL DEFAULT '',
    comment         TEXT NOT NULL DEFAULT '',
    metadata        JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

COMMENT ON TABLE product_versions IS 'Immutable product snapshots for diff, rollback, and audit.';
COMMENT ON COLUMN product_versions.snapshot IS 'JSON snapshot of product + variants + locales + attrs at version time.';
COMMENT ON TABLE approval_actions IS 'Workflow audit: submit / approve / reject / publish / rollback.';
COMMENT ON COLUMN approval_actions.actor_role IS 'author | reviewer | approver | publisher (enforced at BFF/IAM).';
