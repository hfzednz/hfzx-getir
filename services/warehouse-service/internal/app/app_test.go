package app_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/warehouse-service/internal/app"
	"github.com/nexora/warehouse-service/internal/app/memory"
	"github.com/nexora/warehouse-service/internal/app/ports"
	"github.com/nexora/warehouse-service/internal/domain"
)

func testDeps(t *testing.T) (*app.Deps, *memory.Store, *memory.InventoryClient) {
	t.Helper()
	store := memory.NewStore()
	ff, tasks, picks, packs, disp, stations, wf, eq, qc, labels := memory.NewRepos(store)
	inv := &memory.InventoryClient{S: store}
	deps := &app.Deps{
		Fulfillments: ff, Tasks: tasks, Picks: picks, Packs: packs,
		Dispatches: disp, Stations: stations, Workforce: wf, Equipment: eq,
		QC: qc, Labels: labels,
		Inventory: inv,
		RouteAI:   memory.RouteOptimizer{},
		Events:    &memory.EventPublisher{S: store},
		Clock:     &memory.Clock{T: time.Date(2026, 8, 6, 12, 0, 0, 0, time.UTC)},
		IDs:       memory.IDGen{},
	}
	return deps, store, inv
}

func receiveSample(t *testing.T, d *app.Deps, tenant, wh uuid.UUID, orderID string, lines []app.ReceiveLineInput) domain.FulfillmentOrder {
	t.Helper()
	fo, err := d.ReceiveFulfillment(context.Background(), app.ReceiveFulfillmentCmd{
		TenantID: tenant, WarehouseID: wh, ExternalOrderID: orderID,
		Priority: 10, Lines: lines, IdempotencyKey: "idem-" + orderID,
	})
	if err != nil {
		t.Fatalf("receive: %v", err)
	}
	return fo
}

func TestHappyPathReceivePickPackHandoff(t *testing.T) {
	d, _, inv := testDeps(t)
	ctx := context.Background()
	tenant, wh := uuid.New(), uuid.New()
	picker, packer := uuid.New(), uuid.New()
	v1, v2 := uuid.New(), uuid.New()

	fo := receiveSample(t, d, tenant, wh, "ORD-1", []app.ReceiveLineInput{
		{VariantID: v1, SKUCode: "SKU-A", Barcode: "BC-A", LocationCode: "A-01", Qty: 2},
		{VariantID: v2, SKUCode: "SKU-B", Barcode: "BC-B", LocationCode: "B-02", Qty: 1},
	})
	if fo.Status != domain.FulfillmentStatusPickQueued {
		t.Fatalf("want pick_queued got %s", fo.Status)
	}
	if fo.ReservationID == nil {
		t.Fatal("expected reservation id")
	}
	if inv.SoftReserveCallCount() != 1 {
		t.Fatalf("SoftReserve want 1 got %d", inv.SoftReserveCallCount())
	}

	pickType := domain.TaskTypePick
	queued := domain.TaskStatusQueued
	tasks, _, err := d.ListTasks(ctx, ports.TaskFilter{
		TenantID: tenant, WarehouseID: wh, Type: &pickType, Status: &queued, Limit: 10,
	})
	if err != nil || len(tasks) != 1 {
		t.Fatalf("pick tasks: %v len=%d", err, len(tasks))
	}
	pickTask := tasks[0]

	claimed, err := d.ClaimPickTask(ctx, app.ClaimPickTaskCmd{
		TenantID: tenant, WarehouseID: wh, TaskID: &pickTask.ID, PickerID: picker,
	})
	if err != nil {
		t.Fatalf("claim pick: %v", err)
	}
	session, err := d.StartPick(ctx, app.StartPickCmd{
		TenantID: tenant, TaskID: claimed.ID, PickerID: picker,
	})
	if err != nil {
		t.Fatalf("start pick: %v", err)
	}

	for _, pl := range session.Lines {
		session, err = d.ScanPickLine(ctx, app.ScanPickLineCmd{
			TenantID: tenant, SessionID: session.ID, LineID: pl.ID,
			Barcode: pl.Barcode, Qty: pl.QtyRequired, PickerID: picker,
		})
		if err != nil {
			t.Fatalf("scan: %v", err)
		}
	}

	packTask, err := d.CompletePick(ctx, app.CompletePickCmd{
		TenantID: tenant, TaskID: claimed.ID, PickerID: picker,
	})
	if err != nil {
		t.Fatalf("complete pick: %v", err)
	}
	if packTask.Type != domain.TaskTypePack {
		t.Fatalf("want pack task got %s", packTask.Type)
	}

	station, err := d.CreateStation(ctx, app.CreateStationCmd{
		TenantID: tenant, WarehouseID: wh, Code: "PACK-1", Name: "Pack 1", Type: domain.StationTypePack,
	})
	if err != nil {
		t.Fatal(err)
	}
	pack, err := d.ClaimPack(ctx, app.ClaimPackCmd{
		TenantID: tenant, TaskID: packTask.ID, PackerID: packer, StationID: station.ID,
	})
	if err != nil {
		t.Fatalf("claim pack: %v", err)
	}
	pack, err = d.VerifyWeight(ctx, app.VerifyWeightCmd{
		TenantID: tenant, PackSessionID: pack.ID, ActualWeightG: pack.ExpectedWeightG, PackerID: packer,
	})
	if err != nil {
		t.Fatalf("verify weight: %v", err)
	}
	pack, err = d.Seal(ctx, app.SealPackCmd{TenantID: tenant, PackSessionID: pack.ID, PackerID: packer})
	if err != nil {
		t.Fatalf("seal: %v", err)
	}
	label, unit, err := d.GenerateLabel(ctx, app.GenerateLabelCmd{
		TenantID: tenant, PackSessionID: pack.ID, PackerID: packer,
	})
	if err != nil {
		t.Fatalf("label: %v", err)
	}
	if label.TrackingCode == "" || unit.Status != domain.DispatchStatusQueued {
		t.Fatalf("label/unit unexpected: %+v %+v", label, unit)
	}

	unit, err = d.DispatchVerify(ctx, app.DispatchVerifyCmd{
		TenantID: tenant, DispatchUnitID: unit.ID, TrackingCode: unit.TrackingCode,
	})
	if err != nil {
		t.Fatalf("dispatch verify: %v", err)
	}
	unit, err = d.HandoffConfirm(ctx, app.HandoffConfirmCmd{
		TenantID: tenant, DispatchUnitID: unit.ID, CourierRef: "CR-99",
	})
	if err != nil {
		t.Fatalf("handoff: %v", err)
	}
	if unit.Status != domain.DispatchStatusHandedOff {
		t.Fatalf("want handed_off got %s", unit.Status)
	}
	if inv.ConsumeCallCount() != 1 {
		t.Fatalf("Consume want 1 got %d", inv.ConsumeCallCount())
	}

	fo2, err := d.GetFulfillment(ctx, tenant, fo.ID)
	if err != nil {
		t.Fatal(err)
	}
	if fo2.Status != domain.FulfillmentStatusDispatched {
		t.Fatalf("want dispatched got %s", fo2.Status)
	}
}

