package app_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/pricing-service/internal/app"
	"github.com/nexora/pricing-service/internal/app/memory"
	"github.com/nexora/pricing-service/internal/domain"
)

type testEnv struct {
	Deps     *app.Deps
	Store    *memory.Store
	Promo    *memory.PromoClient
	Hints    *memory.HintClient
	Clock    *memory.Clock
	Tenant   uuid.UUID
	Variant  uuid.UUID
	Region   uuid.UUID
	Warehouse uuid.UUID
	BookID   uuid.UUID
}

func testDeps(t *testing.T) *testEnv {
	t.Helper()
	store := memory.NewStore()
	repos := memory.NewRepos(store)
	promo := &memory.PromoClient{}
	hints := &memory.HintClient{}
	clock := &memory.Clock{T: time.Date(2026, 8, 6, 14, 0, 0, 0, time.UTC)} // 14:00 UTC
	deps := &app.Deps{
		Prices: repos.Prices, Taxes: repos.Taxes, Dynamics: repos.Dynamics,
		Audits: repos.Audits, Outbox: repos.Outbox,
		Promo: promo, Hints: hints,
		Publisher: &memory.EventPublisher{S: store},
		Clock: clock, IDs: memory.IDGen{},
	}
	return &testEnv{
		Deps: deps, Store: store, Promo: promo, Hints: hints, Clock: clock,
		Tenant:    uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		Variant:   uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
		Region:    uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"),
		Warehouse: uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc"),
	}
}

func (e *testEnv) seedBook(t *testing.T) uuid.UUID {
	t.Helper()
	book, err := e.Deps.UpsertPriceBook(context.Background(), app.UpsertPriceBookInput{
		TenantID: e.Tenant, Name: "default", Currency: "TRY",
	})
	if err != nil {
		t.Fatalf("book: %v", err)
	}
	e.BookID = book.ID
	return book.ID
}

func (e *testEnv) upsert(t *testing.T, scope domain.PriceScope, scopeID *uuid.UUID, amount int64) domain.PriceEntry {
	t.Helper()
	if e.BookID == uuid.Nil {
		e.seedBook(t)
	}
	entry, err := e.Deps.UpsertPrice(context.Background(), app.UpsertPriceInput{
		TenantID: e.Tenant, PriceBookID: e.BookID, VariantID: e.Variant,
		Scope: scope, ScopeID: scopeID, AmountMinor: amount, Currency: "TRY",
	})
	if err != nil {
		t.Fatalf("upsert %s: %v", scope, err)
	}
	return entry
}

func TestBasePrice(t *testing.T) {
	env := testDeps(t)
	env.upsert(t, domain.ScopeBase, nil, 1000)

	res, err := env.Deps.GetPrice(context.Background(), app.GetPriceInput{
		TenantID: env.Tenant, VariantID: env.Variant, Currency: "TRY",
	})
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if res.AmountMinor != 1000 || res.Scope != domain.ScopeBase {
		t.Fatalf("got amount=%d scope=%s", res.AmountMinor, res.Scope)
	}
}

func TestRegionalOverridesBase(t *testing.T) {
	env := testDeps(t)
	env.upsert(t, domain.ScopeBase, nil, 1000)
	region := env.Region
	env.upsert(t, domain.ScopeRegional, &region, 900)

	res, err := env.Deps.GetPrice(context.Background(), app.GetPriceInput{
		TenantID: env.Tenant, VariantID: env.Variant, Currency: "TRY", RegionID: &region,
	})
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if res.AmountMinor != 900 || res.Scope != domain.ScopeRegional {
		t.Fatalf("got amount=%d scope=%s want 900 regional", res.AmountMinor, res.Scope)
	}
}

func TestWarehouseOverridesRegional(t *testing.T) {
	env := testDeps(t)
	env.upsert(t, domain.ScopeBase, nil, 1000)
	region := env.Region
	wh := env.Warehouse
	env.upsert(t, domain.ScopeRegional, &region, 900)
	env.upsert(t, domain.ScopeWarehouse, &wh, 850)

	res, err := env.Deps.GetPrice(context.Background(), app.GetPriceInput{
		TenantID: env.Tenant, VariantID: env.Variant, Currency: "TRY",
		RegionID: &region, WarehouseID: &wh,
	})
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if res.AmountMinor != 850 || res.Scope != domain.ScopeWarehouse {
		t.Fatalf("got amount=%d scope=%s want 850 warehouse", res.AmountMinor, res.Scope)
	}
}

