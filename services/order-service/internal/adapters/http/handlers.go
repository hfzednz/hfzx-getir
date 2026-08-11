package httpadapter

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/order-service/internal/app"
	"github.com/nexora/order-service/internal/app/ports"
	"github.com/nexora/order-service/internal/domain"
	"github.com/nexora/order-service/internal/ratelimit"
)

// Handler serves order REST endpoints.
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
	const base = "/v1/orders"

	mux.HandleFunc("GET /health", h.health)
	mux.HandleFunc("GET /ready", h.ready)
	mux.HandleFunc("GET "+base+"/health", h.health)
	mux.HandleFunc("GET "+base+"/ready", h.ready)

	mux.HandleFunc("POST "+base, tenant(h.createOrder))
	mux.HandleFunc("GET "+base, tenant(h.listOrders))
	mux.HandleFunc("GET "+base+"/search", tenant(h.searchOrders))
	mux.HandleFunc("GET "+base+"/{id}", tenant(h.getOrder))
	mux.HandleFunc("POST "+base+"/{id}/place", tenant(h.placeOrder))
	mux.HandleFunc("POST "+base+"/{id}/cancel", tenant(h.cancelOrder))
	mux.HandleFunc("POST "+base+"/{id}/returns", tenant(h.requestReturn))
	mux.HandleFunc("POST "+base+"/{id}/refunds", tenant(h.requestRefund))
	mux.HandleFunc("POST "+base+"/{id}/events/warehouse", tenant(h.applyWarehouseEvent))
	mux.HandleFunc("POST "+base+"/{id}/events/dispatch", tenant(h.applyDispatchEvent))
	mux.HandleFunc("GET "+base+"/{id}/timeline", tenant(h.timeline))
	mux.HandleFunc("POST "+base+"/{id}/admin/priority", tenant(h.setPriority))
	mux.HandleFunc("POST "+base+"/{id}/admin/split", tenant(h.splitOrder))
	mux.HandleFunc("POST "+base+"/{id}/admin/intervene", tenant(h.intervene))
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

func parseUUIDParam(r *http.Request, name string) (uuid.UUID, error) {
	return uuid.Parse(r.PathValue(name))
}

func idemKey(r *http.Request, bodyKey string) string {
	if v := r.Header.Get("Idempotency-Key"); v != "" {
		return v
	}
	return bodyKey
}

func orderDTO(o domain.Order) map[string]any {
	lines := make([]map[string]any, 0, len(o.Lines))
	for _, l := range o.Lines {
		lines = append(lines, map[string]any{
			"id":             l.ID,
			"variantId":      l.VariantID,
			"skuCode":        l.SKUCode,
			"titleSnapshot":  l.TitleSnapshot,
			"qty":            l.Qty,
			"unitPriceMinor": l.UnitPriceMinor,
			"discountsMinor": l.DiscountsMinor,
			"taxMinor":       l.TaxMinor,
			"lineTotalMinor": l.LineTotalMinor,
			"warehouseId":    l.WarehouseID,
			"sortOrder":      l.SortOrder,
		})
	}
	return map[string]any{
		"id":                  o.ID,
		"tenantId":            o.TenantID,
		"customerPrincipalId": o.CustomerPrincipalID,
		"status":              o.Status,
		"type":                o.Type,
		"currency":            o.Currency,
		"subtotalMinor":       o.SubtotalMinor,
		"discountMinor":       o.DiscountMinor,
		"taxMinor":            o.TaxMinor,
		"shippingMinor":       o.ShippingMinor,
		"tipMinor":            o.TipMinor,
		"totalMinor":          o.TotalMinor,
		"addressSnapshot":     o.AddressSnapshot,
		"notes":               o.Notes,
		"gift":                o.Gift,
		"priority":            o.Priority,
		"warehouseIds":        o.WarehouseIDs,
		"version":             o.Version,
		"idempotencyKey":      o.IdempotencyKey,
		"paymentIntentRef":    o.PaymentIntentRef,
		"reservationRef":      o.ReservationRef,
		"courierRef":          o.CourierRef,
		"lines":               lines,
		"createdAt":           o.CreatedAt,
		"updatedAt":           o.UpdatedAt,
		"placedAt":            o.PlacedAt,
		"cancelledAt":         o.CancelledAt,
		"completedAt":         o.CompletedAt,
	}
}

