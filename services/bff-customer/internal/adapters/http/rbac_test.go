package httpadapter

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nexora/bff-customer/internal/app"
	"github.com/nexora/bff-customer/internal/authz"
)

func TestCustomerHomeRBAC(t *testing.T) {
	tid := "11111111-1111-1111-1111-111111111111"
	v := authz.Static{
		"cust": {ID: "c1", TenantID: tid, Roles: []string{"customer"}},
		"adm":  {ID: "a1", TenantID: tid, Roles: []string{"admin"}},
	}
	srv := NewServerWithAuth(":0", &app.Deps{}, v)
	ts := httptest.NewServer(srv.Handler)
	defer ts.Close()
	hit := func(token, path string) int {
		r, _ := http.NewRequest(http.MethodGet, ts.URL+path, nil)
		r.Header.Set("X-Tenant-Id", tid)
		if token != "" {
			r.Header.Set("Authorization", "Bearer "+token)
		}
		res, err := http.DefaultClient.Do(r)
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		return res.StatusCode
	}
	if hit("", "/health") != 200 {
		t.Fatal("health")
	}
	if hit("adm", "/v1/customer/home") != 403 {
		t.Fatal("admin must not use customer APIs")
	}
	code := hit("cust", "/v1/customer/home")
	if code == 401 || code == 403 {
		t.Fatalf("customer denied home %d", code)
	}
}
