package app_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/promotion-service/internal/app"
	"github.com/nexora/promotion-service/internal/app/memory"
	"github.com/nexora/promotion-service/internal/domain"
)

type testEnv struct {
	Deps  *app.Deps
	Clock *memory.Clock
	Store *memory.Store
}

func testDeps(t *testing.T) *testEnv {
	t.Helper()
	store := memory.NewStore()
	camps, promos, rules, coupons, vouchers, usage, sims, outbox := memory.NewRepos(store)
	clock := &memory.Clock{T: time.Date(2026, 6, 15, 12, 0, 0, 0, time.UTC)}
	deps := &app.Deps{
		Campaigns:   camps,
		Promotions:  promos,
		Rules:       rules,
		Coupons:     coupons,
		Vouchers:    vouchers,
		Usage:       usage,
		Simulations: sims,
		Outbox:      outbox,
		Publisher:   memory.NoopPublisher{},
		Clock:       clock,
		IDs:         app.UUIDGen{},
	}
	return &testEnv{Deps: deps, Clock: clock, Store: store}
}

func activateCampaign(t *testing.T, env *testEnv, tenantID uuid.UUID, name string) domain.Campaign {
	t.Helper()
	c, err := env.Deps.CreateCampaign(context.Background(), app.CreateCampaignInput{
		TenantID: tenantID, Name: name,
	})
	if err != nil {
		t.Fatal(err)
	}
	c, err = env.Deps.ActivateCampaign(context.Background(), tenantID, c.ID)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestPercentDiscount(t *testing.T) {
	env := testDeps(t)
	tenant := uuid.New()
	camp := activateCampaign(t, env, tenant, "pct")
	pr, err := env.Deps.CreatePromotion(context.Background(), app.CreatePromotionInput{
		TenantID: tenant, CampaignID: camp.ID, Name: "10%", Type: domain.PromoPercent,
		PercentOff: 10, Priority: 10,
	})
	if err != nil {
		t.Fatal(err)
	}
	res, err := env.Deps.EvaluateCart(context.Background(), domain.EvaluateContext{
		TenantID: tenant, Currency: "TRY",
		Lines: []domain.CartLine{{LineID: "1", VariantID: "v1", Quantity: 2, UnitPriceMinor: 5000}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.TotalDiscountMinor != 1000 { // 10% of 10000
		t.Fatalf("got discount %d want 1000", res.TotalDiscountMinor)
	}
	if len(res.Discounts) != 1 || res.Discounts[0].PromotionID != pr.Promotion.ID {
		t.Fatalf("unexpected discounts: %+v", res.Discounts)
	}
}

func TestFixedDiscount(t *testing.T) {
	env := testDeps(t)
	tenant := uuid.New()
	camp := activateCampaign(t, env, tenant, "fixed")
	_, err := env.Deps.CreatePromotion(context.Background(), app.CreatePromotionInput{
		TenantID: tenant, CampaignID: camp.ID, Name: "50 off", Type: domain.PromoFixed,
		FixedOffMinor: 5000, Priority: 5,
	})
	if err != nil {
		t.Fatal(err)
	}
	res, err := env.Deps.EvaluateCart(context.Background(), domain.EvaluateContext{
		TenantID: tenant, Currency: "TRY",
		Lines: []domain.CartLine{{LineID: "1", VariantID: "v1", Quantity: 1, UnitPriceMinor: 20000}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.TotalDiscountMinor != 5000 {
		t.Fatalf("got %d want 5000", res.TotalDiscountMinor)
	}
}

func TestBOGO(t *testing.T) {
	env := testDeps(t)
	tenant := uuid.New()
	camp := activateCampaign(t, env, tenant, "bogo")
	_, err := env.Deps.CreatePromotion(context.Background(), app.CreatePromotionInput{
		TenantID: tenant, CampaignID: camp.ID, Name: "B1G1", Type: domain.PromoBOGO,
		BuyQty: 1, GetQty: 1, Priority: 10,
	})
	if err != nil {
		t.Fatal(err)
	}
	res, err := env.Deps.EvaluateCart(context.Background(), domain.EvaluateContext{
		TenantID: tenant, Currency: "TRY",
		Lines: []domain.CartLine{{LineID: "1", VariantID: "v1", Quantity: 2, UnitPriceMinor: 3000}},
	})
	if err != nil {
		t.Fatal(err)
	}
	// One free unit at 3000
	if res.TotalDiscountMinor != 3000 {
		t.Fatalf("got %d want 3000", res.TotalDiscountMinor)
	}
}

func TestStackConflictPicksHigherPriority(t *testing.T) {
	env := testDeps(t)
	tenant := uuid.New()
	camp := activateCampaign(t, env, tenant, "stack")
	low, err := env.Deps.CreatePromotion(context.Background(), app.CreatePromotionInput{
		TenantID: tenant, CampaignID: camp.ID, Name: "low", Type: domain.PromoFixed,
		FixedOffMinor: 1000, Priority: 1,
		Rule: &app.CreateRuleInput{Priority: 1, StackGroup: "cart", Stackable: false},
	})
	if err != nil {
		t.Fatal(err)
	}
	high, err := env.Deps.CreatePromotion(context.Background(), app.CreatePromotionInput{
		TenantID: tenant, CampaignID: camp.ID, Name: "high", Type: domain.PromoFixed,
		FixedOffMinor: 2500, Priority: 100,
		Rule: &app.CreateRuleInput{Priority: 100, StackGroup: "cart", Stackable: false},
	})
	if err != nil {
		t.Fatal(err)
	}
	res, err := env.Deps.EvaluateCart(context.Background(), domain.EvaluateContext{
		TenantID: tenant, Currency: "TRY",
		Lines: []domain.CartLine{{LineID: "1", VariantID: "v1", Quantity: 1, UnitPriceMinor: 10000}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Discounts) != 1 {
		t.Fatalf("want 1 discount, got %d", len(res.Discounts))
	}
	if res.Discounts[0].PromotionID != high.Promotion.ID {
		t.Fatalf("want high priority %s, got %s (low=%s)", high.Promotion.ID, res.Discounts[0].PromotionID, low.Promotion.ID)
	}
	if res.TotalDiscountMinor != 2500 {
		t.Fatalf("got %d want 2500", res.TotalDiscountMinor)
	}
}

func TestExclusionBlocks(t *testing.T) {
	env := testDeps(t)
	tenant := uuid.New()
	camp := activateCampaign(t, env, tenant, "excl")
	blocked, err := env.Deps.CreatePromotion(context.Background(), app.CreatePromotionInput{
		TenantID: tenant, CampaignID: camp.ID, Name: "blocked", Type: domain.PromoFixed,
		FixedOffMinor: 3000, Priority: 10,
		Rule: &app.CreateRuleInput{Priority: 10, Stackable: true},
	})
	if err != nil {
		t.Fatal(err)
	}
	winner, err := env.Deps.CreatePromotion(context.Background(), app.CreatePromotionInput{
		TenantID: tenant, CampaignID: camp.ID, Name: "winner", Type: domain.PromoFixed,
		FixedOffMinor: 1000, Priority: 90,
		Rule: &app.CreateRuleInput{
			Priority: 90, Stackable: true,
			ExcludePromotionIDs: []uuid.UUID{blocked.Promotion.ID},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	res, err := env.Deps.EvaluateCart(context.Background(), domain.EvaluateContext{
		TenantID: tenant, Currency: "TRY",
		Lines: []domain.CartLine{{LineID: "1", VariantID: "v1", Quantity: 1, UnitPriceMinor: 20000}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Discounts) != 1 || res.Discounts[0].PromotionID != winner.Promotion.ID {
		t.Fatalf("exclusion failed: %+v", res.Discounts)
	}
	if res.TotalDiscountMinor != 1000 {
		t.Fatalf("got %d want 1000", res.TotalDiscountMinor)
	}
}

func TestPerUserLimit(t *testing.T) {
	env := testDeps(t)
	tenant := uuid.New()
	user := uuid.New()
	camp := activateCampaign(t, env, tenant, "limit")
	pr, err := env.Deps.CreatePromotion(context.Background(), app.CreatePromotionInput{
		TenantID: tenant, CampaignID: camp.ID, Name: "once", Type: domain.PromoFixed,
		FixedOffMinor: 1000, Priority: 10,
		Rule: &app.CreateRuleInput{Priority: 10, PerUserLimit: 1},
	})
	if err != nil {
		t.Fatal(err)
	}
	eval := domain.EvaluateContext{
		TenantID: tenant, PrincipalID: user, Currency: "TRY",
		Lines: []domain.CartLine{{LineID: "1", VariantID: "v1", Quantity: 1, UnitPriceMinor: 5000}},
	}
	res, err := env.Deps.EvaluateCart(context.Background(), eval)
	if err != nil {
		t.Fatal(err)
	}
	if res.TotalDiscountMinor != 1000 {
		t.Fatalf("first eval want 1000 got %d", res.TotalDiscountMinor)
	}
	if err := env.Deps.CommitEvaluation(context.Background(), tenant, user, "", "ord1", []uuid.UUID{pr.Promotion.ID}); err != nil {
		t.Fatal(err)
	}
	res2, err := env.Deps.EvaluateCart(context.Background(), eval)
	if err != nil {
		t.Fatal(err)
	}
	if res2.TotalDiscountMinor != 0 {
		t.Fatalf("after limit want 0 got %d", res2.TotalDiscountMinor)
	}
}

func TestCouponRedeemOnce(t *testing.T) {
	env := testDeps(t)
	tenant := uuid.New()
	user := uuid.New()
	camp := activateCampaign(t, env, tenant, "coupon")
	pr, err := env.Deps.CreatePromotion(context.Background(), app.CreatePromotionInput{
		TenantID: tenant, CampaignID: camp.ID, Name: "c", Type: domain.PromoFixed,
		FixedOffMinor: 500, Priority: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	c, err := env.Deps.GenerateCoupon(context.Background(), app.GenerateCouponInput{
		TenantID: tenant, PromotionID: pr.Promotion.ID, Code: "ONCEONLY", Kind: domain.CouponSingle,
	})
	if err != nil {
		t.Fatal(err)
	}
	r1, err := env.Deps.RedeemCoupon(context.Background(), app.RedeemCouponInput{
		TenantID: tenant, Code: c.Code, PrincipalID: user, IdempotencyKey: "idem-1",
		DiscountMinor: 500, Currency: "TRY",
	})
	if err != nil {
		t.Fatal(err)
	}
	// Idempotent replay
	r2, err := env.Deps.RedeemCoupon(context.Background(), app.RedeemCouponInput{
		TenantID: tenant, Code: c.Code, PrincipalID: user, IdempotencyKey: "idem-1",
		DiscountMinor: 500, Currency: "TRY",
	})
	if err != nil {
		t.Fatal(err)
	}
	if r1.ID != r2.ID {
		t.Fatalf("idempotency failed")
	}
	// Second distinct redeem must fail
	_, err = env.Deps.RedeemCoupon(context.Background(), app.RedeemCouponInput{
		TenantID: tenant, Code: c.Code, PrincipalID: user, IdempotencyKey: "idem-2",
		DiscountMinor: 500, Currency: "TRY",
	})
	if !errors.Is(err, domain.ErrCouponRedeemed) && !errors.Is(err, domain.ErrCouponExhausted) && !errors.Is(err, domain.ErrCouponInvalid) {
		t.Fatalf("want coupon exhausted/redeemed, got %v", err)
	}
}

func TestExpiredCampaignIgnored(t *testing.T) {
	env := testDeps(t)
	tenant := uuid.New()
	camp := activateCampaign(t, env, tenant, "exp")
	_, err := env.Deps.CreatePromotion(context.Background(), app.CreatePromotionInput{
		TenantID: tenant, CampaignID: camp.ID, Name: "gone", Type: domain.PromoPercent,
		PercentOff: 50, Priority: 99,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := env.Deps.ExpireCampaign(context.Background(), tenant, camp.ID); err != nil {
		t.Fatal(err)
	}
	res, err := env.Deps.EvaluateCart(context.Background(), domain.EvaluateContext{
		TenantID: tenant, Currency: "TRY",
		Lines: []domain.CartLine{{LineID: "1", VariantID: "v1", Quantity: 1, UnitPriceMinor: 10000}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.TotalDiscountMinor != 0 {
		t.Fatalf("expired campaign should yield 0, got %d", res.TotalDiscountMinor)
	}
}

func TestScheduleWindow(t *testing.T) {
	env := testDeps(t)
	tenant := uuid.New()
	start := env.Clock.T.Add(24 * time.Hour)
	end := env.Clock.T.Add(48 * time.Hour)
	c, err := env.Deps.CreateCampaign(context.Background(), app.CreateCampaignInput{
		TenantID: tenant, Name: "future", StartsAt: &start, EndsAt: &end,
	})
	if err != nil {
		t.Fatal(err)
	}
	// Activate even if scheduled window is future — IsActiveAt still checks window.
	c, err = env.Deps.ActivateCampaign(context.Background(), tenant, c.ID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = env.Deps.CreatePromotion(context.Background(), app.CreatePromotionInput{
		TenantID: tenant, CampaignID: c.ID, Name: "win", Type: domain.PromoFixed,
		FixedOffMinor: 1000, Priority: 10,
	})
	if err != nil {
		t.Fatal(err)
	}
	res, err := env.Deps.EvaluateCart(context.Background(), domain.EvaluateContext{
		TenantID: tenant, Currency: "TRY",
		Lines: []domain.CartLine{{LineID: "1", VariantID: "v1", Quantity: 1, UnitPriceMinor: 5000}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.TotalDiscountMinor != 0 {
		t.Fatalf("before window want 0 got %d", res.TotalDiscountMinor)
	}
	env.Clock.Advance(25 * time.Hour)
	res2, err := env.Deps.EvaluateCart(context.Background(), domain.EvaluateContext{
		TenantID: tenant, Currency: "TRY",
		Lines: []domain.CartLine{{LineID: "1", VariantID: "v1", Quantity: 1, UnitPriceMinor: 5000}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res2.TotalDiscountMinor != 1000 {
		t.Fatalf("inside window want 1000 got %d", res2.TotalDiscountMinor)
	}
	env.Clock.Advance(48 * time.Hour)
	res3, err := env.Deps.EvaluateCart(context.Background(), domain.EvaluateContext{
		TenantID: tenant, Currency: "TRY",
		Lines: []domain.CartLine{{LineID: "1", VariantID: "v1", Quantity: 1, UnitPriceMinor: 5000}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res3.TotalDiscountMinor != 0 {
		t.Fatalf("after window want 0 got %d", res3.TotalDiscountMinor)
	}
}
