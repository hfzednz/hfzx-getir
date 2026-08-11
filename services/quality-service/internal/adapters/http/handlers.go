package httpadapter

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/quality-service/internal/app"
	"github.com/nexora/quality-service/internal/domain"
	"github.com/nexora/quality-service/internal/ratelimit"
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
	const base = "/v1/quality"

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		writeOK(w, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /ready", func(w http.ResponseWriter, _ *http.Request) {
		writeOK(w, map[string]string{"status": "ready"})
	})

	mux.HandleFunc("POST "+base+"/bootstrap", tenant(h.bootstrap))
	mux.HandleFunc("POST "+base+"/suites", tenant(h.upsertSuite))
	mux.HandleFunc("GET "+base+"/suites", tenant(h.listSuites))
	mux.HandleFunc("POST "+base+"/runs", tenant(h.startRun))
	mux.HandleFunc("POST "+base+"/runs/{id}/complete", tenant(h.completeRun))
	mux.HandleFunc("GET "+base+"/runs", tenant(h.listRuns))
	mux.HandleFunc("POST "+base+"/coverage", tenant(h.coverage))
	mux.HandleFunc("POST "+base+"/perf", tenant(h.perf))
	mux.HandleFunc("POST "+base+"/security/findings", tenant(h.secFinding))
	mux.HandleFunc("POST "+base+"/gates/evaluate", tenant(h.evalGates))
	mux.HandleFunc("POST "+base+"/certifications", tenant(h.issueCert))
	mux.HandleFunc("GET "+base+"/certifications", tenant(h.listCerts))
	mux.HandleFunc("GET "+base+"/flaky", tenant(h.listFlaky))
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
	if err := h.Deps.SeedDefaultSuites(r.Context(), tid); err != nil {
		writeErr(w, r, err)
		return
	}
	if err := h.Deps.SeedDefaultPolicies(r.Context(), tid); err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]string{"status": "bootstrapped"})
}

func (h *Handler) upsertSuite(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.Suite
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	s, err := h.Deps.UpsertSuite(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, s)
}

func (h *Handler) listSuites(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	list, err := h.Deps.Suites.List(r.Context(), tid)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": list})
}

func (h *Handler) startRun(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.TestRun
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	run, err := h.Deps.StartRun(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, run)
}

func (h *Handler) completeRun(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		Summary domain.RunSummary       `json:"summary"`
		Cases   []domain.TestCaseResult `json:"cases"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	run, err := h.Deps.CompleteRun(r.Context(), tid, id, body.Summary, body.Cases)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, run)
}

func (h *Handler) listRuns(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	list, err := h.Deps.ListRuns(r.Context(), tid)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": list})
}

func (h *Handler) coverage(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.CoverageReport
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	c, err := h.Deps.IngestCoverage(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, c)
}

func (h *Handler) perf(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.PerfMetric
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	p, err := h.Deps.IngestPerf(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, p)
}

func (h *Handler) secFinding(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.SecurityFinding
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	f, err := h.Deps.IngestSecurityFinding(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, f)
}

func (h *Handler) evalGates(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body struct {
		RunID uuid.UUID `json:"runId"`
	}
	if err := decodeJSON(r, &body); err != nil || body.RunID == uuid.Nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	evals, err := h.Deps.EvaluateGates(r.Context(), tid, body.RunID)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"evaluations": evals})
}

func (h *Handler) issueCert(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body struct {
		Kind      domain.CertKind `json:"kind"`
		Version   string          `json:"version"`
		CommitSHA string          `json:"commitSha"`
		Notes     string          `json:"notes"`
		RunIDs    []uuid.UUID     `json:"runIds"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	if body.Kind == "" {
		body.Kind = domain.CertReleaseReadiness
	}
	c, err := h.Deps.IssueCertification(r.Context(), tid, body.Kind, body.Version, body.CommitSHA, body.Notes, body.RunIDs)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, c)
}

func (h *Handler) listCerts(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	list, err := h.Deps.ListCerts(r.Context(), tid)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": list})
}

func (h *Handler) listFlaky(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	list, err := h.Deps.ListFlaky(r.Context(), tid)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": list})
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
