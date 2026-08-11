package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/loyalty-service/internal/app/ports"
	"github.com/nexora/loyalty-service/internal/domain"
)

// StreakRepo persists streaks.
type StreakRepo struct{ DB *sql.DB }

func (r *StreakRepo) GetStreak(ctx context.Context, tenantID, accountID uuid.UUID) (domain.Streak, error) {
	var s domain.Streak
	err := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, account_id, current_count, longest_count, last_active_date,
		       broken, recovery_used, updated_at
		FROM streaks WHERE tenant_id=$1 AND account_id=$2`, tenantID, accountID).Scan(
		&s.ID, &s.TenantID, &s.AccountID, &s.CurrentCount, &s.LongestCount, &s.LastActiveDate,
		&s.Broken, &s.RecoveryUsed, &s.UpdatedAt)
	if isNoRows(err) {
		return domain.Streak{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Streak{}, err
	}
	s.UpdatedAt = s.UpdatedAt.UTC()
	return s, nil
}

func (r *StreakRepo) UpsertStreak(ctx context.Context, s domain.Streak) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO streaks
		  (id, tenant_id, account_id, current_count, longest_count, last_active_date,
		   broken, recovery_used, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		ON CONFLICT (tenant_id, account_id) DO UPDATE SET
		  current_count = EXCLUDED.current_count,
		  longest_count = EXCLUDED.longest_count,
		  last_active_date = EXCLUDED.last_active_date,
		  broken = EXCLUDED.broken,
		  recovery_used = EXCLUDED.recovery_used,
		  updated_at = EXCLUDED.updated_at`,
		s.ID, s.TenantID, s.AccountID, s.CurrentCount, s.LongestCount, s.LastActiveDate,
		s.Broken, s.RecoveryUsed, s.UpdatedAt.UTC())
	return err
}

var _ ports.StreakRepo = (*StreakRepo)(nil)
