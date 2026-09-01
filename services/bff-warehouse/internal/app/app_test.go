package app_test

import (
	"context"
	"github.com/nexora/bff-warehouse/internal/app"
	"testing"
)

func TestWarehouseJourney(t *testing.T) {
	d := app.Deps{}
	if _, err := d.Pick(context.Background(), "t", "task1"); err != nil {
		t.Fatal(err)
	}
	if _, err := d.Pack(context.Background(), "t", "task1"); err != nil {
		t.Fatal(err)
	}
	if _, err := d.DispatchReady(context.Background(), "t", "task1"); err != nil {
		t.Fatal(err)
	}
	items, err := d.ListTasks(context.Background(), "t")
	if err != nil {
		t.Fatal(err)
	}
	if items == nil {
		t.Fatal("queue must be a list, not nil")
	}
	if len(items) != 0 {
		t.Fatalf("empty ORDER_URL must not invent tasks, got %+v", items)
	}
	scanned, err := d.ScanLine(context.Background(), "t", "task1", "line-1", "MILK-1L", 1)
	if err != nil {
		t.Fatal(err)
	}
	lines, _ := scanned["lines"].([]map[string]any)
	if len(lines) == 0 {
		t.Fatal("scan must persist a pick line")
	}
}
