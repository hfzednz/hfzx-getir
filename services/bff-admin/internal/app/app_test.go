package app_test

import (
	"context"
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