func (h *Handler) createOrder(w http.ResponseWriter, r *http.Request) {
	var body struct {
		CustomerPrincipalID uuid.UUID              `json:"customerPrincipalId"`
		Type                domain.OrderType       `json:"type"`
		Currency            string                 `json:"currency"`
		IdempotencyKey      string                 `json:"idempotencyKey"`
		FromCheckout        bool                   `json:"fromCheckout"`
		AddressSnapshot     map[string]any         `json:"addressSnapshot"`
		Notes               string                 `json:"notes"`
		Gift                map[string]any         `json:"gift"`
		Priority            int                    `json:"priority"`
		WarehouseIDs        []uuid.UUID            `json:"warehouseIds"`
		DiscountMinor       int64                  `json:"discountMinor"`
		ShippingMinor       int64                  `json:"shippingMinor"`
		TipMinor            int64                  `json:"tipMinor"`
		Lines               []app.CreateLineInput  `json:"lines"`
		Metadata            map[string]any         `json:"metadata"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	in := app.CreateDraftInput{
		TenantID: h.tenant(r), CustomerPrincipalID: body.CustomerPrincipalID,
		Type: body.Type, Currency: body.Currency,
		IdempotencyKey: idemKey(r, body.IdempotencyKey),
		AddressSnapshot: body.AddressSnapshot, Notes: body.Notes, Gift: body.Gift,
		Priority: body.Priority, WarehouseIDs: body.WarehouseIDs,
		DiscountMinor: body.DiscountMinor, ShippingMinor: body.ShippingMinor,
		TipMinor: body.TipMinor, Lines: body.Lines, Metadata: body.Metadata,
	}
	var (
		out domain.Order
		err error
	)
	if body.FromCheckout {
		out, err = h.Deps.CreateFromCheckout(r.Context(), app.CreateFromCheckoutInput{CreateDraftInput: in})
	} else {
		out, err = h.Deps.CreateDraft(r.Context(), in)
	}
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, orderDTO(out))
}

func (h *Handler) listOrders(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	var status *domain.OrderStatus
	if s := r.URL.Query().Get("status"); s != "" {
		st := domain.OrderStatus(s)
		status = &st
	}
	items, total, err := h.Deps.ListOrders(r.Context(), h.tenant(r), status, limit, offset)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	dtos := make([]map[string]any, 0, len(items))
	for _, o := range items {
		dtos = append(dtos, orderDTO(o))
	}
	writeOK(w, map[string]any{"items": dtos, "total": total})
}

func (h *Handler) getOrder(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	o, err := h.Deps.GetOrder(r.Context(), h.tenant(r), id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, orderDTO(o))
}

func (h *Handler) placeOrder(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		IdempotencyKey string `json:"idempotencyKey"`
	}
	_ = decodeJSON(r, &body)
	o, err := h.Deps.PlaceOrder(r.Context(), app.PlaceOrderInput{
		TenantID: h.tenant(r), OrderID: id, IdempotencyKey: idemKey(r, body.IdempotencyKey),
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, orderDTO(o))
}

func (h *Handler) cancelOrder(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		Reason         string `json:"reason"`
		IdempotencyKey string `json:"idempotencyKey"`
	}
	_ = decodeJSON(r, &body)
	o, err := h.Deps.CancelOrder(r.Context(), app.CancelOrderInput{
		TenantID: h.tenant(r), OrderID: id, Reason: body.Reason,
		IdempotencyKey: idemKey(r, body.IdempotencyKey),
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, orderDTO(o))
}

func (h *Handler) requestReturn(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		Reason      string                 `json:"reason"`
		Notes       string                 `json:"notes"`
		Disposition domain.ReturnDisposition `json:"disposition"`
		Lines       []app.ReturnLineInput  `json:"lines"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	ret, err := h.Deps.RequestReturn(r.Context(), app.RequestReturnInput{
		TenantID: h.tenant(r), OrderID: id, Reason: body.Reason,
		Notes: body.Notes, Disposition: body.Disposition, Lines: body.Lines,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, ret)
}

func (h *Handler) requestRefund(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		AmountMinor int64               `json:"amountMinor"`
		Currency    string              `json:"currency"`
		Method      domain.RefundMethod `json:"method"`
		Reason      string              `json:"reason"`
		ReturnID    *uuid.UUID          `json:"returnId"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	ref, err := h.Deps.RequestRefund(r.Context(), app.RequestRefundInput{
		TenantID: h.tenant(r), OrderID: id, AmountMinor: body.AmountMinor,
		Currency: body.Currency, Method: body.Method, Reason: body.Reason, ReturnID: body.ReturnID,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, ref)
}

func (h *Handler) applyWarehouseEvent(w http.ResponseWriter, r *http.Request) {
	h.applyEvent(w, r, true)
}

func (h *Handler) applyDispatchEvent(w http.ResponseWriter, r *http.Request) {
	h.applyEvent(w, r, false)
}

func (h *Handler) applyEvent(w http.ResponseWriter, r *http.Request, warehouse bool) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		EventType  string         `json:"eventType"`
		Payload    map[string]any `json:"payload"`
		CourierRef string         `json:"courierRef"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	in := app.ApplyLifecycleEventInput{
		TenantID: h.tenant(r), OrderID: id, EventType: body.EventType,
		Payload: body.Payload, CourierRef: body.CourierRef,
	}
	var o domain.Order
	if warehouse {
		o, err = h.Deps.ApplyWarehouseEvent(r.Context(), in)
	} else {
		o, err = h.Deps.ApplyDispatchEvent(r.Context(), in)
	}
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, orderDTO(o))
}

func (h *Handler) timeline(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	evs, err := h.Deps.Timeline(r.Context(), h.tenant(r), id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"events": evs})
}

