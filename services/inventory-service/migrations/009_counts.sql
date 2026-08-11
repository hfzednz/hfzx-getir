-- Cycle / full / blind / audit / spot count sessions.
CREATE TYPE count_session_type AS ENUM (
    'cycle',
    'full',
    'blind',
    'audit',
    'spot'
);

CREATE TYPE count_session_status AS ENUM (
    'draft',
    'in_progress',
    'submitted',
    'pending_approval',
    'approved',
    'rejected',
    'cancelled'
);

CREATE TABLE count_sessions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    warehouse_id    UUID NOT NULL REFERENCES warehouses (id) ON DELETE RESTRICT,
    location_id     UUID REFERENCES locations (id) ON DELETE RESTRICT,
    type            count_session_type NOT NULL,
    status          count_session_status NOT NULL DEFAULT 'draft',
    started_by      UUID,
    submitted_by    UUID,
    approved_by     UUID,
    notes           TEXT NOT NULL DEFAULT '',
    metadata        JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    started_at      TIMESTAMPTZ,
    submitted_at    TIMESTAMPTZ,
    approved_at     TIMESTAMPTZ,
    cancelled_at    TIMESTAMPTZ
);

CREATE TABLE count_lines (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id      UUID NOT NULL REFERENCES count_sessions (id) ON DELETE CASCADE,
    variant_id      UUID NOT NULL,
    sku_code        TEXT NOT NULL DEFAULT '',
    location_id     UUID REFERENCES locations (id) ON DELETE RESTRICT,
    lot_id          UUID REFERENCES stock_lots (id) ON DELETE RESTRICT,
    system_qty      BIGINT NOT NULL CHECK (system_qty >= 0),
    counted_qty     BIGINT CHECK (counted_qty IS NULL OR counted_qty >= 0),
    -- variance = counted_qty - system_qty (null until counted)
    variance        BIGINT,
    approved        BOOLEAN,
    notes           TEXT NOT NULL DEFAULT '',
    metadata        JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_count_lines_sku CHECK (char_length(sku_code) <= 128),
    CONSTRAINT chk_count_lines_variance CHECK (
        counted_qty IS NULL
        OR variance = counted_qty - system_qty
    )
);

COMMENT ON TABLE count_sessions IS 'Inventory count session; variance approval may post adjust movements.';
COMMENT ON TABLE count_lines IS 'Counted vs system qty; variance drives approval workflow.';
COMMENT ON COLUMN count_sessions.type IS 'cycle | full | blind | audit | spot.';
COMMENT ON COLUMN count_sessions.status IS 'draft | in_progress | submitted | pending_approval | approved | rejected | cancelled.';
