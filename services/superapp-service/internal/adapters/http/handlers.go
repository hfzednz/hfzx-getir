package httpadapter

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/superapp-service/internal/app"
	"github.com/nexora/superapp-service/internal/domain"
	"github.com/nexora/superapp-service/internal/ratelimit"
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
	const base = "/v1/superapp"

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		writeOK(w, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /ready", func(w http.ResponseWriter, _ *http.Request) {
		writeOK(w, map[string]string{"status": "ready"})
	})

	mux.HandleFunc("POST "+base+"/bootstrap/mini-apps", tenant(h.seedMiniApps))
	mux.HandleFunc("POST "+base+"/modules", tenant(h.upsertModule))
	mux.HandleFunc("GET "+base+"/modules", tenant(h.listModules))
	mux.HandleFunc("POST "+base+"/modules/{id}/manifests", tenant(h.publishManifest))
	mux.HandleFunc("POST "+base+"/installs", tenant(h.install))
	mux.HandleFunc("POST "+base+"/installs/update", tenant(h.updateInstall))
	mux.HandleFunc("POST "+base+"/installs/remove", tenant(h.removeInstall))
	mux.HandleFunc("POST "+base+"/installs/rollback", tenant(h.rollback))
	mux.HandleFunc("POST "+base+"/permissions/grant", tenant(h.grantPerm))
	mux.HandleFunc("GET "+base+"/store", tenant(h.browseStore))
	mux.HandleFunc("POST "+base+"/store/rate", tenant(h.rate))
	mux.HandleFunc("POST "+base+"/widgets", tenant(h.addWidget))
	mux.HandleFunc("POST "+base+"/mini-apps/launch", tenant(h.launch))
	mux.HandleFunc("POST "+base+"/monetization", tenant(h.monetization))
	mux.HandleFunc("GET "+base+"/shell/resolve", tenant(h.resolve))
	mux.HandleFunc("GET "+base+"/recommend", tenant(h.recommend))
	mux.HandleFunc("GET "+base+"/search", tenant(h.search))
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

func (h *Handler) seedMiniApps(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	if err := h.Deps.SeedMiniApps(r.Context(), tid); err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]string{"status": "mini_apps_seeded"})
}

func (h *Handler) upsertModule(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.Module
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	m, err := h.Deps.UpsertModule(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, m)
}

func (h *Handler) listModules(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	list, err := h.Deps.Modules.List(r.Context(), tid)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": list})
}

func (h *Handler) publishManifest(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	moduleID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body domain.ModuleManifest
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	body.ModuleID = moduleID
	man, err := h.Deps.PublishManifest(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, man)
}

func (h *Handler) install(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body struct {
		SubjectID string `json:"subjectId"`
		ModuleKey string `json:"moduleKey"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	inst, err := h.Deps.InstallModule(r.Context(), tid, body.SubjectID, body.ModuleKey)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, inst)
}

func (h *Handler) updateInstall(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body struct {
		SubjectID string `json:"subjectId"`
		ModuleKey string `json:"moduleKey"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	inst, err := h.Deps.UpdateModule(r.Context(), tid, body.SubjectID, body.ModuleKey)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, inst)
}

func (h *Handler) removeInstall(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body struct {
		SubjectID string `json:"subjectId"`
		ModuleKey string `json:"moduleKey"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	inst, err := h.Deps.RemoveModule(r.Context(), tid, body.SubjectID, body.ModuleKey)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, inst)
}

func (h *Handler) rollback(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body struct {
		SubjectID string `json:"subjectId"`
		ModuleKey string `json:"moduleKey"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	inst, err := h.Deps.RollbackInstall(r.Context(), tid, body.SubjectID, body.ModuleKey)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, inst)
}

func (h *Handler) grantPerm(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body struct {
		SubjectID  string    `json:"subjectId"`
		ModuleID   uuid.UUID `json:"moduleId"`
		Permission string    `json:"permission"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	g, err := h.Deps.GrantPermission(r.Context(), tid, body.SubjectID, body.ModuleID, body.Permission)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, g)
}

func (h *Handler) browseStore(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	items, err := h.Deps.BrowseStore(r.Context(), tid, r.URL.Query().Get("category"))
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": items})
}

func (h *Handler) rate(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.StoreRating
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	rating, err := h.Deps.RateModule(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, rating)
}

func (h *Handler) addWidget(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.WidgetPlacement
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	widget, err := h.Deps.AddWidget(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, widget)
}

func (h *Handler) launch(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body struct {
		SubjectID string `json:"subjectId"`
		ModuleKey string `json:"moduleKey"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	ev, err := h.Deps.LaunchMiniApp(r.Context(), tid, body.SubjectID, body.ModuleKey)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, ev)
}

func (h *Handler) monetization(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.MonetizationRule
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	rule, err := h.Deps.UpsertMonetization(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, rule)
}

func (h *Handler) resolve(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	subject := r.URL.Query().Get("subjectId")
	shell := r.URL.Query().Get("shellVersion")
	res, err := h.Deps.ResolveShell(r.Context(), tid, subject, shell)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, res)
}

func (h *Handler) recommend(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 5
	}
	keys, err := h.Deps.Recommend(r.Context(), tid, r.URL.Query().Get("subjectId"), limit)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"keys": keys})
}

func (h *Handler) search(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	items, err := h.Deps.SearchModules(r.Context(), tid, r.URL.Query().Get("q"))
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": items})
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
