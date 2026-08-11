package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/pricing-service/internal/app/ports"
	"github.com/nexora/pricing-service/internal/domain"
)

// TaxRepo persists tax display rules.
type TaxRepo struct{ DB *sql.DB }

func (r *TaxRepo) Upsert(ctx context.Context, t domain.TaxRule) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO tax_rules (
			id, tenant_id, code, name, rate_bps, inclusive, region_id, active, priority, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		ON CONFLICT (id) DO UPDATE SET
			code=EXCLUDED.code, name=EXCLUDED.name, rate_bps=EXCLUDED.rate_bps,
			inclusive=EXCLUDED.inclusive, region_id=EXCLUDED.region_id, active=EXCLUDED.active,
			priority=EXCLUDED.priority, updated_at=EXCLUDED.updated_at`,
		t.ID, t.TenantID, t.Code, t.Name, t.RateBps, t.Inclusive, nullUUID(t.RegionID),
		t.Active, t.Priority, t.CreatedAt.UTC(), t.UpdatedAt.UTC())
	return mapUniqueViolation(err)
}

func (r *TaxRepo) Get(ctx context.Context, tenantID, id uuid.UUID) (domain.TaxRule, error) {
	t, err := scanTaxRule(r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, code, name, rate_bps, inclusive, region_id, active, priority, created_at, updated_at
		FROM tax_rules WHERE id=$1 AND tenant_id=$2`, id, tenantID))
	if isNoRows(err) {
		return domain.TaxRule{}, fmt.Errorf("%w: tax rule", domain.ErrNotFound)
	}
	return t, err
}

func (r *TaxRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.TaxRule, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, code, name, rate_bps, inclusive, region_id, active, priority, created_at, updated_at
		FROM tax_rules WHERE tenant_id=$1 ORDER BY priority DESC, created_at ASC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.TaxRule, 0)
	for rows.Next() {
		t, err := scanTaxRule(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func scanTaxRule(row scanner) (domain.TaxRule, error) {
	var t domain.TaxRule
	var regionID uuid.NullUUID
	err := row.Scan(
		&t.ID, &t.TenantID, &t.Code, &t.Name, &t.RateBps, &t.Inclusive, &regionID,
		&t.Active, &t.Priority, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return domain.TaxRule{}, err
	}
	t.RegionID = scanNullUUID(regionID)
	t.CreatedAt = t.CreatedAt.UTC()
	t.UpdatedAt = t.UpdatedAt.UTC()
	return t, nil
}

var _ ports.TaxRepo = (*TaxRepo)(nil)
