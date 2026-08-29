package httpadapter

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nexora/finance-ledger-service/internal/app"
	"github.com/nexora/finance-ledger-service/internal/authz"
)

func TestLedgerRBAC(t *testing.T) {
	tid := "11111111-1111-1111-1111-111111111111"
	v := authz.Static{
		"cust": {ID: "c1", TenantID: tid, Roles: []string{"customer"}},
		"cour": {ID: "o1", TenantID: tid, Roles: []string{"courier"}},
		"wh":   {ID: "w1", TenantID: tid, Roles: []string{"picker"}},
		"fin":  {ID: "f1", TenantID: tid, Roles: []string{"finance_analyst"}},
	}
	h := NewHandler(ServerConfig{Deps: &app.Deps{}, Auth: v, CORSOrigins: []string{"*"}})
	ts := httptest.NewServer(h)
	defer ts.Close()
	hit := func(token string) int {
		r, _ := http.NewRequest(http.MethodGet, ts.URL+"/v1/ledger/journals", nil)
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
	if hit("cust") != 403 || hit("cour") != 403 || hit("wh") != 403 {
		t.Fatalf("cross-role finance must be 403")
	}
	if hit("") != 401 {
		t.Fatal("missing token")
	}
	code := hit("fin")
	if code == 401 || code == 403 {
		t.Fatalf("finance_analyst denied ledger: %d", code)
	}
}
