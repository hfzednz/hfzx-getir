package httpadapter

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/cart-service/internal/app"
	"github.com/nexora/cart-service/internal/domain"
	"github.com/nexora/cart-service/internal/ratelimit"
)

// Handler serves cart REST endpoints.
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
	const base = "/v1/cart"

	mux.HandleFunc("GET /health", h.health)
	mux.HandleFunc("GET /ready", h.ready)
	mux.HandleFunc("GET "+base+"/health", h.health)
	mux.HandleFunc("GET "+base+"/ready", h.ready)

	mux.HandleFunc("POST "+base, tenant(h.createCart))
	mux.HandleFunc("GET "+base, tenant(h.getCurrentCart))
	mux.HandleFunc("GET "+base+"/{id}", tenant(h.getCart))
	mux.HandleFunc("POST "+base+"/{id}/lines", tenant(h.addLine))
	mux.HandleFunc("PATCH "+base+"/{id}/lines/{lineId}", tenant(h.updateQty))
	mux.HandleFunc("DELETE "+base+"/{id}/lines/{lineId}", tenant(h.removeLine))
	mux.HandleFunc("POST "+base+"/{id}/coupons", tenant(h.applyCoupon))
	mux.HandleFunc("DELETE "+base+"/{id}/coupons/{code}", tenant(h.removeCoupon))
	mux.HandleFunc("POST "+base+"/{id}/refresh-quote", tenant(h.refreshQuote))
	mux.HandleFunc("POST "+base+"/merge", tenant(h.mergeCarts))
	mux.HandleFunc("POST "+base+"/{id}/soft-reserve", tenant(h.softReserve))
	mux.HandleFunc("POST "+base+"/{id}/abandon", tenant(h.abandon))
	mux.HandleFunc("POST "+base+"/{id}/recover", tenant(h.recover))
	mux.HandleFunc("GET "+base+"/{id}/recommendations", tenant(h.recommendations))
	mux.HandleFunc("POST "+base+"/{id}/save", tenant(h.saveCart))

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

func (h *Handler) principalPtr(r *http.Request) *uuid.UUID {
	if pid, ok := PrincipalIDFromContext(r.Context()); ok {
		return &pid
	}
	return nil
}

func (h *Handler) guest(r *http.Request) string {
	return GuestTokenFromContext(r.Context())
}

func parseUUIDParam(r *http.Request, name string) (uuid.UUID, error) {
	return uuid.Parse(r.PathValue(name))
}

