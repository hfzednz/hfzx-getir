package httpadapter

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/recommendation-service/internal/app"
	"github.com/nexora/recommendation-service/internal/domain"
	"github.com/nexora/recommendation-service/internal/ratelimit"
)

type Handler struct {
	Deps *app.Deps
}

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
	const base = "/v1/recommendations"

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		writeOK(w, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /ready", func(w http.ResponseWriter, r *http.Request) {
		writeOK(w, map[string]string{"status": "ready"})
	})

	mux.HandleFunc("POST "+base+"/rails", tenant(h.recommend))
	mux.HandleFunc("POST "+base+"/similar", tenant(h.similar))
	mux.HandleFunc("POST "+base+"/for-you", tenant(h.forYou))
	mux.HandleFunc("POST "+base+"/fbt", tenant(h.fbt))
	mux.HandleFunc("POST "+base+"/next-best", tenant(h.nextBest))
	mux.HandleFunc("POST "+base+"/signals", tenant(h.signal))
	mux.HandleFunc("POST "+base+"/features", tenant(h.features))
	mux.HandleFunc("POST "+base+"/clicks", tenant(h.click))
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

func (h *Handler) recommend(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.RecommendRequest
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	rail, err := h.Deps.Recommend(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, rail)
}

func (h *Handler) similar(w http.ResponseWriter, r *http.Request) {
	h.strategyProxy(w, r, domain.StrategyContent)
}

func (h *Handler) forYou(w http.ResponseWriter, r *http.Request) {
	h.strategyProxy(w, r, domain.StrategyPersonalized)
}

func (h *Handler) fbt(w http.ResponseWriter, r *http.Request) {
	h.strategyProxy(w, r, domain.StrategyFBT)
}

func (h *Handler) nextBest(w http.ResponseWriter, r *http.Request) {
	h.strategyProxy(w, r, domain.StrategyHybrid)
}

func (h *Handler) strategyProxy(w http.ResponseWriter, r *http.Request, strategy string) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.RecommendRequest
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	body.Strategy = strategy
	rail, err := h.Deps.Recommend(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, rail)
}

func (h *Handler) signal(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.BehaviorSignal
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	if body.UserID == uuid.Nil {
		if uid, ok := UserIDFromContext(r.Context()); ok {
			body.UserID = uid
		}
	}
	s, err := h.Deps.IngestSignal(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, s)
}

func (h *Handler) features(w http.ResponseWriter, r *http.Request) {
	_, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.ProductFeatures
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	if err := h.Deps.UpsertFeatures(r.Context(), body); err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, body)
}

func (h *Handler) click(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body struct {
		RailID    uuid.UUID  `json:"railId"`
		ProductID uuid.UUID  `json:"productId"`
		UserID    *uuid.UUID `json:"userId"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	if err := h.Deps.RecordClick(r.Context(), tid, body.RailID, body.ProductID, body.UserID); err != nil {
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
