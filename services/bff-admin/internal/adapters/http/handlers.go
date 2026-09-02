package httpadapter

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/nexora/bff-admin/internal/app"
	"github.com/nexora/bff-admin/internal/authz"
)

func writeErr(w http.ResponseWriter, status int, code, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]any{"code": code, "message": msg, "traceId": "", "retriable": status >= 500},
	})
}

func writeOK(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func tenant(r *http.Request) string {
	t := r.Header.Get("X-Tenant-Id")
	if t == "" {
		t = r.Header.Get("X-Nexora-Tenant")
	}
	return t
}

func NewServer(addr string, d *app.Deps) *http.Server {
	return NewServerWithAuth(addr, d, authz.FromEnv())
}

func NewServerWithAuth(addr string, d *app.Deps, v authz.Validator) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		writeOK(w, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /ready", func(w http.ResponseWriter, _ *http.Request) {
		writeOK(w, map[string]string{"status": "ready"})
	})
	mux.HandleFunc("GET /v1/admin/dashboard", func(w http.ResponseWriter, r *http.Request) {
		res, err := d.Dashboard(r.Context(), tenant(r))
		if err != nil {
			writeErr(w, 400, "invalid_argument", err.Error())
			return
		}
		writeOK(w, res)
	})
	mux.HandleFunc("GET /v1/admin/orders", func(w http.ResponseWriter, r *http.Request) {
		res, err := d.ListOrders(r.Context(), tenant(r), r.URL.Query())
		if err != nil {
			writeErr(w, 502, "upstream_error", err.Error())
			return
		}
		writeOK(w, res)
	})
	mux.HandleFunc("GET /v1/admin/orders/{id}", func(w http.ResponseWriter, r *http.Request) {
		res, err := d.GetOrder(r.Context(), tenant(r), r.PathValue("id"))
		if err != nil {
			writeErr(w, 502, "upstream_error", err.Error())
			return
		}
		writeOK(w, res)
	})
	mux.HandleFunc("POST /v1/admin/orders/{id}/actions", func(w http.ResponseWriter, r *http.Request) {
		var b struct {
			Action      string `json:"action"`
			Reason      string `json:"reason"`
			CourierID   string `json:"courierId"`
			RefundMinor int64  `json:"refundMinor"`
		}
		_ = json.NewDecoder(r.Body).Decode(&b)
		res, err := d.OrderAction(r.Context(), tenant(r), r.PathValue("id"), b.Action, b.Reason, b.CourierID, b.RefundMinor)
		if err != nil {
			writeErr(w, 400, "invalid_argument", err.Error())
			return
		}
		writeOK(w, res)
	})
	mux.HandleFunc("GET /v1/admin/catalog/products", func(w http.ResponseWriter, r *http.Request) {
		res, err := d.ListCatalogProducts(r.Context(), tenant(r), r.URL.Query())
		if err != nil {
			writeErr(w, 502, "upstream_error", err.Error())
			return
		}
		writeOK(w, res)
	})
	mux.HandleFunc("GET /v1/admin/catalog/products/{id}", func(w http.ResponseWriter, r *http.Request) {
		res, err := d.GetCatalogProduct(r.Context(), tenant(r), r.PathValue("id"))
		if err != nil {
			writeErr(w, 502, "upstream_error", err.Error())
			return
		}
		writeOK(w, res)
	})
	mux.HandleFunc("GET /v1/admin/support/tickets", func(w http.ResponseWriter, r *http.Request) {
		res, err := d.ListSupportTickets(r.Context(), tenant(r))
		if err != nil {
			writeErr(w, 502, "upstream_error", err.Error())
			return
		}
		writeOK(w, res)
	})
	mux.HandleFunc("GET /v1/admin/support/tickets/{id}", func(w http.ResponseWriter, r *http.Request) {
		res, err := d.GetSupportTicket(r.Context(), tenant(r), r.PathValue("id"))
		if err != nil {
			writeErr(w, 502, "upstream_error", err.Error())
			return
		}
		writeOK(w, res)
	})
	mux.HandleFunc("POST /v1/admin/support/tickets/{id}/escalate", func(w http.ResponseWriter, r *http.Request) {
		var b struct {
			Reason string `json:"reason"`
		}
		_ = json.NewDecoder(r.Body).Decode(&b)
		res, err := d.EscalateSupportTicket(r.Context(), tenant(r), r.PathValue("id"), b.Reason)
		if err != nil {
			writeErr(w, 502, "upstream_error", err.Error())
			return
		}
		writeOK(w, res)
	})
	mux.HandleFunc("POST /v1/admin/support/tickets/{id}/resolve", func(w http.ResponseWriter, r *http.Request) {
		var b struct {
			Note string `json:"note"`
		}
		_ = json.NewDecoder(r.Body).Decode(&b)
		res, err := d.ResolveSupportTicket(r.Context(), tenant(r), r.PathValue("id"), b.Note)
		if err != nil {
			writeErr(w, 502, "upstream_error", err.Error())
			return
		}
		writeOK(w, res)
	})
	mux.HandleFunc("GET /v1/admin/finance/journals", func(w http.ResponseWriter, r *http.Request) {
		res, err := d.FinanceSnapshot(r.Context(), tenant(r))
		if err != nil {
			writeErr(w, 502, "upstream_error", err.Error())
			return
		}
		writeOK(w, res)
	})
	mux.HandleFunc("POST /v1/admin/finance/payouts/{id}/approve", func(w http.ResponseWriter, r *http.Request) {
		res, _ := d.FinanceMutation(r.Context(), "payout_approve", r.PathValue("id"))
		writeErr(w, 422, "not_supported", fmt.Sprint(res["message"]))
	})
	mux.HandleFunc("POST /v1/admin/finance/settlements/{id}/settle", func(w http.ResponseWriter, r *http.Request) {
		res, _ := d.FinanceMutation(r.Context(), "courier_settle", r.PathValue("id"))
		writeErr(w, 422, "not_supported", fmt.Sprint(res["message"]))
	})
	mux.HandleFunc("POST /v1/admin/finance/refunds/{id}/approve", func(w http.ResponseWriter, r *http.Request) {
		res, _ := d.FinanceMutation(r.Context(), "refund_approve", r.PathValue("id"))
		writeErr(w, 422, "not_supported", fmt.Sprint(res["message"]))
	})
	mux.HandleFunc("GET /v1/admin/campaigns", func(w http.ResponseWriter, r *http.Request) {
		res, err := d.ListCampaigns(r.Context(), tenant(r), r.URL.Query())
		if err != nil {
			writeErr(w, 502, "upstream_error", err.Error())
			return
		}
		writeOK(w, res)
	})
	mux.HandleFunc("GET /v1/admin/campaigns/{id}", func(w http.ResponseWriter, r *http.Request) {
		res, err := d.GetCampaign(r.Context(), tenant(r), r.PathValue("id"))
		if err != nil {
			writeErr(w, 502, "upstream_error", err.Error())
			return
		}
		writeOK(w, res)
	})
	mux.HandleFunc("POST /v1/admin/campaigns", func(w http.ResponseWriter, r *http.Request) {
		var b struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			StartsAt    any    `json:"startsAt"`
			EndsAt      any    `json:"endsAt"`
		}
		_ = json.NewDecoder(r.Body).Decode(&b)
		res, err := d.CreateCampaign(r.Context(), tenant(r), b.Name, b.Description, b.StartsAt, b.EndsAt)
		if err != nil {
			writeErr(w, 400, "invalid_argument", err.Error())
			return
		}
		writeOK(w, res)
	})
	mux.HandleFunc("GET /v1/admin/pricing", func(w http.ResponseWriter, r *http.Request) {
		res, err := d.ListPricing(r.Context(), tenant(r))
		if err != nil {
			writeErr(w, 502, "upstream_error", err.Error())
			return
		}
		writeOK(w, res)
	})
	mux.HandleFunc("GET /v1/admin/inventory", func(w http.ResponseWriter, r *http.Request) {
		res, err := d.InventorySnapshot(r.Context(), tenant(r))
		if err != nil {
			writeErr(w, 502, "upstream_error", err.Error())
			return
		}
		writeOK(w, res)
	})
	mux.HandleFunc("GET /v1/admin/warehouses", func(w http.ResponseWriter, r *http.Request) {
		res, err := d.ListWarehouses(r.Context(), tenant(r))
		if err != nil {
			writeErr(w, 502, "upstream_error", err.Error())
			return
		}
		writeOK(w, res)
	})
	mux.HandleFunc("GET /v1/admin/warehouses/{id}", func(w http.ResponseWriter, r *http.Request) {
		res, err := d.GetWarehouse(r.Context(), tenant(r), r.PathValue("id"))
		if err != nil {
			writeErr(w, 502, "upstream_error", err.Error())
			return
		}
		writeOK(w, res)
	})
	mux.HandleFunc("GET /v1/admin/live/snapshot", func(w http.ResponseWriter, r *http.Request) {
		res, err := d.LiveSnapshot(r.Context(), tenant(r))
		if err != nil {
			writeErr(w, 502, "upstream_error", err.Error())
			return
		}
		writeOK(w, res)
	})
	mux.HandleFunc("GET /v1/admin/audit", func(w http.ResponseWriter, r *http.Request) {
		res, err := d.AuditSnapshot(r.Context(), tenant(r))
		if err != nil {
			writeErr(w, 502, "upstream_error", err.Error())
			return
		}
		writeOK(w, res)
	})
	mux.HandleFunc("GET /v1/admin/rbac", func(w http.ResponseWriter, r *http.Request) {
		res, err := d.RbacSnapshot(r.Context(), tenant(r))
		if err != nil {
			writeErr(w, 502, "upstream_error", err.Error())
			return
		}
		writeOK(w, res)
	})
	mux.HandleFunc("GET /v1/admin/customers", func(w http.ResponseWriter, r *http.Request) {
		res, err := d.ListCustomers(r.Context(), tenant(r), r.URL.Query())
		if err != nil {
			writeErr(w, 502, "upstream_error", err.Error())
			return
		}
		writeOK(w, res)
	})
	mux.HandleFunc("GET /v1/admin/customers/{id}", func(w http.ResponseWriter, r *http.Request) {
		res, err := d.GetCustomer(r.Context(), tenant(r), r.PathValue("id"))
		if err != nil {
			writeErr(w, 404, "not_found", err.Error())
			return
		}
		writeOK(w, res)
	})
	mux.HandleFunc("POST /v1/admin/customers/{id}/adjustments", func(w http.ResponseWriter, r *http.Request) {
		var b struct {
			Type string `json:"type"`
			Note string `json:"note"`
		}
		_ = json.NewDecoder(r.Body).Decode(&b)
		res, err := d.CustomerAdjustment(r.Context(), r.PathValue("id"), b.Type, b.Note)
		if err != nil {
			writeErr(w, 400, "invalid_argument", err.Error())
			return
		}
		writeOK(w, res)
	})
	mux.HandleFunc("POST /v1/admin/flags/{key}", func(w http.ResponseWriter, r *http.Request) {
		var b struct {
			Enabled bool `json:"enabled"`
		}
		_ = json.NewDecoder(r.Body).Decode(&b)
		res, err := d.KillSwitch(r.Context(), tenant(r), r.PathValue("key"), b.Enabled)
		if err != nil {
			writeErr(w, 400, "invalid_argument", err.Error())
			return
		}
		writeOK(w, res)
	})
	h := authz.Gate(v, authz.Options{
		Public: []string{"/health", "/ready"},
		Rules: []authz.Rule{
			{Prefix: "/v1/admin/flags", Roles: []string{"admin", "super_admin"}},
			{Prefix: "/v1/admin/finance", Roles: []string{"admin", "super_admin", "finance_analyst"}},
			{Prefix: "/v1/admin/catalog", Roles: []string{"admin", "super_admin", "city_ops"}},
			{Prefix: "/v1/admin", Roles: []string{"admin", "super_admin", "support_agent", "city_ops"}},
		},
	})(mux)
	return &http.Server{Addr: addr, Handler: h, ReadHeaderTimeout: 5 * time.Second}
}
