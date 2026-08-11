package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/recommendation-service/internal/app/ports"
	"github.com/nexora/recommendation-service/internal/domain"
)

// CoOccurRepo persists product co-occurrence counts.
type CoOccurRepo struct{ DB *sql.DB }

func (r *CoOccurRepo) Bump(ctx context.Context, tenantID, a, b uuid.UUID, delta int, now time.Time) error {
	if a == b {
		return nil
	}
	pa, pb := a, b
	if a.String() > b.String() {
		pa, pb = b, a
	}
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO co_occurrences (tenant_id, product_a, product_b, count, updated_at)
		VALUES ($1,$2,$3,$4,$5)
		ON CONFLICT (tenant_id, product_a, product_b) DO UPDATE SET
		  count = co_occurrences.count + EXCLUDED.count,
		  updated_at = EXCLUDED.updated_at`,
		tenantID, pa, pb, delta, now.UTC())
	return err
}

func (r *CoOccurRepo) TopFor(ctx context.Context, tenantID, productID uuid.UUID, limit int) ([]domain.CoOccurrence, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := r.DB.QueryContext(ctx, `
		SELECT tenant_id, product_a, product_b, count, updated_at
		FROM co_occurrences
		WHERE tenant_id=$1 AND (product_a=$2 OR product_b=$2)
		ORDER BY count DESC
		LIMIT $3`, tenantID, productID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.CoOccurrence, 0)
	for rows.Next() {
		var c domain.CoOccurrence
		if err := rows.Scan(&c.TenantID, &c.ProductA, &c.ProductB, &c.Count, &c.UpdatedAt); err != nil {
			return nil, err
		}
		c.UpdatedAt = c.UpdatedAt.UTC()
		out = append(out, c)
	}
	return out, rows.Err()
}

var _ ports.CoOccurRepo = (*CoOccurRepo)(nil)
