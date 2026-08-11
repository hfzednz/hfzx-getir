package httpadapter

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/autonomy-service/internal/app"
	"github.com/nexora/autonomy-service/internal/domain"
	"github.com/nexora/autonomy-service/internal/ratelimit"
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
	const base = "/v1/autonomy"

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		writeOK(w, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /ready", func(w http.ResponseWriter, _ *http.Request) {
		writeOK(w, map[string]string{"status": "ready"})
	})

	mux.HandleFunc("POST "+base+"/bootstrap", tenant(h.bootstrap))
	mux.HandleFunc("POST "+base+"/audits", tenant(h.runAudit))
	mux.HandleFunc("GET "+base+"/audits", tenant(h.listAudits))
	mux.HandleFunc("POST "+base+"/heal", tenant(h.heal))
	mux.HandleFunc("GET "+base+"/heal", tenant(h.listHeals))
	mux.HandleFunc("POST "+base+"/reviews", tenant(h.createReview))
	mux.HandleFunc("GET "+base+"/reviews", tenant(h.listReviews))
	mux.HandleFunc("POST "+base+"/evolution", tenant(h.createEvolution))
	mux.HandleFunc("GET "+base+"/evolution", tenant(h.listEvolution))
	mux.HandleFunc("POST "+base+"/releases/score", tenant(h.scoreRelease))
	mux.HandleFunc("GET "+base+"/releases", tenant(h.listReleases))
	mux.HandleFunc("POST "+base+"/governance", tenant(h.upsertGovernance))
	mux.HandleFunc("GET "+base+"/governance", tenant(h.listGovernance))
	mux.HandleFunc("GET "+base+"/dependencies", tenant(h.listDeps))
	mux.HandleFunc("GET "+base+"/assistants", tenant(h.listAssistants))
	mux.HandleFunc("GET "+base+"/org/teams", tenant(h.listTeams))
	mux.HandleFunc("GET "+base+"/gates", tenant(h.gates))
	mux.HandleFunc("POST "+base+"/genesis", tenant(h.issueGenesis))
	mux.HandleFunc("GET "+base+"/genesis/latest", tenant(h.latestGenesis))
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
	if err := h.Deps.BootstrapAutonomy(r.Context(), tid); err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]string{"status": "bootstrapped"})
}

func (h *Handler) runAudit(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body struct {
		Scope domain.AuditScope `json:"scope"`
	}
	if err := decodeJSON(r, &body); err != nil || body.Scope == "" {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	a, err := h.Deps.RunAutonomyAudit(r.Context(), tid, body.Scope)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, a)
}

func (h *Handler) listAudits(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	list, err := h.Deps.Audits.List(r.Context(), tid)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": list})
}

func (h *Handler) heal(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.HealAction
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	a, err := h.Deps.ExecuteHeal(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, a)
}

func (h *Handler) listHeals(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	list, err := h.Deps.Heals.List(r.Context(), tid)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": list})
}

func (h *Handler) createReview(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.AICTOReview
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	rev, err := h.Deps.CreateCTOReview(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, rev)
}

func (h *Handler) listReviews(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	list, err := h.Deps.Reviews.List(r.Context(), tid)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": list})
}

func (h *Handler) createEvolution(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.EvolutionTask
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	t, err := h.Deps.CreateEvolution(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, t)
}

func (h *Handler) listEvolution(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	list, err := h.Deps.Evolution.List(r.Context(), tid)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": list})
}

func (h *Handler) scoreRelease(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.ReleasePlan
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	p, err := h.Deps.ScoreRelease(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, p)
}

func (h *Handler) listReleases(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	list, err := h.Deps.Releases.List(r.Context(), tid)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": list})
}

func (h *Handler) upsertGovernance(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.GovernanceLoop
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	g, err := h.Deps.UpsertGovernance(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, g)
}

func (h *Handler) listGovernance(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	list, err := h.Deps.Governance.List(r.Context(), tid)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": list})
}

func (h *Handler) listDeps(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	list, err := h.Deps.Dependencies.List(r.Context(), tid)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": list})
}

func (h *Handler) listAssistants(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	list, err := h.Deps.Assistants.List(r.Context(), tid)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": list})
}

func (h *Handler) listTeams(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	list, err := h.Deps.Teams.List(r.Context(), tid)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": list})
}

func (h *Handler) gates(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	g, err := h.Deps.EvaluateGenesisGates(r.Context(), tid)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"gates": g})
}

func (h *Handler) issueGenesis(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body struct {
		Version string `json:"version"`
	}
	_ = decodeJSON(r, &body)
	c, err := h.Deps.IssueGenesis(r.Context(), tid, body.Version)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, c)
}

func (h *Handler) latestGenesis(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	c, err := h.Deps.Genesis.Latest(r.Context(), tid)
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
