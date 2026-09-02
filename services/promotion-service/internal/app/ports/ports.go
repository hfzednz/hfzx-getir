package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/promotion-service/internal/domain"
)

// Clock provides the current time.
type Clock interface {
	Now() time.Time
}

// IDGen generates UUIDs.
type IDGen interface {
	New() uuid.UUID
}

// EventPublisher publishes domain events.
type EventPublisher interface {
	Publish(ctx context.Context, topic, key string, payload any) error
}

// CampaignRepository persists campaigns.
type CampaignRepository interface {
	Create(ctx context.Context, c domain.Campaign) error
	Update(ctx context.Context, c domain.Campaign) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (domain.Campaign, error)
	List(ctx context.Context, tenantID uuid.UUID, status *domain.CampaignStatus, limit, offset int) ([]domain.Campaign, error)
	ListActive(ctx context.Context, tenantID uuid.UUID, now time.Time) ([]domain.Campaign, error)
}

// PromotionRepository persists promotions.
type PromotionRepository interface {
	Create(ctx context.Context, p domain.Promotion) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (domain.Promotion, error)
	ListByCampaign(ctx context.Context, tenantID, campaignID uuid.UUID) ([]domain.Promotion, error)
	ListByIDs(ctx context.Context, tenantID uuid.UUID, ids []uuid.UUID) ([]domain.Promotion, error)
}

// RuleRepository persists promotion rules.
type RuleRepository interface {
	Create(ctx context.Context, r domain.Rule) error
	GetByPromotionID(ctx context.Context, tenantID, promotionID uuid.UUID) (domain.Rule, error)
	ListByPromotionIDs(ctx context.Context, tenantID uuid.UUID, promotionIDs []uuid.UUID) ([]domain.Rule, error)
}

// CouponRepository persists coupons and redemptions.
type CouponRepository interface {
	Create(ctx context.Context, c domain.Coupon) error
	Update(ctx context.Context, c domain.Coupon) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (domain.Coupon, error)
	GetByCode(ctx context.Context, tenantID uuid.UUID, code string) (domain.Coupon, error)
	List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]domain.Coupon, error)
	CreateRedemption(ctx context.Context, r domain.CouponRedemption) error
	GetRedemptionByIdempotency(ctx context.Context, tenantID uuid.UUID, key string) (domain.CouponRedemption, error)
}

// VoucherRepository persists vouchers and redemptions.
type VoucherRepository interface {
	Create(ctx context.Context, v domain.Voucher) error
	Update(ctx context.Context, v domain.Voucher) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (domain.Voucher, error)
	GetByCode(ctx context.Context, tenantID uuid.UUID, code string) (domain.Voucher, error)
	CreateRedemption(ctx context.Context, r domain.VoucherRedemption) error
	GetRedemptionByIdempotency(ctx context.Context, tenantID uuid.UUID, key string) (domain.VoucherRedemption, error)
}

// UsageRepository tracks usage counters.
type UsageRepository interface {
	Get(ctx context.Context, tenantID, promotionID uuid.UUID, scope domain.UsageScope, scopeKey string) (domain.UsageCounter, error)
	Increment(ctx context.Context, tenantID, promotionID uuid.UUID, scope domain.UsageScope, scopeKey string) (domain.UsageCounter, error)
}

// SimulationRepository stores evaluate simulations.
type SimulationRepository interface {
	Create(ctx context.Context, s domain.Simulation) error
	List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]domain.Simulation, error)
}

// OutboxRepository is the transactional outbox.
type OutboxRepository interface {
	Enqueue(ctx context.Context, msg domain.OutboxMessage) error
	ListPending(ctx context.Context, limit int) ([]domain.OutboxMessage, error)
	MarkPublished(ctx context.Context, id uuid.UUID, at time.Time) error
}
