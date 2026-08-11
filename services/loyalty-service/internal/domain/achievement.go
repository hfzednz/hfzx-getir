package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// AchievementRuleType classifies unlock rules.
type AchievementRuleType string

const (
	RulePurchaseCount AchievementRuleType = "purchase_count"
	RuleReferralCount AchievementRuleType = "referral_count"
	RuleSpendMinor    AchievementRuleType = "spend_minor"
	RuleMissionCode   AchievementRuleType = "mission_code"
)

// Achievement is an unlockable badge.
type Achievement struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	Code      string
	Title     string
	RuleType  AchievementRuleType
	Threshold int64
	Active    bool
	CreatedAt time.Time
}

// AchievementUnlock records when an account unlocked an achievement.
type AchievementUnlock struct {
	ID            uuid.UUID
	TenantID      uuid.UUID
	AccountID     uuid.UUID
	AchievementID uuid.UUID
	Code          string
	UnlockedAt    time.Time
}

// Validate checks unlock invariants.
func (u AchievementUnlock) Validate() error {
	if u.ID == uuid.Nil || u.TenantID == uuid.Nil || u.AccountID == uuid.Nil || u.AchievementID == uuid.Nil {
		return fmt.Errorf("%w: unlock ids required", ErrInvalidArgument)
	}
	return nil
}

// Collectible is a digital collectible definition.
type Collectible struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	Code      string
	Title     string
	Rarity    string
	Active    bool
	CreatedAt time.Time
}

// OwnedCollectible is ownership of a collectible.
type OwnedCollectible struct {
	ID            uuid.UUID
	TenantID      uuid.UUID
	AccountID     uuid.UUID
	CollectibleID uuid.UUID
	AcquiredAt    time.Time
}

// AIScore is a stub churn/LTV projection.
type AIScore struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	AccountID   uuid.UUID
	PrincipalID uuid.UUID
	ChurnScore  float64
	LTVScore    float64
	UpdatedAt   time.Time
}
