package app_test

import (
	"context"
	"testing"

	"github.com/nexora/bff-customer/internal/app"
	"github.com/nexora/bff-customer/internal/app/memory"
	"github.com/nexora/bff-customer/internal/domain"
)

func TestCustomerJourney(t *testing.T) {
	ctx := context.Background()
	stubs := memory.NewStubs()
	d := &app.Deps{
		Identity: stubs, Catalog: stubs, Recs: stubs, Cart: stubs, Checkout: stubs,
		Orders: memory.OrderStub{S: stubs}, Tracking: stubs, Location: stubs,
		Notify: stubs, CRM: stubs, Reviews: stubs,
	}
	ch, err := d.StartOTP(ctx, "t1", "+905551112233")
	if err != nil || ch == "" {
		t.Fatal(err)
	}
	sess, err := d.Login(ctx, "t1", ch, "123456")
	if err != nil || sess.AccessToken == "" {
		t.Fatal(err)
	}
	home, err := d.Home(ctx, "t1", sess.CustomerID, "water", 41.01, 28.97)
	if err != nil || !home.Serviceable || len(home.Products) == 0 {
		t.Fatalf("%+v %v", home, err)
	}
	_, err = d.AddToCart(ctx, "t1", "cart1", "SKU1", 2, 1000)
	if err != nil {
		t.Fatal(err)
	}
	prev, err := d.PreviewCheckout(ctx, "t1", "cart1")
	if err != nil || prev.TotalMinor <= 0 {
		t.Fatal(err)
	}
	oid, err := d.PlaceOrder(ctx, "t1", "cart1", "card", prev.SessionID, domain.CheckoutAddress{
		Line1: "Istanbul", City: "Istanbul", Country: "TR", Lat: 41.0082, Lng: 28.9784,
	})
	if err != nil || oid == "" {
		t.Fatal(err)
	}
	tr, err := d.TrackOrder(ctx, "t1", oid)
	if err != nil || tr.Status == "" {
		t.Fatal(err)
	}
	_, err = d.OpenSupport(ctx, "t1", sess.CustomerID, "late delivery")
	if err != nil {
		t.Fatal(err)
	}
	if err := d.SubmitReview(ctx, "t1", oid, 5, "great"); err != nil {
		t.Fatal(err)
	}
}

type failingRecs struct{}

func (failingRecs) ForYou(context.Context, string, string) ([]map[string]any, error) {
	return nil, domain.ErrUpstream
}

func TestHomeSurvivesRecsOutage(t *testing.T) {
	stubs := memory.NewStubs()
	d := &app.Deps{Catalog: stubs, Recs: failingRecs{}, Location: stubs}
	home, err := d.Home(context.Background(), "t1", "cust_1", "water", 41.01, 28.97)
	if err != nil {
		t.Fatalf("home must not fail when recommendations are down: %v", err)
	}
	if len(home.Products) == 0 {
		t.Fatal("products must still be served when recommendations are down")
	}
	if len(home.Rails) != 0 {
		t.Fatalf("rails must be empty on recs failure, got %v", home.Rails)
	}
}

func TestHomeFailsWhenCatalogDown(t *testing.T) {
	stubs := memory.NewStubs()
	d := &app.Deps{Catalog: failingCatalog{}, Location: stubs}
	if _, err := d.Home(context.Background(), "t1", "", "water", 41.01, 28.97); err == nil {
		t.Fatal("catalog outage must surface as an error")
	}
}

type failingCatalog struct{}

func (failingCatalog) Search(context.Context, string, string) ([]map[string]any, error) {
	return nil, domain.ErrUpstream
}
