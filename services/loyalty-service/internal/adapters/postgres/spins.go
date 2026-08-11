package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/loyalty-service/internal/app/ports"
	"github.com/nexora/loyalty-service/internal/domain"
)

// SpinRepo persists spin campaigns and results.
type SpinRepo struct{ DB *sql.DB }

func (r *SpinRepo) CreateCampaign(ctx context.Context, c domain.SpinCampaign) error {
	prizes := JSONPrizes(c.Prizes)
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO spin_campaigns (id, tenant_id, code, title, prizes, active, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		c.ID, c.TenantID, c.Code, c.Title, prizes, c.Active, c.CreatedAt.UTC())
	return mapUniqueViolation(err)
}

func (r *SpinRepo) GetCampaign(ctx context.Context, tenantID, campaignID uuid.UUID) (domain.SpinCampaign, error) {
	return r.scanCampaign(r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, code, title, prizes, active, created_at
		FROM spin_campaigns WHERE id=$1 AND tenant_id=$2`, campaignID, tenantID))
}

func (r *SpinRepo) GetCampaignByCode(ctx context.Context, tenantID uuid.UUID, code string) (domain.SpinCampaign, error) {
	return r.scanCampaign(r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, code, title, prizes, active, created_at
		FROM spin_campaigns WHERE tenant_id=$1 AND code=$2`, tenantID, code))
}

func (r *SpinRepo) CreateSpin(ctx context.Context, s domain.SpinResult) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO spins (id, tenant_id, account_id, campaign_id, prize_code, points_won, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		s.ID, s.TenantID, s.AccountID, s.CampaignID, s.PrizeCode, s.PointsWon, s.CreatedAt.UTC())
	return err
}

func (r *SpinRepo) scanCampaign(row *sql.Row) (domain.SpinCampaign, error) {
	var c domain.SpinCampaign
	var prizes JSONPrizes
	err := row.Scan(&c.ID, &c.TenantID, &c.Code, &c.Title, &prizes, &c.Active, &c.CreatedAt)
	if isNoRows(err) {
		return domain.SpinCampaign{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.SpinCampaign{}, err
	}
	c.Prizes = []domain.SpinPrize(prizes)
	c.CreatedAt = c.CreatedAt.UTC()
	return c, nil
}

var _ ports.SpinRepo = (*SpinRepo)(nil)
