package httpadapter

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nexora/platform-ops-service/internal/app"
	"github.com/nexora/platform-ops-service/internal/app/memory"
	"github.com/nexora/platform-ops-service/internal/authz"
)

func TestPlatformRegistryRBAC(t *testing.T) {
	tid := "11111111-1111-1111-1111-111111111111"
	v := authz.Static{
		"cust": {ID: "c1", TenantID: tid, Roles: []string{"customer"}},
		"adm":  {ID: "a1", TenantID: tid, Roles: []string{"admin"}},
		"sa":   {ID: "s1", TenantID: tid, Roles: []string{"super_admin"}},
	}
	deps := &app.Deps{Clock: app.SystemClock{}, IDs: app.UUIDGen{}, Registry: memory.NewRegistry()}
	h := NewHandler(ServerConfig{Addr: ":0", Deps: deps, Auth: v, RateLimitPerMinute: 10000})
	ts := httptest.NewServer(h)
	defer ts.Close()

	req := func(method, token, path, body string) int {
		var r *http.Request
		if body != "" {
			r, _ = http.NewRequest(method, ts.URL+path, strings.NewReader(body))
			r.Header.Set("Content-Type", "application/json")
		} else {
			r, _ = http.NewRequest(method, ts.URL+path, nil)
		}
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
	if req(http.MethodGet, "", "/v1/platform/tenants", "") != 401 {
		t.Fatal("anonymous must fail")
	}
	if req(http.MethodGet, "cust", "/v1/platform/tenants", "") != 403 {
		t.Fatal("customer must fail")
	}
	if req(http.MethodGet, "adm", "/v1/platform/tenants", "") != 403 {
		t.Fatal("admin must not get super-admin tenants")
	}
	if st := req(http.MethodGet, "sa", "/v1/platform/tenants", ""); st != 200 {
		t.Fatalf("super_admin list tenants got %d", st)
	}
	if st := req(http.MethodPost, "sa", "/v1/platform/companies", `{"legalName":"Acme A.S.","tradeName":"Acme","countryCode":"TR","primaryCurrency":"TRY"}`); st != 201 {
		t.Fatalf("create company got %d", st)
	}
	if st := req(http.MethodGet, "sa", "/v1/platform/roles", ""); st != 200 {
		t.Fatalf("roles got %d", st)
	}
	if st := req(http.MethodGet, "sa", "/v1/platform/org", ""); st != 200 {
		t.Fatalf("org got %d", st)
	}
	if st := req(http.MethodGet, "sa", "/v1/platform/audit", ""); st != 200 {
		t.Fatalf("audit got %d", st)
	}
}
