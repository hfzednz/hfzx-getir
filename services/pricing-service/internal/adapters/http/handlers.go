package httpadapter

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/pricing-service/internal/app"
	"github.com/nexora/pricing-service/internal/domain"
	"github.com/nexora/pricing-service/internal/ratelimit"
)

// Handler serves pricing REST endpoints.
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
	const base = "/v1/pricing"

	mux.HandleFunc("GET /health", h.health)
	mux.HandleFunc("GET /ready", h.ready)
	mux.HandleFunc("GET "+base+"/health", h.health)
	mux.HandleFunc("GET "+base+"/ready", h.ready)

	mux.HandleFunc("POST "+base+"/books", tenant(h.upsertBook))
	mux.HandleFunc("GET "+base+"/books", tenant(h.listBooks))
	mux.HandleFunc("POST "+base+"/prices", tenant(h.upsertPrice))
	mux.HandleFunc("GET "+base+"/prices", tenant(h.getPrice))
	mux.HandleFunc("POST "+base+"/quote", tenant(h.quoteCart))
	mux.HandleFunc("POST "+base+"/simulate", tenant(h.simulateQuote))
	mux.HandleFunc("POST "+base+"/tax/calculate", tenant(h.taxCalculate))
	mux.HandleFunc("POST "+base+"/dynamic/apply", tenant(h.applyDynamic))
	mux.HandleFunc("POST "+base+"/tax-rules", tenant(h.upsertTaxRule))
	mux.HandleFunc("POST "+base+"/dynamic-rules", tenant(h.upsertDynamicRule))
	mux.HandleFunc("GET "+base+"/admin", tenant(h.adminList))
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

func bookDTO(b domain.PriceBook) map[string]any {
	return map[string]any{
		"id": b.ID, "tenantId": b.TenantID, "name": b.Name,
		"currency": b.Currency, "active": b.Active,
		"createdAt": b.CreatedAt, "updatedAt": b.UpdatedAt,
	}
}

func entryDTO(e domain.PriceEntry) map[string]any {
	return map[string]any{
		"id": e.ID, "tenantId": e.TenantID, "priceBookId": e.PriceBookID,
		"variantId": e.VariantID, "scope": e.Scope, "scopeId": e.ScopeID,
		"amountMinor": e.AmountMinor, "currency": e.Currency,
		"validFrom": e.ValidFrom, "validTo": e.ValidTo,
		"createdAt": e.CreatedAt, "updatedAt": e.UpdatedAt,
	}
}

func quoteDTO(q domain.Quote) map[string]any {
	lines := make([]map[string]any, 0, len(q.Lines))
	for _, ln := range q.Lines {
		lines = append(lines, map[string]any{
			"variantId": ln.VariantID, "qty": ln.Qty,
			"unitPriceMinor": ln.UnitPriceMinor, "baseUnitMinor": ln.BaseUnitMinor,
			"dynamicAdjMinor": ln.DynamicAdjMinor, "lineSubtotalMinor": ln.LineSubtotalMinor,
			"discountMinor": ln.DiscountMinor, "lineTotalMinor": ln.LineTotalMinor,
			"resolvedScope": ln.ResolvedScope, "priceEntryId": ln.PriceEntryID,
		})
	}
	promos := make([]map[string]any, 0, len(q.Promos))
	for _, p := range q.Promos {
		promos = append(promos, map[string]any{
			"promotionId": p.PromotionID, "code": p.Code,
			"discountMinor": p.DiscountMinor, "description": p.Description,
		})
	}
	return map[string]any{
		"id": q.ID, "tenantId": q.TenantID, "cartId": q.CartID, "currency": q.Currency,
		"lines": lines, "promos": promos,
		"subtotalMinor": q.SubtotalMinor, "discountMinor": q.DiscountMinor,
		"taxMinor": q.TaxMinor, "deliveryMinor": q.DeliveryMinor,
		"serviceMinor": q.ServiceMinor, "packagingMinor": q.PackagingMinor,
		"tipMinor": q.TipMinor, "totalMinor": q.TotalMinor,
		"taxRuleCode": q.TaxRuleCode, "simulated": q.Simulated, "quotedAt": q.QuotedAt,
	}
}