func (h *Handler) setPriority(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		Priority int `json:"priority"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	o, err := h.Deps.SetPriority(r.Context(), app.SetPriorityInput{
		TenantID: h.tenant(r), OrderID: id, Priority: body.Priority,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, orderDTO(o))
}

func (h *Handler) splitOrder(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		Splits []app.SplitUnitInput `json:"splits"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	fs, err := h.Deps.SplitOrder(r.Context(), app.SplitOrderInput{
		TenantID: h.tenant(r), OrderID: id, Splits: body.Splits,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"fulfillments": fs})
}

func (h *Handler) intervene(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		NextStatus domain.OrderStatus `json:"nextStatus"`
		Reason     string             `json:"reason"`
		Force      bool               `json:"force"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	o, err := h.Deps.InterveneStatus(r.Context(), app.InterveneStatusInput{
		TenantID: h.tenant(r), OrderID: id, NextStatus: body.NextStatus,
		Reason: body.Reason, Force: body.Force,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, orderDTO(o))
}

func (h *Handler) searchOrders(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	q := ports.SearchQuery{
		TenantID: h.tenant(r),
		Query:    r.URL.Query().Get("q"),
		Limit:    limit,
		Offset:   offset,
	}
	if s := r.URL.Query().Get("status"); s != "" {
		st := domain.OrderStatus(s)
		q.Status = &st
	}
	res, err := h.Deps.SearchOrders(r.Context(), q)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, res)
}

func (h *Handler) publishOutbox(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	n, err := h.Deps.PublishPending(r.Context(), limit)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"published": n})
}
