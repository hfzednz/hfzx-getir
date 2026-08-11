package httpadapter

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/payment-service/internal/app"
	"github.com/nexora/payment-service/internal/domain"
	"github.com/nexora/payment-service/internal/ratelimit"
)

// Handler serves payment REST endpoints.
type Handler struct {
	Deps  *app.Deps
	Ready func(*http.Request) error
	Live  func(*http.Request) error
}

// ServerConfig configures the HTTP server.
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

// NewHandler returns a fully wired http.Handler.
func NewHandler(cfg ServerConfig) http.Handler {
	h := &Handler{Deps: cfg.Deps, Ready: cfg.Ready, Live: cfg.Live}
	mux := http.NewServeMux()
	const base = "/v1/payments"

	mux.HandleFunc("GET /health", h.health)
	mux.HandleFunc("GET /ready", h.ready)
	mux.HandleFunc("GET "+base+"/health", h.health)
	mux.HandleFunc("GET "+base+"/ready", h.ready)

	mux.HandleFunc("POST "+base+"/intents", tenant(h.createIntent))
	mux.HandleFunc("GET "+base+"/intents", tenant(h.listIntents))
	mux.HandleFunc("GET "+base+"/intents/{id}", tenant(h.getIntent))
	mux.HandleFunc("POST "+base+"/intents/{id}/authorize", tenant(h.authorize))
	mux.HandleFunc("POST "+base+"/intents/{id}/capture", tenant(h.capture))
	mux.HandleFunc("POST "+base+"/intents/{id}/void", tenant(h.void))
	mux.HandleFunc("POST "+base+"/intents/{id}/refund", tenant(h.refund))
	mux.HandleFunc("POST "+base+"/eligibility", tenant(h.eligibility))
	mux.HandleFunc("POST "+base+"/routes", tenant(h.upsertRoute))
	mux.HandleFunc("GET "+base+"/routes", tenant(h.listRoutes))
	mux.HandleFunc("POST "+base+"/intents/{id}/chargebacks", tenant(h.chargeback))
	mux.HandleFunc("POST "+base+"/outbox/publish", tenant(h.publishOutbox))

	return chain(mux,
		requestIDMiddleware,
		recoverMiddleware(cfg.Log),
		loggingMiddleware(cfg.Log),
		corsMiddleware(cfg.CORSOrigins),
		rateLimitMiddleware(cfg.Limiter, cfg.RateLimitPerMinute),
	)
}

