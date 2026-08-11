package httpadapter

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/search-service/internal/app"
	"github.com/nexora/search-service/internal/domain"
	"github.com/nexora/search-service/internal/ratelimit"
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
	const base = "/v1/search"

	mux.HandleFunc("GET /health", h.health)
	mux.HandleFunc("GET /ready", h.ready)
	mux.HandleFunc("GET "+base+"/health", h.health)

	mux.HandleFunc("POST "+base+"/query", tenant(h.query))
	mux.HandleFunc("GET "+base+"/autocomplete", tenant(h.autocomplete))
	mux.HandleFunc("POST "+base+"/suggest/click", tenant(h.suggestClick))
	mux.HandleFunc("POST "+base+"/voice", tenant(h.voice))
	mux.HandleFunc("POST "+base+"/image", tenant(h.image))
	mux.HandleFunc("GET "+base+"/trends", tenant(h.trends))
	mux.HandleFunc("POST "+base+"/trends/refresh", tenant(h.refreshTrends))

	mux.HandleFunc("POST "+base+"/index/documents", tenant(h.indexDoc))
	mux.HandleFunc("POST "+base+"/index/reindex/{productId}", tenant(h.reindex))
	mux.HandleFunc("POST "+base+"/index/jobs", tenant(h.startJob))

	mux.HandleFunc("POST "+base+"/synonyms", tenant(h.upsertSynonym))
	mux.HandleFunc("GET "+base+"/synonyms", tenant(h.listSynonyms))
	mux.HandleFunc("POST "+base+"/boosts", tenant(h.upsertBoost))
	mux.HandleFunc("GET "+base+"/boosts", tenant(h.listBoosts))

	mux.HandleFunc("POST "+base+"/summarize", tenant(h.summarize))
	mux.HandleFunc("GET "+base+"/admin/stats", tenant(h.adminStats))
	mux.HandleFunc("POST "+base+"/outbox/publish", tenant(h.publishOutbox))

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

func requireTenant(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	tid, ok := TenantIDFromContext(r.Context())
	if !ok {
		writeErr(w, r, domain.ErrInvalidArgument)
		return uuid.Nil, false
	}
	return tid, true
}

func (h *Handler) query(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body struct {
		Query         string              `json:"query"`
		Locale        string              `json:"locale"`
		UserID        *uuid.UUID          `json:"userId"`
		CityID        *uuid.UUID          `json:"cityId"`
		Sort          string              `json:"sort"`
		Page          int                 `json:"page"`
		PageSize      int                 `json:"pageSize"`
		Hybrid        bool                `json:"hybrid"`
		Personalize   bool                `json:"personalize"`
		IncludeFacets bool                `json:"includeFacets"`
		Filters       domain.SearchFilters `json:"filters"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	res, err := h.Deps.Search(r.Context(), domain.SearchQuery{
		TenantID: tid, RawQuery: body.Query, Locale: body.Locale, UserID: body.UserID,
		CityID: body.CityID, Sort: body.Sort, Page: body.Page, PageSize: body.PageSize,
		Hybrid: body.Hybrid, Personalize: body.Personalize, IncludeFacets: body.IncludeFacets,
		Filters: body.Filters,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, res)
}

func (h *Handler) autocomplete(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	q := r.URL.Query().Get("q")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	items, err := h.Deps.Autocomplete(r.Context(), tid, q, limit)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": items})
}

func (h *Handler) suggestClick(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body struct {
		Text      string     `json:"text"`
		ProductID *uuid.UUID `json:"productId"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	if err := h.Deps.RecordSuggestionClick(r.Context(), tid, body.Text, body.ProductID); err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]string{"status": "recorded"})
}

func (h *Handler) voice(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body struct {
		Transcript string     `json:"transcript"`
		Locale     string     `json:"locale"`
		UserID     *uuid.UUID `json:"userId"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	res, err := h.Deps.VoiceSearch(r.Context(), tid, body.Transcript, body.Locale, body.UserID)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, res)
}

func (h *Handler) image(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body struct {
		ImageRef string `json:"imageRef"`
		Limit    int    `json:"limit"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	res, err := h.Deps.ImageSearch(r.Context(), tid, body.ImageRef, body.Limit)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, res)
}

func (h *Handler) trends(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	kind := r.URL.Query().Get("kind")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	items, err := h.Deps.ListTrends(r.Context(), tid, kind, limit)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": items})
}

func (h *Handler) refreshTrends(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	if err := h.Deps.RefreshTrends(r.Context(), tid); err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]string{"status": "ok"})
}

func (h *Handler) indexDoc(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var doc domain.ProductDocument
	if err := decodeJSON(r, &doc); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	doc.TenantID = tid
	if err := h.Deps.IndexDocument(r.Context(), doc); err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, map[string]any{"productId": doc.ProductID.String(), "status": "indexed"})
}

func (h *Handler) reindex(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	pid, err := uuid.Parse(r.PathValue("productId"))
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	if err := h.Deps.ReindexFromCatalog(r.Context(), tid, pid); err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]string{"status": "reindexed"})
}

func (h *Handler) startJob(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body struct {
		Mode string `json:"mode"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	job, err := h.Deps.StartIndexJob(r.Context(), tid, body.Mode)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, job)
}

func (h *Handler) upsertSynonym(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.SynonymRule
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	s, err := h.Deps.UpsertSynonym(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, s)
}

func (h *Handler) listSynonyms(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	items, err := h.Deps.Synonyms.List(r.Context(), tid, r.URL.Query().Get("locale"))
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": items})
}

func (h *Handler) upsertBoost(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.BoostRule
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	b, err := h.Deps.UpsertBoost(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, b)
}

func (h *Handler) listBoosts(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	items, err := h.Deps.Boosts.ListActive(r.Context(), tid, time.Now().UTC())
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": items})
}

func (h *Handler) summarize(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body struct {
		Query string `json:"query"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	res, err := h.Deps.Search(r.Context(), domain.SearchQuery{
		TenantID: tid, RawQuery: body.Query, Hybrid: true, Page: 1, PageSize: 5,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	sum, err := h.Deps.SummarizeSearch(r.Context(), tid, body.Query, res.Hits)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"summary": sum, "hits": res.Hits})
}

func (h *Handler) adminStats(w http.ResponseWriter, r *http.Request) {
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

func (h *Handler) publishOutbox(w http.ResponseWriter, r *http.Request) {
	n, err := h.Deps.PublishPending(r.Context(), 100)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"published": n})
}
