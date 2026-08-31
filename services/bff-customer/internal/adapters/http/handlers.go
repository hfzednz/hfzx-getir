package httpadapter

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/bff-customer/internal/app"
	"github.com/nexora/bff-customer/internal/authz"
	"github.com/nexora/bff-customer/internal/domain"
	"github.com/nexora/bff-customer/internal/reqctx"
)

type Handler struct {
	Deps *app.Deps
	Book *CustomerBook
}

func NewServer(addr string, deps *app.Deps) *http.Server {
	return NewServerWithAuth(addr, deps, authz.FromEnv())
}

func NewServerWithAuth(addr string, deps *app.Deps, v authz.Validator) *http.Server {
	h := &Handler{Deps: deps, Book: NewCustomerBook()}
	mux := http.NewServeMux()
	const base = "/v1/customer"
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) { writeJSON(w, 200, map[string]string{"status": "ok"}) })
	mux.HandleFunc("GET /ready", func(w http.ResponseWriter, _ *http.Request) { writeJSON(w, 200, map[string]string{"status": "ready"}) })
	mux.HandleFunc("POST "+base+"/auth/otp/start", h.otpStart)
	mux.HandleFunc("POST "+base+"/auth/otp/verify", h.login)
	mux.HandleFunc("GET "+base+"/home", h.home)
	mux.HandleFunc("GET "+base+"/search", h.search)
	mux.HandleFunc("GET "+base+"/search/semantic", h.search)
	mux.HandleFunc("GET "+base+"/search/suggestions", h.searchSuggestions)
	mux.HandleFunc("GET "+base+"/products/{id}", h.getProduct)
	mux.HandleFunc("GET "+base+"/categories", h.listCategories)
	mux.HandleFunc("GET "+base+"/categories/{id}", h.getCategory)
	mux.HandleFunc("GET "+base+"/stores", h.listStores)
	mux.HandleFunc("GET "+base+"/stores/{id}", h.getStore)
	mux.HandleFunc("POST "+base+"/cart/items", h.addCart)
	mux.HandleFunc("POST "+base+"/checkout/preview", h.preview)
	mux.HandleFunc("POST "+base+"/checkout/place", h.place)
	mux.HandleFunc("GET "+base+"/orders", h.listOrders)
	mux.HandleFunc("GET "+base+"/orders/{id}", h.getOrder)
	mux.HandleFunc("GET "+base+"/orders/{id}/track", h.track)
	mux.HandleFunc("GET "+base+"/orders/{id}/tracking", h.track)
	mux.HandleFunc("POST "+base+"/orders/{id}/realtime-ticket", h.realtimeTicket)
	mux.HandleFunc("GET "+base+"/orders/{id}/realtime-ticket", h.realtimeTicket)
	mux.HandleFunc("POST "+base+"/orders/{id}/cancel", h.cancelOrder)
	mux.HandleFunc("POST "+base+"/orders/{id}/reorder", h.reorder)
	mux.HandleFunc("GET "+base+"/addresses", h.listAddresses)
	mux.HandleFunc("POST "+base+"/addresses", h.createAddress)
	mux.HandleFunc("GET "+base+"/addresses/validate-zone", h.validateZone)
	mux.HandleFunc("GET "+base+"/addresses/{id}", h.getAddress)
	mux.HandleFunc("PATCH "+base+"/addresses/{id}", h.patchAddress)
	mux.HandleFunc("DELETE "+base+"/addresses/{id}", h.deleteAddress)
	mux.HandleFunc("POST "+base+"/addresses/{id}/default", h.defaultAddress)
	mux.HandleFunc("POST "+base+"/addresses/{id}/favorite", h.favoriteAddress)
	mux.HandleFunc("GET "+base+"/favorites", h.listFavorites)
	mux.HandleFunc("POST "+base+"/favorites", h.addFavorite)
	mux.HandleFunc("DELETE "+base+"/favorites/{id}", h.deleteFavorite)
	mux.HandleFunc("POST "+base+"/support/tickets", h.ticket)
	mux.HandleFunc("GET "+base+"/support/tickets", h.listTickets)
	mux.HandleFunc("GET "+base+"/support/tickets/{id}", h.getTicket)
	mux.HandleFunc("GET "+base+"/support/faq", h.listFaq)
	mux.HandleFunc("POST "+base+"/support/assistant/message", h.assistantMessage)
	mux.HandleFunc("GET "+base+"/profile", h.getProfile)
	mux.HandleFunc("PATCH "+base+"/profile", h.patchProfile)
	mux.HandleFunc("GET "+base+"/notifications", h.listNotifications)
	mux.HandleFunc("POST "+base+"/notifications/{id}/read", h.markNotificationRead)
	mux.HandleFunc("POST "+base+"/notifications/read-all", h.markAllNotificationsRead)
	mux.HandleFunc("GET "+base+"/notifications/preferences", h.getNotificationPrefs)
	mux.HandleFunc("PUT "+base+"/notifications/preferences", h.putNotificationPrefs)
	mux.HandleFunc("POST "+base+"/notifications/fcm-token", h.registerFcm)
	mux.HandleFunc("GET "+base+"/coupons", h.listCoupons)
	mux.HandleFunc("GET "+base+"/coupons/{code}", h.getCoupon)
	mux.HandleFunc("POST "+base+"/coupons/validate", h.validateCoupon)
	mux.HandleFunc("POST "+base+"/coupons/apply", h.validateCoupon)
	mux.HandleFunc("GET "+base+"/payment-methods/cards", h.listPaymentCards)
	mux.HandleFunc("POST "+base+"/payment-methods/wallet/pay", h.walletPay)
	mux.HandleFunc("POST "+base+"/payment-methods/retry", h.retryPayment)
	mux.HandleFunc("POST "+base+"/reviews", h.review)
	gated := authz.Gate(v, authz.Options{
		Public: []string{"/health", "/ready", "/v1/customer/auth/otp/"},
		Rules: []authz.Rule{
			{Prefix: "/v1/customer", Roles: []string{"customer"}},
		},
	})(requestIDMiddleware(mux))
	return &http.Server{Addr: addr, Handler: gated, ReadHeaderTimeout: 5 * time.Second}
}

func requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-Id")
		if id == "" {
			id = uuid.NewString()
		}
		w.Header().Set("X-Request-Id", id)
		ctx := reqctx.WithRequestID(r.Context(), id)
		if uid := r.Header.Get("X-Nexora-User"); uid != "" {
			ctx = reqctx.WithUserID(ctx, uid)
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, err error) {
	code, status := "internal_error", 500
	switch {
	case errors.Is(err, domain.ErrInvalidArgument):
		code, status = "invalid_argument", 400
	case errors.Is(err, domain.ErrUnauthorized):
		code, status = "unauthorized", 401
	case errors.Is(err, domain.ErrNotFound):
		code, status = "not_found", 404
	case errors.Is(err, domain.ErrConflict):
		code, status = "conflict", 409
	case errors.Is(err, domain.ErrUpstream):
		code, status = "upstream_failure", 502
	}
	writeJSON(w, status, map[string]any{"error": map[string]any{"code": code, "message": err.Error(), "retriable": status >= 500}})
}

func tenant(r *http.Request) string {
	t := r.Header.Get("X-Tenant-Id")
	if t == "" {
		t = r.Header.Get("X-Nexora-Tenant")
	}
	return t
}

func (h *Handler) otpStart(w http.ResponseWriter, r *http.Request) {
	var body struct{ Phone string }
	_ = json.NewDecoder(r.Body).Decode(&body)
	id, err := h.Deps.StartOTP(r.Context(), tenant(r), body.Phone)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, 200, map[string]any{"challengeId": id, "expiresIn": 300})
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ChallengeID string `json:"challengeId"`
		Code        string `json:"code"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	s, err := h.Deps.Login(r.Context(), tenant(r), body.ChallengeID, body.Code)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, 200, s)
}

func (h *Handler) home(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	var lat, lng float64
	fmt.Sscanf(q.Get("lat"), "%f", &lat)
	fmt.Sscanf(q.Get("lng"), "%f", &lng)
	feed, err := h.Deps.Home(r.Context(), tenant(r), q.Get("customerId"), q.Get("q"), lat, lng)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, 200, feed)
}

func (h *Handler) addCart(w http.ResponseWriter, r *http.Request) {
	var body struct {
		CartID    string `json:"cartId"`
		SKU       string `json:"sku"`
		Qty       int64  `json:"qty"`
		UnitMinor int64  `json:"unitMinor"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	if pid := callerID(r); pid != "" {
		r = r.WithContext(reqctx.WithUserID(r.Context(), pid))
	}
	res, err := h.Deps.AddToCart(r.Context(), tenant(r), body.CartID, body.SKU, body.Qty, body.UnitMinor)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, 201, res)
}