// NewServer builds an *http.Server with sensible timeouts.
func NewServer(cfg ServerConfig) *http.Server {
	return &http.Server{
		Addr:              cfg.Addr,
		Handler:           NewHandler(cfg),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
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

func (h *Handler) tenantID(r *http.Request) uuid.UUID {
	tid, _ := TenantIDFromContext(r.Context())
	return tid
}

func (h *Handler) principal(r *http.Request, bodyPrincipal *uuid.UUID) uuid.UUID {
	if uid, ok := UserIDFromContext(r.Context()); ok {
		return uid
	}
	if bodyPrincipal != nil {
		return *bodyPrincipal
	}
	return uuid.Nil
}

func idemKey(r *http.Request, bodyKey string) string {
	if v := r.Header.Get("Idempotency-Key"); v != "" {
		return v
	}
	return bodyKey
}

func intentDTO(i domain.PaymentIntent) map[string]any {
	return map[string]any{
		"id":                i.ID,
		"tenantId":          i.TenantID,
		"principalId":       i.PrincipalID,
		"orderId":           i.OrderID,
		"status":            i.Status,
		"amountMinor":       i.AmountMinor,
		"capturedMinor":     i.CapturedMinor,
		"refundedMinor":     i.RefundedMinor,
		"currency":          i.Currency,
		"methodType":        i.MethodType,
		"paymentMethodId":   i.PaymentMethodID,
		"provider":          i.Provider,
		"providerIntentRef": i.ProviderIntentRef,
		"idempotencyKey":    i.IdempotencyKey,
		"fraudScore":        i.FraudScore,
		"fraudDecision":     i.FraudDecision,
		"failureReason":     i.FailureReason,
		"version":           i.Version,
		"createdAt":         i.CreatedAt,
		"updatedAt":         i.UpdatedAt,
	}
}

func (h *Handler) createIntent(w http.ResponseWriter, r *http.Request) {
	var body struct {
		PrincipalID     *uuid.UUID               `json:"principalId"`
		OrderID         string                   `json:"orderId"`
		AmountMinor     int64                    `json:"amountMinor"`
		Currency        string                   `json:"currency"`
		MethodType      domain.PaymentMethodType `json:"methodType"`
		PaymentMethodID *uuid.UUID               `json:"paymentMethodId"`
		IdempotencyKey  string                   `json:"idempotencyKey"`
		Metadata        map[string]any           `json:"metadata"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	intent, err := h.Deps.CreateIntent(r.Context(), app.CreateIntentInput{
		TenantID: h.tenantID(r), PrincipalID: h.principal(r, body.PrincipalID),
		OrderID: body.OrderID, AmountMinor: body.AmountMinor, Currency: body.Currency,
		MethodType: body.MethodType, PaymentMethodID: body.PaymentMethodID,
		IdempotencyKey: idemKey(r, body.IdempotencyKey), Metadata: body.Metadata,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, intentDTO(intent))
}

func (h *Handler) getIntent(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	intent, err := h.Deps.GetIntent(r.Context(), h.tenantID(r), id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, intentDTO(intent))
}

func (h *Handler) listIntents(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	var status *domain.IntentStatus
	if s := q.Get("status"); s != "" {
		st := domain.IntentStatus(s)
		status = &st
	}
	var principal *uuid.UUID
	if p := q.Get("principalId"); p != "" {
		if id, err := uuid.Parse(p); err == nil {
			principal = &id
		}
	}
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))
	items, total, err := h.Deps.AdminList(r.Context(), app.AdminListInput{
		TenantID: h.tenantID(r), PrincipalID: principal, Status: status,
		OrderID: q.Get("orderId"), Limit: limit, Offset: offset,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	out := make([]map[string]any, 0, len(items))
	for _, i := range items {
		out = append(out, intentDTO(i))
	}
	writeOK(w, map[string]any{"items": out, "total": total})
}

func (h *Handler) authorize(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		IdempotencyKey string `json:"idempotencyKey"`
		Token          string `json:"token"`
	}
	_ = decodeJSON(r, &body)
	intent, err := h.Deps.Authorize(r.Context(), app.AuthorizeInput{
		TenantID: h.tenantID(r), IntentID: id,
		IdempotencyKey: idemKey(r, body.IdempotencyKey), Token: body.Token,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, intentDTO(intent))
}

func (h *Handler) capture(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		AmountMinor    int64  `json:"amountMinor"`
		IdempotencyKey string `json:"idempotencyKey"`
	}
	_ = decodeJSON(r, &body)
	intent, err := h.Deps.Capture(r.Context(), app.CaptureInput{
		TenantID: h.tenantID(r), IntentID: id,
		AmountMinor: body.AmountMinor, IdempotencyKey: idemKey(r, body.IdempotencyKey),
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, intentDTO(intent))
}

func (h *Handler) void(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		IdempotencyKey string `json:"idempotencyKey"`
	}
	_ = decodeJSON(r, &body)
	intent, err := h.Deps.Void(r.Context(), app.VoidInput{
		TenantID: h.tenantID(r), IntentID: id, IdempotencyKey: idemKey(r, body.IdempotencyKey),
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, intentDTO(intent))
}

func (h *Handler) refund(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		AmountMinor    int64  `json:"amountMinor"`
		Reason         string `json:"reason"`
		IdempotencyKey string `json:"idempotencyKey"`
	}
	_ = decodeJSON(r, &body)
	refund, intent, err := h.Deps.Refund(r.Context(), app.RefundInput{
		TenantID: h.tenantID(r), IntentID: id,
		AmountMinor: body.AmountMinor, Reason: body.Reason,
		IdempotencyKey: idemKey(r, body.IdempotencyKey),
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{
		"refund": map[string]any{
			"id": refund.ID, "intentId": refund.IntentID, "amountMinor": refund.AmountMinor,
			"currency": refund.Currency, "status": refund.Status, "reason": refund.Reason,
		},
		"intent": intentDTO(intent),
	})
}

func (h *Handler) eligibility(w http.ResponseWriter, r *http.Request) {
	var body struct {
		PrincipalID *uuid.UUID `json:"principalId"`
		AmountMinor int64      `json:"amountMinor"`
		Currency    string     `json:"currency"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	res, err := h.Deps.CheckEligibility(r.Context(), app.EligibilityInput{
		TenantID: h.tenantID(r), PrincipalID: h.principal(r, body.PrincipalID),
		AmountMinor: body.AmountMinor, Currency: body.Currency,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, res)
}

func (h *Handler) upsertRoute(w http.ResponseWriter, r *http.Request) {
	var body struct {
		MethodType domain.PaymentMethodType `json:"methodType"`
		Currency   string                   `json:"currency"`
		Providers  []string                 `json:"providers"`
		Priority   int                      `json:"priority"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	route, err := h.Deps.UpsertRoute(r.Context(), app.UpsertRouteInput{
		TenantID: h.tenantID(r), MethodType: body.MethodType,
		Currency: body.Currency, Providers: body.Providers, Priority: body.Priority,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, route)
}

func (h *Handler) listRoutes(w http.ResponseWriter, r *http.Request) {
	method := domain.PaymentMethodType(r.URL.Query().Get("methodType"))
	currency := r.URL.Query().Get("currency")
	providers, err := h.Deps.RouteProvider(r.Context(), app.RouteProviderInput{
		TenantID: h.tenantID(r), MethodType: method, Currency: currency,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"providers": providers})
}

func (h *Handler) chargeback(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		AmountMinor int64  `json:"amountMinor"`
		ReasonCode  string `json:"reasonCode"`
		Reason      string `json:"reason"`
		ProviderRef string `json:"providerRef"`
	}
	_ = decodeJSON(r, &body)
	cb, err := h.Deps.RecordChargeback(r.Context(), app.RecordChargebackInput{
		TenantID: h.tenantID(r), IntentID: id, AmountMinor: body.AmountMinor,
		ReasonCode: body.ReasonCode, Reason: body.Reason, ProviderRef: body.ProviderRef,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, cb)
}

func (h *Handler) publishOutbox(w http.ResponseWriter, r *http.Request) {
	n, err := h.Deps.PublishPending(r.Context(), 100)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"published": n})
}
