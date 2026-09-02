package httpadapter

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/promotion-service/internal/app"
	"github.com/nexora/promotion-service/internal/domain"
	"github.com/nexora/promotion-service/internal/ratelimit"
)

// Handler serves promotion REST endpoints.
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
	const base = "/v1/promo"

	mux.HandleFunc("GET /health", h.health)
	mux.HandleFunc("GET /ready", h.ready)
	mux.HandleFunc("GET "+base+"/health", h.health)
	mux.HandleFunc("GET "+base+"/ready", h.ready)

	mux.HandleFunc("POST "+base+"/campaigns", tenant(h.createCampaign))
	mux.HandleFunc("GET "+base+"/campaigns", tenant(h.listCampaigns))
	mux.HandleFunc("GET "+base+"/campaigns/{id}", tenant(h.getCampaign))
	mux.HandleFunc("POST "+base+"/campaigns/{id}/activate", tenant(h.activateCampaign))
	mux.HandleFunc("POST "+base+"/campaigns/{id}/pause", tenant(h.pauseCampaign))
	mux.HandleFunc("POST "+base+"/campaigns/{id}/expire", tenant(h.expireCampaign))

	mux.HandleFunc("POST "+base+"/promotions", tenant(h.createPromotion))
	mux.HandleFunc("GET "+base+"/campaigns/{id}/promotions", tenant(h.listPromotions))

	mux.HandleFunc("POST "+base+"/coupons", tenant(h.generateCoupon))
	mux.HandleFunc("GET "+base+"/coupons", tenant(h.listCoupons))
	mux.HandleFunc("GET "+base+"/coupons/{code}", tenant(h.getCoupon))
	mux.HandleFunc("PATCH "+base+"/coupons/{code}", tenant(h.updateCoupon))
	mux.HandleFunc("POST "+base+"/coupons/redeem", tenant(h.redeemCoupon))

	mux.HandleFunc("POST "+base+"/vouchers", tenant(h.issueVoucher))
	mux.HandleFunc("POST "+base+"/vouchers/redeem", tenant(h.redeemVoucher))

	mux.HandleFunc("POST "+base+"/evaluate", tenant(h.evaluate))
	mux.HandleFunc("POST "+base+"/simulate", tenant(h.simulate))
	mux.HandleFunc("GET "+base+"/simulations", tenant(h.listSimulations))
	mux.HandleFunc("GET "+base+"/admin/overview", tenant(h.adminOverview))
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