func (h *Handler) preview(w http.ResponseWriter, r *http.Request) {
	var body struct {
		CartID      string `json:"cartId"`
		PrincipalID string `json:"principalId"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	if pid := firstNonEmpty(callerID(r), body.PrincipalID); pid != "" {
		r = r.WithContext(reqctx.WithUserID(r.Context(), pid))
	}
	p, err := h.Deps.PreviewCheckout(r.Context(), tenant(r), body.CartID)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, 200, p)
}

func (h *Handler) place(w http.ResponseWriter, r *http.Request) {
	var body struct {
		CartID        string                 `json:"cartId"`
		PaymentMethod string                 `json:"paymentMethod"`
		SessionID     string                 `json:"sessionId"`
		PrincipalID   string                 `json:"principalId"`
		Address       domain.CheckoutAddress `json:"address"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	pid := firstNonEmpty(callerID(r), body.PrincipalID)
	if pid != "" {
		r = r.WithContext(reqctx.WithUserID(r.Context(), pid))
	}
	id, err := h.Deps.PlaceOrder(r.Context(), tenant(r), body.CartID, body.PaymentMethod, body.SessionID, body.Address)
	if err != nil {
		writeErr(w, err)
		return
	}
	if h.Book != nil && pid != "" {
		h.Book.RememberOrder(tenant(r), pid, map[string]any{
			"id": id, "orderId": id, "status": "confirmed", "cartId": body.CartID,
			"customerPrincipalId": pid,
			"address":             body.Address.Map(),
		})
		h.Book.AddNotification(tenant(r), pid, map[string]any{
			"type": "transactional", "title": "Order confirmed",
			"body": "Your order is confirmed.", "deep_link": "/orders/" + id,
			"orderId": id,
		})
	}
	writeJSON(w, 201, map[string]string{"orderId": id})
}

func (h *Handler) getOrder(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if h.Deps.Orders != nil {
		out, err := h.Deps.Orders.Get(r.Context(), tenant(r), id)
		if err == nil && ownedByCaller(r, out) {
			writeJSON(w, 200, out)
			return
		}
	}
	if h.Book != nil {
		pid := callerID(r)
		for _, o := range h.Book.ListOrders(tenant(r), pid) {
			if asString(o["id"]) == id || asString(o["orderId"]) == id {
				writeJSON(w, 200, o)
				return
			}
		}
	}
	if h.Deps.Orders == nil {
		writeErr(w, domain.ErrUpstream)
		return
	}
	writeErr(w, domain.ErrNotFound)
}

func (h *Handler) track(w http.ResponseWriter, r *http.Request) {
	if err := h.requireOwnedOrder(r); err != nil {
		writeErr(w, err)
		return
	}
	t, err := h.Deps.TrackOrder(r.Context(), tenant(r), r.PathValue("id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, 200, t)
}

func (h *Handler) realtimeTicket(w http.ResponseWriter, r *http.Request) {
	if err := h.requireOwnedOrder(r); err != nil {
		writeErr(w, err)
		return
	}
	p, ok := authz.PrincipalFrom(r.Context())
	if !ok {
		writeErr(w, domain.ErrUnauthorized)
		return
	}
	secret := os.Getenv("SSE_TICKET_SECRET")
	topic := "order:" + r.PathValue("id")
	ticket, err := authz.IssueSSETicket(secret, p.TenantID, p.ID, topic, 2*time.Minute)
	if err != nil {
		writeErr(w, domain.ErrUnauthorized)
		return
	}
	writeJSON(w, 200, map[string]any{"ticket": ticket, "expiresIn": 120, "topic": topic})
}

func (h *Handler) requireOwnedOrder(r *http.Request) error {
	id := r.PathValue("id")
	if h.Deps.Orders != nil {
		out, err := h.Deps.Orders.Get(r.Context(), tenant(r), id)
		if err == nil && ownedByCaller(r, out) {
			return nil
		}
		if err != nil && h.Book == nil {
			return err
		}
	}
	if h.Book != nil {
		pid := callerID(r)
		if pid == "" {
			return domain.ErrUnauthorized
		}
		for _, o := range h.Book.ListOrders(tenant(r), pid) {
			if asString(o["id"]) == id || asString(o["orderId"]) == id {
				return nil
			}
		}
		return domain.ErrNotFound
	}
	return domain.ErrUpstream
}

func ownedByCaller(r *http.Request, order map[string]any) bool {
	p, ok := authz.PrincipalFrom(r.Context())
	if !ok || p.ID == "" {
		return false
	}
	owner := firstNonEmpty(
		asString(order["customerPrincipalId"]),
		asString(order["CustomerPrincipalId"]),
		asString(order["principalId"]),
	)
	return owner == p.ID
}

func callerID(r *http.Request) string {
	if p, ok := authz.PrincipalFrom(r.Context()); ok && p.ID != "" {
		return p.ID
	}
	return reqctx.UserID(r.Context())
}

func requirePrincipal(w http.ResponseWriter, r *http.Request) (string, bool) {
	id := callerID(r)
	if id == "" {
		writeErr(w, domain.ErrUnauthorized)
		return "", false
	}
	return id, true
}

func asString(v any) string {
	s, _ := v.(string)
	return s
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

func (h *Handler) ticket(w http.ResponseWriter, r *http.Request) {
	pid, ok := requirePrincipal(w, r)
	if !ok {
		return
	}
	var body struct {
		Subject, Category, OrderID, Message, Body string
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	subject := firstNonEmpty(body.Subject, "Help request")
	ticketID := ""
	if h.Deps.CRM != nil {
		if id, err := h.Deps.OpenSupport(r.Context(), tenant(r), pid, subject); err == nil {
			ticketID = id
		}
	}
	if h.Book != nil {
		created := h.Book.AddTicket(tenant(r), pid, map[string]any{
			"id": ticketID, "subject": subject, "category": body.Category,
			"order_id": firstNonEmpty(body.OrderID), "status": "open",
			"body": firstNonEmpty(body.Body, body.Message),
		})
		writeJSON(w, 201, created)
		return
	}
	if ticketID == "" {
		writeErr(w, domain.ErrUpstream)
		return
	}
	writeJSON(w, 201, map[string]any{"id": ticketID, "ticketId": ticketID, "subject": subject, "status": "open"})
}

func (h *Handler) review(w http.ResponseWriter, r *http.Request) {
	var body struct {
		OrderID, Body string
		Rating        int
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	if err := h.Deps.SubmitReview(r.Context(), tenant(r), body.OrderID, body.Rating, body.Body); err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, 201, map[string]string{"status": "accepted"})
}
