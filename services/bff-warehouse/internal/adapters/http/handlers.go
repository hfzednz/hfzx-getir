package httpadapter

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/nexora/bff-warehouse/internal/app"
	"github.com/nexora/bff-warehouse/internal/authz"
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, err error) {
	if errors.Is(err, app.ErrInvalid) {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error": map[string]any{"code": "invalid_argument", "message": err.Error()},
		})
		return
	}
	writeJSON(w, http.StatusBadGateway, map[string]any{
		"error": map[string]any{"code": "upstream_error", "message": err.Error()},
	})
}

func NewServer(addr string, d *app.Deps) *http.Server {
	return NewServerWithAuth(addr, d, authz.FromEnv())
}

func NewServerWithAuth(addr string, d *app.Deps, v authz.Validator) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /ready", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
	})
	mux.HandleFunc("POST /v1/warehouse/tasks/{id}/pick", func(w http.ResponseWriter, r *http.Request) {
		res, err := d.Pick(r.Context(), r.Header.Get("X-Tenant-Id"), r.PathValue("id"))
		if err != nil {
			writeErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, res)
	})
	mux.HandleFunc("POST /v1/warehouse/tasks/{id}/pack", func(w http.ResponseWriter, r *http.Request) {
		res, err := d.Pack(r.Context(), r.Header.Get("X-Tenant-Id"), r.PathValue("id"))
		if err != nil {
			writeErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, res)
	})
	mux.HandleFunc("POST /v1/warehouse/tasks/{id}/ready", func(w http.ResponseWriter, r *http.Request) {
		res, err := d.DispatchReady(r.Context(), r.Header.Get("X-Tenant-Id"), r.PathValue("id"))
		if err != nil {
			writeErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, res)
	})
	h := authz.Gate(v, authz.Options{
		Public: []string{"/health", "/ready"},
		Rules: []authz.Rule{
			{Prefix: "/v1/warehouse", Roles: []string{"picker", "packer", "dispatcher"}},
		},
	})(mux)
	return &http.Server{Addr: addr, Handler: h, ReadHeaderTimeout: 5 * time.Second}
}
