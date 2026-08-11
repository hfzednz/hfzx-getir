-- Audit trail of pre-pick order modifications (items, address, schedule, notes, gift).
CREATE TYPE modification_kind AS ENUM (
    'items',
    'address',
    'schedule',
    'notes',
    'gift',
    'priority',
    'split',
    'merge',
    'other'
);

CREATE TABLE modifications (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id        UUID NOT NULL REFERENCES orders (id) ON DELETE CASCADE,
    tenant_id       UUID NOT NULL,
    kind            modification_kind NOT NULL,
    before_snapshot JSONB NOT NULL DEFAULT '{}'::jsonb,
    after_snapshot  JSONB NOT NULL DEFAULT '{}'::jsonb,
    reason          TEXT NOT NULL DEFAULT '',
    actor_id        UUID,
    actor_type      TEXT NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

COMMENT ON TABLE modifications IS 'Append-only audit of order modifications under cancel/modify policy.';
COMMENT ON COLUMN modifications.before_snapshot IS 'State before the modification.';
COMMENT ON COLUMN modifications.after_snapshot IS 'State after the modification.';
