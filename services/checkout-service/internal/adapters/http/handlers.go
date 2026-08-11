package httpadapter

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/checkout-service/internal/app"
	"github.com/nexora/checkout-service/internal/domain"
	"github.com/nexora/checkout-service/internal/ratelimit"
)

// Handler serves checkout REST endpoints.
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
	const base = "/v1/checkout"

	mux.HandleFunc("GET /health", h.health)
	mux.HandleFunc("GET /ready", h.ready)
	mux.HandleFunc("GET "+base+"/health", h.health)
	mux.HandleFunc("GET "+base+"/ready", h.ready)

	mux.HandleFunc("POST "+base+"/sessions", tenant(h.startSession))
	mux.HandleFunc("GET "+base+"/sessions", tenant(h.listSessions))
	mux.HandleFunc("GET "+base+"/sessions/{id}", tenant(h.getSession))
	mux.HandleFunc("PATCH "+base+"/sessions/{id}", tenant(h.patchSession))
	mux.HandleFunc("POST "+base+"/sessions/{id}/validate", tenant(h.validateSession))
	mux.HandleFunc("POST "+base+"/sessions/{id}/refresh-quote", tenant(h.refreshQuote))
	mux.HandleFunc("POST "+base+"/sessions/{id}/complete", tenant(h.completeSession))
	mux.HandleFunc("POST "+base+"/sessions/{id}/abandon", tenant(h.abandonSession))
	mux.HandleFunc("POST "+base+"/recover", tenant(h.recoverSession))
	mux.HandleFunc("GET "+base+"/admin/metrics", tenant(h.adminMetrics))
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

