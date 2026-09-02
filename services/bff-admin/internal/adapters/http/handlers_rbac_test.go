package httpadapter

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/nexora/bff-admin/internal/app"
	"github.com/nexora/bff-admin/internal/authz"
)

func TestAdminProductRoutesRBAC(t *testing.T) {
	tid := "11111111-1111-1111-1111-111111111111"
	v := authz.Static{
		"anon":   {},
		"cust":   {ID: "c1", TenantID: tid, Roles: []string{"customer"}},
		"cour":   {ID: "o1", TenantID: tid, Roles: []string{"courier"}},
		"merch":  {ID: "m1", TenantID: tid, Roles: []string{"merchant"}},
		"sup":    {ID: "s1", TenantID: tid, Roles: []string{"supplier"}},
		"fin":    {ID: "f1", TenantID: tid, Roles: []string{"finance_analyst"}},
		"admin":  {ID: "a1", TenantID: tid, Roles: []string{"admin"}},
		"sa":     {ID: "x1", TenantID: tid, Roles: []string{"super_admin"}},
		"wrongt": {ID: "a2", TenantID: "22222222-2222-2222-2222-222222222222", Roles: []string{"admin"}},
	}
	d := &app.Deps{Orders: orderListStub{}, Promo: couponOKStub{}, LiveOps: flagsOKStub{}, Settlement: settleOKStub{}}
	srv := NewServerWithAuth(":0", d, v)
	ts := httptest.NewServer(srv.Handler)
	t.Cleanup(ts.Close)

	hit := func(token, method, path string) int {
		t.Helper()
		req, _ := http.NewRequest(method, ts.URL+path, nil)
		req.Header.Set("X-Tenant-Id", tid)
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		return res.StatusCode
	}

	if code := hit("", http.MethodGet, "/v1/admin/reports"); code != 401 {
		t.Fatalf("anonymous reports %d", code)
	}
	if code := hit("cust", http.MethodGet, "/v1/admin/customers"); code != 403 {
		t.Fatalf("customer directory %d", code)
	}
	if code := hit("merch", http.MethodGet, "/v1/admin/monitoring"); code != 403 {
		t.Fatalf("merchant monitoring %d", code)
	}
	if code := hit("sup", http.MethodGet, "/v1/admin/loyalty"); code != 403 {
		t.Fatalf("supplier loyalty %d", code)
	}
	if code := hit("fin", http.MethodGet, "/v1/admin/reports"); code != 200 {
		t.Fatalf("finance reports %d", code)
	}
	if code := hit("fin", http.MethodGet, "/v1/admin/system"); code != 403 {
		t.Fatalf("finance system %d", code)
	}
	if code := hit("admin", http.MethodGet, "/v1/admin/reports"); code != 200 {
		t.Fatalf("admin reports %d", code)
	}
	if code := hit("admin", http.MethodGet, "/v1/admin/monitoring"); code != 200 {
		t.Fatalf("admin monitoring %d", code)
	}
	if code := hit("admin", http.MethodGet, "/v1/admin/ai"); code != 200 {
		t.Fatalf("admin ai %d", code)
	}
	if code := hit("admin", http.MethodGet, "/v1/admin/loyalty"); code != 200 {
		t.Fatalf("admin loyalty %d", code)
	}
	if code := hit("admin", http.MethodGet, "/v1/admin/notifications"); code != 200 {
		t.Fatalf("admin notifications %d", code)
	}
	if code := hit("admin", http.MethodGet, "/v1/admin/coupons"); code != 200 {
		t.Fatalf("admin coupons %d", code)
	}
	if code := hit("admin", http.MethodGet, "/v1/admin/customers"); code != 200 {
		t.Fatalf("admin customers %d", code)
	}
	if code := hit("wrongt", http.MethodGet, "/v1/admin/reports"); code != 404 {
		t.Fatalf("wrong tenant %d", code)
	}
	if code := hit("sa", http.MethodPost, "/v1/admin/finance/payouts/b1/approve"); code != 200 {
		t.Fatalf("super-admin payout %d", code)
	}
}

type orderListStub struct{}

func (orderListStub) List(_ context.Context, _ string, _ url.Values) (map[string]any, error) {
	return map[string]any{"items": []any{}}, nil
}
func (orderListStub) Get(_ context.Context, _, id string) (map[string]any, error) {
	return map[string]any{"id": id}, nil
}
func (orderListStub) Cancel(_ context.Context, _, id, _ string) (map[string]any, error) {
	return map[string]any{"id": id}, nil
}
func (orderListStub) Refund(_ context.Context, _, id string, _ map[string]any) (map[string]any, error) {
	return map[string]any{"id": id}, nil
}
func (orderListStub) DispatchEvent(_ context.Context, _, id, _, _ string) (map[string]any, error) {
	return map[string]any{"id": id}, nil
}

type couponOKStub struct{}

func (couponOKStub) ListCampaigns(_ context.Context, _ string, _ url.Values) (map[string]any, error) {
	return map[string]any{"items": []any{}}, nil
}
func (couponOKStub) GetCampaign(_ context.Context, _, id string) (map[string]any, error) {
	return map[string]any{"id": id}, nil
}
func (couponOKStub) CreateCampaign(_ context.Context, _ string, _ map[string]any) (map[string]any, error) {
	return map[string]any{}, nil
}
func (couponOKStub) ListCoupons(_ context.Context, _ string, _ url.Values) (map[string]any, error) {
	return map[string]any{"items": []any{}}, nil
}
func (couponOKStub) GetCoupon(_ context.Context, _, code string) (map[string]any, error) {
	return map[string]any{"code": code}, nil
}
func (couponOKStub) CreateCoupon(_ context.Context, _ string, _ map[string]any) (map[string]any, error) {
	return map[string]any{}, nil
}
func (couponOKStub) UpdateCoupon(_ context.Context, _, _ string, _ map[string]any) (map[string]any, error) {
	return map[string]any{}, nil
}

type flagsOKStub struct{}

func (flagsOKStub) SetFlag(_ context.Context, _, key string, enabled bool) (map[string]any, error) {
	return map[string]any{"key": key, "enabled": enabled}, nil
}
func (flagsOKStub) ListFlags(_ context.Context, _ string) (map[string]any, error) {
	return map[string]any{"items": []any{}}, nil
}

type settleOKStub struct{}

func (settleOKStub) ListBatches(_ context.Context, _ string) (map[string]any, error) {
	return map[string]any{"items": []any{}}, nil
}
func (settleOKStub) Approve(_ context.Context, _, id string) (map[string]any, error) {
	return map[string]any{"id": id, "status": "approved"}, nil
}
func (settleOKStub) Execute(_ context.Context, _, id string) (map[string]any, error) {
	return map[string]any{"id": id, "status": "processing"}, nil
}
