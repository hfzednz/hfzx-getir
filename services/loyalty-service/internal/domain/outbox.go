package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

const (
	EventPointsEarned         = "PointsEarned"
	EventPointsRedeemed       = "PointsRedeemed"
	EventRewardUnlocked       = "RewardUnlocked"
	EventRewardRedeemed       = "RewardRedeemed"
	EventMembershipUpgraded   = "MembershipUpgraded"
	EventMembershipDowngraded = "MembershipDowngraded"
	EventReferralCompleted    = "ReferralCompleted"
	EventCashbackIssued       = "CashbackIssued"
	EventAchievementUnlocked  = "AchievementUnlocked"
	EventMissionCompleted     = "MissionCompleted"
	EventStreakUpdated        = "StreakUpdated"
)

const (
	TopicLoyaltyPoints     = "loyalty.points"
	TopicLoyaltyRewards    = "loyalty.rewards"
	TopicLoyaltyMembership = "loyalty.membership"
	TopicLoyaltyReferral   = "loyalty.referral"
	TopicLoyaltyCashback   = "loyalty.cashback"
	TopicLoyaltyGame       = "loyalty.game"
)

// OutboxStatus is the transactional outbox row lifecycle.
type OutboxStatus string

const (
	OutboxStatusPending   OutboxStatus = "pending"
	OutboxStatusPublished OutboxStatus = "published"
	OutboxStatusFailed    OutboxStatus = "failed"
)

// Valid reports whether the outbox status is recognized.
func (s OutboxStatus) Valid() bool {
	switch s {
	case OutboxStatusPending, OutboxStatusPublished, OutboxStatusFailed:
		return true
	default:
		return false
	}
}

// OutboxMessage is a transactional outbox row awaiting publish.
type OutboxMessage struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	AccountID   uuid.UUID
	Topic       string
	Key         string
	Payload     map[string]any
	Status      OutboxStatus
	Attempts    int
	LastError   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	PublishedAt *time.Time
}

// Validate checks outbox message invariants.
func (m OutboxMessage) Validate() error {
	if m.ID == uuid.Nil {
		return fmt.Errorf("%w: outbox id required", ErrInvalidArgument)
	}
	if m.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if m.Topic == "" {
		return fmt.Errorf("%w: topic required", ErrInvalidArgument)
	}
	if !m.Status.Valid() {
		return fmt.Errorf("%w: invalid outbox status %q", ErrInvalidArgument, m.Status)
	}
	return nil
}

// TopicForEvent maps an event type to its Kafka topic.
func TopicForEvent(eventType string) string {
	switch eventType {
	case EventPointsEarned, EventPointsRedeemed:
		return TopicLoyaltyPoints
	case EventRewardUnlocked, EventRewardRedeemed:
		return TopicLoyaltyRewards
	case EventMembershipUpgraded, EventMembershipDowngraded:
		return TopicLoyaltyMembership
	case EventReferralCompleted:
		return TopicLoyaltyReferral
	case EventCashbackIssued:
		return TopicLoyaltyCashback
	case EventAchievementUnlocked, EventMissionCompleted, EventStreakUpdated:
		return TopicLoyaltyGame
	default:
		return TopicLoyaltyPoints
	}
}

// AuditEntry records an admin mutation.
type AuditEntry struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	AccountID uuid.UUID
	Action    string
	ActorID   *uuid.UUID
	Detail    map[string]any
	CreatedAt time.Time
}
