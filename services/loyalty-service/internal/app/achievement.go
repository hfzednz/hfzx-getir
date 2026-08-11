package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/loyalty-service/internal/domain"
)

// UnlockAchievementInput unlocks by code (idempotent).
type UnlockAchievementInput struct {
	TenantID  uuid.UUID
	AccountID uuid.UUID
	Code      string
}

// UnlockAchievement unlocks an achievement if not already unlocked.
func (d *Deps) UnlockAchievement(ctx context.Context, in UnlockAchievementInput) (domain.AchievementUnlock, error) {
	ach, err := d.Achievements.GetAchievementByCode(ctx, in.TenantID, in.Code)
	if err != nil {
		return domain.AchievementUnlock{}, err
	}
	if existing, err := d.Achievements.GetUnlock(ctx, in.TenantID, in.AccountID, ach.ID); err == nil {
		return existing, nil
	}
	now := d.now()
	u := domain.AchievementUnlock{
		ID: d.newID(), TenantID: in.TenantID, AccountID: in.AccountID,
		AchievementID: ach.ID, Code: ach.Code, UnlockedAt: now,
	}
	if err := d.Achievements.CreateUnlock(ctx, u); err != nil {
		return domain.AchievementUnlock{}, err
	}
	acct, _ := d.Accounts.GetAccount(ctx, in.TenantID, in.AccountID)
	d.emit(ctx, acct, domain.EventAchievementUnlocked, map[string]any{
		"code": ach.Code, "achievementId": ach.ID.String(),
	})
	return u, nil
}

func (d *Deps) evaluateAchievements(ctx context.Context, acct domain.Account) error {
	list, err := d.Achievements.ListAchievements(ctx, acct.TenantID)
	if err != nil {
		return err
	}
	for _, ach := range list {
		if !ach.Active {
			continue
		}
		if _, err := d.Achievements.GetUnlock(ctx, acct.TenantID, acct.ID, ach.ID); err == nil {
			continue
		}
		var current int64
		switch ach.RuleType {
		case domain.RulePurchaseCount:
			current, _ = d.Accounts.GetStat(ctx, acct.TenantID, acct.ID, "purchase_count")
		case domain.RuleReferralCount:
			current, _ = d.Accounts.GetStat(ctx, acct.TenantID, acct.ID, "referral_count")
		case domain.RuleSpendMinor:
			current, _ = d.Accounts.GetStat(ctx, acct.TenantID, acct.ID, "spend_minor")
		case domain.RuleMissionCode:
			continue // unlocked via mission complete path
		default:
			continue
		}
		if current >= ach.Threshold {
			_, _ = d.UnlockAchievement(ctx, UnlockAchievementInput{
				TenantID: acct.TenantID, AccountID: acct.ID, Code: ach.Code,
			})
		}
	}
	return nil
}

// CheckAchievementRules evaluates rule thresholds for an account.
func (d *Deps) CheckAchievementRules(ctx context.Context, tenantID, accountID uuid.UUID) error {
	acct, err := d.Accounts.GetAccount(ctx, tenantID, accountID)
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	return d.evaluateAchievements(ctx, acct)
}
