package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/cart-service/internal/app/ports"
	"github.com/nexora/cart-service/internal/domain"
)

// SavedCartRepo persists save-for-later snapshots.
type SavedCartRepo struct{ DB *sql.DB }

var _ ports.SavedCartRepository = (*SavedCartRepo)(nil)

func (r *SavedCartRepo) Create(ctx context.Context, s domain.SavedCart) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO saved_carts (
			id, tenant_id, principal_id, source_cart_id, name, snapshot, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		s.ID, s.TenantID, s.PrincipalID, nullUUID(s.SourceCartID), s.Name, JSONMap(s.Snapshot),
		s.CreatedAt, s.UpdatedAt,
	)
	return mapUniqueViolation(err)
}

func (r *SavedCartRepo) ListByPrincipal(ctx context.Context, tenantID, principalID uuid.UUID) ([]domain.SavedCart, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, principal_id, source_cart_id, name, snapshot, created_at, updated_at
		FROM saved_carts
		WHERE tenant_id=$1 AND principal_id=$2
		ORDER BY created_at DESC`, tenantID, principalID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.SavedCart, 0)
	for rows.Next() {
		var s domain.SavedCart
		var source uuid.NullUUID
		var snap JSONMap
		if err := rows.Scan(
			&s.ID, &s.TenantID, &s.PrincipalID, &source, &s.Name, &snap, &s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, err
		}
		s.SourceCartID = scanNullUUID(source)
		s.Snapshot = map[string]any(snap)
		if s.Snapshot == nil {
			s.Snapshot = map[string]any{}
		}
		s.CreatedAt = s.CreatedAt.UTC()
		s.UpdatedAt = s.UpdatedAt.UTC()
		out = append(out, s)
	}
	return out, rows.Err()
}
