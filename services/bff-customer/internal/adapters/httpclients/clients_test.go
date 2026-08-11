package httpclients_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nexora/bff-customer/internal/adapters/httpclients"
	"github.com/nexora/bff-customer/internal/domain"
)

func TestIdentityOTP(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /v1/identity/auth/otp/start", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"challengeId": "ch-1", "expiresIn": 300})
	})
	mux.HandleFunc("POST /v1/identity/auth/otp/verify", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"accessToken": "atk", "refreshToken": "rtk", "principalId": "cust-1", "expiresIn": 3600,
		})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := httpclients.NewIdentity(srv.URL)
	ch, err := c.StartOTP(context.Background(), "t1", "+90555")
	if err != nil || ch != "ch-1" {
		t.Fatal(err, ch)
	}
	sess, err := c.VerifyOTP(context.Background(), "t1", ch, "123456")
	if err != nil {
		t.Fatal(err)
	}
	if sess.AccessToken != "atk" || sess.CustomerID != "cust-1" {
		t.Fatalf("%+v", sess)
	}
}

func TestCatalogSearchAndCheckoutPlace(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/catalog/search", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"Hits": []map[string]any{{"SKU": "SKU1", "Title": "water"}},
		})
	})
	mux.HandleFunc("POST /v1/checkout/sessions", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(201)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id": "sess-1", "cartId": "cart1", "currency": "TRY",
			"quote": map[string]any{"subtotalMinor": 1000, "totalMinor": 1000},
		})
	})
	mux.HandleFunc("POST /v1/checkout/sessions/sess-1/complete", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"orderId": "ord-9", "id": "sess-1"})
	})
	mux.HandleFunc("POST /v1/payments/eligibility", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"eligible": true, "methods": []string{"card"}})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	cat := httpclients.NewCatalog(srv.URL)
	items, err := cat.Search(context.Background(), "t1", "water")
	if err != nil || len(items) != 1 || items[0]["sku"] != "SKU1" {
		t.Fatalf("%v %v", items, err)
	}

	chk := httpclients.NewCheckout(srv.URL, srv.URL)
	prev, err := chk.Preview(context.Background(), "t1", "cart1")
	if err != nil || prev.TotalMinor != 1000 {
		t.Fatal(err, prev)
	}
	oid, err := chk.Place(context.Background(), "t1", "cart1", "card", prev.SessionID)
	if err != nil || oid != "ord-9" {
		t.Fatal(err, oid)
	}
}

func TestUpstreamMapsToDomain(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(502)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": map[string]any{"code": "internal_error", "message": "down"}})
	}))
	defer srv.Close()
	c := httpclients.NewCatalog(srv.URL)
	_, err := c.Search(context.Background(), "t1", "x")
	if err != domain.ErrUpstream {
		t.Fatalf("got %v", err)
	}
}
