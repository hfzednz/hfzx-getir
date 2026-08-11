package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/pricing-service/internal/app/ports"
	"github.com/nexora/pricing-service/internal/domain"
)

// DynamicRepo persists dynamic pricing rules.
type DynamicRepo struct{ DB *sql.DB }

func (r *DynamicRepo) Upsert(ctx context.Context, d domain.DynamicRule) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO dynamic_rules (
			id, tenant_id, code, kind, "trigger", adjustment_bps, adjustment_minor,
			start_hour, end_hour, inventory_threshold, active, priority, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
		ON CONFLICT (id) DO UPDATE SET
			code=EXCLUDED.code, kind=EXCLUDED.kind, "trigger"=EXCLUDED."trigger",
			adjustment_bps=EXCLUDED.adjustment_bps, adjustment_minor=EXCLUDED.adjustment_minor,
			start_hour=EXCLUDED.start_hour, end_hour=EXCLUDED.end_hour,
			inventory_threshold=EXCLUDED.inventory_threshold, active=EXCLUDED.active,
			priority=EXCLUDED.priority, updated_at=EXCLUDED.updated_at`,
		d.ID, d.TenantID, d.Code, string(d.Kind), string(d.Trigger), d.AdjustmentBps, d.AdjustmentMinor,
		d.StartHour, d.EndHour, d.InventoryThreshold, d.Active, d.Priority,
		d.CreatedAt.UTC(), d.UpdatedAt.UTC())
	return mapUniqueViolation(err)
}

func (r *DynamicRepo) Get(ctx context.Context, tenantID, id uuid.UUID) (domain.DynamicRule, error) {
	d, err := scanDynamicRule(r.DB.QueryRowContext(ctx, dynamicSelect+` WHERE id=$1 AND tenant_id=$2`, id, tenantID))
	if isNoRows(err) {
		return domain.DynamicRule{}, fmt.Errorf("%w: dynamic rule", domain.ErrNotFound)
	}
	return d, err
}

func (r *DynamicRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.DynamicRule, error) {
	rows, err := r.DB.QueryContext(ctx, dynamicSelect+` WHERE tenant_id=$1 ORDER BY priority DESC, created_at ASC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanDynamicRules(rows)
}

func (r *DynamicRepo) ListActive(ctx context.Context, tenantID uuid.UUID) ([]domain.DynamicRule, error) {
	rows, err := r.DB.QueryContext(ctx,
		dynamicSelect+` WHERE tenant_id=$1 AND active=true ORDER BY priority DESC, created_at ASC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanDynamicRules(rows)
}

const dynamicSelect = `
	SELECT id, tenant_id, code, kind, "trigger", adjustment_bps, adjustment_minor,
		start_hour, end_hour, inventory_threshold, active, priority, created_at, updated_at
	FROM dynamic_rules`

func scanDynamicRule(row scanner) (domain.DynamicRule, error) {
	var d domain.DynamicRule
	var kind, trigger string
	err := row.Scan(
		&d.ID, &d.TenantID, &d.Code, &kind, &trigger, &d.AdjustmentBps, &d.AdjustmentMinor,
		&d.StartHour, &d.EndHour, &d.InventoryThreshold, &d.Active, &d.Priority,
		&d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		return domain.DynamicRule{}, err
	}
	d.Kind = domain.DynamicKind(kind)
	d.Trigger = domain.DynamicTrigger(trigger)
	d.CreatedAt = d.CreatedAt.UTC()
	d.UpdatedAt = d.UpdatedAt.UTC()
	return d, nil
}

func scanDynamicRules(rows *sql.Rows) ([]domain.DynamicRule, error) {
	out := make([]domain.DynamicRule, 0)
	for rows.Next() {
		d, err := scanDynamicRule(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

var _ ports.DynamicRepo = (*DynamicRepo)(nil)
