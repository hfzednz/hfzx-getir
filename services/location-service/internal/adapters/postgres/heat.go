package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/location-service/internal/app/ports"
	"github.com/nexora/location-service/internal/domain"
)

// HeatRepo persists demand heat cells.
type HeatRepo struct{ DB *sql.DB }

func (r *HeatRepo) Upsert(ctx context.Context, c domain.HeatCell) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO heat_cells (id, tenant_id, grid_cell, demand_score, density, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6)
		ON CONFLICT (tenant_id, grid_cell) DO UPDATE SET
		  demand_score=EXCLUDED.demand_score, density=EXCLUDED.density,
		  updated_at=EXCLUDED.updated_at`,
		c.ID, c.TenantID, c.GridCell, c.DemandScore, c.Density, c.UpdatedAt.UTC())
	return err
}

func (r *HeatRepo) List(ctx context.Context, tenantID uuid.UUID, limit int) ([]domain.HeatCell, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, grid_cell, demand_score, density, updated_at
		FROM heat_cells WHERE tenant_id=$1
		ORDER BY updated_at DESC
		LIMIT $2`, tenantID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.HeatCell, 0)
	for rows.Next() {
		var c domain.HeatCell
		if err := rows.Scan(&c.ID, &c.TenantID, &c.GridCell, &c.DemandScore, &c.Density, &c.UpdatedAt); err != nil {
			return nil, err
		}
		c.UpdatedAt = c.UpdatedAt.UTC()
		out = append(out, c)
	}
	return out, rows.Err()
}

var _ ports.HeatRepo = (*HeatRepo)(nil)
