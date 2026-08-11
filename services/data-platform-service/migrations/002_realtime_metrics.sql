CREATE TABLE IF NOT EXISTS realtime_metrics (
    tenant_id  UUID NOT NULL,
    key        TEXT NOT NULL,
    value      DOUBLE PRECISION NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (tenant_id, key)
);

CREATE INDEX IF NOT EXISTS idx_realtime_tenant ON realtime_metrics (tenant_id);