func (h *Handler) createCampaign(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name        string     `json:"name"`
		Description string     `json:"description"`
		StartsAt    *time.Time `json:"startsAt"`
		EndsAt      *time.Time `json:"endsAt"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	c, err := h.Deps.CreateCampaign(r.Context(), app.CreateCampaignInput{
		TenantID: h.tenantID(r), Name: body.Name, Description: body.Description,
		StartsAt: body.StartsAt, EndsAt: body.EndsAt,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, campaignDTO(c))
}

func (h *Handler) listCampaigns(w http.ResponseWriter, r *http.Request) {
	var status *domain.CampaignStatus
	if s := r.URL.Query().Get("status"); s != "" {
		st := domain.CampaignStatus(s)
		status = &st
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	list, err := h.Deps.ListCampaigns(r.Context(), h.tenantID(r), status, limit, offset)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	out := make([]any, 0, len(list))
	for _, c := range list {
		out = append(out, campaignDTO(c))
	}
	writeOK(w, map[string]any{"items": out})
}

func (h *Handler) getCampaign(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	c, err := h.Deps.GetCampaign(r.Context(), h.tenantID(r), id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, campaignDTO(c))
}

func (h *Handler) activateCampaign(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	c, err := h.Deps.ActivateCampaign(r.Context(), h.tenantID(r), id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, campaignDTO(c))
}

func (h *Handler) pauseCampaign(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	c, err := h.Deps.PauseCampaign(r.Context(), h.tenantID(r), id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, campaignDTO(c))
}

func (h *Handler) expireCampaign(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	c, err := h.Deps.ExpireCampaign(r.Context(), h.tenantID(r), id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, campaignDTO(c))
}

func (h *Handler) createPromotion(w http.ResponseWriter, r *http.Request) {
	var body struct {
		CampaignID       string   `json:"campaignId"`
		Name             string   `json:"name"`
		Type             string   `json:"type"`
		PercentOff       int      `json:"percentOff"`
		FixedOffMinor    int64    `json:"fixedOffMinor"`
		BuyQty           int      `json:"buyQty"`
		GetQty           int      `json:"getQty"`
		ThresholdMinor   int64    `json:"thresholdMinor"`
		GiftVariantID    string   `json:"giftVariantId"`
		MaxDiscountMinor int64    `json:"maxDiscountMinor"`
		Priority         int      `json:"priority"`
		StackGroup       string   `json:"stackGroup"`
		Stackable        bool     `json:"stackable"`
		ExcludePromoIDs  []string `json:"excludePromotionIds"`
		VariantIDs       []string `json:"variantIds"`
		CategoryIDs      []string `json:"categoryIds"`
		BrandIDs         []string `json:"brandIds"`
		SegmentIDs       []string `json:"segmentIds"`
		GlobalLimit      int      `json:"globalLimit"`
		PerUserLimit     int      `json:"perUserLimit"`
		PerOrderLimit    int      `json:"perOrderLimit"`
		PerDeviceLimit   int      `json:"perDeviceLimit"`
		MinQty           int      `json:"minQty"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	cid, err := uuid.Parse(body.CampaignID)
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var excl []uuid.UUID
	for _, s := range body.ExcludePromoIDs {
		id, err := uuid.Parse(s)
		if err != nil {
			writeErr(w, r, domain.ErrInvalidArgument)
			return
		}
		excl = append(excl, id)
	}
	res, err := h.Deps.CreatePromotion(r.Context(), app.CreatePromotionInput{
		TenantID: h.tenantID(r), CampaignID: cid, Name: body.Name,
		Type: domain.PromotionType(body.Type), PercentOff: body.PercentOff,
		FixedOffMinor: body.FixedOffMinor, BuyQty: body.BuyQty, GetQty: body.GetQty,
		ThresholdMinor: body.ThresholdMinor, GiftVariantID: body.GiftVariantID,
		MaxDiscountMinor: body.MaxDiscountMinor, Priority: body.Priority,
		Rule: &app.CreateRuleInput{
			Priority: body.Priority, StackGroup: body.StackGroup, Stackable: body.Stackable,
			ExcludePromotionIDs: excl, VariantIDs: body.VariantIDs, CategoryIDs: body.CategoryIDs,
			BrandIDs: body.BrandIDs, SegmentIDs: body.SegmentIDs,
			GlobalLimit: body.GlobalLimit, PerUserLimit: body.PerUserLimit,
			PerOrderLimit: body.PerOrderLimit, PerDeviceLimit: body.PerDeviceLimit,
			MinQty: body.MinQty,
		},
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, map[string]any{
		"promotion": promotionDTO(res.Promotion),
		"rule":      ruleDTO(res.Rule),
	})
}

func (h *Handler) listPromotions(w http.ResponseWriter, r *http.Request) {
	cid, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	list, err := h.Deps.ListPromotionsByCampaign(r.Context(), h.tenantID(r), cid)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	out := make([]any, 0, len(list))
	for _, p := range list {
		out = append(out, promotionDTO(p))
	}
	writeOK(w, map[string]any{"items": out})
}

func (h *Handler) generateCoupon(w http.ResponseWriter, r *http.Request) {
	var body struct {
		PromotionID    string     `json:"promotionId"`
		Code           string     `json:"code"`
		Kind           string     `json:"kind"`
		MaxRedemptions int        `json:"maxRedemptions"`
		PrincipalID    string     `json:"principalId"`
		StartsAt       *time.Time `json:"startsAt"`
		EndsAt         *time.Time `json:"endsAt"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	pid, err := uuid.Parse(body.PromotionID)
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var principal *uuid.UUID
	if body.PrincipalID != "" {
		p, err := uuid.Parse(body.PrincipalID)
		if err != nil {
			writeErr(w, r, domain.ErrInvalidArgument)
			return
		}
		principal = &p
	}
	c, err := h.Deps.GenerateCoupon(r.Context(), app.GenerateCouponInput{
		TenantID: h.tenantID(r), PromotionID: pid, Code: body.Code,
		Kind: domain.CouponKind(body.Kind), MaxRedemptions: body.MaxRedemptions,
		PrincipalID: principal, StartsAt: body.StartsAt, EndsAt: body.EndsAt,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, couponDTO(c))
}

func (h *Handler) listCoupons(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	list, err := h.Deps.ListCoupons(r.Context(), h.tenantID(r), limit, offset)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	out := make([]any, 0, len(list))
	for _, c := range list {
		out = append(out, couponDTO(c))
	}
	writeOK(w, map[string]any{"items": out, "total": len(out)})
}

func (h *Handler) getCoupon(w http.ResponseWriter, r *http.Request) {
	c, err := h.Deps.GetCoupon(r.Context(), h.tenantID(r), r.PathValue("code"))
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, couponDTO(c))
}

func (h *Handler) updateCoupon(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Kind           string     `json:"kind"`
		MaxRedemptions *int       `json:"maxRedemptions"`
		StartsAt       *time.Time `json:"startsAt"`
		EndsAt         *time.Time `json:"endsAt"`
		Enabled        *bool      `json:"enabled"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	in := app.UpdateCouponInput{
		TenantID: h.tenantID(r), Code: r.PathValue("code"),
		MaxRedemptions: body.MaxRedemptions, StartsAt: body.StartsAt, EndsAt: body.EndsAt,
		Enabled: body.Enabled,
	}
	if body.Kind != "" {
		k := domain.CouponKind(body.Kind)
		in.Kind = &k
	}
	c, err := h.Deps.UpdateCoupon(r.Context(), in)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, couponDTO(c))
}

func (h *Handler) redeemCoupon(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Code           string `json:"code"`
		PrincipalID    string `json:"principalId"`
		IdempotencyKey string `json:"idempotencyKey"`
		OrderRef       string `json:"orderRef"`
		DiscountMinor  int64  `json:"discountMinor"`
		Currency       string `json:"currency"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	if body.IdempotencyKey == "" {
		body.IdempotencyKey = r.Header.Get("Idempotency-Key")
	}
	pid, _ := uuid.Parse(body.PrincipalID)
	if pid == uuid.Nil {
		pid, _ = PrincipalIDFromContext(r.Context())
	}
	red, err := h.Deps.RedeemCoupon(r.Context(), app.RedeemCouponInput{
		TenantID: h.tenantID(r), Code: body.Code, PrincipalID: pid,
		IdempotencyKey: body.IdempotencyKey, OrderRef: body.OrderRef,
		DiscountMinor: body.DiscountMinor, Currency: body.Currency,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{
		"id": red.ID.String(), "couponId": red.CouponID.String(),
		"discountMinor": red.DiscountMinor, "currency": red.Currency,
		"idempotencyKey": red.IdempotencyKey,
	})
}

func (h *Handler) issueVoucher(w http.ResponseWriter, r *http.Request) {
	var body struct {
		PrincipalID  string     `json:"principalId"`
		PromotionID  string     `json:"promotionId"`
		Code         string     `json:"code"`
		ValueMinor   int64      `json:"valueMinor"`
		Currency     string     `json:"currency"`
		StartsAt     *time.Time `json:"startsAt"`
		EndsAt       *time.Time `json:"endsAt"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	pid, err := uuid.Parse(body.PrincipalID)
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var promoID *uuid.UUID
	if body.PromotionID != "" {
		p, err := uuid.Parse(body.PromotionID)
		if err != nil {
			writeErr(w, r, domain.ErrInvalidArgument)
			return
		}
		promoID = &p
	}
	v, err := h.Deps.IssueVoucher(r.Context(), app.IssueVoucherInput{
		TenantID: h.tenantID(r), PrincipalID: pid, PromotionID: promoID,
		Code: body.Code, ValueMinor: body.ValueMinor, Currency: body.Currency,
		StartsAt: body.StartsAt, EndsAt: body.EndsAt,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, voucherDTO(v))
}

func (h *Handler) redeemVoucher(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Code           string `json:"code"`
		PrincipalID    string `json:"principalId"`
		IdempotencyKey string `json:"idempotencyKey"`
		OrderRef       string `json:"orderRef"`
		AmountMinor    int64  `json:"amountMinor"`
		Currency       string `json:"currency"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	if body.IdempotencyKey == "" {
		body.IdempotencyKey = r.Header.Get("Idempotency-Key")
	}
	pid, _ := uuid.Parse(body.PrincipalID)
	red, err := h.Deps.RedeemVoucher(r.Context(), app.RedeemVoucherInput{
		TenantID: h.tenantID(r), Code: body.Code, PrincipalID: pid,
		IdempotencyKey: body.IdempotencyKey, OrderRef: body.OrderRef,
		AmountMinor: body.AmountMinor, Currency: body.Currency,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{
		"id": red.ID.String(), "voucherId": red.VoucherID.String(),
		"amountMinor": red.AmountMinor, "currency": red.Currency,
	})
}

func (h *Handler) evaluate(w http.ResponseWriter, r *http.Request) {
	eval, err := decodeEvaluate(r)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	eval.TenantID = h.tenantID(r)
	res, err := h.Deps.EvaluateCart(r.Context(), eval)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, evaluateDTO(res))
}

func (h *Handler) simulate(w http.ResponseWriter, r *http.Request) {
	eval, err := decodeEvaluate(r)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	eval.TenantID = h.tenantID(r)
	res, sim, err := h.Deps.SimulateCart(r.Context(), eval)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{
		"simulationId": sim.ID.String(),
		"result":       evaluateDTO(res),
	})
}

func (h *Handler) listSimulations(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	list, err := h.Deps.ListSimulations(r.Context(), h.tenantID(r), limit, offset)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": list})
}

func (h *Handler) adminOverview(w http.ResponseWriter, r *http.Request) {
	ov, err := h.Deps.AdminListOverview(r.Context(), h.tenantID(r))
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, ov)
}

func (h *Handler) publishOutbox(w http.ResponseWriter, r *http.Request) {
	n, err := h.Deps.PublishPendingOutbox(r.Context(), 100)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"published": n})
}

func decodeEvaluate(r *http.Request) (domain.EvaluateContext, error) {
	var body struct {
		PrincipalID   string   `json:"principalId"`
		DeviceID      string   `json:"deviceId"`
		Currency      string   `json:"currency"`
		CouponCodes   []string `json:"couponCodes"`
		ShippingMinor int64    `json:"shippingMinor"`
		SegmentIDs    []string `json:"segmentIds"`
		OrderRef      string   `json:"orderRef"`
		Lines         []struct {
			LineID         string `json:"lineId"`
			VariantID      string `json:"variantId"`
			CategoryID     string `json:"categoryId"`
			BrandID        string `json:"brandId"`
			Quantity       int    `json:"quantity"`
			UnitPriceMinor int64  `json:"unitPriceMinor"`
		} `json:"lines"`
	}
	if err := decodeJSON(r, &body); err != nil {
		return domain.EvaluateContext{}, domain.ErrInvalidArgument
	}
	pid, _ := uuid.Parse(body.PrincipalID)
	lines := make([]domain.CartLine, 0, len(body.Lines))
	for _, l := range body.Lines {
		lines = append(lines, domain.CartLine{
			LineID: l.LineID, VariantID: l.VariantID, CategoryID: l.CategoryID,
			BrandID: l.BrandID, Quantity: l.Quantity, UnitPriceMinor: l.UnitPriceMinor,
		})
	}
	return domain.EvaluateContext{
		PrincipalID: pid, DeviceID: body.DeviceID, Currency: body.Currency,
		CouponCodes: body.CouponCodes, ShippingMinor: body.ShippingMinor,
		SegmentIDs: body.SegmentIDs, OrderRef: body.OrderRef, Lines: lines,
	}, nil
}

func campaignDTO(c domain.Campaign) map[string]any {
	return map[string]any{
		"id": c.ID.String(), "tenantId": c.TenantID.String(),
		"name": c.Name, "description": c.Description, "status": string(c.Status),
		"startsAt": c.StartsAt, "endsAt": c.EndsAt,
		"createdAt": c.CreatedAt, "updatedAt": c.UpdatedAt, "version": c.Version,
	}
}

func promotionDTO(p domain.Promotion) map[string]any {
	return map[string]any{
		"id": p.ID.String(), "campaignId": p.CampaignID.String(),
		"name": p.Name, "type": string(p.Type),
		"percentOff": p.PercentOff, "fixedOffMinor": p.FixedOffMinor,
		"buyQty": p.BuyQty, "getQty": p.GetQty, "thresholdMinor": p.ThresholdMinor,
		"giftVariantId": p.GiftVariantID, "maxDiscountMinor": p.MaxDiscountMinor,
		"priority": p.Priority, "createdAt": p.CreatedAt,
	}
}

func ruleDTO(r domain.Rule) map[string]any {
	excl := make([]string, 0, len(r.ExcludePromotionIDs))
	for _, id := range r.ExcludePromotionIDs {
		excl = append(excl, id.String())
	}
	return map[string]any{
		"id": r.ID.String(), "promotionId": r.PromotionID.String(),
		"priority": r.Priority, "stackGroup": r.StackGroup, "stackable": r.Stackable,
		"excludePromotionIds": excl, "variantIds": r.VariantIDs,
		"categoryIds": r.CategoryIDs, "brandIds": r.BrandIDs, "segmentIds": r.SegmentIDs,
		"globalLimit": r.GlobalLimit, "perUserLimit": r.PerUserLimit,
		"perOrderLimit": r.PerOrderLimit, "perDeviceLimit": r.PerDeviceLimit,
		"minQty": r.MinQty,
	}
}

func couponDTO(c domain.Coupon) map[string]any {
	now := time.Now().UTC()
	enabled := c.IsValidAt(now)
	status := "active"
	if !enabled {
		status = "disabled"
		if c.EndsAt != nil && !now.Before(*c.EndsAt) {
			status = "expired"
		}
		if c.MaxRedemptions > 0 && c.RedeemedCount >= c.MaxRedemptions {
			status = "exhausted"
		}
	}
	m := map[string]any{
		"id": c.ID.String(), "promotionId": c.PromotionID.String(),
		"code": c.Code, "kind": string(c.Kind),
		"maxRedemptions": c.MaxRedemptions, "redeemedCount": c.RedeemedCount,
		"startsAt": c.StartsAt, "endsAt": c.EndsAt,
		"enabled": enabled, "active": enabled, "status": status,
		"createdAt": c.CreatedAt, "updatedAt": c.UpdatedAt,
	}
	if c.PrincipalID != nil {
		m["principalId"] = c.PrincipalID.String()
	}
	return m
}

func voucherDTO(v domain.Voucher) map[string]any {
	m := map[string]any{
		"id": v.ID.String(), "code": v.Code, "status": string(v.Status),
		"principalId": v.PrincipalID.String(), "valueMinor": v.ValueMinor,
		"remainingMinor": v.RemainingMinor, "currency": v.Currency,
		"startsAt": v.StartsAt, "endsAt": v.EndsAt,
	}
	if v.PromotionID != nil {
		m["promotionId"] = v.PromotionID.String()
	}
	return m
}

func evaluateDTO(res domain.EvaluateResult) map[string]any {
	discounts := make([]map[string]any, 0, len(res.Discounts))
	for _, d := range res.Discounts {
		discounts = append(discounts, map[string]any{
			"promotionId": d.PromotionID.String(), "campaignId": d.CampaignID.String(),
			"type": string(d.Type), "amountMinor": d.AmountMinor, "currency": d.Currency,
			"description": d.Description, "stackGroup": d.StackGroup, "priority": d.Priority,
			"appliedLineIds": d.AppliedLineIDs, "couponCode": d.CouponCode,
		})
	}
	return map[string]any{
		"discounts": discounts, "totalDiscountMinor": res.TotalDiscountMinor,
		"shippingDiscountMinor": res.ShippingDiscountMinor, "currency": res.Currency,
	}
}
