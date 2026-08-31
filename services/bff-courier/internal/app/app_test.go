package app_test

import (
	"context"
	"github.com/nexora/bff-courier/internal/app"
	"testing"
)

func TestCourierJourney(t *testing.T) {
	d := app.Deps{}
	on, err := d.Duty(context.Background(), "t", "c1", true)
	if err != nil {
		t.Fatal(err)
	}
	if on["onDuty"] != true {
		t.Fatalf("duty %+v", on)
	}
	got, err := d.GetDuty(context.Background(), "t", "c1")
	if err != nil {
		t.Fatal(err)
	}
	if got["onDuty"] != true {
		t.Fatalf("duty not persisted %+v", got)
	}
	if _, err := d.Offer(context.Background(), "t", "c1", "j1", true); err != nil {
		t.Fatal(err)
	}
	if _, err := d.Complete(context.Background(), "t", "j1"); err != nil {
		t.Fatal(err)
	}
	offers, err := d.ListOffers(context.Background(), "t", "c1")
	if err != nil {
		t.Fatal(err)
	}
	if offers == nil {
		t.Fatal("offers must be a list, not nil")
	}
	if len(offers) != 0 {
		t.Fatalf("empty ORDER_URL must not invent offers, got %+v", offers)
	}
}
