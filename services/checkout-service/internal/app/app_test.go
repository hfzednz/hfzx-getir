package app_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/checkout-service/internal/app"
	"github.com/nexora/checkout-service/internal/app/memory"
	"github.com/nexora/checkout-service/internal/app/ports"
	"github.com/nexora/checkout-service/internal/domain"
)

type testEnv struct {
	Deps      *app.Deps
	Store     *memory.Store
	Inventory *memory.InventoryClient
	Geofence  *memory.GeofenceClient
	Orders    *memory.OrderClient
	Tenant    uuid.UUID
	Principal uuid.UUID
	CartID    uuid.UUID
	Variant   uuid.UUID
}

func testDeps(t *testing.T) *testEnv {
	t.Helper()
	store := memory.NewStore()
	sessions, events, outbox := memory.NewRepos(store)
	clock := &memory.Clock{T: time.Date(2026, 8, 6, 12, 0, 0, 0, time.UTC)}
	inv := memory.NewInventoryClient()
	geo := &memory.GeofenceClient{}
	orders := &memory.OrderClient{}
	tenant := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	principal := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	cartID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	variant := uuid.MustParse("44444444-4444-4444-4444-444444444444")

	store.SeedCart(ports.CartView{
		ID: cartID, TenantID: tenant, PrincipalID: principal,
		CityID: "istanbul", Currency: "TRY", Active: true,
		Lines: []ports.CartLine{{
			VariantID: variant, SKUCode: "SKU-1", TitleSnapshot: "Milk",
			Qty: 2, UnitPriceMinor: 1500,
		}},
	})

	deps := &app.Deps{
		Sessions: sessions, Events: events, Outbox: outbox,
		Publisher: &memory.EventPublisher{S: store},
		Cart:      &memory.CartClient{S: store},
		Pricing:   &memory.PricingClient{},
		Inventory: inv,
		Geofence:  geo,
		Fraud:     &memory.FraudClient{},
		PayElig:   &memory.PaymentEligibilityClient{},
		Orders:    orders,
		Promo:     &memory.PromoClient{},
		Customer:  &memory.CustomerClient{},
		Clock:     clock,
		IDs:       memory.IDGen{},
		CompleteLock: &memory.CompleteLocker{S: store},
	}
	return &testEnv{
		Deps: deps, Store: store, Inventory: inv, Geofence: geo, Orders: orders,
		Tenant: tenant, Principal: principal, CartID: cartID, Variant: variant,
	}
}

func startAndAddress(t *testing.T, env *testEnv, idem string) domain.Session {
	t.Helper()
	ctx := context.Background()
	s, err := env.Deps.StartFromCart(ctx, app.StartFromCartInput{
		TenantID: env.Tenant, CartID: env.CartID, PrincipalID: env.Principal,
		IdempotencyKey: idem, DeliveryOption: domain.DeliveryInstant,
	})
	if err != nil {
		t.Fatalf("start: %v", err)
	}
	s, err = env.Deps.Patch(ctx, app.PatchInput{
		TenantID: env.Tenant, SessionID: s.ID,
		Address: &domain.AddressSnapshot{
			Line1: "Test St 1", City: "Istanbul", Lat: 41.0, Lng: 29.0,
		},
	})
	if err != nil {
		t.Fatalf("patch: %v", err)
	}
	return s
}

func TestStartValidateReady(t *testing.T) {
	env := testDeps(t)
	ctx := context.Background()
	s := startAndAddress(t, env, "start-validate-ready")

	s, err := env.Deps.Validate(ctx, app.ValidateInput{TenantID: env.Tenant, SessionID: s.ID})
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	if s.Status != domain.StatusReady {
		t.Fatalf("want ready got %s issues=%v", s.Status, s.Validation.Issues)
	}
	if !s.Validation.Passed {
		t.Fatalf("want passed")
	}
	if s.Quote.TotalMinor <= 0 {
		t.Fatalf("want quote total > 0")
	}
}

