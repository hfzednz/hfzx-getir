CREATE TABLE heat_cells (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id     UUID NOT NULL,
    grid_cell     TEXT NOT NULL,
    demand_score  DOUBLE PRECISION NOT NULL DEFAULT 0,
    density       DOUBLE PRECISION NOT NULL DEFAULT 0,
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_heat_cells_tenant_grid UNIQUE (tenant_id, grid_cell),
    CONSTRAINT chk_heat_demand CHECK (demand_score >= 0),
    CONSTRAINT chk_heat_density CHECK (density >= 0)
);

COMMENT ON TABLE heat_cells IS 'Demand/density aggregates per grid cell for heatmap analytics.';
