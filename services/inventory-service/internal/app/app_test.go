package app_test

import (
	"context"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/inventory-service/internal/app"
	"github.com/nexora/inventory-service/internal/app/memory"
	"github.com/nexora/inventory-service/internal/app/ports"
	"github.com/nexora/inventory-service/internal/domain"
)

func testDeps(t *testing.T) (*app.Deps, *memory.Clock) {
	t.Helper()
	store := memory.NewStore()
	wh, loc, bal, lots, res, mov, xfer, counts, rets, fc := memory.NewRepos(store)
	clock := &memory.Clock{T: time.Date(2026, 8, 6, 12, 0, 0, 0, time.UTC)}
	deps := &app.Deps{
		Warehouses: wh, Locations: loc, Balances: bal, Lots: lots,
		Reservations: res, Movements: mov, Transfers: xfer, Counts: counts,
		Returns: rets, Forecasts: fc,
		Search:      &memory.SearchIndexer{S: store},
		Events:      &memory.EventPublisher{S: store},
		AI:          memory.ForecastAIClient{},
		Idempotency: &memory.IdempotencyStore{S: store},
		Locker:      &memory.StockLocker{S: store},
		Clock:       clock,
		IDs:         memory.IDGen{},
		SoftReserveTTL: 15 * time.Minute,
	}
	return deps, clock
}

