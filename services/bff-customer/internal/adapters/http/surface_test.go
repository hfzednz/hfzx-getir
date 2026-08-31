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

	code, kadikoy := hit(http.MethodGet, "/v1/customer/stores/store-kadikoy/products", "")
	if code != 200 {
		t.Fatalf("kadikoy products %d %+v", code, kadikoy)
	}
	if !storeHasSKU(kadikoy, "SKU1") {
		t.Fatalf("kadikoy must carry milk SKU1: %+v", kadikoy)
	}
	code, besiktas := hit(http.MethodGet, "/v1/customer/stores/store-besiktas/products", "")
	if code != 200 {
		t.Fatalf("besiktas products %d %+v", code, besiktas)
	}
	if storeHasSKU(besiktas, "SKU1") {
		t.Fatalf("besiktas must not carry milk SKU1: %+v", besiktas)
	}
	code, scoped := hit(http.MethodGet, "/v1/customer/search?q=SKU1&storeId=store-besiktas", "")
	if code != 200 {
		t.Fatalf("store-scoped search %d %+v", code, scoped)
	}
	if n, _ := scoped["total_count"].(float64); n != 0 {
		if items, _ := scoped["items"].([]any); len(items) != 0 {
			t.Fatalf("besiktas search must hide milk, got %+v", scoped)
		}
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

	code, _ = hit(http.MethodPost, "/v1/customer/cart/items", `{"cartId":"cart-history","sku":"SKU1","qty":7,"unitMinor":1999}`)
	if code != 201 && code != 200 {
		t.Fatalf("cart add for history %d", code)
	}
	prevHist, err := stubs.Preview(nil, tid, "cart-history")
	if err != nil {
		t.Fatal(err)
	}
	code, histPlaced := hit(http.MethodPost, "/v1/customer/checkout/place", `{"cartId":"cart-history","paymentMethod":"cash","sessionId":"`+prevHist.SessionID+`","principalId":"c1","address":{"line1":"Moda Cd 12","lat":40.98,"lng":29.02}}`)
	if code != 201 {
		t.Fatalf("history place %d %+v", code, histPlaced)
	}
	code, homeAfter := hit(http.MethodGet, "/v1/customer/home", "")
	if code != 200 {
		t.Fatalf("home after order %d", code)
	}
	if !homeHasWidget(homeAfter, "recently-ordered") {
		t.Fatalf("home must include recently-ordered after a real order: %+v", homeAfter)
	}
}

func storeHasSKU(body map[string]any, sku string) bool {
	raw, _ := body["items"].([]any)
	if raw == nil {
		raw, _ = body["products"].([]any)
	}
	for _, item := range raw {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if m["sku"] == sku || m["id"] == sku || m["productId"] == sku {
			return true
		}
	}
	return false
}

func homeHasWidget(body map[string]any, id string) bool {
	raw, _ := body["widgets"].([]any)
	for _, item := range raw {
		m, ok := item.(map[string]any)
		if ok && m["id"] == id {
			return true
		}
	}
	return false
}
