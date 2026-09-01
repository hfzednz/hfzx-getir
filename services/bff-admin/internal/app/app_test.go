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
