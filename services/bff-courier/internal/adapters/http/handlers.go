package httpadapter

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/nexora/bff-courier/internal/app"
	"github.com/nexora/bff-courier/internal/authz"
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

func courierID(r *http.Request, bodyID string) string {
	if bodyID != "" {
		return bodyID
	}
	if q := r.URL.Query().Get("courierId"); q != "" {
		return q
	}
	return r.Header.Get("X-Nexora-User")
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
	mux.HandleFunc("POST /v1/courier/duty", func(w http.ResponseWriter, r *http.Request) {
		var b struct {
			CourierID string `json:"courierId"`
			On        bool   `json:"on"`
		}
		_ = json.NewDecoder(r.Body).Decode(&b)
		res, err := d.Duty(r.Context(), tenant(r), courierID(r, b.CourierID), b.On)
		if err != nil {
			writeErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, res)
	})
	mux.HandleFunc("GET /v1/courier/duty", func(w http.ResponseWriter, r *http.Request) {
		res, err := d.GetDuty(r.Context(), tenant(r), courierID(r, ""))
		if err != nil {
			writeErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, res)
	})
	list := func(w http.ResponseWriter, r *http.Request) {
		res, err := d.ListOffers(r.Context(), tenant(r), courierID(r, ""))
		if err != nil {
			writeErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": res, "deliveries": res, "total": len(res)})
	}
	mux.HandleFunc("GET /v1/courier/offers", list)
	mux.HandleFunc("GET /v1/courier/deliveries", list)
	mux.HandleFunc("GET /v1/courier/deliveries/{id}", func(w http.ResponseWriter, r *http.Request) {
		res, err := d.GetJob(r.Context(), tenant(r), r.PathValue("id"))
		if err != nil {
			writeErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, res)
	})
	mux.HandleFunc("POST /v1/courier/offers/{id}", func(w http.ResponseWriter, r *http.Request) {
		var b struct {
			CourierID string `json:"courierId"`
			Accept    bool   `json:"accept"`
		}
		_ = json.NewDecoder(r.Body).Decode(&b)
		res, err := d.Offer(r.Context(), tenant(r), courierID(r, b.CourierID), r.PathValue("id"), b.Accept)
		if err != nil {
			writeErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, res)
	})
	mux.HandleFunc("POST /v1/courier/offers/{id}/enroute", func(w http.ResponseWriter, r *http.Request) {
		res, err := d.Enroute(r.Context(), tenant(r), r.PathValue("id"))
		if err != nil {
			writeErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, res)
	})
	mux.HandleFunc("POST /v1/courier/offers/{id}/complete", func(w http.ResponseWriter, r *http.Request) {
		res, err := d.Complete(r.Context(), tenant(r), r.PathValue("id"))
		if err != nil {
			writeErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, res)
	})
	mux.HandleFunc("POST /v1/courier/deliveries/{id}/status", func(w http.ResponseWriter, r *http.Request) {
		var b struct {
			Status    string `json:"status"`
			CourierID string `json:"courierId"`
		}
		_ = json.NewDecoder(r.Body).Decode(&b)
		res, err := d.Transition(r.Context(), tenant(r), courierID(r, b.CourierID), r.PathValue("id"), b.Status)
		if err != nil {
			writeErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, res)
	})
	mux.HandleFunc("POST /v1/courier/deliveries/{id}/pickup", func(w http.ResponseWriter, r *http.Request) {
		res, err := d.Enroute(r.Context(), tenant(r), r.PathValue("id"))
		if err != nil {
			writeErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, res)
	})
	mux.HandleFunc("POST /v1/courier/deliveries/{id}/pod", func(w http.ResponseWriter, r *http.Request) {
		res, err := d.Complete(r.Context(), tenant(r), r.PathValue("id"))
		if err != nil {
			writeErr(w, err)
			return
		}
		res["podStored"] = false
		writeJSON(w, http.StatusOK, res)
	})
	mux.HandleFunc("POST /v1/courier/deliveries/{id}/fail", func(w http.ResponseWriter, r *http.Request) {
		var b struct {
			ReasonCode string `json:"reason_code"`
			Note       string `json:"note"`
		}
		_ = json.NewDecoder(r.Body).Decode(&b)
		reason := b.ReasonCode
		if b.Note != "" {
			reason = reason + ": " + b.Note
		}
		res, err := d.Fail(r.Context(), tenant(r), r.PathValue("id"), reason)
		if err != nil {
			writeErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, res)
	})
	h := authz.Gate(v, authz.Options{
		Public: []string{"/health", "/ready"},
		Rules: []authz.Rule{
			{Prefix: "/v1/courier", Roles: []string{"courier"}},
		},
	})(mux)
	return &http.Server{Addr: addr, Handler: h, ReadHeaderTimeout: 5 * time.Second}
}
