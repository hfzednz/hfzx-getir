package httpadapter

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nexora/bff-admin/internal/app"
	"github.com/nexora/bff-admin/internal/authz"
)

func TestAdminRBAC(t *testing.T) {
	tid := "11111111-1111-1111-1111-111111111111"
	v := authz.Static{
		"cust": {ID: "c1", TenantID: tid, Roles: []string{"customer"}},
		"adm":  {ID: "a1", TenantID: tid, Roles: []string{"admin"}},
		"fin":  {ID: "f1", TenantID: tid, Roles: []string{"finance_analyst"}},
	}
	srv := NewServerWithAuth(":0", &app.Deps{}, v)
	ts := httptest.NewServer(srv.Handler)
	defer ts.Close()

	req := func(token, path string) int {
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
	if req("cust", "/v1/admin/dashboard") != 403 {
		t.Fatal("customer must be denied admin")
	}
	if req("fin", "/v1/admin/dashboard") != 403 {
		t.Fatal("finance must be denied admin")
	}
	if req("adm", "/v1/admin/dashboard") != 200 {
		t.Fatal("admin must pass")
	}
	if req("", "/health") != 200 {
		t.Fatal("health public")
	}
}
