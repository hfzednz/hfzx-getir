// Package ports defines application-layer dependency interfaces.
package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/loyalty-service/internal/domain"
)

// Clock abstracts time for deterministic tests.
type Clock interface {
	Now() time.Time
}

// IDGen abstracts UUID generation.
type IDGen interface {
	New() uuid.UUID
}

// Rand abstracts random ints for spin (injectable for tests).
type Rand interface {
	Intn(n int) int
}

// EventPublisher publishes domain events.
type EventPublisher interface {
	Publish(ctx context.Context, topic string, key string, payload any) error
}

// WalletCreditRequest credits cashback/promo via wallet-service.
type WalletCreditRequest struct {
	TenantID       uuid.UUID
	PrincipalID    uuid.UUID
	AmountMinor    int64
	Currency       string
	AccountType    string // cashback | promo
	IdempotencyKey string
	Reference      string
}

// WalletCreditResult is the wallet credit response.
type WalletCreditResult struct {
	WalletID  string
	EntryID   string
	Credited  bool
}

// WalletClient credits promo/cashback accounts only (no ledger ownership here).
type WalletClient interface {
	Credit(ctx context.Context, req WalletCreditRequest) (WalletCreditResult, error)
}

// AccountRepo persists loyalty accounts and point ledger.
type AccountRepo interface {
	CreateAccount(ctx context.Context, a domain.Account) error
	GetAccount(ctx context.Context, tenantID, accountID uuid.UUID) (domain.Account, error)
	GetAccountByPrincipal(ctx context.Context, tenantID, principalID uuid.UUID) (domain.Account, error)
	UpdateAccount(ctx context.Context, a domain.Account) error

	CreateLedgerEntry(ctx context.Context, e domain.PointLedgerEntry) error
	GetLedgerByIdempotency(ctx context.Context, tenantID uuid.UUID, key string) (domain.PointLedgerEntry, error)
	GetLedgerByOrder(ctx context.Context, tenantID, accountID, orderID uuid.UUID) (domain.PointLedgerEntry, error)
	ListLedger(ctx context.Context, tenantID, accountID uuid.UUID, limit, offset int) ([]domain.PointLedgerEntry, int, error)

	CreateAudit(ctx context.Context, a domain.AuditEntry) error
	IncrStat(ctx context.Context, tenantID, accountID uuid.UUID, key string, delta int64) (int64, error)
	GetStat(ctx context.Context, tenantID, accountID uuid.UUID, key string) (int64, error)
}

// MembershipRepo persists memberships and tier config.
type MembershipRepo interface {
	ListTiers(ctx context.Context, tenantID uuid.UUID) ([]domain.TierConfig, error)
	GetMembership(ctx context.Context, tenantID, accountID uuid.UUID) (domain.Membership, error)
	UpsertMembership(ctx context.Context, m domain.Membership) error
}

// RewardRepo persists rewards and redemptions.
type RewardRepo interface {
	CreateReward(ctx context.Context, r domain.Reward) error
	GetReward(ctx context.Context, tenantID, rewardID uuid.UUID) (domain.Reward, error)
	GetRewardByCode(ctx context.Context, tenantID uuid.UUID, code string) (domain.Reward, error)
	ListRewards(ctx context.Context, tenantID uuid.UUID) ([]domain.Reward, error)
	CreateRedemption(ctx context.Context, r domain.Redemption) error
	GetRedemption(ctx context.Context, tenantID, redemptionID uuid.UUID) (domain.Redemption, error)
	UpdateRedemption(ctx context.Context, r domain.Redemption) error
	ListRedemptions(ctx context.Context, tenantID, accountID uuid.UUID) ([]domain.Redemption, error)
}

