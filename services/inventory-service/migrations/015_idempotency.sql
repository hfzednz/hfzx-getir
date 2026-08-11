-- Command idempotency cache for multi-instance inventory mutations.
CREATE TABLE IF NOT EXISTS inventory_idempotency (
    key          TEXT PRIMARY KEY,
    payload      JSONB NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at   TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_inventory_idempotency_expires
    ON inventory_idempotency (expires_at)
    WHERE expires_at IS NOT NULL;
