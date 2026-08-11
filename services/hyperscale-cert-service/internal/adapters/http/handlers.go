package httpadapter

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/hyperscale-cert-service/internal/app"
	"github.com/nexora/hyperscale-cert-service/internal/domain"
	"github.com/nexora/hyperscale-cert-service/internal/ratelimit"
)

type Handler struct{ Deps *app.Deps }

type ServerConfig struct {
	Addr               string
	Deps               *app.Deps
	Limiter            ratelimit.Limiter
	RateLimitPerMinute int
	CORSOrigins        []string
	Log                *slog.Logger
}

func NewHandler(cfg ServerConfig) http.Handler {
	h := &Handler{Deps: cfg.Deps}
	mux := http.NewServeMux()
	const base = "/v1/hyperscale"

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		writeOK(w, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /ready", func(w http.ResponseWriter, _ *http.Request) {
		writeOK(w, map[string]string{"status": "ready"})
	})

	mux.HandleFunc("POST "+base+"/bootstrap", tenant(h.bootstrap))
	mux.HandleFunc("POST "+base+"/audits", tenant(h.startAudit))
	mux.HandleFunc("POST "+base+"/audits/{id}/complete", tenant(h.completeAudit))
	mux.HandleFunc("POST "+base+"/findings", tenant(h.addFinding))
	mux.HandleFunc("POST "+base+"/findings/resolve", tenant(h.resolveFinding))
	mux.HandleFunc("POST "+base+"/benchmarks", tenant(h.recordBench))
	mux.HandleFunc("GET "+base+"/benchmarks", tenant(h.listBench))
	mux.HandleFunc("GET "+base+"/capacity", tenant(h.listCapacity))
	mux.HandleFunc("POST "+base+"/chaos", tenant(h.runChaos))
	mux.HandleFunc("GET "+base+"/chaos", tenant(h.listChaos))
	mux.HandleFunc("POST "+base+"/tuning", tenant(h.upsertTuning))
	mux.HandleFunc("POST "+base+"/tuning/apply", tenant(h.applyTuning))
	mux.HandleFunc("GET "+base+"/gates", tenant(h.gates))
	mux.HandleFunc("POST "+base+"/certificates", tenant(h.issueCert))
	mux.HandleFunc("GET "+base+"/certificates/latest", tenant(h.latestCert))
	mux.HandleFunc("GET "+base+"/admin/stats", tenant(h.stats))
	mux.HandleFunc("POST "+base+"/outbox/publish", tenant(h.outbox))

	return chain(mux, requestIDMiddleware, recoverMiddleware(cfg.Log), loggingMiddleware(cfg.Log),
		corsMiddleware(cfg.CORSOrigins), rateLimitMiddleware(cfg.Limiter, cfg.RateLimitPerMinute))
}

func NewServer(cfg ServerConfig) *http.Server {
	return &http.Server{
		Addr: cfg.Addr, Handler: NewHandler(cfg),
		ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second,
		WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second,
	}
}

func tenant(next func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tenantMiddleware(http.HandlerFunc(next)).ServeHTTP(w, r)
	}
}

func requireTenant(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	tid, ok := TenantIDFromContext(r.Context())
	if !ok {
		writeErr(w, r, domain.ErrInvalidArgument)
		return uuid.Nil, false
	}
	return tid, true
}

func (h *Handler) bootstrap(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	if err := h.Deps.BootstrapHyperscale(r.Context(), tid); err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]string{"status": "bootstrapped"})
}

func (h *Handler) startAudit(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.Audit
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	a, err := h.Deps.StartAudit(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, a)
}

func (h *Handler) completeAudit(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	a, err := h.Deps.CompleteAudit(r.Context(), tid, id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, a)
}

func (h *Handler) addFinding(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.Finding
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	f, err := h.Deps.AddFinding(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, f)
}

func (h *Handler) resolveFinding(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body struct {
		Code       string `json:"code"`
		Resolution string `json:"resolution"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	f, err := h.Deps.ResolveFinding(r.Context(), tid, body.Code, body.Resolution)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, f)
}

func (h *Handler) recordBench(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.BenchmarkRun
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	b, err := h.Deps.RecordBenchmark(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, b)
}

func (h *Handler) listBench(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	list, err := h.Deps.Benchmarks.List(r.Context(), tid)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": list})
}

func (h *Handler) listCapacity(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	list, err := h.Deps.Capacity.List(r.Context(), tid)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": list})
}

func (h *Handler) runChaos(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.ChaosExperiment
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	c, err := h.Deps.RunChaos(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, c)
}

func (h *Handler) listChaos(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	list, err := h.Deps.Chaos.List(r.Context(), tid)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": list})
}

func (h *Handler) upsertTuning(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.TuningProfile
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	t, err := h.Deps.UpsertTuning(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, t)
}

func (h *Handler) applyTuning(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body struct {
		Key string `json:"key"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	t, err := h.Deps.ApplyTuning(r.Context(), tid, body.Key)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, t)
}

func (h *Handler) gates(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	g, err := h.Deps.EvaluateGates(r.Context(), tid)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"gates": g})
}

func (h *Handler) issueCert(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body struct {
		Version string `json:"version"`
	}
	_ = decodeJSON(r, &body)
	c, err := h.Deps.IssueCertificate(r.Context(), tid, body.Version)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, c)
}

func (h *Handler) latestCert(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	c, err := h.Deps.Certificates.Latest(r.Context(), tid)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, c)
}

func (h *Handler) stats(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	st, err := h.Deps.AdminStats(r.Context(), tid)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, st)
}

func (h *Handler) outbox(w http.ResponseWriter, r *http.Request) {
	n, err := h.Deps.PublishPending(r.Context(), 100)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"published": n})
}
