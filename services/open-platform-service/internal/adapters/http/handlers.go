package httpadapter

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/open-platform-service/internal/app"
	"github.com/nexora/open-platform-service/internal/domain"
	"github.com/nexora/open-platform-service/internal/ratelimit"
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
	const base = "/v1/open"

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		writeOK(w, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /ready", func(w http.ResponseWriter, _ *http.Request) {
		writeOK(w, map[string]string{"status": "ready"})
	})

	mux.HandleFunc("POST "+base+"/bootstrap/catalog", tenant(h.seedCatalog))
	mux.HandleFunc("POST "+base+"/apps", tenant(h.createApp))
	mux.HandleFunc("GET "+base+"/apps", tenant(h.listApps))
	mux.HandleFunc("POST "+base+"/apps/{id}/keys", tenant(h.createKey))
	mux.HandleFunc("GET "+base+"/catalog", tenant(h.listCatalog))
	mux.HandleFunc("POST "+base+"/gateway/policies", tenant(h.upsertPolicy))
	mux.HandleFunc("GET "+base+"/gateway/policies/export", tenant(h.exportPolicies))
	mux.HandleFunc("POST "+base+"/webhooks", tenant(h.registerWebhook))
	mux.HandleFunc("POST "+base+"/webhooks/enqueue", tenant(h.enqueueWebhook))
	mux.HandleFunc("POST "+base+"/webhooks/process", tenant(h.processWebhooks))
	mux.HandleFunc("POST "+base+"/webhooks/deliveries/{id}/replay", tenant(h.replay))
	mux.HandleFunc("POST "+base+"/sdks", tenant(h.publishSdk))
	mux.HandleFunc("GET "+base+"/sdks", tenant(h.listSdks))
	mux.HandleFunc("POST "+base+"/integrations", tenant(h.connectInt))
	mux.HandleFunc("GET "+base+"/integrations", tenant(h.listInts))
	mux.HandleFunc("POST "+base+"/usage", tenant(h.recordUsage))
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

func (h *Handler) seedCatalog(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	if err := h.Deps.SeedCatalog(r.Context(), tid); err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]string{"status": "catalog_seeded"})
}

func (h *Handler) createApp(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.DeveloperApp
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	a, err := h.Deps.CreateApp(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, a)
}

func (h *Handler) listApps(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	list, err := h.Deps.Apps.List(r.Context(), tid)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": list})
}

func (h *Handler) createKey(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	appID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		Name   string   `json:"name"`
		Scopes []string `json:"scopes"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	k, err := h.Deps.CreateApiKey(r.Context(), tid, appID, body.Name, body.Scopes)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, k)
}

func (h *Handler) listCatalog(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	list, err := h.Deps.ListCatalog(r.Context(), tid, r.URL.Query().Get("surface"))
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": list})
}

func (h *Handler) upsertPolicy(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.GatewayPolicy
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	p, err := h.Deps.UpsertPolicy(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, p)
}

func (h *Handler) exportPolicies(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	list, err := h.Deps.ExportGatewayPolicies(r.Context(), tid)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"policies": list})
}

func (h *Handler) registerWebhook(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.WebhookEndpoint
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	wh, err := h.Deps.RegisterWebhook(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, map[string]any{"id": wh.ID, "url": wh.URL, "events": wh.Events, "version": wh.Version, "active": wh.Active})
}

func (h *Handler) enqueueWebhook(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body struct {
		EventType string         `json:"eventType"`
		Payload   map[string]any `json:"payload"`
	}
	if err := decodeJSON(r, &body); err != nil || body.EventType == "" {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	n, err := h.Deps.EnqueueWebhook(r.Context(), tid, body.EventType, body.Payload)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"queued": n})
}

func (h *Handler) processWebhooks(w http.ResponseWriter, r *http.Request) {
	n, err := h.Deps.ProcessDeliveries(r.Context(), 100)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"delivered": n})
}

func (h *Handler) replay(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	d, err := h.Deps.ReplayDelivery(r.Context(), tid, id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, d)
}

func (h *Handler) publishSdk(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.SdkPackage
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	s, err := h.Deps.PublishSdk(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, s)
}

func (h *Handler) listSdks(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	list, err := h.Deps.SDKs.List(r.Context(), tid)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": list})
}

func (h *Handler) connectInt(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.IntegrationConnector
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	i, err := h.Deps.ConnectIntegration(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, i)
}

func (h *Handler) listInts(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	list, err := h.Deps.Integrations.List(r.Context(), tid)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": list})
}

func (h *Handler) recordUsage(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body struct {
		AppID uuid.UUID `json:"appId"`
		OK    bool      `json:"ok"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	if err := h.Deps.RecordUsage(r.Context(), tid, body.AppID, body.OK); err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]string{"status": "recorded"})
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
