package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/loyalty-service/internal/app/ports"
	"github.com/nexora/loyalty-service/internal/domain"
)

// MembershipRepo persists memberships and tier config.
type MembershipRepo struct{ DB *sql.DB }

func (r *MembershipRepo) ListTiers(ctx context.Context, tenantID uuid.UUID) ([]domain.TierConfig, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT code, name, threshold_points, rank, benefits
		FROM membership_tiers WHERE tenant_id=$1 ORDER BY rank ASC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.TierConfig
	for rows.Next() {
		var t domain.TierConfig
		var code string
		var benefits JSONMap
		if err := rows.Scan(&code, &t.Name, &t.ThresholdPoints, &t.Rank, &benefits); err != nil {
			return nil, err
		}
		t.Code = domain.TierCode(code)
		t.Benefits = map[string]any(benefits)
		out = append(out, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(out) == 0 {
		return domain.DefaultTiers(), nil
	}
	return out, nil
}

func (r *MembershipRepo) GetMembership(ctx context.Context, tenantID, accountID uuid.UUID) (domain.Membership, error) {
	var m domain.Membership
	var tier string
	err := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, account_id, tier, since, updated_at
		FROM memberships WHERE tenant_id=$1 AND account_id=$2`, tenantID, accountID).Scan(
		&m.ID, &m.TenantID, &m.AccountID, &tier, &m.Since, &m.UpdatedAt)
	if isNoRows(err) {
		return domain.Membership{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Membership{}, err
	}
	m.Tier = domain.TierCode(tier)
	m.Since = m.Since.UTC()
	m.UpdatedAt = m.UpdatedAt.UTC()
	return m, nil
}

func (r *MembershipRepo) UpsertMembership(ctx context.Context, m domain.Membership) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO memberships (id, tenant_id, account_id, tier, since, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6)
		ON CONFLICT (tenant_id, account_id) DO UPDATE SET
		  tier = EXCLUDED.tier,
		  since = EXCLUDED.since,
		  updated_at = EXCLUDED.updated_at`,
		m.ID, m.TenantID, m.AccountID, string(m.Tier), m.Since.UTC(), m.UpdatedAt.UTC())
	return err
}

var _ ports.MembershipRepo = (*MembershipRepo)(nil)
