package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/loyalty-service/internal/app/ports"
	"github.com/nexora/loyalty-service/internal/domain"
)

// CollectibleRepo persists collectibles and ownership.
type CollectibleRepo struct{ DB *sql.DB }

func (r *CollectibleRepo) CreateCollectible(ctx context.Context, c domain.Collectible) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO collectibles (id, tenant_id, code, title, rarity, active, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		c.ID, c.TenantID, c.Code, c.Title, c.Rarity, c.Active, c.CreatedAt.UTC())
	return mapUniqueViolation(err)
}

func (r *CollectibleRepo) Grant(ctx context.Context, o domain.OwnedCollectible) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO owned_collectibles (id, tenant_id, account_id, collectible_id, acquired_at)
		VALUES ($1,$2,$3,$4,$5)`,
		o.ID, o.TenantID, o.AccountID, o.CollectibleID, o.AcquiredAt.UTC())
	return err
}

func (r *CollectibleRepo) ListOwned(ctx context.Context, tenantID, accountID uuid.UUID) ([]domain.OwnedCollectible, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, account_id, collectible_id, acquired_at
		FROM owned_collectibles WHERE tenant_id=$1 AND account_id=$2 ORDER BY acquired_at DESC`,
		tenantID, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.OwnedCollectible
	for rows.Next() {
		var o domain.OwnedCollectible
		if err := rows.Scan(&o.ID, &o.TenantID, &o.AccountID, &o.CollectibleID, &o.AcquiredAt); err != nil {
			return nil, err
		}
		o.AcquiredAt = o.AcquiredAt.UTC()
		out = append(out, o)
	}
	return out, rows.Err()
}

var _ ports.CollectibleRepo = (*CollectibleRepo)(nil)
