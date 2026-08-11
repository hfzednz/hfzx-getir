package httpadapter

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/liveops-service/internal/app"
	"github.com/nexora/liveops-service/internal/domain"
	"github.com/nexora/liveops-service/internal/ratelimit"
)

type Handler struct {
	Deps  *app.Deps
	Ready func(*http.Request) error
	Live  func(*http.Request) error
}

type ServerConfig struct {
	Addr               string
	Deps               *app.Deps
	Limiter            ratelimit.Limiter
	RateLimitPerMinute int
	CORSOrigins        []string
	Log                *slog.Logger
	Ready              func(*http.Request) error
	Live               func(*http.Request) error
}

func NewHandler(cfg ServerConfig) http.Handler {
	h := &Handler{Deps: cfg.Deps, Ready: cfg.Ready, Live: cfg.Live}
	mux := http.NewServeMux()
	const base = "/v1/liveops"

	mux.HandleFunc("GET /health", h.health)
	mux.HandleFunc("GET /ready", h.ready)

	mux.HandleFunc("POST "+base+"/flags", tenant(h.upsertFlag))
	mux.HandleFunc("GET "+base+"/flags", tenant(h.listFlags))
	mux.HandleFunc("POST "+base+"/flags/evaluate", tenant(h.evalFlags))

	mux.HandleFunc("POST "+base+"/configs", tenant(h.publishConfig))
	mux.HandleFunc("GET "+base+"/configs/resolve", tenant(h.resolveConfig))

	mux.HandleFunc("POST "+base+"/experiments", tenant(h.upsertExperiment))
	mux.HandleFunc("POST "+base+"/experiments/{key}/start", tenant(h.startExperiment))
	mux.HandleFunc("POST "+base+"/experiments/{key}/assign", tenant(h.assignExperiment))
	mux.HandleFunc("POST "+base+"/experiments/{key}/complete", tenant(h.completeExperiment))
	mux.HandleFunc("GET "+base+"/experiments", tenant(h.listExperiments))

	mux.HandleFunc("POST "+base+"/events", tenant(h.upsertEvent))
	mux.HandleFunc("POST "+base+"/events/{key}/activate", tenant(h.activateEvent))
	mux.HandleFunc("GET "+base+"/events", tenant(h.listEvents))

	mux.HandleFunc("POST "+base+"/changes", tenant(h.requestChange))
	mux.HandleFunc("POST "+base+"/changes/{id}/decide", tenant(h.decideChange))
	mux.HandleFunc("POST "+base+"/rollbacks", tenant(h.rollback))
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

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	if h.Live != nil {
		if err := h.Live(r); err != nil {
			writeErr(w, r, err)
			return
		}
	}
	writeOK(w, map[string]string{"status": "ok"})
}

func (h *Handler) ready(w http.ResponseWriter, r *http.Request) {
	if h.Ready != nil {
		if err := h.Ready(r); err != nil {
			writeErr(w, r, err)
			return
		}
	}
	writeOK(w, map[string]string{"status": "ready"})
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

func (h *Handler) upsertFlag(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.FeatureFlag
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	f, err := h.Deps.UpsertFlag(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, f)
}

func (h *Handler) listFlags(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	items, err := h.Deps.Flags.List(r.Context(), tid)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": items})
}

func (h *Handler) evalFlags(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body struct {
		Keys    []string           `json:"keys"`
		Context domain.EvalContext `json:"context"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	ev, err := h.Deps.EvaluateFlags(r.Context(), tid, body.Keys, body.Context)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"evaluations": ev})
}

func (h *Handler) publishConfig(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.ConfigDocument
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	c, err := h.Deps.PublishConfig(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, c)
}

func (h *Handler) resolveConfig(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	eval := domain.EvalContext{
		SubjectID: r.URL.Query().Get("subjectId"),
		Country:   r.URL.Query().Get("country"),
		City:      r.URL.Query().Get("city"),
		OS:        r.URL.Query().Get("os"),
		AppVersion: r.URL.Query().Get("appVersion"),
	}
	cfg, err := h.Deps.ResolveConfig(r.Context(), tid, r.URL.Query().Get("namespace"), eval)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"config": cfg})
}

func (h *Handler) upsertExperiment(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.Experiment
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	e, err := h.Deps.UpsertExperiment(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, e)
}

func (h *Handler) startExperiment(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	e, err := h.Deps.StartExperiment(r.Context(), tid, r.PathValue("key"))
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, e)
}

func (h *Handler) assignExperiment(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body struct {
		SubjectID string `json:"subjectId"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	a, err := h.Deps.AssignExperiment(r.Context(), tid, r.PathValue("key"), body.SubjectID)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, a)
}

func (h *Handler) completeExperiment(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body struct {
		Rates       map[string]float64 `json:"rates"`
		AutoRollout bool               `json:"autoRollout"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	e, err := h.Deps.CompleteExperiment(r.Context(), tid, r.PathValue("key"), body.Rates, body.AutoRollout)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, e)
}

func (h *Handler) listExperiments(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	items, err := h.Deps.Experiments.List(r.Context(), tid)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": items})
}

func (h *Handler) upsertEvent(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.LiveOpsEvent
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	e, err := h.Deps.UpsertLiveEvent(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, e)
}

func (h *Handler) activateEvent(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	e, err := h.Deps.ActivateEvent(r.Context(), tid, r.PathValue("key"))
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, e)
}

func (h *Handler) listEvents(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	items, err := h.Deps.Events.List(r.Context(), tid, r.URL.Query().Get("status"))
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": items})
}

func (h *Handler) requestChange(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.ChangeRequest
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	c, err := h.Deps.RequestChange(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, c)
}

func (h *Handler) decideChange(w http.ResponseWriter, r *http.Request) {
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
		Approve bool `json:"approve"`
	}
	_ = decodeJSON(r, &body)
	c, err := h.Deps.DecideChange(r.Context(), tid, id, body.Approve)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, c)
}

func (h *Handler) rollback(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body struct {
		Kind, SubjectKey, Reason string
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	rec, err := h.Deps.Rollback(r.Context(), tid, body.Kind, body.SubjectKey, body.Reason)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, rec)
}

func (h *Handler) stats(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	s, err := h.Deps.AdminStats(r.Context(), tid)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, s)
}

func (h *Handler) outbox(w http.ResponseWriter, r *http.Request) {
	n, err := h.Deps.PublishPending(r.Context(), 100)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"published": n})
}
