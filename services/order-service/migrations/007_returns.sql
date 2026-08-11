-- Customer/admin returns: OMS return headers + lines. Disposition is policy only;
-- stock write-back happens in inventory-service via ports/events.
CREATE TYPE return_status AS ENUM (
    'requested',
    'approved',
    'rejected',
    'in_transit',
    'received',
    'completed',
    'cancelled'
);

CREATE TYPE return_disposition AS ENUM (
    'restock',
    'quarantine',
    'waste',
    'resell',
    'pending'
);

CREATE TABLE returns (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id        UUID NOT NULL REFERENCES orders (id) ON DELETE CASCADE,
    tenant_id       UUID NOT NULL,
    status          return_status NOT NULL DEFAULT 'requested',
    disposition     return_disposition NOT NULL DEFAULT 'pending',
    reason          TEXT NOT NULL DEFAULT '',
    notes           TEXT NOT NULL DEFAULT '',
    actor_id        UUID,
    refund_id       UUID,
    metadata        JSONB NOT NULL DEFAULT '{}'::jsonb,
    requested_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    decided_at      TIMESTAMPTZ,
    completed_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE return_lines (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    return_id       UUID NOT NULL REFERENCES returns (id) ON DELETE CASCADE,
    order_line_id   UUID NOT NULL REFERENCES order_lines (id) ON DELETE RESTRICT,
    variant_id      UUID NOT NULL,
    qty             INT NOT NULL,
    disposition     return_disposition NOT NULL DEFAULT 'pending',
    reason          TEXT NOT NULL DEFAULT '',
    metadata        JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_return_lines_qty CHECK (qty > 0)
);

COMMENT ON TABLE returns IS 'OMS return requests; inventory restock is out of band via ports.';
COMMENT ON COLUMN returns.disposition IS 'Intended stock disposition hint for inventory-service.';
COMMENT ON TABLE return_lines IS 'Return lines referencing original order lines.';
