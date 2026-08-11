-- Forecast projections (AI / demand-planning output store).
CREATE TABLE stock_forecasts (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id           UUID NOT NULL,
    warehouse_id        UUID NOT NULL REFERENCES warehouses (id) ON DELETE CASCADE,
    variant_id          UUID NOT NULL,
    sku_code            TEXT NOT NULL DEFAULT '',
    -- Horizon end (or bucket start) for the prediction window.
    horizon_start       DATE NOT NULL,
    horizon_end         DATE NOT NULL,
    predicted_demand    NUMERIC(18, 4) NOT NULL CHECK (predicted_demand >= 0),
    predicted_atp       NUMERIC(18, 4),
    confidence          NUMERIC(5, 4) CHECK (confidence IS NULL OR (confidence >= 0 AND confidence <= 1)),
    model_id            TEXT NOT NULL DEFAULT '',
    model_version       TEXT NOT NULL DEFAULT '',
    metadata            JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_stock_forecasts_key UNIQUE (warehouse_id, variant_id, horizon_start, horizon_end, model_id),
    CONSTRAINT chk_stock_forecasts_horizon CHECK (horizon_end >= horizon_start),
    CONSTRAINT chk_stock_forecasts_sku CHECK (char_length(sku_code) <= 128)
);

COMMENT ON TABLE stock_forecasts IS 'AI/demand projection store; not a write path for live stock mutations.';
COMMENT ON COLUMN stock_forecasts.variant_id IS 'Opaque catalog variant UUID.';
COMMENT ON COLUMN stock_forecasts.predicted_demand IS 'Projected demand units over the horizon window.';
COMMENT ON COLUMN stock_forecasts.model_id IS 'Upstream model identifier that produced this row.';
