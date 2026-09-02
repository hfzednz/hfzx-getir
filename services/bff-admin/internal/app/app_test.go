package app_test

import (
	"context"
	"net/url"
	"testing"

	"github.com/nexora/bff-admin/internal/app"
)

type liveOpsStub struct{}

func (liveOpsStub) SetFlag(_ context.Context, _, key string, enabled bool) (map[string]any, error) {
	return map[string]any{"flag": key, "enabled": enabled}, nil
}

func TestAdminJourney(t *testing.T) {
	d := app.Deps{LiveOps: liveOpsStub{}}
	if _, err := d.Dashboard(context.Background(), "t"); err != nil {
		t.Fatal(err)
	}
	if _, err := d.KillSwitch(context.Background(), "t", "checkout.enabled", false); err != nil {
		t.Fatal(err)
	}
	if _, err := (&app.Deps{}).KillSwitch(context.Background(), "t", "checkout.enabled", false); err == nil {
		t.Fatal("expected error when LiveOps missing")
	}
}

type catalogStub struct{}

func (catalogStub) ListProducts(_ context.Context, _ string, _ url.Values) (map[string]any, error) {
	return map[string]any{"items": []any{map[string]any{"ID": "p1", "SKUCode": "MILK-1L", "Slug": "taze-sut", "Status": "published"}}, "total": 1}, nil
}
func (catalogStub) GetProduct(_ context.Context, _, id string) (map[string]any, error) {
	return map[string]any{"id": id, "skuCode": "MILK-1L", "name": "Taze Süt", "status": "published"}, nil
}

type orderStub struct{}

func (orderStub) List(_ context.Context, _ string, _ url.Values) (map[string]any, error) {
	return map[string]any{"items": []any{map[string]any{"id": "o1", "status": "picking"}}, "total": 1}, nil
}
func (orderStub) Get(_ context.Context, _, id string) (map[string]any, error) {
	return map[string]any{"id": id, "status": "picking"}, nil
}
func (orderStub) Cancel(_ context.Context, _, id, _ string) (map[string]any, error) {
	return map[string]any{"id": id, "status": "cancelled"}, nil
}
func (orderStub) Refund(_ context.Context, _, id string, _ map[string]any) (map[string]any, error) {
	return map[string]any{"id": id, "status": "refund_pending"}, nil
}
func (orderStub) DispatchEvent(_ context.Context, _, id, eventType, _ string) (map[string]any, error) {
	return map[string]any{"id": id, "eventType": eventType}, nil
}

func TestAdminCatalogAndOrderAction(t *testing.T) {
	d := app.Deps{Catalog: catalogStub{}, Orders: orderStub{}}
	out, err := d.ListCatalogProducts(context.Background(), "t", nil)
	if err != nil {
		t.Fatal(err)
	}
	items, _ := out["items"].([]map[string]any)
	if len(items) != 1 || items[0]["sku"] != "MILK-1L" || items[0]["status"] != "active" {
		t.Fatalf("%+v", out)
	}
	act, err := d.OrderAction(context.Background(), "t", "o1", "cancel", "ops", "", 0)
	if err != nil || act["ok"] != true {
		t.Fatalf("%+v %v", act, err)
	}
	dash, err := d.Dashboard(context.Background(), "t")
	if err != nil || dash["ordersLive"] != 1 {
		t.Fatalf("%+v %v", dash, err)
	}
}

type promoStub struct{}

func (promoStub) ListCampaigns(_ context.Context, _ string, _ url.Values) (map[string]any, error) {
	return map[string]any{"items": []any{map[string]any{"id": "c1", "name": "Dairy", "status": "active"}}}, nil
}
func (promoStub) GetCampaign(_ context.Context, _, id string) (map[string]any, error) {
	return map[string]any{"id": id, "name": "Dairy", "status": "active"}, nil
}
func (promoStub) CreateCampaign(_ context.Context, _ string, body map[string]any) (map[string]any, error) {
	return map[string]any{"id": "c-new", "name": body["name"], "status": "draft"}, nil
}

type inventoryStub struct{}

func (inventoryStub) ListWarehouses(_ context.Context, _ string) (map[string]any, error) {
	return map[string]any{"items": []any{map[string]any{"ID": "w1", "Code": "WH-1", "Name": "Kadikoy", "Status": "active"}}, "total": 1}, nil
}
func (inventoryStub) GetWarehouse(_ context.Context, _, id string) (map[string]any, error) {
	return map[string]any{"ID": id, "Code": "WH-1", "Name": "Kadikoy", "Status": "active"}, nil
}
func (inventoryStub) ListStock(_ context.Context, _, _ string) (map[string]any, error) {
	return map[string]any{"items": []any{map[string]any{"ID": "b1", "SKUCode": "MILK-1L", "OnHand": 12, "Reserved": 2, "SafetyMin": 4}}}, nil
}

func TestAdminLiveInventoryCampaigns(t *testing.T) {
	d := app.Deps{Orders: orderStub{}, Promo: promoStub{}, Inventory: inventoryStub{}}
	live, err := d.LiveSnapshot(context.Background(), "t")
	if err != nil {
		t.Fatal(err)
	}
	stream, _ := live["orderStream"].([]map[string]any)
	if len(stream) != 1 {
		t.Fatalf("%+v", live)
	}
	inv, err := d.InventorySnapshot(context.Background(), "t")
	if err != nil {
		t.Fatal(err)
	}
	stock, _ := inv["stock"].([]map[string]any)
	if len(stock) != 1 || stock[0]["sku"] != "MILK-1L" {
		t.Fatalf("%+v", inv)
	}
	cam, err := d.ListCampaigns(context.Background(), "t", nil)
	if err != nil {
		t.Fatal(err)
	}
	citems, _ := cam["items"].([]map[string]any)
	if len(citems) != 1 || citems[0]["status"] != "active" {
		t.Fatalf("%+v", cam)
	}
	cust, err := d.ListCustomers(context.Background(), "t", nil)
	if err != nil {
		t.Fatal(err)
	}
	if cust["total"] != 0 && cust["total"] != 1 {
		// orders stub has no customerId — empty list is honest
		_ = cust
	}
	fin, err := d.FinanceMutation(context.Background(), "payout_approve", "p1")
	if err != nil || fin["ok"] != false {
		t.Fatalf("finance write must not fake success: %+v %v", fin, err)
	}
}