func TestPromoDiscountAppliedViaMock(t *testing.T) {
	env := testDeps(t)
	env.upsert(t, domain.ScopeBase, nil, 1000)
	env.Promo.DiscountMinor = 200

	q, err := env.Deps.QuoteCart(context.Background(), app.QuoteCartInput{
		TenantID: env.Tenant, Currency: "TRY",
		Lines: []app.QuoteLineInput{{VariantID: env.Variant, Qty: 2}},
	})
	if err != nil {
		t.Fatalf("quote: %v", err)
	}
	if env.Promo.Calls.Load() != 1 {
		t.Fatalf("promo calls=%d", env.Promo.Calls.Load())
	}
	if q.SubtotalMinor != 2000 {
		t.Fatalf("subtotal=%d", q.SubtotalMinor)
	}
	if q.DiscountMinor != 200 {
		t.Fatalf("discount=%d want 200", q.DiscountMinor)
	}
	if len(q.Promos) == 0 {
		t.Fatal("expected promo attribution")
	}
}

func TestTaxAdded(t *testing.T) {
	env := testDeps(t)
	env.upsert(t, domain.ScopeBase, nil, 1000)
	_, err := env.Deps.UpsertTaxRule(context.Background(), app.UpsertTaxRuleInput{
		TenantID: env.Tenant, Code: "VAT18", Name: "VAT", RateBps: 1800, Priority: 1,
	})
	if err != nil {
		t.Fatalf("tax: %v", err)
	}

	q, err := env.Deps.QuoteCart(context.Background(), app.QuoteCartInput{
		TenantID: env.Tenant, Currency: "TRY",
		Lines: []app.QuoteLineInput{{VariantID: env.Variant, Qty: 1}},
	})
	if err != nil {
		t.Fatalf("quote: %v", err)
	}
	// 1000 * 18% = 180
	if q.TaxMinor != 180 {
		t.Fatalf("tax=%d want 180", q.TaxMinor)
	}
	if q.TotalMinor != 1180 {
		t.Fatalf("total=%d want 1180", q.TotalMinor)
	}
}

func TestDynamicTimeBump(t *testing.T) {
	env := testDeps(t)
	env.upsert(t, domain.ScopeBase, nil, 1000)
	// Clock is 14:00; bump +10% in window 12-18
	_, err := env.Deps.UpsertDynamicRule(context.Background(), app.UpsertDynamicRuleInput{
		TenantID: env.Tenant, Code: "PEAK", Kind: domain.DynamicKindPercent,
		Trigger: domain.TriggerTimeOfDay, AdjustmentBps: 1000,
		StartHour: 12, EndHour: 18, Priority: 10,
	})
	if err != nil {
		t.Fatalf("dyn: %v", err)
	}

	dyn, err := env.Deps.ApplyDynamic(context.Background(), app.ApplyDynamicInput{
		TenantID: env.Tenant, VariantID: env.Variant, UnitMinor: 1000, Currency: "TRY",
	})
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	if dyn.UnitMinor != 1100 {
		t.Fatalf("unit=%d want 1100", dyn.UnitMinor)
	}

	q, err := env.Deps.QuoteCart(context.Background(), app.QuoteCartInput{
		TenantID: env.Tenant, Currency: "TRY",
		Lines: []app.QuoteLineInput{{VariantID: env.Variant, Qty: 1}},
	})
	if err != nil {
		t.Fatalf("quote: %v", err)
	}
	if q.Lines[0].UnitPriceMinor != 1100 {
		t.Fatalf("line unit=%d want 1100", q.Lines[0].UnitPriceMinor)
	}
}

func TestMissingPriceError(t *testing.T) {
	env := testDeps(t)
	_, err := env.Deps.GetPrice(context.Background(), app.GetPriceInput{
		TenantID: env.Tenant, VariantID: env.Variant, Currency: "TRY",
	})
	if !errors.Is(err, domain.ErrPriceNotFound) {
		t.Fatalf("expected ErrPriceNotFound, got %v", err)
	}

	_, err = env.Deps.QuoteCart(context.Background(), app.QuoteCartInput{
		TenantID: env.Tenant, Currency: "TRY",
		Lines: []app.QuoteLineInput{{VariantID: env.Variant, Qty: 1}},
	})
	if !errors.Is(err, domain.ErrPriceNotFound) {
		t.Fatalf("quote expected ErrPriceNotFound, got %v", err)
	}
}

func TestSimulateQuoteNoEvent(t *testing.T) {
	env := testDeps(t)
	env.upsert(t, domain.ScopeBase, nil, 500)
	q, err := env.Deps.SimulateQuote(context.Background(), app.QuoteCartInput{
		TenantID: env.Tenant, Currency: "TRY",
		Lines: []app.QuoteLineInput{{VariantID: env.Variant, Qty: 1}},
	})
	if err != nil {
		t.Fatalf("sim: %v", err)
	}
	if !q.Simulated {
		t.Fatal("expected simulated")
	}
}
