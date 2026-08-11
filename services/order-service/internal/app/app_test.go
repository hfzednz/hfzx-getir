package app_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/order-service/internal/app"
	"github.com/nexora/order-service/internal/app/memory"
	"github.com/nexora/order-service/internal/domain"
)

type testEnv struct {
	Deps      *app.Deps
	Clock     *memory.Clock
	Inventory *memory.InventoryClient
	Payment   *memory.PaymentClient
	Warehouse *memory.WarehouseClient
	Dispatch  *memory.DispatchClient
	Store     *memory.Store
}

func testDeps(t *testing.T) *testEnv {
	t.Helper()
	store := memory.NewStore()
	orders, events, sagas, outbox, fulfills, returns, refunds := memory.NewRepos(store)
	clock := &memory.Clock{T: time.Date(2026, 8, 6, 12, 0, 0, 0, time.UTC)}
	inv := memory.NewInventoryClient()
	pay := &memory.PaymentClient{}
	wh := &memory.WarehouseClient{}
	disp := &memory.DispatchClient{}
	deps := &app.Deps{
		Orders: orders, Events: events, Sagas: sagas, Outbox: outbox,
		Fulfillments: fulfills, Returns: returns, Refunds: refunds,
		Search: &memory.SearchIndexer{S: store},
		Publisher: &memory.EventPublisher{S: store},
		Inventory: inv, Payment: pay, Warehouse: wh, Dispatch: disp,
		Clock: clock, IDs: memory.IDGen{},
		PlaceLock: &memory.PlaceLocker{S: store},
	}
	return &testEnv{
		Deps: deps, Clock: clock, Inventory: inv, Payment: pay,
		Warehouse: wh, Dispatch: disp, Store: store,
	}
}

func seedDraft(t *testing.T, env *testEnv, idemKey string) domain.Order {
	t.Helper()
	tenant := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	customer := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	wh := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	variant := uuid.MustParse("44444444-4444-4444-4444-444444444444")
	o, err := env.Deps.CreateDraft(context.Background(), app.CreateDraftInput{
		TenantID: tenant, CustomerPrincipalID: customer,
		Type: domain.OrderTypeInstant, Currency: "TRY",
		IdempotencyKey: idemKey,
		WarehouseIDs:   []uuid.UUID{wh},
		Lines: []app.CreateLineInput{{
			VariantID: variant, SKUCode: "SKU-1", TitleSnapshot: "Milk",
			Qty: 2, UnitPriceMinor: 1500, WarehouseID: &wh,
		}},
	})
	if err != nil {
		t.Fatalf("create draft: %v", err)
	}
	return o
}

func TestPlaceOrderHappyPath(t *testing.T) {
	env := testDeps(t)
	ctx := context.Background()
	o := seedDraft(t, env, "place-happy-1")

	placed, err := env.Deps.PlaceOrder(ctx, app.PlaceOrderInput{
		TenantID: o.TenantID, OrderID: o.ID, IdempotencyKey: "place-saga-1",
	})
	if err != nil {
		t.Fatalf("place: %v", err)
	}
	if placed.Status != domain.OrderStatusWarehouseAssigned {
		t.Fatalf("want warehouse_assigned got %s", placed.Status)
	}
	if placed.ReservationRef == "" || placed.PaymentIntentRef == "" {
		t.Fatalf("expected reservation and payment refs")
	}
	if env.Inventory.SoftReserveCalls.Load() != 1 {
		t.Fatalf("soft reserve calls=%d", env.Inventory.SoftReserveCalls.Load())
	}
	if env.Inventory.ConfirmHardCalls.Load() != 1 {
		t.Fatalf("confirm hard calls=%d", env.Inventory.ConfirmHardCalls.Load())
	}
	if env.Warehouse.ReceiveCalls.Load() != 1 {
		t.Fatalf("warehouse receive calls=%d", env.Warehouse.ReceiveCalls.Load())
	}
}

func TestPaymentFailReleasesAndCancels(t *testing.T) {
	env := testDeps(t)
	ctx := context.Background()
	o := seedDraft(t, env, "place-pay-fail")
	env.Payment.FailAuthorize = true

	placed, err := env.Deps.PlaceOrder(ctx, app.PlaceOrderInput{
		TenantID: o.TenantID, OrderID: o.ID, IdempotencyKey: "place-saga-payfail",
	})
	if err == nil {
		t.Fatal("expected payment failure")
	}
	if placed.Status != domain.OrderStatusCancelled && placed.Status != domain.OrderStatusPaymentFailed {
		t.Fatalf("want payment_failed/cancelled got %s", placed.Status)
	}
	if env.Inventory.ReleaseCalls.Load() < 1 {
		t.Fatalf("expected Release after payment fail, got %d", env.Inventory.ReleaseCalls.Load())
	}
	// Final status should be cancelled via compensate path.
	got, _ := env.Deps.GetOrder(ctx, o.TenantID, o.ID)
	if got.Status != domain.OrderStatusCancelled {
		t.Fatalf("want cancelled got %s", got.Status)
	}
}