func ensureWarehouse(t *testing.T, d *app.Deps, tenant, wh uuid.UUID, code string) {
	t.Helper()
	ctx := context.Background()
	if _, err := d.Warehouses.GetByID(ctx, tenant, wh); err == nil {
		return
	}
	now := d.Clock.Now()
	if err := d.Warehouses.Create(ctx, domain.Warehouse{
		ID: wh, TenantID: tenant, Code: code, Name: code,
		Timezone: "UTC", Status: domain.WarehouseStatusActive,
		Metadata: map[string]any{}, CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatalf("create warehouse: %v", err)
	}
}

func seedStock(t *testing.T, d *app.Deps, tenant, wh, variant uuid.UUID, qty int64) {
	t.Helper()
	ensureWarehouse(t, d, tenant, wh, "WH-"+wh.String()[:8])
	_, _, err := d.Receive(context.Background(), app.ReceiveStockCmd{
		TenantID: tenant, WarehouseID: wh, VariantID: variant, SKUCode: "SKU-1",
		Qty: qty, IdempotencyKey: "seed-" + uuid.NewString(),
	})
	if err != nil {
		t.Fatalf("seed receive: %v", err)
	}
}

func TestSoftReserveReducesAvailable(t *testing.T) {
	d, _ := testDeps(t)
	ctx := context.Background()
	tenant, wh, variant := uuid.New(), uuid.New(), uuid.New()
	seedStock(t, d, tenant, wh, variant, 100)

	_, err := d.SoftReserve(ctx, app.SoftReserveCmd{
		TenantID: tenant, WarehouseID: &wh, ExternalRef: "cart-1",
		IdempotencyKey: "idem-soft-1",
		Lines: []app.SoftReserveLine{{
			WarehouseID: wh, VariantID: variant, SKUCode: "SKU-1", Qty: 30,
		}},
	})
	if err != nil {
		t.Fatalf("soft reserve: %v", err)
	}
	bal, err := d.Balances.GetByKey(ctx, ports.BalanceKey{
		TenantID: tenant, WarehouseID: wh, VariantID: variant,
	})
	if err != nil {
		t.Fatal(err)
	}
	if bal.Available() != 70 {
		t.Fatalf("available want 70 got %d (on_hand=%d reserved=%d)", bal.Available(), bal.OnHand, bal.Reserved)
	}
}

func TestOverReserveFails(t *testing.T) {
	d, _ := testDeps(t)
	ctx := context.Background()
	tenant, wh, variant := uuid.New(), uuid.New(), uuid.New()
	seedStock(t, d, tenant, wh, variant, 10)

	_, err := d.SoftReserve(ctx, app.SoftReserveCmd{
		TenantID: tenant, IdempotencyKey: "over-1",
		Lines: []app.SoftReserveLine{{
			WarehouseID: wh, VariantID: variant, Qty: 11,
		}},
	})
	if !errors.Is(err, domain.ErrInsufficientStock) {
		t.Fatalf("want insufficient stock, got %v", err)
	}
}

func TestExpireReleaseRestores(t *testing.T) {
	d, clock := testDeps(t)
	ctx := context.Background()
	tenant, wh, variant := uuid.New(), uuid.New(), uuid.New()
	seedStock(t, d, tenant, wh, variant, 50)

	res, err := d.SoftReserve(ctx, app.SoftReserveCmd{
		TenantID: tenant, IdempotencyKey: "exp-1", TTL: time.Minute,
		Lines: []app.SoftReserveLine{{WarehouseID: wh, VariantID: variant, Qty: 20}},
	})
	if err != nil {
		t.Fatal(err)
	}
	clock.T = clock.T.Add(2 * time.Minute)
	got, err := d.GetReservation(ctx, tenant, res.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != domain.ReservationStatusExpired {
		t.Fatalf("want expired got %s", got.Status)
	}
	bal, err := d.Balances.GetByKey(ctx, ports.BalanceKey{
		TenantID: tenant, WarehouseID: wh, VariantID: variant,
	})
	if err != nil {
		t.Fatal(err)
	}
	if bal.Available() != 50 {
		t.Fatalf("available restored want 50 got %d", bal.Available())
	}

	res2, err := d.SoftReserve(ctx, app.SoftReserveCmd{
		TenantID: tenant, IdempotencyKey: "rel-1",
		Lines: []app.SoftReserveLine{{WarehouseID: wh, VariantID: variant, Qty: 10}},
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = d.Release(ctx, app.ReleaseReservationCmd{
		TenantID: tenant, ReservationID: res2.ID, IdempotencyKey: "rel-key-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	bal, _ = d.Balances.GetByKey(ctx, ports.BalanceKey{
		TenantID: tenant, WarehouseID: wh, VariantID: variant,
	})
	if bal.Available() != 50 {
		t.Fatalf("after release want 50 got %d", bal.Available())
	}
}

func TestConcurrentSoftReservesNoOversell(t *testing.T) {
	d, _ := testDeps(t)
	ctx := context.Background()
	tenant, wh, variant := uuid.New(), uuid.New(), uuid.New()
	seedStock(t, d, tenant, wh, variant, 100)

	var okCount atomic.Int64
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := d.SoftReserve(ctx, app.SoftReserveCmd{
				TenantID: tenant, IdempotencyKey: "c-" + uuid.NewString(),
				Lines: []app.SoftReserveLine{{WarehouseID: wh, VariantID: variant, Qty: 10}},
			})
			if err == nil {
				okCount.Add(1)
			}
		}()
	}
	wg.Wait()
	if okCount.Load() != 10 {
		t.Fatalf("want exactly 10 successful reserves of 10, got %d", okCount.Load())
	}
	bal, _ := d.Balances.GetByKey(ctx, ports.BalanceKey{
		TenantID: tenant, WarehouseID: wh, VariantID: variant,
	})
	if bal.Available() != 0 || bal.Reserved != 100 {
		t.Fatalf("want reserved=100 available=0 got reserved=%d available=%d", bal.Reserved, bal.Available())
	}
}

func TestFEFOPicksEarlierLot(t *testing.T) {
	d, clock := testDeps(t)
	ctx := context.Background()
	tenant, wh, variant := uuid.New(), uuid.New(), uuid.New()
	seedStock(t, d, tenant, wh, variant, 100)
	bal, err := d.Balances.GetByKey(ctx, ports.BalanceKey{
		TenantID: tenant, WarehouseID: wh, VariantID: variant,
	})
	if err != nil {
		t.Fatal(err)
	}
	early := clock.T.AddDate(0, 0, 3)
	late := clock.T.AddDate(0, 0, 30)
	lotEarly := domain.Lot{
		ID: uuid.New(), TenantID: tenant, BalanceID: bal.ID, WarehouseID: wh, VariantID: variant,
		LotCode: "L-EARLY", Qty: 40, ExpiryDate: &early, Status: domain.LotStatusGood,
		Metadata: map[string]any{}, CreatedAt: clock.T, UpdatedAt: clock.T,
	}
	lotLate := domain.Lot{
		ID: uuid.New(), TenantID: tenant, BalanceID: bal.ID, WarehouseID: wh, VariantID: variant,
		LotCode: "L-LATE", Qty: 40, ExpiryDate: &late, Status: domain.LotStatusGood,
		Metadata: map[string]any{}, CreatedAt: clock.T, UpdatedAt: clock.T,
	}
	if err := d.Lots.Create(ctx, lotLate); err != nil {
		t.Fatal(err)
	}
	if err := d.Lots.Create(ctx, lotEarly); err != nil {
		t.Fatal(err)
	}
	allocs, err := d.AllocateFEFO(ctx, app.AllocateFEFOCmd{
		TenantID: tenant, WarehouseID: wh, VariantID: variant, Qty: 25,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(allocs) == 0 || allocs[0].LotID != lotEarly.ID {
		t.Fatalf("FEFO should pick early lot first, got %+v", allocs)
	}

	res, err := d.SoftReserve(ctx, app.SoftReserveCmd{
		TenantID: tenant, IdempotencyKey: "fefo-1",
		Lines: []app.SoftReserveLine{{
			WarehouseID: wh, VariantID: variant, Qty: 5, UseFEFO: true,
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Lines[0].LotID == nil || *res.Lines[0].LotID != lotEarly.ID {
		t.Fatalf("soft reserve FEFO lot want %s got %v", lotEarly.ID, res.Lines[0].LotID)
	}
}

func TestTransferCompletes(t *testing.T) {
	d, _ := testDeps(t)
	ctx := context.Background()
	tenant := uuid.New()
	from, to := uuid.New(), uuid.New()
	variant := uuid.New()
	seedStock(t, d, tenant, from, variant, 40)
	ensureWarehouse(t, d, tenant, to, "WH-TO")

	tr, err := d.CreateTransfer(ctx, app.CreateTransferCmd{
		TenantID: tenant, FromWarehouseID: from, ToWarehouseID: to,
		Lines: []app.CreateTransferLine{{VariantID: variant, SKUCode: "SKU-1", Qty: 15}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := d.ApproveTransfer(ctx, app.ApproveTransferCmd{TenantID: tenant, TransferID: tr.ID}); err != nil {
		t.Fatal(err)
	}
	done, err := d.CompleteTransfer(ctx, app.CompleteTransferCmd{
		TenantID: tenant, TransferID: tr.ID, IdempotencyKey: "xfer-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if done.Status != domain.TransferStatusCompleted {
		t.Fatalf("want completed got %s", done.Status)
	}
	src, _ := d.Balances.GetByKey(ctx, ports.BalanceKey{TenantID: tenant, WarehouseID: from, VariantID: variant})
	dst, _ := d.Balances.GetByKey(ctx, ports.BalanceKey{TenantID: tenant, WarehouseID: to, VariantID: variant})
	if src.OnHand != 25 || dst.OnHand != 15 {
		t.Fatalf("src=%d dst=%d", src.OnHand, dst.OnHand)
	}
}

func TestIllegalConsumeWithoutHardFails(t *testing.T) {
	d, _ := testDeps(t)
	ctx := context.Background()
	tenant, wh, variant := uuid.New(), uuid.New(), uuid.New()
	seedStock(t, d, tenant, wh, variant, 20)
	res, err := d.SoftReserve(ctx, app.SoftReserveCmd{
		TenantID: tenant, IdempotencyKey: "soft-only",
		Lines: []app.SoftReserveLine{{WarehouseID: wh, VariantID: variant, Qty: 5}},
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = d.Consume(ctx, app.ConsumeReservationCmd{
		TenantID: tenant, ReservationID: res.ID, IdempotencyKey: "consume-soft",
	})
	if err == nil {
		t.Fatal("expected consume of soft reservation to fail")
	}
	if !errors.Is(err, domain.ErrInvalidTransition) && !strings.Contains(err.Error(), "only hard") {
		t.Fatalf("want invalid transition, got %v", err)
	}

	if _, err := d.ConfirmHard(ctx, app.ConfirmHardCmd{
		TenantID: tenant, ReservationID: res.ID, IdempotencyKey: "to-hard",
	}); err != nil {
		t.Fatal(err)
	}
	consumed, err := d.Consume(ctx, app.ConsumeReservationCmd{
		TenantID: tenant, ReservationID: res.ID, IdempotencyKey: "consume-hard",
	})
	if err != nil {
		t.Fatal(err)
	}
	if consumed.Status != domain.ReservationStatusConsumed {
		t.Fatalf("want consumed got %s", consumed.Status)
	}
}
