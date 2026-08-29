package httpadapter

import (
	"encoding/json"
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

func NewServer(addr string, d *app.Deps) *http.Server {
	return NewServerWithAuth(addr, d, authz.FromEnv())
}

func NewServerWithAuth(addr string, d *app.Deps, v authz.Validator) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /ready", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ready"})
	})
	mux.HandleFunc("GET /v1/admin/dashboard", func(w http.ResponseWriter, r *http.Request) {
		res, err := d.Dashboard(r.Context(), r.Header.Get("X-Tenant-Id"))
		if err != nil {
			writeErr(w, 400, "invalid_argument", err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(res)
	})
	mux.HandleFunc("GET /v1/admin/orders", func(w http.ResponseWriter, r *http.Request) {
		res, err := d.ListOrders(r.Context(), r.Header.Get("X-Tenant-Id"), r.URL.Query())
		if err != nil {
			writeErr(w, 502, "upstream_error", err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(res)
	})
	mux.HandleFunc("GET /v1/admin/orders/{id}", func(w http.ResponseWriter, r *http.Request) {
		res, err := d.GetOrder(r.Context(), r.Header.Get("X-Tenant-Id"), r.PathValue("id"))
		if err != nil {
			writeErr(w, 502, "upstream_error", err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(res)
	})
	mux.HandleFunc("POST /v1/admin/flags/{key}", func(w http.ResponseWriter, r *http.Request) {
		var b struct {
			Enabled bool `json:"enabled"`
		}
		_ = json.NewDecoder(r.Body).Decode(&b)
		res, err := d.KillSwitch(r.Context(), r.Header.Get("X-Tenant-Id"), r.PathValue("key"), b.Enabled)
		if err != nil {
			writeErr(w, 400, "invalid_argument", err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(res)
	})
	h := authz.Gate(v, authz.Options{
		Public: []string{"/health", "/ready"},
		Rules: []authz.Rule{
			{Prefix: "/v1/admin/flags", Roles: []string{"admin", "super_admin"}},
			{Prefix: "/v1/admin", Roles: []string{"admin", "super_admin", "support_agent", "city_ops"}},
		},
	})(mux)
	return &http.Server{Addr: addr, Handler: h, ReadHeaderTimeout: 5 * time.Second}
}
