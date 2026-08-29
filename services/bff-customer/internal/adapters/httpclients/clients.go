package httpclients

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/nexora/bff-customer/internal/app/ports"
	"github.com/nexora/bff-customer/internal/domain"
	"github.com/nexora/bff-customer/internal/reqctx"
)

// Identity implements ports.IdentityClient against identity-service.
type Identity struct{ Base }

func NewIdentity(baseURL string) *Identity { return &Identity{Base: newBase(baseURL)} }

func (c *Identity) StartOTP(ctx context.Context, tenantID, phone string) (string, error) {
	var start struct {
		ChallengeID string `json:"challengeId"`
	}
	if err := c.post(ctx, "/v1/identity/auth/otp/start", tenantID, map[string]any{
		"phone": phone, "tenantId": tenantID,
	}, &start); err != nil {
		return "", err
	}
	if start.ChallengeID == "" {
		return "", domain.ErrUpstream
	}
	return start.ChallengeID, nil
}

func (c *Identity) VerifyOTP(ctx context.Context, tenantID, challengeID, code string) (domain.SessionView, error) {
	var res map[string]any
	if err := c.post(ctx, "/v1/identity/auth/otp/verify", tenantID, map[string]any{
		"challengeId": challengeID, "code": code,
	}, &res); err != nil {
		return domain.SessionView{}, err
	}
	if asBool(res["mfaRequired"]) {
		return domain.SessionView{}, domain.ErrUnauthorized
	}
	return domain.SessionView{
		AccessToken:  asString(res["accessToken"]),
		RefreshToken: asString(res["refreshToken"]),
		CustomerID:   firstNonEmpty(asString(res["principalId"]), asString(res["customerId"])),
		ExpiresIn:    int(asInt64(res["expiresIn"])),
	}, nil
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

var _ ports.IdentityClient = (*Identity)(nil)

// Catalog implements ports.CatalogClient against catalog-service.
type Catalog struct{ Base }

func NewCatalog(baseURL string) *Catalog { return &Catalog{Base: newBase(baseURL)} }

func (c *Catalog) Search(ctx context.Context, tenantID, query string) ([]map[string]any, error) {
	q := url.Values{"q": {query}}.Encode()
	var raw map[string]any
	if err := c.get(ctx, "/v1/catalog/search?"+q, tenantID, &raw); err != nil {
		return nil, err
	}
	hits := raw["hits"]
	if hits == nil {
		hits = raw["Hits"]
	}
	arr, _ := hits.([]any)
	out := make([]map[string]any, 0, len(arr))
	for _, h := range arr {
		if m, ok := h.(map[string]any); ok {
			pid := firstNonEmpty(asString(m["ProductID"]), asString(m["productId"]))
			sku := firstNonEmpty(pid, asString(m["SKU"]), asString(m["sku"]))
			item := map[string]any{
				"id":   sku,
				"sku":  sku,
				"name": firstNonEmpty(asString(m["Title"]), asString(m["title"]), asString(m["name"]), "Product"),
			}
			if pid != "" {
				item["productId"] = pid
			}
			if pm, ok := m["priceMinor"]; ok {
				item["priceMinor"] = asInt64(pm)
			}
			out = append(out, item)
		}
	}
	return out, nil
}

var _ ports.CatalogClient = (*Catalog)(nil)

// Recs implements ports.RecClient against recommendation-service.
type Recs struct{ Base }

func NewRecs(baseURL string) *Recs { return &Recs{Base: newBase(baseURL)} }

func (c *Recs) ForYou(ctx context.Context, tenantID, customerID string) ([]map[string]any, error) {
	body := map[string]any{"context": "home", "limit": 20}
	if customerID != "" {
		body["userId"] = customerID
	}
	var res map[string]any
	if err := c.post(ctx, "/v1/recommendations/for-you", tenantID, body, &res); err != nil {
		return nil, err
	}
	return []map[string]any{res}, nil
}

var _ ports.RecClient = (*Recs)(nil)

// Cart implements ports.CartClient against cart-service.
type Cart struct{ Base }

func NewCart(baseURL string) *Cart { return &Cart{Base: newBase(baseURL)} }

func (c *Cart) Get(ctx context.Context, tenantID, cartID string) (map[string]any, error) {
	var out map[string]any
	if err := c.get(ctx, "/v1/cart/"+url.PathEscape(cartID), tenantID, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Cart) AddItem(ctx context.Context, tenantID, cartID, sku string, qty, unitMinor int64) (map[string]any, error) {
	create := func() error {
		var created map[string]any
		if err := c.post(ctx, "/v1/cart", tenantID, map[string]any{
			"guestToken": fmt.Sprintf("web-%d", time.Now().UnixNano()), "currency": "TRY",
		}, &created); err != nil {
			return err
		}
		cartID = firstNonEmpty(asString(created["ID"]), asString(created["id"]), asString(created["cartId"]))
		if cartID == "" {
			return domain.ErrUpstream
		}
		return nil
	}
	if cartID == "" {
		if err := create(); err != nil {
			return nil, err
		}
	}
	body := map[string]any{"qty": qty, "sku": sku, "unitMinor": unitMinor, "variantId": sku}
	var out map[string]any
	err := c.post(ctx, "/v1/cart/"+url.PathEscape(cartID)+"/lines", tenantID, body, &out)
	if err == domain.ErrNotFound {
		if err := create(); err != nil {
			return nil, err
		}
		err = c.post(ctx, "/v1/cart/"+url.PathEscape(cartID)+"/lines", tenantID, body, &out)
	}
	if err != nil {
		return nil, err
	}
	if _, ok := out["cartId"]; !ok {
		out["cartId"] = cartID
	}
	if _, ok := out["sku"]; !ok {
		out["sku"] = sku
	}
	if _, ok := out["qty"]; !ok {
		out["qty"] = qty
	}
	if _, ok := out["lineTotalMinor"]; !ok && unitMinor > 0 {
		out["lineTotalMinor"] = qty * unitMinor
	}
	return out, nil
}

var _ ports.CartClient = (*Cart)(nil)

// Checkout implements ports.CheckoutClient; Place also calls payment eligibility.
type Checkout struct {
	Base
	Payment Base
}

func NewCheckout(checkoutURL, paymentURL string) *Checkout {
	return &Checkout{Base: newBase(checkoutURL), Payment: newBase(paymentURL)}
}

func (c *Checkout) Preview(ctx context.Context, tenantID, cartID string) (domain.CheckoutPreview, error) {
	var sess map[string]any
	body := map[string]any{
		"cartId": cartID, "currency": "TRY",
		"idempotencyKey": "bff-preview-" + cartID,
	}
	if uid := reqctx.UserID(ctx); uid != "" {
		body["principalId"] = uid
	}
	if err := c.post(ctx, "/v1/checkout/sessions", tenantID, body, &sess); err != nil {
		return domain.CheckoutPreview{}, err
	}
	return previewFromSession(cartID, sess), nil
}

func (c *Checkout) Place(ctx context.Context, tenantID, cartID, paymentMethod, sessionID string, addr domain.CheckoutAddress) (string, error) {
	if addr.Empty() {
		return "", fmt.Errorf("%w: delivery address required", domain.ErrInvalidArgument)
	}
	if sessionID == "" {
		if cartID == "" {
			return "", fmt.Errorf("%w: cart required", domain.ErrInvalidArgument)
		}
		var sess map[string]any
		body := map[string]any{
			"cartId": cartID, "currency": "TRY",
			"idempotencyKey": "bff-place-" + cartID,
		}
		if uid := reqctx.UserID(ctx); uid != "" {
			body["principalId"] = uid
		}
		if err := c.post(ctx, "/v1/checkout/sessions", tenantID, body, &sess); err != nil {
			return "", err
		}
		sessionID = asString(sess["id"])
		if sessionID == "" {
			return "", domain.ErrUpstream
		}
	}
	_ = paymentMethod
	if err := c.patch(ctx, "/v1/checkout/sessions/"+url.PathEscape(sessionID), tenantID, map[string]any{
		"address": addr.Map(),
	}, nil); err != nil {
		return "", err
	}
	var validated map[string]any
	if err := c.post(ctx, "/v1/checkout/sessions/"+url.PathEscape(sessionID)+"/validate", tenantID, map[string]any{}, &validated); err != nil {
		return "", err
	}
	if asString(validated["status"]) != "ready" {
		return "", fmt.Errorf("%w: checkout not ready (status=%s issues=%s)",
			domain.ErrInvalidArgument, asString(validated["status"]), validationIssueSummary(validated))
	}
	quote, _ := validated["quote"].(map[string]any)
	amount := asInt64(quote["totalMinor"])
	currency := firstNonEmpty(asString(validated["currency"]), asString(quote["currency"]), "TRY")
	if c.Payment.BaseURL != "" && amount > 0 {
		payload := map[string]any{"currency": currency, "amountMinor": amount}
		var elig map[string]any
		if err := c.Payment.post(ctx, "/v1/payments/eligibility", tenantID, payload, &elig); err != nil {
			return "", err
		}
		if v, ok := elig["eligible"]; ok && !asBool(v) {
			return "", domain.ErrUpstream
		}
		if v, ok := elig["Eligible"]; ok && !asBool(v) {
			return "", domain.ErrUpstream
		}
	}
	var done map[string]any
	if err := c.post(ctx, "/v1/checkout/sessions/"+url.PathEscape(sessionID)+"/complete", tenantID, map[string]any{
		"placeOrder":     true,
		"idempotencyKey": "bff-complete-" + sessionID,
	}, &done); err != nil {
		return "", err
	}
	oid := asString(done["orderId"])
	if oid == "" {
		return "", domain.ErrUpstream
	}
	return oid, nil
}

func validationIssueSummary(sess map[string]any) string {
	val, _ := sess["validation"].(map[string]any)
	if val == nil {
		return ""
	}
	raw, ok := val["issues"].([]any)
	if !ok || len(raw) == 0 {
		return ""
	}
	parts := make([]string, 0, len(raw))
	for _, item := range raw {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		msg := firstNonEmpty(asString(m["message"]), asString(m["code"]))
		if msg != "" {
			parts = append(parts, msg)
		}
	}
	if len(parts) == 0 {
		return ""
	}
	return fmt.Sprintf("%v", parts)
}

func previewFromSession(cartID string, sess map[string]any) domain.CheckoutPreview {
	quote, _ := sess["quote"].(map[string]any)
	val, _ := sess["validation"].(map[string]any)
	ready := true
	if val != nil {
		if v, ok := val["paymentReady"]; ok {
			ready = asBool(v)
		} else if v, ok := val["ok"]; ok {
			ready = asBool(v)
		}
	}
	cid := firstNonEmpty(asString(sess["cartId"]), cartID)
	return domain.CheckoutPreview{
		SessionID:     asString(sess["id"]),
		CartID:        cid,
		Currency:      firstNonEmpty(asString(sess["currency"]), "TRY"),
		SubtotalMinor: asInt64(quote["subtotalMinor"]),
		DiscountMinor: asInt64(quote["discountMinor"]),
		TotalMinor:    asInt64(quote["totalMinor"]),
		PaymentReady:  ready,
	}
}

var _ ports.CheckoutClient = (*Checkout)(nil)

// Orders implements ports.OrderClient against order-service.
type Orders struct{ Base }

func NewOrders(baseURL string) *Orders { return &Orders{Base: newBase(baseURL)} }

func (c *Orders) Get(ctx context.Context, tenantID, orderID string) (map[string]any, error) {
	var out map[string]any
	if err := c.get(ctx, "/v1/orders/"+url.PathEscape(orderID), tenantID, &out); err != nil {
		return nil, err
	}
	return out, nil
}

var _ ports.OrderClient = (*Orders)(nil)

// Tracking implements ports.TrackingClient against tracking-service.
type Tracking struct{ Base }

func NewTracking(baseURL string) *Tracking { return &Tracking{Base: newBase(baseURL)} }

func (c *Tracking) Track(ctx context.Context, tenantID, orderID string) (domain.OrderTrack, error) {
	var raw map[string]any
	if err := c.get(ctx, "/v1/tracking/orders/"+url.PathEscape(orderID)+"/timeline", tenantID, &raw); err != nil {
		return domain.OrderTrack{}, err
	}
	tr := domain.OrderTrack{OrderID: orderID, Status: "unknown"}
	items, _ := raw["items"].([]any)
	if len(items) == 0 {
		return tr, nil
	}
	last, _ := items[len(items)-1].(map[string]any)
	var meta map[string]any
	if m, ok := last["meta"].(map[string]any); ok {
		meta = m
	}
	tr.Status = firstNonEmpty(asString(meta["status"]), asString(last["message"]), asString(last["status"]), asString(last["type"]), "unknown")
	if tr.Status == "Custom" {
		tr.Status = firstNonEmpty(asString(meta["status"]), asString(last["message"]), "unknown")
	}
	tr.CourierID = asString(last["courierId"])
	tr.Lat = asFloat64(last["lat"])
	tr.Lng = asFloat64(last["lon"])
	if tr.Lng == 0 {
		tr.Lng = asFloat64(last["lng"])
	}
	if eta := asInt64(meta["etaSeconds"]); eta > 0 {
		tr.ETASeconds = int(eta)
	}
	return tr, nil
}

var _ ports.TrackingClient = (*Tracking)(nil)

// Location implements ports.LocationClient against location-service.
type Location struct{ Base }

func NewLocation(baseURL string) *Location { return &Location{Base: newBase(baseURL)} }

func (c *Location) Serviceable(ctx context.Context, tenantID string, lat, lng float64) (bool, error) {
	var res map[string]any
	if err := c.post(ctx, "/v1/location/zones/serviceability", tenantID, map[string]any{
		"lat": lat, "lng": lng,
	}, &res); err != nil {
		return false, err
	}
	if v, ok := res["serviceable"]; ok {
		return asBool(v), nil
	}
	return true, nil
}

var _ ports.LocationClient = (*Location)(nil)

// Notify implements ports.NotificationClient against notification-service.
type Notify struct{ Base }

func NewNotify(baseURL string) *Notify { return &Notify{Base: newBase(baseURL)} }

func (c *Notify) RegisterDevice(ctx context.Context, tenantID, customerID, token string) error {
	return c.post(ctx, "/v1/notifications/devices", tenantID, map[string]any{
		"principalId": customerID, "token": token, "platform": "unknown",
	}, nil)
}

var _ ports.NotificationClient = (*Notify)(nil)

// CRM implements ports.CRMClient against crm-service.
type CRM struct{ Base }

func NewCRM(baseURL string) *CRM { return &CRM{Base: newBase(baseURL)} }

func (c *CRM) OpenTicket(ctx context.Context, tenantID, customerID, subject string) (string, error) {
	var res map[string]any
	if err := c.post(ctx, "/v1/crm/tickets", tenantID, map[string]any{
		"customerId": customerID, "subject": subject, "description": subject,
	}, &res); err != nil {
		return "", err
	}
	id := firstNonEmpty(asString(res["id"]), asString(res["ticketId"]))
	if id == "" {
		return "", domain.ErrUpstream
	}
	return id, nil
}

var _ ports.CRMClient = (*CRM)(nil)

// Reviews implements ports.ReviewClient against review-service.
type Reviews struct{ Base }

func NewReviews(baseURL string) *Reviews { return &Reviews{Base: newBase(baseURL)} }

func (c *Reviews) Submit(ctx context.Context, tenantID, orderID string, rating int, body string) error {
	rv := float64(rating)
	return c.post(ctx, "/v1/reviews", tenantID, map[string]any{
		"targetType": "order", "targetId": orderID, "orderId": orderID,
		"body": body, "ratingValue": rv, "scheme": "stars",
	}, nil)
}

var _ ports.ReviewClient = (*Reviews)(nil)