func (h *Handler) upsertBook(w http.ResponseWriter, r *http.Request) {
	var body struct {
		BookID   *uuid.UUID `json:"bookId"`
		Name     string     `json:"name"`
		Currency string     `json:"currency"`
		Active   *bool      `json:"active"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	book, err := h.Deps.UpsertPriceBook(r.Context(), app.UpsertPriceBookInput{
		TenantID: h.tenantID(r), BookID: body.BookID, Name: body.Name,
		Currency: body.Currency, Active: body.Active,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	if body.BookID == nil {
		writeCreated(w, bookDTO(book))
		return
	}
	writeOK(w, bookDTO(book))
}

func (h *Handler) listBooks(w http.ResponseWriter, r *http.Request) {
	list, err := h.Deps.AdminList(r.Context(), app.AdminListInput{TenantID: h.tenantID(r)})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	out := make([]map[string]any, 0, len(list.Books))
	for _, b := range list.Books {
		out = append(out, bookDTO(b))
	}
	writeOK(w, map[string]any{"books": out})
}

func (h *Handler) upsertPrice(w http.ResponseWriter, r *http.Request) {
	var body struct {
		EntryID     *uuid.UUID         `json:"entryId"`
		PriceBookID uuid.UUID          `json:"priceBookId"`
		VariantID   uuid.UUID          `json:"variantId"`
		Scope       domain.PriceScope  `json:"scope"`
		ScopeID     *uuid.UUID         `json:"scopeId"`
		AmountMinor int64              `json:"amountMinor"`
		Currency    string             `json:"currency"`
		ValidFrom   *time.Time         `json:"validFrom"`
		ValidTo     *time.Time         `json:"validTo"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	entry, err := h.Deps.UpsertPrice(r.Context(), app.UpsertPriceInput{
		TenantID: h.tenantID(r), EntryID: body.EntryID, PriceBookID: body.PriceBookID,
		VariantID: body.VariantID, Scope: body.Scope, ScopeID: body.ScopeID,
		AmountMinor: body.AmountMinor, Currency: body.Currency,
		ValidFrom: body.ValidFrom, ValidTo: body.ValidTo,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	if body.EntryID == nil {
		writeCreated(w, entryDTO(entry))
		return
	}
	writeOK(w, entryDTO(entry))
}

func (h *Handler) getPrice(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	variantID, err := uuid.Parse(q.Get("variantId"))
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	parseOpt := func(key string) *uuid.UUID {
		raw := q.Get(key)
		if raw == "" {
			return nil
		}
		id, err := uuid.Parse(raw)
		if err != nil {
			return nil
		}
		return &id
	}
	res, err := h.Deps.GetPrice(r.Context(), app.GetPriceInput{
		TenantID: h.tenantID(r), VariantID: variantID, Currency: q.Get("currency"),
		RegionID: parseOpt("regionId"), WarehouseID: parseOpt("warehouseId"),
		CustomerID: parseOpt("customerId"), VIPID: parseOpt("vipId"), CorporateID: parseOpt("corporateId"),
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{
		"amountMinor": res.AmountMinor, "currency": res.Currency,
		"scope": res.Scope, "entry": entryDTO(res.Entry),
	})
}

func parseQuoteBody(r *http.Request) (app.QuoteCartInput, error) {
	var body struct {
		CartID         *uuid.UUID `json:"cartId"`
		Currency       string     `json:"currency"`
		RegionID       *uuid.UUID `json:"regionId"`
		WarehouseID    *uuid.UUID `json:"warehouseId"`
		CustomerID     *uuid.UUID `json:"customerId"`
		VIPID          *uuid.UUID `json:"vipId"`
		CorporateID    *uuid.UUID `json:"corporateId"`
		CouponCodes    []string   `json:"couponCodes"`
		DeliveryMinor  int64      `json:"deliveryMinor"`
		ServiceMinor   int64      `json:"serviceMinor"`
		PackagingMinor int64      `json:"packagingMinor"`
		TipMinor       int64      `json:"tipMinor"`
		Lines          []struct {
			VariantID uuid.UUID `json:"variantId"`
			Qty       int       `json:"qty"`
		} `json:"lines"`
	}
	if err := decodeJSON(r, &body); err != nil {
		return app.QuoteCartInput{}, domain.ErrInvalidArgument
	}
	lines := make([]app.QuoteLineInput, 0, len(body.Lines))
	for _, ln := range body.Lines {
		lines = append(lines, app.QuoteLineInput{VariantID: ln.VariantID, Qty: ln.Qty})
	}
	return app.QuoteCartInput{
		CartID: body.CartID, Currency: body.Currency,
		RegionID: body.RegionID, WarehouseID: body.WarehouseID,
		CustomerID: body.CustomerID, VIPID: body.VIPID, CorporateID: body.CorporateID,
		CouponCodes: body.CouponCodes, Lines: lines,
		DeliveryMinor: body.DeliveryMinor, ServiceMinor: body.ServiceMinor,
		PackagingMinor: body.PackagingMinor, TipMinor: body.TipMinor,
	}, nil
}

func (h *Handler) quoteCart(w http.ResponseWriter, r *http.Request) {
	in, err := parseQuoteBody(r)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	in.TenantID = h.tenantID(r)
	q, err := h.Deps.QuoteCart(r.Context(), in)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, quoteDTO(q))
}

func (h *Handler) simulateQuote(w http.ResponseWriter, r *http.Request) {
	in, err := parseQuoteBody(r)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	in.TenantID = h.tenantID(r)
	q, err := h.Deps.SimulateQuote(r.Context(), in)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, quoteDTO(q))
}

func (h *Handler) taxCalculate(w http.ResponseWriter, r *http.Request) {
	var body struct {
		RegionID  *uuid.UUID `json:"regionId"`
		BaseMinor int64      `json:"baseMinor"`
		Currency  string     `json:"currency"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	res, err := h.Deps.TaxCalculate(r.Context(), app.TaxCalculateInput{
		TenantID: h.tenantID(r), RegionID: body.RegionID,
		BaseMinor: body.BaseMinor, Currency: body.Currency,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{
		"taxMinor": res.TaxMinor, "rateBps": res.RateBps,
		"ruleCode": res.RuleCode, "inclusive": res.Inclusive,
		"taxableBase": res.TaxableBase,
	})
}

func (h *Handler) applyDynamic(w http.ResponseWriter, r *http.Request) {
	var body struct {
		VariantID   uuid.UUID  `json:"variantId"`
		UnitMinor   int64      `json:"unitMinor"`
		Currency    string     `json:"currency"`
		WarehouseID *uuid.UUID `json:"warehouseId"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	res, err := h.Deps.ApplyDynamic(r.Context(), app.ApplyDynamicInput{
		TenantID: h.tenantID(r), VariantID: body.VariantID,
		UnitMinor: body.UnitMinor, Currency: body.Currency, WarehouseID: body.WarehouseID,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{
		"unitMinor": res.UnitMinor, "baseUnitMinor": res.BaseUnitMinor,
		"adjustmentMinor": res.AdjustmentMinor, "appliedRules": res.AppliedRules,
	})
}

func (h *Handler) upsertTaxRule(w http.ResponseWriter, r *http.Request) {
	var body struct {
		RuleID    *uuid.UUID `json:"ruleId"`
		Code      string     `json:"code"`
		Name      string     `json:"name"`
		RateBps   int        `json:"rateBps"`
		Inclusive bool       `json:"inclusive"`
		RegionID  *uuid.UUID `json:"regionId"`
		Active    *bool      `json:"active"`
		Priority  int        `json:"priority"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	rule, err := h.Deps.UpsertTaxRule(r.Context(), app.UpsertTaxRuleInput{
		TenantID: h.tenantID(r), RuleID: body.RuleID, Code: body.Code, Name: body.Name,
		RateBps: body.RateBps, Inclusive: body.Inclusive, RegionID: body.RegionID,
		Active: body.Active, Priority: body.Priority,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	dto := map[string]any{
		"id": rule.ID, "code": rule.Code, "name": rule.Name,
		"rateBps": rule.RateBps, "inclusive": rule.Inclusive,
		"regionId": rule.RegionID, "active": rule.Active, "priority": rule.Priority,
	}
	if body.RuleID == nil {
		writeCreated(w, dto)
		return
	}
	writeOK(w, dto)
}

func (h *Handler) upsertDynamicRule(w http.ResponseWriter, r *http.Request) {
	var body struct {
		RuleID             *uuid.UUID            `json:"ruleId"`
		Code               string                `json:"code"`
		Kind               domain.DynamicKind    `json:"kind"`
		Trigger            domain.DynamicTrigger `json:"trigger"`
		AdjustmentBps      int                   `json:"adjustmentBps"`
		AdjustmentMinor    int64                 `json:"adjustmentMinor"`
		StartHour          int                   `json:"startHour"`
		EndHour            int                   `json:"endHour"`
		InventoryThreshold int                   `json:"inventoryThreshold"`
		Active             *bool                 `json:"active"`
		Priority           int                   `json:"priority"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	rule, err := h.Deps.UpsertDynamicRule(r.Context(), app.UpsertDynamicRuleInput{
		TenantID: h.tenantID(r), RuleID: body.RuleID, Code: body.Code,
		Kind: body.Kind, Trigger: body.Trigger,
		AdjustmentBps: body.AdjustmentBps, AdjustmentMinor: body.AdjustmentMinor,
		StartHour: body.StartHour, EndHour: body.EndHour,
		InventoryThreshold: body.InventoryThreshold, Active: body.Active, Priority: body.Priority,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	dto := map[string]any{
		"id": rule.ID, "code": rule.Code, "kind": rule.Kind, "trigger": rule.Trigger,
		"adjustmentBps": rule.AdjustmentBps, "adjustmentMinor": rule.AdjustmentMinor,
		"startHour": rule.StartHour, "endHour": rule.EndHour,
		"inventoryThreshold": rule.InventoryThreshold, "active": rule.Active, "priority": rule.Priority,
	}
	if body.RuleID == nil {
		writeCreated(w, dto)
		return
	}
	writeOK(w, dto)
}

func (h *Handler) adminList(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	var bookID, variantID *uuid.UUID
	if raw := q.Get("bookId"); raw != "" {
		if id, err := uuid.Parse(raw); err == nil {
			bookID = &id
		}
	}
	if raw := q.Get("variantId"); raw != "" {
		if id, err := uuid.Parse(raw); err == nil {
			variantID = &id
		}
	}
	limit := 50
	if raw := q.Get("limit"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil {
			limit = n
		}
	}
	list, err := h.Deps.AdminList(r.Context(), app.AdminListInput{
		TenantID: h.tenantID(r), BookID: bookID, VariantID: variantID, AuditLimit: limit,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	books := make([]map[string]any, 0, len(list.Books))
	for _, b := range list.Books {
		books = append(books, bookDTO(b))
	}
	entries := make([]map[string]any, 0, len(list.Entries))
	for _, e := range list.Entries {
		entries = append(entries, entryDTO(e))
	}
	writeOK(w, map[string]any{
		"books": books, "entries": entries,
		"taxRules": list.TaxRules, "dynamicRules": list.Dynamics,
		"audits": list.Audits,
	})
}

func (h *Handler) publishOutbox(w http.ResponseWriter, r *http.Request) {
	n, err := h.Deps.PublishPending(r.Context(), 100)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"published": n})
}
