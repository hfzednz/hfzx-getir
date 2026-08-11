package memory

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/promotion-service/internal/domain"
)

// Store is the in-memory aggregate for all promotion entities.
type Store struct {
	mu sync.RWMutex

	Campaigns         map[uuid.UUID]domain.Campaign
	Promotions        map[uuid.UUID]domain.Promotion
	Rules             map[uuid.UUID]domain.Rule // keyed by rule id
	RulesByPromo      map[uuid.UUID]uuid.UUID   // promotionID -> ruleID
	Coupons           map[uuid.UUID]domain.Coupon
	CouponsByCode     map[string]uuid.UUID // tenantID|code -> couponID
	CouponRedemptions map[uuid.UUID]domain.CouponRedemption
	CouponIdem        map[string]uuid.UUID // tenantID|key -> redemptionID
	Vouchers          map[uuid.UUID]domain.Voucher
	VouchersByCode    map[string]uuid.UUID
	VoucherRedemptions map[uuid.UUID]domain.VoucherRedemption
	VoucherIdem       map[string]uuid.UUID
	Usage             map[string]domain.UsageCounter // tenant|promo|scope|key
	Simulations       map[uuid.UUID]domain.Simulation
	Outbox            map[uuid.UUID]domain.OutboxMessage
}

// NewStore creates an empty store.
func NewStore() *Store {
	return &Store{
		Campaigns:          make(map[uuid.UUID]domain.Campaign),
		Promotions:         make(map[uuid.UUID]domain.Promotion),
		Rules:              make(map[uuid.UUID]domain.Rule),
		RulesByPromo:       make(map[uuid.UUID]uuid.UUID),
		Coupons:            make(map[uuid.UUID]domain.Coupon),
		CouponsByCode:      make(map[string]uuid.UUID),
		CouponRedemptions:  make(map[uuid.UUID]domain.CouponRedemption),
		CouponIdem:         make(map[string]uuid.UUID),
		Vouchers:           make(map[uuid.UUID]domain.Voucher),
		VouchersByCode:     make(map[string]uuid.UUID),
		VoucherRedemptions: make(map[uuid.UUID]domain.VoucherRedemption),
		VoucherIdem:        make(map[string]uuid.UUID),
		Usage:              make(map[string]domain.UsageCounter),
		Simulations:        make(map[uuid.UUID]domain.Simulation),
		Outbox:             make(map[uuid.UUID]domain.OutboxMessage),
	}
}

func tenantCodeKey(tenantID uuid.UUID, code string) string {
	return tenantID.String() + "|" + code
}

func idemKey(tenantID uuid.UUID, key string) string {
	return tenantID.String() + "|" + key
}

func usageKey(tenantID, promotionID uuid.UUID, scope domain.UsageScope, scopeKey string) string {
	return tenantID.String() + "|" + promotionID.String() + "|" + string(scope) + "|" + scopeKey
}

// Clock is a controllable clock for tests.
type Clock struct {
	T time.Time
}

// Now returns the fixed or advancing time.
func (c *Clock) Now() time.Time {
	if c.T.IsZero() {
		return time.Now().UTC()
	}
	return c.T.UTC()
}

// Advance moves the clock forward.
func (c *Clock) Advance(d time.Duration) {
	if c.T.IsZero() {
		c.T = time.Now().UTC()
	}
	c.T = c.T.Add(d)
}
