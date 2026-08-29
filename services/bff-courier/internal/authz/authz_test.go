package authz

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGateDeniesCustomerOnAdmin(t *testing.T) {
	v := Static{
		"cust":   {ID: "c1", TenantID: "11111111-1111-1111-1111-111111111111", Roles: []string{"customer"}},
		"admin":  {ID: "a1", TenantID: "11111111-1111-1111-1111-111111111111", Roles: []string{"admin"}},
		"fin":    {ID: "f1", TenantID: "11111111-1111-1111-1111-111111111111", Roles: []string{"finance_analyst"}},
		"cour":   {ID: "o1", TenantID: "11111111-1111-1111-1111-111111111111", Roles: []string{"courier"}},
		"sup":    {ID: "s1", TenantID: "11111111-1111-1111-1111-111111111111", Roles: []string{"supplier"}},
		"supp":   {ID: "p1", TenantID: "11111111-1111-1111-1111-111111111111", Roles: []string{"support_agent"}},
		"wh":     {ID: "w1", TenantID: "11111111-1111-1111-1111-111111111111", Roles: []string{"picker"}},
		"ops":    {ID: "z1", TenantID: "11111111-1111-1111-1111-111111111111", Roles: []string{"city_ops"}},
		"sa":     {ID: "x1", TenantID: "11111111-1111-1111-1111-111111111111", Roles: []string{"super_admin"}},
		"wrongt": {ID: "a2", TenantID: "22222222-2222-2222-2222-222222222222", Roles: []string{"admin"}},
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(200) })
	mux.HandleFunc("GET /v1/admin/dashboard", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(200) })
	mux.HandleFunc("GET /v1/ledger/journals", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(200) })
	mux.HandleFunc("GET /v1/platform/admin/stats", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(200) })
	h := Gate(v, Options{
		Public: []string{"/health"},
		Rules: []Rule{
			{Prefix: "/v1/admin", Roles: []string{"admin", "super_admin", "support_agent", "city_ops"}},
			{Prefix: "/v1/ledger", Roles: []string{"finance_analyst", "admin", "super_admin"}},
			{Prefix: "/v1/platform", Roles: []string{"super_admin"}},
		},
	})(mux)

	tid := "11111111-1111-1111-1111-111111111111"
	assertStatus := func(token, path string, want int) {
		t.Helper()
		req := httptest.NewRequest(http.MethodGet, path, nil)
		req.Header.Set("X-Tenant-Id", tid)
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != want {
			t.Fatalf("%s %s: got %d want %d body=%s", token, path, rec.Code, want, rec.Body.String())
		}
	}
	assertStatus("", "/v1/admin/dashboard", 401)
	assertStatus("cust", "/v1/admin/dashboard", 403)
	assertStatus("cust", "/v1/ledger/journals", 403)
	assertStatus("cour", "/v1/ledger/journals", 403)
	assertStatus("sup", "/v1/admin/dashboard", 403)
	assertStatus("supp", "/v1/platform/admin/stats", 403)
	assertStatus("fin", "/v1/admin/dashboard", 403)
	assertStatus("wh", "/v1/ledger/journals", 403)
	assertStatus("admin", "/v1/admin/dashboard", 200)
	assertStatus("fin", "/v1/ledger/journals", 200)
	assertStatus("supp", "/v1/admin/dashboard", 200)
	assertStatus("ops", "/v1/admin/dashboard", 200)
	assertStatus("sa", "/v1/platform/admin/stats", 200)
	assertStatus("", "/health", 200)

	req := httptest.NewRequest(http.MethodGet, "/v1/admin/dashboard", nil)
	req.Header.Set("Authorization", "Bearer wrongt")
	req.Header.Set("X-Tenant-Id", tid)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != 404 {
		t.Fatalf("wrong tenant authorized role: got %d want 404", rec.Code)
	}
}

func TestSSETicketRoundTrip(t *testing.T) {
	secret := "test-sse-secret"
	tok, err := IssueSSETicket(secret, "t1", "c1", "order:abc", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	c, err := ParseSSETicket(secret, tok)
	if err != nil {
		t.Fatal(err)
	}
	if c.Sub != "c1" || c.Topic != "order:abc" || c.Tenant != "t1" {
		t.Fatalf("%+v", c)
	}
	if _, err := ParseSSETicket(secret, "nope"); err == nil {
		t.Fatal("expected invalid")
	}
	if _, err := ParseSSETicket("other", tok); err == nil {
		t.Fatal("expected mac fail")
	}
	if _, err := ParseSSETicket("", tok); err == nil {
		t.Fatal("empty secret must fail")
	}
}