// ReferralRepo persists referral codes and events.
type ReferralRepo interface {
	CreateCode(ctx context.Context, c domain.ReferralCode) error
	GetCode(ctx context.Context, tenantID uuid.UUID, code string) (domain.ReferralCode, error)
	GetCodeByAccount(ctx context.Context, tenantID, accountID uuid.UUID) (domain.ReferralCode, error)
	CreateEvent(ctx context.Context, e domain.ReferralEvent) error
	GetEventByReferee(ctx context.Context, tenantID, refereeAccount uuid.UUID) (domain.ReferralEvent, error)
	UpdateEvent(ctx context.Context, e domain.ReferralEvent) error
	CountCompletedByReferrer(ctx context.Context, tenantID, referrerAccount uuid.UUID) (int64, error)
}

// MissionRepo persists missions and progress.
type MissionRepo interface {
	CreateMission(ctx context.Context, m domain.Mission) error
	GetMission(ctx context.Context, tenantID, missionID uuid.UUID) (domain.Mission, error)
	GetMissionByCode(ctx context.Context, tenantID uuid.UUID, code string) (domain.Mission, error)
	GetProgress(ctx context.Context, tenantID, accountID, missionID uuid.UUID) (domain.MissionProgress, error)
	UpsertProgress(ctx context.Context, p domain.MissionProgress) error
}

// AchievementRepo persists achievements and unlocks.
type AchievementRepo interface {
	CreateAchievement(ctx context.Context, a domain.Achievement) error
	GetAchievement(ctx context.Context, tenantID, id uuid.UUID) (domain.Achievement, error)
	GetAchievementByCode(ctx context.Context, tenantID uuid.UUID, code string) (domain.Achievement, error)
	ListAchievements(ctx context.Context, tenantID uuid.UUID) ([]domain.Achievement, error)
	CreateUnlock(ctx context.Context, u domain.AchievementUnlock) error
	GetUnlock(ctx context.Context, tenantID, accountID, achievementID uuid.UUID) (domain.AchievementUnlock, error)
	ListUnlocks(ctx context.Context, tenantID, accountID uuid.UUID) ([]domain.AchievementUnlock, error)
}

// StreakRepo persists streaks.
type StreakRepo interface {
	GetStreak(ctx context.Context, tenantID, accountID uuid.UUID) (domain.Streak, error)
	UpsertStreak(ctx context.Context, s domain.Streak) error
}

// SpinRepo persists spin campaigns and results.
type SpinRepo interface {
	CreateCampaign(ctx context.Context, c domain.SpinCampaign) error
	GetCampaign(ctx context.Context, tenantID, campaignID uuid.UUID) (domain.SpinCampaign, error)
	GetCampaignByCode(ctx context.Context, tenantID uuid.UUID, code string) (domain.SpinCampaign, error)
	CreateSpin(ctx context.Context, s domain.SpinResult) error
}

// CollectibleRepo persists collectibles.
type CollectibleRepo interface {
	CreateCollectible(ctx context.Context, c domain.Collectible) error
	Grant(ctx context.Context, o domain.OwnedCollectible) error
	ListOwned(ctx context.Context, tenantID, accountID uuid.UUID) ([]domain.OwnedCollectible, error)
}

// AIScoreRepo persists stub AI scores.
type AIScoreRepo interface {
	Upsert(ctx context.Context, s domain.AIScore) error
	Get(ctx context.Context, tenantID, accountID uuid.UUID) (domain.AIScore, error)
}

// OutboxRepository persists transactional outbox rows.
type OutboxRepository interface {
	Enqueue(ctx context.Context, m domain.OutboxMessage) error
	Update(ctx context.Context, m domain.OutboxMessage) error
	ListPending(ctx context.Context, limit int) ([]domain.OutboxMessage, error)
}

// CashbackRepo persists cashback grants.
type CashbackRepo interface {
	CreateGrant(ctx context.Context, g domain.CashbackGrant) error
	GetGrant(ctx context.Context, tenantID, grantID uuid.UUID) (domain.CashbackGrant, error)
	GetGrantByIdempotency(ctx context.Context, tenantID uuid.UUID, key string) (domain.CashbackGrant, error)
	UpdateGrant(ctx context.Context, g domain.CashbackGrant) error
}