func TestIllegalTransitionRejected(t *testing.T) {
	env := testDeps(t)
	ctx := context.Background()
	o := seedDraft(t, env, "illegal-1")

	_, err := env.Deps.InterveneStatus(ctx, app.InterveneStatusInput{
		TenantID: o.TenantID, OrderID: o.ID,
		NextStatus: domain.OrderStatusDelivered,
		Reason:     "bad jump",
	})
	if !errors.Is(err, domain.ErrInvalidTransition) {
		t.Fatalf("want invalid transition, got %v", err)
	}
}

func TestCancelBeforePickReleasesReservation(t *testing.T) {
	env := testDeps(t)
	ctx := context.Background()
	o := seedDraft(t, env, "cancel-prepick")

	placed, err := env.Deps.PlaceOrder(ctx, app.PlaceOrderInput{
		TenantID: o.TenantID, OrderID: o.ID, IdempotencyKey: "place-then-cancel",
	})
	if err != nil {
		t.Fatal(err)
	}
	if placed.Status != domain.OrderStatusWarehouseAssigned {
		t.Fatalf("setup want warehouse_assigned got %s", placed.Status)
	}
	beforeRelease := env.Inventory.ReleaseCalls.Load()

	cancelled, err := env.Deps.CancelOrder(ctx, app.CancelOrderInput{
		TenantID: o.TenantID, OrderID: o.ID, Reason: "customer changed mind",
	})
	if err != nil {
		t.Fatal(err)
	}
	if cancelled.Status != domain.OrderStatusCancelled {
		t.Fatalf("want cancelled got %s", cancelled.Status)
	}
	if env.Inventory.ReleaseCalls.Load() <= beforeRelease {
		t.Fatalf("expected Release on cancel before pick")
	}
}

func TestIdempotentPlaceOrderSameKey(t *testing.T) {
	env := testDeps(t)
	ctx := context.Background()
	o := seedDraft(t, env, "idem-order-1")

	a, err := env.Deps.PlaceOrder(ctx, app.PlaceOrderInput{
		TenantID: o.TenantID, OrderID: o.ID, IdempotencyKey: "same-place-key",
	})
	if err != nil {
		t.Fatal(err)
	}
	b, err := env.Deps.PlaceOrder(ctx, app.PlaceOrderInput{
		TenantID: o.TenantID, OrderID: o.ID, IdempotencyKey: "same-place-key",
	})
	if err != nil {
		t.Fatal(err)
	}
	if a.ID != b.ID {
		t.Fatalf("want same order id")
	}
	if env.Inventory.SoftReserveCalls.Load() != 1 {
		t.Fatalf("soft reserve should run once, got %d", env.Inventory.SoftReserveCalls.Load())
	}
}

func TestDeliveredToCompletedViaEvent(t *testing.T) {
	env := testDeps(t)
	ctx := context.Background()
	o := seedDraft(t, env, "deliver-1")

	placed, err := env.Deps.PlaceOrder(ctx, app.PlaceOrderInput{
		TenantID: o.TenantID, OrderID: o.ID, IdempotencyKey: "place-deliver",
	})
	if err != nil {
		t.Fatal(err)
	}

	// Advance through warehouse/dispatch events to delivered→completed.
	steps := []string{
		domain.EventPickingStarted,
		domain.EventPackingCompleted,
		domain.EventCourierAssigned,
		domain.EventOutForDelivery,
		domain.EventDelivered,
	}
	cur := placed
	for _, ev := range steps {
		cur, err = env.Deps.ApplyDispatchEvent(ctx, app.ApplyLifecycleEventInput{
			TenantID: cur.TenantID, OrderID: cur.ID, EventType: ev,
			Payload: map[string]any{},
		})
		if err != nil {
			t.Fatalf("event %s: %v", ev, err)
		}
	}
	if cur.Status != domain.OrderStatusCompleted {
		t.Fatalf("want completed got %s", cur.Status)
	}
}

func TestConcurrentPlaceSameIdempotencyNoDoubleReserve(t *testing.T) {
	env := testDeps(t)
	ctx := context.Background()
	o := seedDraft(t, env, "concurrent-place")

	const n = 16
	var wg sync.WaitGroup
	var success atomic.Int64
	ids := make([]uuid.UUID, n)
	wg.Add(n)
	for i := 0; i < n; i++ {
		i := i
		go func() {
			defer wg.Done()
			got, err := env.Deps.PlaceOrder(ctx, app.PlaceOrderInput{
				TenantID: o.TenantID, OrderID: o.ID, IdempotencyKey: "concurrent-key",
			})
			if err != nil {
				return
			}
			ids[i] = got.ID
			success.Add(1)
		}()
	}
	wg.Wait()
	if success.Load() == 0 {
		t.Fatal("expected at least one success")
	}
	if env.Inventory.SoftReserveCalls.Load() != 1 {
		t.Fatalf("soft reserve must run once under concurrency, got %d", env.Inventory.SoftReserveCalls.Load())
	}
	first := uuid.Nil
	for _, id := range ids {
		if id == uuid.Nil {
			continue
		}
		if first == uuid.Nil {
			first = id
			continue
		}
		if id != first {
			t.Fatalf("concurrent place returned different order ids")
		}
	}
}

func TestCreateDraftIdempotent(t *testing.T) {
	env := testDeps(t)
	a := seedDraft(t, env, "draft-idem")
	b := seedDraft(t, env, "draft-idem")
	if a.ID != b.ID {
		t.Fatalf("want same draft")
	}
}
