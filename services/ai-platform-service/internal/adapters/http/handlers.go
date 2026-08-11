package httpadapter

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/ai-platform-service/internal/app"
	"github.com/nexora/ai-platform-service/internal/domain"
	"github.com/nexora/ai-platform-service/internal/ratelimit"
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
	const base = "/v1/ai"

	mux.HandleFunc("GET /health", h.health)
	mux.HandleFunc("GET /ready", h.ready)

	mux.HandleFunc("POST "+base+"/features", tenant(h.upsertFeature))
	mux.HandleFunc("GET "+base+"/features/{entityType}/{entityId}", tenant(h.listFeatures))

	mux.HandleFunc("POST "+base+"/models", tenant(h.registerModel))
	mux.HandleFunc("POST "+base+"/models/{key}/{version}/promote", tenant(h.promoteModel))
	mux.HandleFunc("GET "+base+"/models", tenant(h.listModels))

	mux.HandleFunc("POST "+base+"/infer", tenant(h.infer))
	mux.HandleFunc("POST "+base+"/forecast/demand", tenant(h.forecast))
	mux.HandleFunc("POST "+base+"/fraud/score", tenant(h.fraud))
	mux.HandleFunc("POST "+base+"/pricing/suggest", tenant(h.pricing))
	mux.HandleFunc("POST "+base+"/embeddings", tenant(h.embed))

	mux.HandleFunc("POST "+base+"/prompts", tenant(h.upsertPrompt))
	mux.HandleFunc("POST "+base+"/llm/complete", tenant(h.llm))
	mux.HandleFunc("POST "+base+"/agents/{kind}/run", tenant(h.agent))
	mux.HandleFunc("POST "+base+"/automation/rules", tenant(h.upsertRule))
	mux.HandleFunc("POST "+base+"/automation/evaluate", tenant(h.evalRules))
	mux.HandleFunc("POST "+base+"/drift", tenant(h.drift))

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
		WriteTimeout: 120 * time.Second, IdleTimeout: 60 * time.Second,
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

func (h *Handler) upsertFeature(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.FeatureRecord
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	f, err := h.Deps.UpsertFeature(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, f)
}

func (h *Handler) listFeatures(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	entityType := r.PathValue("entityType")
	entityID, err := uuid.Parse(r.PathValue("entityId"))
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	items, err := h.Deps.Features.ListByEntity(r.Context(), tid, entityType, entityID)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": items})
}

func (h *Handler) registerModel(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.ModelCard
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	m, err := h.Deps.RegisterModel(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, m)
}

func (h *Handler) promoteModel(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	key := r.PathValue("key")
	version := r.PathValue("version")
	var body struct {
		Stage    string     `json:"stage"`
		Approver *uuid.UUID `json:"approverId"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	if body.Approver == nil {
		if uid, ok := UserIDFromContext(r.Context()); ok {
			body.Approver = &uid
		}
	}
	m, err := h.Deps.PromoteModel(r.Context(), tid, key, version, body.Stage, body.Approver)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, m)
}

func (h *Handler) listModels(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	items, err := h.Deps.Models.List(r.Context(), tid, r.URL.Query().Get("key"))
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": items})
}

func (h *Handler) infer(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.InferenceRequest
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	res, err := h.Deps.Infer(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, res)
}

func (h *Handler) forecast(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body struct {
		ProductID   uuid.UUID `json:"productId"`
		HorizonDays int       `json:"horizonDays"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	res, err := h.Deps.ForecastDemand(r.Context(), tid, body.ProductID, body.HorizonDays)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, res)
}

func (h *Handler) fraud(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body struct {
		EntityType string             `json:"entityType"`
		EntityID   uuid.UUID          `json:"entityId"`
		Features   map[string]float64 `json:"features"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	res, err := h.Deps.ScoreFraud(r.Context(), tid, body.EntityType, body.EntityID, body.Features)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, res)
}

func (h *Handler) pricing(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body struct {
		ProductID uuid.UUID          `json:"productId"`
		Features  map[string]float64 `json:"features"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	res, err := h.Deps.SuggestPrice(r.Context(), tid, body.ProductID, body.Features)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, res)
}

func (h *Handler) embed(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body struct {
		Text string `json:"text"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	vec, err := h.Deps.EmbedText(r.Context(), tid, body.Text)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"embedding": vec, "dims": len(vec)})
}

func (h *Handler) upsertPrompt(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.PromptTemplate
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	p, err := h.Deps.UpsertPrompt(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, p)
}

func (h *Handler) llm(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.LLMRequest
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	res, err := h.Deps.CompleteLLM(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, res)
}

func (h *Handler) agent(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	kind := r.PathValue("kind")
	var body struct {
		Input     string     `json:"input"`
		SessionID *uuid.UUID `json:"sessionId"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	run, err := h.Deps.RunAgent(r.Context(), tid, kind, body.Input, body.SessionID)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, run)
}

func (h *Handler) upsertRule(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.AutomationRule
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	rule, err := h.Deps.UpsertAutomationRule(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, rule)
}

func (h *Handler) evalRules(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body struct {
		Features map[string]float64 `json:"features"`
		Approve  bool               `json:"approve"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	runs, err := h.Deps.EvaluateAutomation(r.Context(), tid, body.Features, body.Approve)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"runs": runs})
}

func (h *Handler) drift(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body struct {
		ModelKey  string  `json:"modelKey"`
		Metric    string  `json:"metric"`
		Value     float64 `json:"value"`
		Threshold float64 `json:"threshold"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	rep, err := h.Deps.ReportDrift(r.Context(), tid, body.ModelKey, body.Metric, body.Value, body.Threshold)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, rep)
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
