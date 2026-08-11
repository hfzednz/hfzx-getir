package app_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/cart-service/internal/app"
	"github.com/nexora/cart-service/internal/app/memory"
	"github.com/nexora/cart-service/internal/domain"
)

type testEnv struct {
	Deps      *app.Deps
	Clock     *memory.Clock
	Pricing   *memory.PricingClient
	Inventory *memory.InventoryClient
	Recommend *memory.RecommendClient
	Store     *memory.Store
}

func testDeps(t *testing.T) *testEnv {
	t.Helper()
	store := memory.NewStore()
	carts, events, outbox, saved := memory.NewRepos(store)
	clock := &memory.Clock{T: time.Date(2026, 8, 6, 12, 0, 0, 0, time.UTC)}
	pricing := memory.NewPricingClient()
	inv := memory.NewInventoryClient()
	rec := &memory.RecommendClient{}
	deps := &app.Deps{
		Carts: carts, Events: events, Outbox: outbox, Saved: saved,
		Publisher: &memory.EventPublisher{S: store},
		Pricing:   pricing, Inventory: inv, Recommend: rec,
		Clock: clock, IDs: memory.IDGen{},
	}
	return &testEnv{
		Deps: deps, Clock: clock, Pricing: pricing,
		Inventory: inv, Recommend: rec, Store: store,
	}
}

var (
	tenantID    = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	principalID = uuid.MustParse("22222222-2222-2222-2222-222222222222")
	variantA    = uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	variantB    = uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
)

