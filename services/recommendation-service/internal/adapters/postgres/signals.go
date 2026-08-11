package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/recommendation-service/internal/app/ports"
	"github.com/nexora/recommendation-service/internal/domain"
)

// SignalRepo persists behavior signals.
type SignalRepo struct{ DB *sql.DB }

func (r *SignalRepo) Save(ctx context.Context, s domain.BehaviorSignal) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO behavior_signals (id, tenant_id, user_id, product_id, kind, weight, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		s.ID, s.TenantID, s.UserID, s.ProductID, s.Kind, s.Weight, s.CreatedAt.UTC())
	return err
}

func (r *SignalRepo) ListByUser(ctx context.Context, tenantID, userID uuid.UUID, limit int) ([]domain.BehaviorSignal, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, user_id, product_id, kind, weight, created_at
		FROM behavior_signals
		WHERE tenant_id=$1 AND user_id=$2
		ORDER BY created_at DESC
		LIMIT $3`, tenantID, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.BehaviorSignal, 0)
	for rows.Next() {
		var s domain.BehaviorSignal
		if err := rows.Scan(&s.ID, &s.TenantID, &s.UserID, &s.ProductID, &s.Kind, &s.Weight, &s.CreatedAt); err != nil {
			return nil, err
		}
		s.CreatedAt = s.CreatedAt.UTC()
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *SignalRepo) UsersWhoInteracted(ctx context.Context, tenantID, productID uuid.UUID, limit int) ([]uuid.UUID, error) {
	q := `
		SELECT DISTINCT user_id FROM behavior_signals
		WHERE tenant_id=$1 AND product_id=$2`
	args := []any{tenantID, productID}
	if limit > 0 {
		q += ` LIMIT $3`
		args = append(args, limit)
	}
	rows, err := r.DB.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]uuid.UUID, 0)
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

var _ ports.SignalRepo = (*SignalRepo)(nil)
