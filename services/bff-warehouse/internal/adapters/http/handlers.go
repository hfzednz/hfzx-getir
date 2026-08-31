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
	if errors.Is(err, app.ErrNotSupported) {
		writeJSON(w, http.StatusNotImplemented, map[string]any{
			"error": map[string]any{"code": "not_supported", "message": err.Error()},
		})
		return
	}
	writeJSON(w, http.StatusBadGateway, map[string]any{
		"error": map[string]any{"code": "upstream_error", "message": err.Error()},
	})
}

func tenant(r *http.Request) string {
	t := r.Header.Get("X-Tenant-Id")
	if t == "" {
		t = r.Header.Get("X-Nexora-Tenant")
	}
	return t
}

func writeTaskList(w http.ResponseWriter, items []map[string]any) {
	writeJSON(w, http.StatusOK, map[string]any{"items": items, "total": len(items)})
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
	list := func(w http.ResponseWriter, r *http.Request) {
		res, err := d.ListTasks(r.Context(), tenant(r))
		if err != nil {
			writeErr(w, err)
			return
		}
		writeTaskList(w, res)
	}
	mux.HandleFunc("GET /v1/warehouse/tasks", list)
	mux.HandleFunc("GET /v1/warehouse/picking/queue", list)
	mux.HandleFunc("GET /v1/warehouse/picking/{id}", func(w http.ResponseWriter, r *http.Request) {
		res, err := d.GetTask(r.Context(), tenant(r), r.PathValue("id"))
		if err != nil {
			writeErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, res)
	})
	mux.HandleFunc("POST /v1/warehouse/tasks/{id}/pick", func(w http.ResponseWriter, r *http.Request) {
		res, err := d.Pick(r.Context(), tenant(r), r.PathValue("id"))
		if err != nil {
			writeErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, res)
	})
	mux.HandleFunc("POST /v1/warehouse/picking/{id}/claim", func(w http.ResponseWriter, r *http.Request) {
		res, err := d.Pick(r.Context(), tenant(r), r.PathValue("id"))
		if err != nil {
			writeErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, res)
	})
	mux.HandleFunc("POST /v1/warehouse/picking/{id}/start", func(w http.ResponseWriter, r *http.Request) {
		res, err := d.Pick(r.Context(), tenant(r), r.PathValue("id"))
		if err != nil {
			writeErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, res)
	})
	mux.HandleFunc("POST /v1/warehouse/picking/{id}/complete", func(w http.ResponseWriter, r *http.Request) {
		res, err := d.Pack(r.Context(), tenant(r), r.PathValue("id"))
		if err != nil {
			writeErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, res)
	})
	mux.HandleFunc("POST /v1/warehouse/picking/{id}/stage", func(w http.ResponseWriter, r *http.Request) {
		res, err := d.DispatchReady(r.Context(), tenant(r), r.PathValue("id"))
		if err != nil {
			writeErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, res)
	})
	mux.HandleFunc("POST /v1/warehouse/picking/{id}/lines/{lineId}/scan", func(w http.ResponseWriter, _ *http.Request) {
		writeErr(w, app.ErrNotSupported)
	})
	mux.HandleFunc("POST /v1/warehouse/picking/{id}/lines/{lineId}/short-pick", func(w http.ResponseWriter, _ *http.Request) {
		writeErr(w, app.ErrNotSupported)
	})
	mux.HandleFunc("POST /v1/warehouse/tasks/{id}/pack", func(w http.ResponseWriter, r *http.Request) {
		res, err := d.Pack(r.Context(), tenant(r), r.PathValue("id"))
		if err != nil {
			writeErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, res)
	})
	mux.HandleFunc("POST /v1/warehouse/tasks/{id}/ready", func(w http.ResponseWriter, r *http.Request) {
		res, err := d.DispatchReady(r.Context(), tenant(r), r.PathValue("id"))
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