func (h *Handler) createCart(w http.ResponseWriter, r *http.Request) {
	var body struct {
		GuestToken  string     `json:"guestToken"`
		PrincipalID *uuid.UUID `json:"principalId"`
		CityID      *uuid.UUID `json:"cityId"`
		Currency    string     `json:"currency"`
	}
	_ = decodeJSON(r, &body)
	guest := body.GuestToken
	if guest == "" {
		guest = h.guest(r)
	}
	principal := body.PrincipalID
	if principal == nil {
		principal = h.principalPtr(r)
	}
	out, err := h.Deps.CreateCart(r.Context(), app.CreateCartInput{
		TenantID: h.tenant(r), GuestToken: guest, PrincipalID: principal,
		CityID: body.CityID, Currency: body.Currency,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, out)
}

func (h *Handler) getCurrentCart(w http.ResponseWriter, r *http.Request) {
	out, err := h.Deps.GetOrCreateCart(r.Context(), app.CreateCartInput{
		TenantID: h.tenant(r), GuestToken: h.guest(r), PrincipalID: h.principalPtr(r),
		Currency: "TRY",
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, out)
}

func (h *Handler) getCart(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	out, err := h.Deps.GetCart(r.Context(), app.GetCartInput{
		TenantID: h.tenant(r), CartID: &id,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, out)
}

func (h *Handler) addLine(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		VariantID       uuid.UUID          `json:"variantId"`
		Qty             int                `json:"qty"`
		MaxQty          int                `json:"maxQty"`
		Notes           string             `json:"notes"`
		Addons          []domain.LineAddon `json:"addons"`
		ReplacementPref string             `json:"replacementPref"`
		SKU             string             `json:"sku"`
		UnitMinor       int64              `json:"unitMinor"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	out, err := h.Deps.AddLine(r.Context(), app.AddLineInput{
		TenantID: h.tenant(r), CartID: id, VariantID: body.VariantID,
		Qty: body.Qty, MaxQty: body.MaxQty, Notes: body.Notes,
		Addons: body.Addons, ReplacementPref: body.ReplacementPref,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, out)
}

func (h *Handler) updateQty(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	lineID, err := parseUUIDParam(r, "lineId")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		Qty int `json:"qty"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	out, err := h.Deps.UpdateQty(r.Context(), app.UpdateQtyInput{
		TenantID: h.tenant(r), CartID: id, LineID: lineID, Qty: body.Qty,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, out)
}

func (h *Handler) removeLine(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	lineID, err := parseUUIDParam(r, "lineId")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	out, err := h.Deps.RemoveLine(r.Context(), app.RemoveLineInput{
		TenantID: h.tenant(r), CartID: id, LineID: lineID,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, out)
}

func (h *Handler) applyCoupon(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		Code string `json:"code"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	out, err := h.Deps.ApplyCoupon(r.Context(), app.ApplyCouponInput{
		TenantID: h.tenant(r), CartID: id, Code: body.Code,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, out)
}

func (h *Handler) removeCoupon(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	code := r.PathValue("code")
	out, err := h.Deps.RemoveCoupon(r.Context(), app.RemoveCouponInput{
		TenantID: h.tenant(r), CartID: id, Code: code,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, out)
}

func (h *Handler) refreshQuote(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		SoftReserve bool `json:"softReserve"`
	}
	_ = decodeJSON(r, &body)
	out, err := h.Deps.RefreshQuote(r.Context(), app.RefreshQuoteInput{
		TenantID: h.tenant(r), CartID: id, SoftReserve: body.SoftReserve,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, out)
}

func (h *Handler) mergeCarts(w http.ResponseWriter, r *http.Request) {
	var body struct {
		GuestToken  string             `json:"guestToken"`
		PrincipalID *uuid.UUID         `json:"principalId"`
		Policy      domain.MergePolicy `json:"policy"`
		CityID      *uuid.UUID         `json:"cityId"`
		Currency    string             `json:"currency"`
	}
	_ = decodeJSON(r, &body)
	guest := body.GuestToken
	if guest == "" {
		guest = h.guest(r)
	}
	principal := uuid.Nil
	if body.PrincipalID != nil {
		principal = *body.PrincipalID
	} else if p := h.principalPtr(r); p != nil {
		principal = *p
	}
	out, err := h.Deps.MergeCarts(r.Context(), app.MergeCartsInput{
		TenantID: h.tenant(r), GuestToken: guest, PrincipalID: principal,
		Policy: body.Policy, CityID: body.CityID, Currency: body.Currency,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, out)
}

func (h *Handler) softReserve(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		IdempotencyKey string `json:"idempotencyKey"`
	}
	_ = decodeJSON(r, &body)
	key := body.IdempotencyKey
	if key == "" {
		key = r.Header.Get("Idempotency-Key")
	}
	out, err := h.Deps.SoftReserveLines(r.Context(), app.SoftReserveLinesInput{
		TenantID: h.tenant(r), CartID: id, IdempotencyKey: key,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, out)
}

func (h *Handler) abandon(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	out, err := h.Deps.MarkAbandoned(r.Context(), app.MarkAbandonedInput{
		TenantID: h.tenant(r), CartID: id,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, out)
}

func (h *Handler) recover(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	out, err := h.Deps.Recover(r.Context(), app.RecoverInput{
		TenantID: h.tenant(r), CartID: id,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, out)
}

func (h *Handler) recommendations(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	out, err := h.Deps.Recommendations(r.Context(), app.RecommendationsInput{
		TenantID: h.tenant(r), CartID: id, Limit: limit,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, out)
}

func (h *Handler) saveCart(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		Name string `json:"name"`
	}
	_ = decodeJSON(r, &body)
	principal := h.principalPtr(r)
	if principal == nil {
		writeErr(w, r, domain.ErrUnauthorized)
		return
	}
	out, err := h.Deps.SaveCart(r.Context(), app.SaveCartInput{
		TenantID: h.tenant(r), CartID: id, PrincipalID: *principal, Name: body.Name,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, out)
}
