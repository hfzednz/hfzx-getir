package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/loyalty-service/internal/app/ports"
	"github.com/nexora/loyalty-service/internal/domain"
)

// RewardRepo persists rewards and redemptions.
type RewardRepo struct{ DB *sql.DB }

func (r *RewardRepo) CreateReward(ctx context.Context, rw domain.Reward) error {
	meta := JSONMap(rw.Metadata)
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO rewards (id, tenant_id, code, title, points_cost, active, metadata, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		rw.ID, rw.TenantID, rw.Code, rw.Title, rw.PointsCost, rw.Active, meta, rw.CreatedAt.UTC())
	return mapUniqueViolation(err)
}

func (r *RewardRepo) GetReward(ctx context.Context, tenantID, rewardID uuid.UUID) (domain.Reward, error) {
	return r.scanReward(r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, code, title, points_cost, active, metadata, created_at
		FROM rewards WHERE id=$1 AND tenant_id=$2`, rewardID, tenantID))
}

func (r *RewardRepo) GetRewardByCode(ctx context.Context, tenantID uuid.UUID, code string) (domain.Reward, error) {
	return r.scanReward(r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, code, title, points_cost, active, metadata, created_at
		FROM rewards WHERE tenant_id=$1 AND code=$2`, tenantID, code))
}

func (r *RewardRepo) ListRewards(ctx context.Context, tenantID uuid.UUID) ([]domain.Reward, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, code, title, points_cost, active, metadata, created_at
		FROM rewards WHERE tenant_id=$1 ORDER BY created_at DESC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Reward
	for rows.Next() {
		rw, err := scanRewardRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, rw)
	}
	return out, rows.Err()
}

func (r *RewardRepo) CreateRedemption(ctx context.Context, red domain.Redemption) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO redemptions (id, tenant_id, account_id, reward_id, status, points_paid, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		red.ID, red.TenantID, red.AccountID, red.RewardID, string(red.Status), red.PointsPaid,
		red.CreatedAt.UTC(), red.UpdatedAt.UTC())
	return err
}

func (r *RewardRepo) GetRedemption(ctx context.Context, tenantID, redemptionID uuid.UUID) (domain.Redemption, error) {
	return r.scanRedemption(r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, account_id, reward_id, status, points_paid, created_at, updated_at
		FROM redemptions WHERE id=$1 AND tenant_id=$2`, redemptionID, tenantID))
}

func (r *RewardRepo) UpdateRedemption(ctx context.Context, red domain.Redemption) error {
	res, err := r.DB.ExecContext(ctx, `
		UPDATE redemptions
		SET status=$3, points_paid=$4, updated_at=$5
		WHERE id=$1 AND tenant_id=$2`,
		red.ID, red.TenantID, string(red.Status), red.PointsPaid, red.UpdatedAt.UTC())
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *RewardRepo) ListRedemptions(ctx context.Context, tenantID, accountID uuid.UUID) ([]domain.Redemption, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, account_id, reward_id, status, points_paid, created_at, updated_at
		FROM redemptions WHERE tenant_id=$1 AND account_id=$2 ORDER BY created_at DESC`,
		tenantID, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Redemption
	for rows.Next() {
		red, err := scanRedemptionRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, red)
	}
	return out, rows.Err()
}

func (r *RewardRepo) scanReward(row *sql.Row) (domain.Reward, error) {
	rw, err := scanRewardRow(row)
	if isNoRows(err) {
		return domain.Reward{}, domain.ErrNotFound
	}
	return rw, err
}

func (r *RewardRepo) scanRedemption(row *sql.Row) (domain.Redemption, error) {
	red, err := scanRedemptionRow(row)
	if isNoRows(err) {
		return domain.Redemption{}, domain.ErrNotFound
	}
	return red, err
}

type scannable interface {
	Scan(dest ...any) error
}

func scanRewardRow(row scannable) (domain.Reward, error) {
	var rw domain.Reward
	var meta JSONMap
	if err := row.Scan(
		&rw.ID, &rw.TenantID, &rw.Code, &rw.Title, &rw.PointsCost, &rw.Active, &meta, &rw.CreatedAt); err != nil {
		return domain.Reward{}, err
	}
	rw.Metadata = map[string]any(meta)
	rw.CreatedAt = rw.CreatedAt.UTC()
	return rw, nil
}

func scanRedemptionRow(row scannable) (domain.Redemption, error) {
	var red domain.Redemption
	var status string
	if err := row.Scan(
		&red.ID, &red.TenantID, &red.AccountID, &red.RewardID, &status, &red.PointsPaid,
		&red.CreatedAt, &red.UpdatedAt); err != nil {
		return domain.Redemption{}, err
	}
	red.Status = domain.RewardStatus(status)
	red.CreatedAt = red.CreatedAt.UTC()
	red.UpdatedAt = red.UpdatedAt.UTC()
	return red, nil
}

var _ ports.RewardRepo = (*RewardRepo)(nil)