func (h *Handler) tenant(r *http.Request) uuid.UUID {
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

func parseUUIDParam(r *http.Request, name string) (uuid.UUID, error) {
	return uuid.Parse(r.PathValue(name))
}

func idemKey(r *http.Request, bodyKey string) string {
	if v := r.Header.Get("Idempotency-Key"); v != "" {
		return v
	}
	return bodyKey
}

func sessionDTO(s domain.Session) map[string]any {
	return map[string]any{
		"id":             s.ID,
		"tenantId":       s.TenantID,
		"cartId":         s.CartID,
		"principalId":    s.PrincipalID,
		"status":         s.Status,
		"deliveryOption": s.DeliveryOption,
		"address":        s.Address,
		"slot":           s.Slot,
		"gift":           s.Gift,
		"invoice":        s.Invoice,
		"substitutions":  s.Substitutions,
		"notes":          s.Notes,
		"tipMinor":       s.TipMinor,
		"currency":       s.Currency,
		"validation":     s.Validation,
		"quote":          s.Quote,
		"orderId":        s.OrderID,
		"idempotencyKey": s.IdempotencyKey,
		"recoveryToken":  s.RecoveryToken,
		"cityId":         s.CityID,
		"couponCodes":    s.CouponCodes,
		"version":        s.Version,
		"createdAt":      s.CreatedAt,
		"updatedAt":      s.UpdatedAt,
		"completedAt":    s.CompletedAt,
		"abandonedAt":    s.AbandonedAt,
	}
}

func (h *Handler) startSession(w http.ResponseWriter, r *http.Request) {
	var body struct {
		CartID         uuid.UUID             `json:"cartId"`
		PrincipalID    *uuid.UUID            `json:"principalId"`
		IdempotencyKey string                `json:"idempotencyKey"`
		DeliveryOption domain.DeliveryOption `json:"deliveryOption"`
		Currency       string                `json:"currency"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	principal := h.principal(r, body.PrincipalID)
	s, err := h.Deps.StartFromCart(r.Context(), app.StartFromCartInput{
		TenantID: h.tenant(r), CartID: body.CartID, PrincipalID: principal,
		IdempotencyKey: idemKey(r, body.IdempotencyKey),
		DeliveryOption: body.DeliveryOption, Currency: body.Currency,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, sessionDTO(s))
}

func (h *Handler) getSession(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	s, err := h.Deps.GetSession(r.Context(), h.tenant(r), id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, sessionDTO(s))
}

func (h *Handler) listSessions(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))
	var status *domain.SessionStatus
	if raw := q.Get("status"); raw != "" {
		st := domain.SessionStatus(raw)
		status = &st
	}
	var principal *uuid.UUID
	if raw := q.Get("principalId"); raw != "" {
		if id, err := uuid.Parse(raw); err == nil {
			principal = &id
		}
	}
	items, total, err := h.Deps.ListSessions(r.Context(), app.ListSessionsInput{
		TenantID: h.tenant(r), PrincipalID: principal, Status: status,
		Query: q.Get("q"), Limit: limit, Offset: offset,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	out := make([]map[string]any, 0, len(items))
	for _, s := range items {
		out = append(out, sessionDTO(s))
	}
	writeOK(w, map[string]any{"items": out, "total": total})
}

func (h *Handler) patchSession(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		Address        *domain.AddressSnapshot    `json:"address"`
		Slot           *domain.SlotSnapshot       `json:"slot"`
		Gift           *domain.GiftPrefs          `json:"gift"`
		Invoice        *domain.InvoicePrefs       `json:"invoice"`
		Notes          *string                    `json:"notes"`
		Substitutions  *domain.SubstitutionPolicy `json:"substitutions"`
		TipMinor       *int64                     `json:"tipMinor"`
		DeliveryOption *domain.DeliveryOption     `json:"deliveryOption"`
		CouponCodes    []string                   `json:"couponCodes"`
		ClearCoupons   bool                       `json:"clearCoupons"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	s, err := h.Deps.Patch(r.Context(), app.PatchInput{
		TenantID: h.tenant(r), SessionID: id,
		Address: body.Address, Slot: body.Slot, Gift: body.Gift, Invoice: body.Invoice,
		Notes: body.Notes, Substitutions: body.Substitutions, TipMinor: body.TipMinor,
		DeliveryOption: body.DeliveryOption, CouponCodes: body.CouponCodes, ClearCoupons: body.ClearCoupons,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, sessionDTO(s))
}

func (h *Handler) validateSession(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	s, err := h.Deps.Validate(r.Context(), app.ValidateInput{TenantID: h.tenant(r), SessionID: id})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, sessionDTO(s))
}

func (h *Handler) refreshQuote(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	s, err := h.Deps.RefreshQuote(r.Context(), app.RefreshQuoteInput{TenantID: h.tenant(r), SessionID: id})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, sessionDTO(s))
}

func (h *Handler) completeSession(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		IdempotencyKey string `json:"idempotencyKey"`
		PlaceOrder     bool   `json:"placeOrder"`
	}
	_ = decodeJSON(r, &body)
	s, err := h.Deps.Complete(r.Context(), app.CompleteInput{
		TenantID: h.tenant(r), SessionID: id,
		IdempotencyKey: idemKey(r, body.IdempotencyKey),
		PlaceOrder:     body.PlaceOrder,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, sessionDTO(s))
}

func (h *Handler) abandonSession(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	s, err := h.Deps.Abandon(r.Context(), app.AbandonInput{TenantID: h.tenant(r), SessionID: id})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, sessionDTO(s))
}

func (h *Handler) recoverSession(w http.ResponseWriter, r *http.Request) {
	var body struct {
		RecoveryToken string     `json:"recoveryToken"`
		PrincipalID   *uuid.UUID `json:"principalId"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	principal := h.principal(r, body.PrincipalID)
	s, err := h.Deps.RecoverAbandoned(r.Context(), app.RecoverAbandonedInput{
		RecoveryToken: body.RecoveryToken, PrincipalID: principal,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, sessionDTO(s))
}

func (h *Handler) adminMetrics(w http.ResponseWriter, r *http.Request) {
	m, err := h.Deps.Metrics(r.Context(), h.tenant(r))
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, m)
}

func (h *Handler) publishOutbox(w http.ResponseWriter, r *http.Request) {
	n, err := h.Deps.PublishPending(r.Context(), 100)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"published": n})
}
