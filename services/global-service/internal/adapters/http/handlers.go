package httpadapter

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/global-service/internal/app"
	"github.com/nexora/global-service/internal/domain"
	"github.com/nexora/global-service/internal/ratelimit"
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
	const base = "/v1/global"

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		writeOK(w, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /ready", func(w http.ResponseWriter, _ *http.Request) {
		writeOK(w, map[string]string{"status": "ready"})
	})

	mux.HandleFunc("POST "+base+"/bootstrap/tr", tenant(h.bootstrapTR))
	mux.HandleFunc("POST "+base+"/countries", tenant(h.upsertCountry))
	mux.HandleFunc("POST "+base+"/countries/{id}/activate", tenant(h.activateCountry))
	mux.HandleFunc("GET "+base+"/countries", tenant(h.listCountries))
	mux.HandleFunc("POST "+base+"/places", tenant(h.upsertPlace))
	mux.HandleFunc("POST "+base+"/languages", tenant(h.addLang))
	mux.HandleFunc("POST "+base+"/locales", tenant(h.upsertLocale))
	mux.HandleFunc("POST "+base+"/translations", tenant(h.upsertTranslation))
	mux.HandleFunc("POST "+base+"/translations/ai", tenant(h.aiTranslate))
	mux.HandleFunc("POST "+base+"/currencies", tenant(h.upsertCurrency))
	mux.HandleFunc("POST "+base+"/rates", tenant(h.upsertRate))
	mux.HandleFunc("POST "+base+"/rates/refresh", tenant(h.refreshFX))
	mux.HandleFunc("GET "+base+"/convert", tenant(h.convert))
	mux.HandleFunc("POST "+base+"/holidays", tenant(h.importHoliday))
	mux.HandleFunc("POST "+base+"/rules", tenant(h.upsertRule))
	mux.HandleFunc("POST "+base+"/tax", tenant(h.upsertTax))
	mux.HandleFunc("POST "+base+"/privacy", tenant(h.upsertPrivacy))
	mux.HandleFunc("POST "+base+"/payments-availability", tenant(h.upsertPay))
	mux.HandleFunc("POST "+base+"/logistics-policy", tenant(h.upsertLogistics))
	mux.HandleFunc("POST "+base+"/legal-docs", tenant(h.upsertLegal))
	mux.HandleFunc("GET "+base+"/resolve", tenant(h.resolve))
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

func (h *Handler) bootstrapTR(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	if err := h.Deps.SeedTR(r.Context(), tid); err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]string{"status": "seeded", "country": "TR"})
}

func (h *Handler) upsertCountry(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.Country
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	c, err := h.Deps.UpsertCountry(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, c)
}

func (h *Handler) activateCountry(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	c, err := h.Deps.ActivateCountry(r.Context(), tid, id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, c)
}

func (h *Handler) listCountries(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	list, err := h.Deps.Countries.List(r.Context(), tid)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": list})
}

func (h *Handler) upsertPlace(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.Place
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	p, err := h.Deps.UpsertPlace(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, p)
}

func (h *Handler) addLang(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.Language
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	l, err := h.Deps.AddLanguage(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, l)
}

func (h *Handler) upsertLocale(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.LocaleProfile
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	l, err := h.Deps.UpsertLocale(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, l)
}

func (h *Handler) upsertTranslation(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.Translation
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	t, err := h.Deps.UpsertTranslation(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, t)
}

func (h *Handler) aiTranslate(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body struct {
		Namespace  string `json:"namespace"`
		Key        string `json:"key"`
		FromLocale string `json:"fromLocale"`
		ToLocale   string `json:"toLocale"`
		Source     string `json:"source"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	t, err := h.Deps.AIAssistTranslate(r.Context(), tid, body.Namespace, body.Key, body.FromLocale, body.ToLocale, body.Source)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, t)
}

func (h *Handler) upsertCurrency(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.Currency
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	c, err := h.Deps.UpsertCurrency(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, c)
}

func (h *Handler) upsertRate(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.ExchangeRate
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	rate, err := h.Deps.UpsertExchangeRate(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, rate)
}

func (h *Handler) refreshFX(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body struct {
		Base  string `json:"base"`
		Quote string `json:"quote"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	rate, err := h.Deps.RefreshFX(r.Context(), tid, body.Base, body.Quote)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, rate)
}

func (h *Handler) convert(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	amount, _ := strconv.ParseInt(r.URL.Query().Get("amountMinor"), 10, 64)
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")
	out, err := h.Deps.Convert(r.Context(), tid, amount, from, to)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"amountMinor": out, "from": from, "to": to})
}

func (h *Handler) importHoliday(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.Holiday
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	hol, err := h.Deps.ImportHoliday(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, hol)
}

func (h *Handler) upsertRule(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.RegionalRule
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	rule, err := h.Deps.UpsertRule(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, rule)
}

func (h *Handler) upsertTax(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.TaxRule
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	t, err := h.Deps.UpsertTax(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, t)
}

func (h *Handler) upsertPrivacy(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.PrivacyRegime
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	p, err := h.Deps.UpsertPrivacy(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, p)
}

func (h *Handler) upsertPay(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.PaymentMethodAvailability
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	p, err := h.Deps.UpsertPayAvail(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, p)
}

func (h *Handler) upsertLogistics(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.LogisticsPolicy
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	p, err := h.Deps.UpsertLogistics(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, p)
}

func (h *Handler) upsertLegal(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.LegalDocument
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	d, err := h.Deps.UpsertLegal(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, d)
}

func (h *Handler) resolve(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	bundle, err := h.Deps.Resolve(r.Context(), tid, r.URL.Query().Get("country"), r.URL.Query().Get("locale"), r.URL.Query().Get("ns"))
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, bundle)
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
