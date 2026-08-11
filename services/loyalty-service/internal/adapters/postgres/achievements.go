package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/loyalty-service/internal/app/ports"
	"github.com/nexora/loyalty-service/internal/domain"
)

// AchievementRepo persists achievements and unlocks.
type AchievementRepo struct{ DB *sql.DB }

func (r *AchievementRepo) CreateAchievement(ctx context.Context, a domain.Achievement) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO achievements (id, tenant_id, code, title, rule_type, threshold, active, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		a.ID, a.TenantID, a.Code, a.Title, string(a.RuleType), a.Threshold, a.Active, a.CreatedAt.UTC())
	return mapUniqueViolation(err)
}

func (r *AchievementRepo) GetAchievement(ctx context.Context, tenantID, id uuid.UUID) (domain.Achievement, error) {
	return r.scanAchievement(r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, code, title, rule_type, threshold, active, created_at
		FROM achievements WHERE id=$1 AND tenant_id=$2`, id, tenantID))
}

func (r *AchievementRepo) GetAchievementByCode(ctx context.Context, tenantID uuid.UUID, code string) (domain.Achievement, error) {
	return r.scanAchievement(r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, code, title, rule_type, threshold, active, created_at
		FROM achievements WHERE tenant_id=$1 AND code=$2`, tenantID, code))
}

func (r *AchievementRepo) ListAchievements(ctx context.Context, tenantID uuid.UUID) ([]domain.Achievement, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, code, title, rule_type, threshold, active, created_at
		FROM achievements WHERE tenant_id=$1 ORDER BY created_at DESC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Achievement
	for rows.Next() {
		a, err := scanAchievementRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (r *AchievementRepo) CreateUnlock(ctx context.Context, u domain.AchievementUnlock) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO achievement_unlocks (id, tenant_id, account_id, achievement_id, code, unlocked_at)
		VALUES ($1,$2,$3,$4,$5,$6)`,
		u.ID, u.TenantID, u.AccountID, u.AchievementID, u.Code, u.UnlockedAt.UTC())
	return mapUniqueViolation(err)
}

func (r *AchievementRepo) GetUnlock(ctx context.Context, tenantID, accountID, achievementID uuid.UUID) (domain.AchievementUnlock, error) {
	var u domain.AchievementUnlock
	err := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, account_id, achievement_id, code, unlocked_at
		FROM achievement_unlocks
		WHERE tenant_id=$1 AND account_id=$2 AND achievement_id=$3`,
		tenantID, accountID, achievementID).Scan(
		&u.ID, &u.TenantID, &u.AccountID, &u.AchievementID, &u.Code, &u.UnlockedAt)
	if isNoRows(err) {
		return domain.AchievementUnlock{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.AchievementUnlock{}, err
	}
	u.UnlockedAt = u.UnlockedAt.UTC()
	return u, nil
}

func (r *AchievementRepo) ListUnlocks(ctx context.Context, tenantID, accountID uuid.UUID) ([]domain.AchievementUnlock, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, account_id, achievement_id, code, unlocked_at
		FROM achievement_unlocks WHERE tenant_id=$1 AND account_id=$2 ORDER BY unlocked_at DESC`,
		tenantID, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.AchievementUnlock
	for rows.Next() {
		var u domain.AchievementUnlock
		if err := rows.Scan(&u.ID, &u.TenantID, &u.AccountID, &u.AchievementID, &u.Code, &u.UnlockedAt); err != nil {
			return nil, err
		}
		u.UnlockedAt = u.UnlockedAt.UTC()
		out = append(out, u)
	}
	return out, rows.Err()
}

func (r *AchievementRepo) scanAchievement(row *sql.Row) (domain.Achievement, error) {
	a, err := scanAchievementRow(row)
	if isNoRows(err) {
		return domain.Achievement{}, domain.ErrNotFound
	}
	return a, err
}

func scanAchievementRow(row scannable) (domain.Achievement, error) {
	var a domain.Achievement
	var ruleType string
	if err := row.Scan(
		&a.ID, &a.TenantID, &a.Code, &a.Title, &ruleType, &a.Threshold, &a.Active, &a.CreatedAt); err != nil {
		return domain.Achievement{}, err
	}
	a.RuleType = domain.AchievementRuleType(ruleType)
	a.CreatedAt = a.CreatedAt.UTC()
	return a, nil
}

var _ ports.AchievementRepo = (*AchievementRepo)(nil)