func TestBadBarcodeFails(t *testing.T) {
	d, _, _ := testDeps(t)
	ctx := context.Background()
	tenant, wh, picker := uuid.New(), uuid.New(), uuid.New()
	fo := receiveSample(t, d, tenant, wh, "ORD-BAD", []app.ReceiveLineInput{
		{VariantID: uuid.New(), SKUCode: "SKU", Barcode: "GOOD", LocationCode: "A-01", Qty: 1},
	})
	_ = fo
	pickType := domain.TaskTypePick
	tasks, _, _ := d.ListTasks(ctx, ports.TaskFilter{TenantID: tenant, WarehouseID: wh, Type: &pickType, Limit: 1})
	claimed, _ := d.ClaimPickTask(ctx, app.ClaimPickTaskCmd{
		TenantID: tenant, WarehouseID: wh, TaskID: &tasks[0].ID, PickerID: picker,
	})
	session, _ := d.StartPick(ctx, app.StartPickCmd{TenantID: tenant, TaskID: claimed.ID, PickerID: picker})
	_, err := d.ScanPickLine(ctx, app.ScanPickLineCmd{
		TenantID: tenant, SessionID: session.ID, LineID: session.Lines[0].ID,
		Barcode: "WRONG", Qty: 1, PickerID: picker,
	})
	if !errors.Is(err, domain.ErrBarcodeMismatch) {
		t.Fatalf("want barcode mismatch got %v", err)
	}
}

func TestCannotCompletePickWithRemainingQty(t *testing.T) {
	d, _, _ := testDeps(t)
	ctx := context.Background()
	tenant, wh, picker := uuid.New(), uuid.New(), uuid.New()
	receiveSample(t, d, tenant, wh, "ORD-REM", []app.ReceiveLineInput{
		{VariantID: uuid.New(), SKUCode: "SKU", Barcode: "BC", LocationCode: "A-01", Qty: 3},
	})
	pickType := domain.TaskTypePick
	tasks, _, _ := d.ListTasks(ctx, ports.TaskFilter{TenantID: tenant, WarehouseID: wh, Type: &pickType, Limit: 1})
	claimed, _ := d.ClaimPickTask(ctx, app.ClaimPickTaskCmd{
		TenantID: tenant, WarehouseID: wh, TaskID: &tasks[0].ID, PickerID: picker,
	})
	session, _ := d.StartPick(ctx, app.StartPickCmd{TenantID: tenant, TaskID: claimed.ID, PickerID: picker})
	_, _ = d.ScanPickLine(ctx, app.ScanPickLineCmd{
		TenantID: tenant, SessionID: session.ID, LineID: session.Lines[0].ID,
		Barcode: "BC", Qty: 1, PickerID: picker,
	})
	_, err := d.CompletePick(ctx, app.CompletePickCmd{TenantID: tenant, TaskID: claimed.ID, PickerID: picker})
	if !errors.Is(err, domain.ErrRemainingQty) {
		t.Fatalf("want remaining qty got %v", err)
	}
}