func TestAddLine(t *testing.T) {
	env := testDeps(t)
	ctx := context.Background()
	c, err := env.Deps.CreateCart(ctx, app.CreateCartInput{
		TenantID: tenantID, GuestToken: "guest-1", Currency: "TRY",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	c, err = env.Deps.AddLine(ctx, app.AddLineInput{
		TenantID: tenantID, CartID: c.ID, VariantID: variantA, Qty: 2, MaxQty: 10,
	})
	if err != nil {
		t.Fatalf("add: %v", err)
	}
	if len(c.Lines) != 1 || c.Lines[0].Qty != 2 {
		t.Fatalf("want 1 line qty=2 got %+v", c.Lines)
	}
	c, err = env.Deps.AddLine(ctx, app.AddLineInput{
		TenantID: tenantID, CartID: c.ID, VariantID: variantA, Qty: 3,
	})
	if err != nil {
		t.Fatalf("add again: %v", err)
	}
	if len(c.Lines) != 1 || c.Lines[0].Qty != 5 {
		t.Fatalf("want summed qty=5 got %+v", c.Lines)
	}
}

func TestMergeSumsQty(t *testing.T) {
	env := testDeps(t)
	ctx := context.Background()

	guest, err := env.Deps.CreateCart(ctx, app.CreateCartInput{
		TenantID: tenantID, GuestToken: "g-merge", Currency: "TRY",
	})
	if err != nil {
		t.Fatalf("guest create: %v", err)
	}
	guest, err = env.Deps.AddLine(ctx, app.AddLineInput{
		TenantID: tenantID, CartID: guest.ID, VariantID: variantA, Qty: 2, MaxQty: 20,
	})
	if err != nil {
		t.Fatalf("guest add: %v", err)
	}

	auth, err := env.Deps.CreateCart(ctx, app.CreateCartInput{
		TenantID: tenantID, PrincipalID: &principalID, Currency: "TRY",
	})
	if err != nil {
		t.Fatalf("auth create: %v", err)
	}
	auth, err = env.Deps.AddLine(ctx, app.AddLineInput{
		TenantID: tenantID, CartID: auth.ID, VariantID: variantA, Qty: 3, MaxQty: 20,
	})
	if err != nil {
		t.Fatalf("auth add A: %v", err)
	}
	_, err = env.Deps.AddLine(ctx, app.AddLineInput{
		TenantID: tenantID, CartID: auth.ID, VariantID: variantB, Qty: 1, MaxQty: 20,
	})
	if err != nil {
		t.Fatalf("auth add B: %v", err)
	}

	merged, err := env.Deps.MergeCarts(ctx, app.MergeCartsInput{
		TenantID: tenantID, GuestToken: "g-merge", PrincipalID: principalID,
		Policy: domain.MergePolicySumQty,
	})
	if err != nil {
		t.Fatalf("merge: %v", err)
	}
	if len(merged.Lines) != 2 {
		t.Fatalf("want 2 lines got %d", len(merged.Lines))
	}
	lineA, ok := merged.LineByVariant(variantA)
	if !ok || lineA.Qty != 5 {
		t.Fatalf("want variantA qty=5 got %+v", lineA)
	}
	guestAfter, err := env.Deps.GetCart(ctx, app.GetCartInput{TenantID: tenantID, CartID: &guest.ID})
	if err != nil {
		t.Fatalf("get guest: %v", err)
	}
	if guestAfter.Status != domain.CartStatusMerged {
		t.Fatalf("guest status want merged got %s", guestAfter.Status)
	}
}

func TestApplyCoupon(t *testing.T) {
	env := testDeps(t)
	ctx := context.Background()
	c, err := env.Deps.CreateCart(ctx, app.CreateCartInput{
		TenantID: tenantID, GuestToken: "g-coupon", Currency: "TRY",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	c, err = env.Deps.ApplyCoupon(ctx, app.ApplyCouponInput{
		TenantID: tenantID, CartID: c.ID, Code: "SAVE10",
	})
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	if !c.HasCoupon("SAVE10") {
		t.Fatalf("coupon not applied: %+v", c.Coupons)
	}
	_, err = env.Deps.ApplyCoupon(ctx, app.ApplyCouponInput{
		TenantID: tenantID, CartID: c.ID, Code: "save10",
	})
	if !errors.Is(err, domain.ErrAlreadyExists) {
		t.Fatalf("want already exists got %v", err)
	}
}

func TestRefreshQuoteCallsPricing(t *testing.T) {
	env := testDeps(t)
	ctx := context.Background()
	c, err := env.Deps.CreateCart(ctx, app.CreateCartInput{
		TenantID: tenantID, GuestToken: "g-quote", Currency: "TRY",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	c, err = env.Deps.AddLine(ctx, app.AddLineInput{
		TenantID: tenantID, CartID: c.ID, VariantID: variantA, Qty: 2,
	})
	if err != nil {
		t.Fatalf("add: %v", err)
	}
	c, err = env.Deps.RefreshQuote(ctx, app.RefreshQuoteInput{
		TenantID: tenantID, CartID: c.ID,
	})
	if err != nil {
		t.Fatalf("refresh: %v", err)
	}
	if env.Pricing.QuoteCalls.Load() != 1 {
		t.Fatalf("pricing quote calls=%d", env.Pricing.QuoteCalls.Load())
	}
	if c.Quote == nil || c.Quote.TotalMinor <= 0 {
		t.Fatalf("expected quote snapshot: %+v", c.Quote)
	}
	if env.Inventory.SoftReserveCalls.Load() != 0 {
		t.Fatalf("soft reserve should be optional/off, calls=%d", env.Inventory.SoftReserveCalls.Load())
	}
}

func TestRefreshQuoteSoftReserveOptional(t *testing.T) {
	env := testDeps(t)
	ctx := context.Background()
	c, err := env.Deps.CreateCart(ctx, app.CreateCartInput{
		TenantID: tenantID, GuestToken: "g-sr", Currency: "TRY",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	_, err = env.Deps.AddLine(ctx, app.AddLineInput{
		TenantID: tenantID, CartID: c.ID, VariantID: variantA, Qty: 1,
	})
	if err != nil {
		t.Fatalf("add: %v", err)
	}
	c, err = env.Deps.RefreshQuote(ctx, app.RefreshQuoteInput{
		TenantID: tenantID, CartID: c.ID, SoftReserve: true,
	})
	if err != nil {
		t.Fatalf("refresh+reserve: %v", err)
	}
	if env.Pricing.QuoteCalls.Load() != 1 {
		t.Fatalf("quote calls=%d", env.Pricing.QuoteCalls.Load())
	}
	if env.Inventory.SoftReserveCalls.Load() != 1 {
		t.Fatalf("soft reserve calls=%d", env.Inventory.SoftReserveCalls.Load())
	}
	if len(c.ReservationRefs) == 0 || c.ReservationRefs[0].ReservationRef == "" {
		t.Fatalf("expected reservation ref: %+v", c.ReservationRefs)
	}
}

func TestGuestToAuthMerge(t *testing.T) {
	env := testDeps(t)
	ctx := context.Background()
	guest, err := env.Deps.CreateCart(ctx, app.CreateCartInput{
		TenantID: tenantID, GuestToken: "login-guest", Currency: "TRY",
	})
	if err != nil {
		t.Fatalf("guest: %v", err)
	}
	_, err = env.Deps.AddLine(ctx, app.AddLineInput{
		TenantID: tenantID, CartID: guest.ID, VariantID: variantB, Qty: 4, MaxQty: 10,
	})
	if err != nil {
		t.Fatalf("add: %v", err)
	}
	// No existing auth cart — adopt guest.
	merged, err := env.Deps.MergeCarts(ctx, app.MergeCartsInput{
		TenantID: tenantID, GuestToken: "login-guest", PrincipalID: principalID,
	})
	if err != nil {
		t.Fatalf("merge adopt: %v", err)
	}
	if merged.PrincipalID == nil || *merged.PrincipalID != principalID {
		t.Fatalf("want principal ownership: %+v", merged.PrincipalID)
	}
	if merged.GuestToken != "" {
		t.Fatalf("guest token should be cleared")
	}
	line, ok := merged.LineByVariant(variantB)
	if !ok || line.Qty != 4 {
		t.Fatalf("want variantB qty=4 got %+v", line)
	}
}

func TestMaxQtyExceeded(t *testing.T) {
	env := testDeps(t)
	ctx := context.Background()
	c, err := env.Deps.CreateCart(ctx, app.CreateCartInput{
		TenantID: tenantID, GuestToken: "g-max", Currency: "TRY",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	_, err = env.Deps.AddLine(ctx, app.AddLineInput{
		TenantID: tenantID, CartID: c.ID, VariantID: variantA, Qty: 5, MaxQty: 5,
	})
	if err != nil {
		t.Fatalf("add: %v", err)
	}
	_, err = env.Deps.AddLine(ctx, app.AddLineInput{
		TenantID: tenantID, CartID: c.ID, VariantID: variantA, Qty: 1,
	})
	if !errors.Is(err, domain.ErrMaxQtyExceeded) {
		t.Fatalf("want max qty exceeded got %v", err)
	}
}
