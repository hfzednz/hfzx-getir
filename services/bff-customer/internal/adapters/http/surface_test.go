package httpadapter

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nexora/bff-customer/internal/app"
	"github.com/nexora/bff-customer/internal/app/memory"
	"github.com/nexora/bff-customer/internal/authz"
	"github.com/nexora/bff-customer/internal/domain"
)

func TestCustomerSurfaceSearchAndAddresses(t *testing.T) {
	tid := "11111111-1111-1111-1111-111111111111"
	stubs := memory.NewStubs()
	v := authz.Static{
		"cust": {ID: "c1", TenantID: tid, Roles: []string{"customer"}},
	}
	srv := NewServerWithAuth(":0", &app.Deps{
		Catalog: stubs, Stores: stubs, Location: stubs, Cart: stubs, Checkout: stubs,
		Orders: memory.OrderStub{S: stubs},
	}, v)
	ts := httptest.NewServer(srv.Handler)
	defer ts.Close()

	hit := func(method, path, body string) (int, map[string]any) {
		t.Helper()
		req, _ := http.NewRequest(method, ts.URL+path, strings.NewReader(body))
		req.Header.Set("X-Tenant-Id", tid)
		req.Header.Set("Authorization", "Bearer cust")
		req.Header.Set("Content-Type", "application/json")
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		var out map[string]any
		_ = json.NewDecoder(res.Body).Decode(&out)
		return res.StatusCode, out
	}

	code, search := hit(http.MethodGet, "/v1/customer/search?q=milk", "")
	if code != 200 {
		t.Fatalf("search %d", code)
	}
	if search["total_count"] == nil && search["items"] == nil {
		t.Fatalf("search body %+v", search)
	}

	code, _ = hit(http.MethodGet, "/v1/customer/categories", "")
	if code != 200 {
		t.Fatalf("categories %d", code)
	}
	code, _ = hit(http.MethodGet, "/v1/customer/stores", "")
	if code != 200 {
		t.Fatalf("stores %d", code)
	}

	code, created := hit(http.MethodPost, "/v1/customer/addresses", `{"formatted":"Moda Cd 12","lat":40.98,"lng":29.02}`)
	if code != 201 {
		t.Fatalf("create address %d %+v", code, created)
	}
	code, _ = hit(http.MethodGet, "/v1/customer/addresses", "")
	if code != 200 {
		t.Fatalf("list addresses %d", code)
	}

	code, _ = hit(http.MethodGet, "/v1/customer/orders", "")
	if code != 200 {
		t.Fatalf("orders %d", code)
	}

	prev, err := stubs.Preview(nil, tid, "cart1")
	if err != nil {
		t.Fatal(err)
	}
	code, placed := hit(http.MethodPost, "/v1/customer/checkout/place", `{"cartId":"cart1","paymentMethod":"card","sessionId":"`+prev.SessionID+`","principalId":"c1","address":{"line1":"Moda Cd 12","lat":40.98,"lng":29.02}}`)
	if code != 201 {
		t.Fatalf("place %d %+v", code, placed)
	}
	if domain.ErrInvalidArgument == nil {
		t.Fatal("sanity")
	}

	code, home := hit(http.MethodGet, "/v1/customer/home", "")
	if code != 200 {
		t.Fatalf("home %d", code)
	}
	if home["widgets"] == nil && home["products"] == nil {
		t.Fatalf("home body %+v", home)
	}

	code, profile := hit(http.MethodGet, "/v1/customer/profile", "")
	if code != 200 {
		t.Fatalf("profile %d %+v", code, profile)
	}
	code, patched := hit(http.MethodPatch, "/v1/customer/profile", `{"first_name":"Ada","last_name":"Lovelace"}`)
	if code != 200 {
		t.Fatalf("patch profile %d %+v", code, patched)
	}

	code, _ = hit(http.MethodGet, "/v1/customer/notifications", "")
	if code != 200 {
		t.Fatalf("notifications %d", code)
	}
	code, _ = hit(http.MethodGet, "/v1/customer/support/faq", "")
	if code != 200 {
		t.Fatalf("faq %d", code)
	}
	code, coupon := hit(http.MethodGet, "/v1/customer/coupons", "")
	if code != 200 {
		t.Fatalf("coupons %d %+v", code, coupon)
	}
	code, applied := hit(http.MethodPost, "/v1/customer/coupons/validate", `{"code":"WELCOME10","cart_subtotal_minor":20000}`)
	if code != 200 {
		t.Fatalf("coupon validate %d %+v", code, applied)
	}
	code, expired := hit(http.MethodPost, "/v1/customer/coupons/validate", `{"code":"EXPIRED","cart_subtotal_minor":20000}`)
	if code != 409 {
		t.Fatalf("expired coupon %d %+v", code, expired)
	}
	code, cards := hit(http.MethodGet, "/v1/customer/payment-methods/cards", "")
	if code != 200 {
		t.Fatalf("cards %d %+v", code, cards)
	}

	oid, _ := placed["orderId"].(string)
	if oid != "" {
		code, re := hit(http.MethodPost, "/v1/customer/orders/"+oid+"/reorder", "")
		if code != 200 {
			t.Fatalf("reorder %d %+v", code, re)
		}
	}
}