func TestSoftReserveCalledOnReceive(t *testing.T) {
	d, _, inv := testDeps(t)
	tenant, wh := uuid.New(), uuid.New()
	receiveSample(t, d, tenant, wh, "ORD-INV", []app.ReceiveLineInput{
		{VariantID: uuid.New(), SKUCode: "SKU", Barcode: "BC", LocationCode: "A", Qty: 5},
	})
	if inv.SoftReserveCallCount() != 1 {
		t.Fatalf("want SoftReserve=1 got %d", inv.SoftReserveCallCount())
	}
	calls := inv.SoftReserveCalls()
	if len(calls) != 1 || len(calls[0].Lines) != 1 || calls[0].Lines[0].Qty != 5 {
		t.Fatalf("unexpected soft reserve payload: %+v", calls)
	}
}

func TestConsumeOnHandoff(t *testing.T) {
	d, _, inv := testDeps(t)
	ctx := context.Background()
	tenant, wh := uuid.New(), uuid.New()
	picker, packer := uuid.New(), uuid.New()

	receiveSample(t, d, tenant, wh, "ORD-CON", []app.ReceiveLineInput{
		{VariantID: uuid.New(), SKUCode: "SKU", Barcode: "BC", LocationCode: "A", Qty: 1},
	})
	pickType := domain.TaskTypePick
	tasks, _, _ := d.ListTasks(ctx, ports.TaskFilter{TenantID: tenant, WarehouseID: wh, Type: &pickType, Limit: 1})
	claimed, _ := d.ClaimPickTask(ctx, app.ClaimPickTaskCmd{
		TenantID: tenant, WarehouseID: wh, TaskID: &tasks[0].ID, PickerID: picker,
	})
	session, _ := d.StartPick(ctx, app.StartPickCmd{TenantID: tenant, TaskID: claimed.ID, PickerID: picker})
	_, _ = d.ScanPickLine(ctx, app.ScanPickLineCmd{
		TenantID: tenant, SessionID: session.ID, LineID: session.Lines[0].ID,
		Barcode: "BC", Qty: 1, PickerID: picker,
	})
	packTask, _ := d.CompletePick(ctx, app.CompletePickCmd{TenantID: tenant, TaskID: claimed.ID, PickerID: picker})
	station, _ := d.CreateStation(ctx, app.CreateStationCmd{
		TenantID: tenant, WarehouseID: wh, Code: "P1", Name: "P1",
	})
	pack, _ := d.ClaimPack(ctx, app.ClaimPackCmd{
		TenantID: tenant, TaskID: packTask.ID, PackerID: packer, StationID: station.ID,
	})
	pack, _ = d.VerifyWeight(ctx, app.VerifyWeightCmd{
		TenantID: tenant, PackSessionID: pack.ID, ActualWeightG: pack.ExpectedWeightG, PackerID: packer,
	})
	pack, _ = d.Seal(ctx, app.SealPackCmd{TenantID: tenant, PackSessionID: pack.ID, PackerID: packer})
	_, unit, _ := d.GenerateLabel(ctx, app.GenerateLabelCmd{TenantID: tenant, PackSessionID: pack.ID, PackerID: packer})
	_, _ = d.DispatchVerify(ctx, app.DispatchVerifyCmd{TenantID: tenant, DispatchUnitID: unit.ID})
	_, err := d.HandoffConfirm(ctx, app.HandoffConfirmCmd{TenantID: tenant, DispatchUnitID: unit.ID, CourierRef: "C1"})
	if err != nil {
		t.Fatal(err)
	}
	if inv.ConsumeCallCount() != 1 {
		t.Fatalf("Consume want 1 got %d", inv.ConsumeCallCount())
	}
}

func TestReassignTask(t *testing.T) {
	d, _, _ := testDeps(t)
	ctx := context.Background()
	tenant, wh := uuid.New(), uuid.New()
	a1, a2 := uuid.New(), uuid.New()
	receiveSample(t, d, tenant, wh, "ORD-RE", []app.ReceiveLineInput{
		{VariantID: uuid.New(), SKUCode: "SKU", Barcode: "BC", LocationCode: "A", Qty: 1},
	})
	pickType := domain.TaskTypePick
	tasks, _, _ := d.ListTasks(ctx, ports.TaskFilter{TenantID: tenant, WarehouseID: wh, Type: &pickType, Limit: 1})
	claimed, err := d.ClaimPickTask(ctx, app.ClaimPickTaskCmd{
		TenantID: tenant, WarehouseID: wh, TaskID: &tasks[0].ID, PickerID: a1,
	})
	if err != nil {
		t.Fatal(err)
	}
	re, err := d.ReassignTask(ctx, app.ReassignTaskCmd{
		TenantID: tenant, TaskID: claimed.ID, NewAssignee: a2, Note: "cover",
	})
	if err != nil {
		t.Fatal(err)
	}
	if re.AssigneeID == nil || *re.AssigneeID != a2 {
		t.Fatalf("want assignee %s got %v", a2, re.AssigneeID)
	}
	if len(re.History) < 2 {
		t.Fatalf("want history entries, got %d", len(re.History))
	}
}