func TestZoneFailBlocks(t *testing.T) {
	env := testDeps(t)
	ctx := context.Background()
	env.Geofence.FailInZone = true
	s := startAndAddress(t, env, "zone-fail")

	s, err := env.Deps.Validate(ctx, app.ValidateInput{TenantID: env.Tenant, SessionID: s.ID})
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	if s.Status != domain.StatusBlocked {
		t.Fatalf("want blocked got %s", s.Status)
	}
	found := false
	for _, i := range s.Validation.Issues {
		if i.Code == domain.IssueZoneOutOfRange {
			found = true
		}
	}
	if !found {
		t.Fatalf("want zone_out_of_range issue, got %#v", s.Validation.Issues)
	}
}

func TestInventoryFailBlocks(t *testing.T) {
	env := testDeps(t)
	ctx := context.Background()
	env.Inventory.FailAll = true
	s := startAndAddress(t, env, "inv-fail")

	s, err := env.Deps.Validate(ctx, app.ValidateInput{TenantID: env.Tenant, SessionID: s.ID})
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	if s.Status != domain.StatusBlocked {
		t.Fatalf("want blocked got %s", s.Status)
	}
	found := false
	for _, i := range s.Validation.Issues {
		if i.Code == domain.IssueInventoryInsufficient {
			found = true
		}
	}
	if !found {
		t.Fatalf("want inventory_insufficient, got %#v", s.Validation.Issues)
	}
}

func TestCompleteCreatesOrderID(t *testing.T) {
	env := testDeps(t)
	ctx := context.Background()
	s := startAndAddress(t, env, "complete-order")
	s, err := env.Deps.Validate(ctx, app.ValidateInput{TenantID: env.Tenant, SessionID: s.ID})
	if err != nil {
		t.Fatal(err)
	}
	if s.Status != domain.StatusReady {
		t.Fatalf("setup want ready got %s", s.Status)
	}

	s, err = env.Deps.Complete(ctx, app.CompleteInput{
		TenantID: env.Tenant, SessionID: s.ID, IdempotencyKey: "complete-1",
	})
	if err != nil {
		t.Fatalf("complete: %v", err)
	}
	if s.Status != domain.StatusCompleted {
		t.Fatalf("want completed got %s", s.Status)
	}
	if s.OrderID == "" {
		t.Fatal("want order id")
	}
	if env.Orders.Calls.Load() != 1 {
		t.Fatalf("order create calls=%d", env.Orders.Calls.Load())
	}
}

func TestIdempotentComplete(t *testing.T) {
	env := testDeps(t)
	ctx := context.Background()
	s := startAndAddress(t, env, "idem-complete")
	s, err := env.Deps.Validate(ctx, app.ValidateInput{TenantID: env.Tenant, SessionID: s.ID})
	if err != nil {
		t.Fatal(err)
	}

	a, err := env.Deps.Complete(ctx, app.CompleteInput{
		TenantID: env.Tenant, SessionID: s.ID, IdempotencyKey: "same-complete-key",
	})
	if err != nil {
		t.Fatal(err)
	}
	b, err := env.Deps.Complete(ctx, app.CompleteInput{
		TenantID: env.Tenant, SessionID: s.ID, IdempotencyKey: "same-complete-key",
	})
	if err != nil {
		t.Fatal(err)
	}
	if a.OrderID != b.OrderID || a.ID != b.ID {
		t.Fatalf("want same session/order")
	}
	if env.Orders.Calls.Load() != 1 {
		t.Fatalf("create should run once, got %d", env.Orders.Calls.Load())
	}
}

func TestIllegalCompleteFromBlocked(t *testing.T) {
	env := testDeps(t)
	ctx := context.Background()
	env.Geofence.FailInZone = true
	s := startAndAddress(t, env, "illegal-complete")
	s, err := env.Deps.Validate(ctx, app.ValidateInput{TenantID: env.Tenant, SessionID: s.ID})
	if err != nil {
		t.Fatal(err)
	}
	if s.Status != domain.StatusBlocked {
		t.Fatalf("setup want blocked got %s", s.Status)
	}

	_, err = env.Deps.Complete(ctx, app.CompleteInput{
		TenantID: env.Tenant, SessionID: s.ID, IdempotencyKey: "blocked-complete",
	})
	if !errors.Is(err, domain.ErrInvalidTransition) {
		t.Fatalf("want invalid transition, got %v", err)
	}
	if env.Orders.Calls.Load() != 0 {
		t.Fatalf("order must not be created from blocked")
	}
}
