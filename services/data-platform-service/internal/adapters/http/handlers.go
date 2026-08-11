package httpadapter

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/data-platform-service/internal/app"
	"github.com/nexora/data-platform-service/internal/domain"
	"github.com/nexora/data-platform-service/internal/ratelimit"
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
	const base = "/v1/data"

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		writeOK(w, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /ready", func(w http.ResponseWriter, r *http.Request) {
		writeOK(w, map[string]string{"status": "ready"})
	})

	mux.HandleFunc("POST "+base+"/schemas", tenant(h.registerSchema))
	mux.HandleFunc("GET "+base+"/schemas", tenant(h.listSchemas))
	mux.HandleFunc("POST "+base+"/events", tenant(h.ingestEvent))
	mux.HandleFunc("GET "+base+"/events", tenant(h.listEvents))

	mux.HandleFunc("POST "+base+"/streams/jobs", tenant(h.upsertJob))
	mux.HandleFunc("GET "+base+"/streams/jobs", tenant(h.listJobs))
	mux.HandleFunc("GET "+base+"/lake/datasets", tenant(h.listLake))
	mux.HandleFunc("POST "+base+"/warehouse/refresh", tenant(h.refreshMarts))
	mux.HandleFunc("GET "+base+"/kpis", tenant(h.listKPIs))
	mux.HandleFunc("GET "+base+"/realtime", tenant(h.realtime))

	mux.HandleFunc("POST "+base+"/experiments", tenant(h.upsertExperiment))
	mux.HandleFunc("POST "+base+"/experiments/{key}/assign", tenant(h.assignExperiment))
	mux.HandleFunc("POST "+base+"/experiments/{key}/decide", tenant(h.decideExperiment))

	mux.HandleFunc("POST "+base+"/reports", tenant(h.upsertReport))
	mux.HandleFunc("POST "+base+"/reports/{id}/run", tenant(h.runReport))

	mux.HandleFunc("POST "+base+"/obs/signals", tenant(h.ingestObs))
	mux.HandleFunc("POST "+base+"/alerts/rules", tenant(h.upsertAlert))
	mux.HandleFunc("GET "+base+"/alerts/events", tenant(h.listAlerts))

	mux.HandleFunc("POST "+base+"/catalog/assets", tenant(h.upsertCatalog))
	mux.HandleFunc("GET "+base+"/catalog/assets", tenant(h.listCatalog))
	mux.HandleFunc("POST "+base+"/quality/run", tenant(h.runQuality))
	mux.HandleFunc("POST "+base+"/privacy/erase-user", tenant(h.gdprErase))

	mux.HandleFunc("GET "+base+"/admin/stats", tenant(h.stats))
	mux.HandleFunc("POST "+base+"/outbox/publish", tenant(h.outbox))

	return chain(mux,
		requestIDMiddleware,
		recoverMiddleware(cfg.Log),
		loggingMiddleware(cfg.Log),
		corsMiddleware(cfg.CORSOrigins),
		rateLimitMiddleware(cfg.Limiter, cfg.RateLimitPerMinute),
	)
}

func NewServer(cfg ServerConfig) *http.Server {
	return &http.Server{
		Addr: cfg.Addr, Handler: NewHandler(cfg),
		ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 30 * time.Second,
		WriteTimeout: 60 * time.Second, IdleTimeout: 60 * time.Second,
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

func (h *Handler) registerSchema(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.EventSchema
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	s, err := h.Deps.RegisterSchema(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, s)
}

func (h *Handler) listSchemas(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	items, err := h.Deps.Schemas.List(r.Context(), tid, r.URL.Query().Get("family"))
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": items})
}

func (h *Handler) ingestEvent(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.AnalyticsEvent
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	if key := r.Header.Get("Idempotency-Key"); key != "" && body.IdempotencyKey == "" {
		body.IdempotencyKey = key
	}
	e, err := h.Deps.IngestEvent(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, e)
}

func (h *Handler) listEvents(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	items, err := h.Deps.Events.List(r.Context(), tid, r.URL.Query().Get("name"), 100)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": items})
}

func (h *Handler) upsertJob(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.StreamJob
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	j, err := h.Deps.UpsertStreamJob(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, j)
}

func (h *Handler) listJobs(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	items, err := h.Deps.Streams.ListJobs(r.Context(), tid)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": items})
}

func (h *Handler) listLake(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	items, err := h.Deps.Lake.ListDatasets(r.Context(), tid, r.URL.Query().Get("layer"))
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": items})
}

func (h *Handler) refreshMarts(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	kpis, err := h.Deps.RefreshMarts(r.Context(), tid)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"kpis": kpis})
}

func (h *Handler) listKPIs(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	items, err := h.Deps.Warehouse.ListKPIs(r.Context(), tid)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": items})
}

func (h *Handler) realtime(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	items, err := h.Deps.Realtime.List(r.Context(), tid)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": items})
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

func (h *Handler) assignExperiment(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	key := r.PathValue("key")
	var body struct {
		SubjectID uuid.UUID `json:"subjectId"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	a, err := h.Deps.AssignExperiment(r.Context(), tid, key, body.SubjectID)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, a)
}

func (h *Handler) decideExperiment(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	key := r.PathValue("key")
	var body struct {
		Scores map[string]float64 `json:"scores"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	e, err := h.Deps.DecideExperiment(r.Context(), tid, key, body.Scores)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, e)
}

func (h *Handler) upsertReport(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.ReportDef
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	def, err := h.Deps.UpsertReportDef(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, def)
}

func (h *Handler) runReport(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	run, err := h.Deps.RunReport(r.Context(), tid, id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, run)
}

func (h *Handler) ingestObs(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.ObservabilitySignal
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	s, err := h.Deps.IngestObs(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, s)
}

func (h *Handler) upsertAlert(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.AlertRule
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	rule, err := h.Deps.UpsertAlertRule(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, rule)
}

func (h *Handler) listAlerts(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	items, err := h.Deps.Alerts.ListEvents(r.Context(), tid, 100)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": items})
}

func (h *Handler) upsertCatalog(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.CatalogAsset
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	a, err := h.Deps.UpsertCatalogAsset(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, a)
}

func (h *Handler) listCatalog(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	items, err := h.Deps.Catalog.ListAssets(r.Context(), tid, r.URL.Query().Get("type"))
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": items})
}

func (h *Handler) runQuality(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body struct {
		AssetName string `json:"assetName"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	checks, err := h.Deps.RunQualityChecks(r.Context(), tid, body.AssetName)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"checks": checks})
}

func (h *Handler) gdprErase(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body struct {
		UserID uuid.UUID `json:"userId"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	n, err := h.Deps.GDPRDeleteUser(r.Context(), tid, body.UserID)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"redactedEvents": n})
}

func (h *Handler) stats(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	stats, err := h.Deps.AdminStats(r.Context(), tid)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, stats)
}

func (h *Handler) outbox(w http.ResponseWriter, r *http.Request) {
	n, err := h.Deps.PublishPending(r.Context(), 100)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"published": n})
}
