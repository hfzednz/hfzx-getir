package httpadapter

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/nexora/bff-customer/internal/domain"
)

func (h *Handler) getProfile(w http.ResponseWriter, r *http.Request) {
	pid, ok := requirePrincipal(w, r)
	if !ok {
		return
	}
	if h.Book == nil {
		writeJSON(w, 200, map[string]any{"id": pid})
		return
	}
	writeJSON(w, 200, h.Book.Profile(tenant(r), pid))
}

func (h *Handler) patchProfile(w http.ResponseWriter, r *http.Request) {
	pid, ok := requirePrincipal(w, r)
	if !ok {
		return
	}
	if h.Book == nil {
		writeJSON(w, 200, map[string]any{"id": pid})
		return
	}
	var body map[string]any
	_ = json.NewDecoder(r.Body).Decode(&body)
	writeJSON(w, 200, h.Book.PutProfile(tenant(r), pid, body))
}

func (h *Handler) listNotifications(w http.ResponseWriter, r *http.Request) {
	pid, ok := requirePrincipal(w, r)
	if !ok {
		return
	}
	if h.Book == nil {
		writeJSON(w, 200, map[string]any{"items": []any{}})
		return
	}
	items := h.Book.ListNotifications(tenant(r), pid)
	writeJSON(w, 200, map[string]any{"items": items, "notifications": items})
}

func (h *Handler) markNotificationRead(w http.ResponseWriter, r *http.Request) {
	pid, ok := requirePrincipal(w, r)
	if !ok {
		return
	}
	n, found := h.Book.MarkNotificationRead(tenant(r), pid, r.PathValue("id"))
	if !found {
		writeErr(w, domain.ErrNotFound)
		return
	}
	writeJSON(w, 200, n)
}

func (h *Handler) markAllNotificationsRead(w http.ResponseWriter, r *http.Request) {
	pid, ok := requirePrincipal(w, r)
	if !ok {
		return
	}
	h.Book.MarkAllNotificationsRead(tenant(r), pid)
	writeJSON(w, 200, map[string]any{"status": "ok"})
}

func (h *Handler) getNotificationPrefs(w http.ResponseWriter, r *http.Request) {
	pid, ok := requirePrincipal(w, r)
	if !ok {
		return
	}
	writeJSON(w, 200, h.Book.NotificationPrefs(tenant(r), pid))
}

func (h *Handler) putNotificationPrefs(w http.ResponseWriter, r *http.Request) {
	pid, ok := requirePrincipal(w, r)
	if !ok {
		return
	}
	var body map[string]any
	_ = json.NewDecoder(r.Body).Decode(&body)
	writeJSON(w, 200, h.Book.PutNotificationPrefs(tenant(r), pid, body))
}

func (h *Handler) registerFcm(w http.ResponseWriter, r *http.Request) {
	if _, ok := requirePrincipal(w, r); !ok {
		return
	}
	var body map[string]any
	_ = json.NewDecoder(r.Body).Decode(&body)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) listFaq(w http.ResponseWriter, r *http.Request) {
	lang := r.Header.Get("Accept-Language")
	tr := strings.HasPrefix(strings.ToLower(lang), "tr")
	items := customerFaq(tr)
	writeJSON(w, 200, map[string]any{"items": items, "faq": items})
}

func customerFaq(tr bool) []map[string]any {
	if tr {
		return []map[string]any{
			{"id": "late", "category": "delivery", "question": "Siparişim gecikti, ne yapmalıyım?", "answer": "Sipariş detayından canlı takibi açın. 10 dakikadan fazla gecikmede destek bileti oluşturun."},
			{"id": "missing", "category": "order", "question": "Ürün eksik geldi, nasıl bildirebilirim?", "answer": "Sipariş detayından destek bileti açın ve eksik ürünü seçin. Uygun durumlarda iade başlatılır."},
			{"id": "wrong", "category": "order", "question": "Yanlış ürün geldi.", "answer": "Yanlış ürün için sipariş numarasıyla destek talebi açın. Fotoğraf eklemeniz süreci hızlandırır."},
			{"id": "damaged", "category": "order", "question": "Ürün hasarlı geldi.", "answer": "Hasar fotoğrafıyla destek talebi açın. Onay sonrası iade veya değişim yapılır."},
			{"id": "cancel", "category": "order", "question": "Siparişi iptal edebilir miyim?", "answer": "Hazırlık başlamadan sipariş detayından iptal edebilirsiniz. Kurye yola çıktıktan sonra iptal kapalıdır."},
			{"id": "pay", "category": "payment", "question": "Ödeme alınmadı görünüyor.", "answer": "Ödeme durumunu sipariş özetinden kontrol edin. Çift çekim şüphesinde destek bileti açın."},
			{"id": "coupon", "category": "promo", "question": "Kuponum uygulanmadı.", "answer": "Kuponun süresi dolmamış ve minimum sepet tutarını karşılıyor olmalıdır. Geçersiz kodlar reddedilir."},
		}
	}
	return []map[string]any{
		{"id": "late", "category": "delivery", "question": "My order is late. What should I do?", "answer": "Open live tracking from the order. If it is more than 10 minutes late, create a support ticket."},
		{"id": "missing", "category": "order", "question": "An item is missing.", "answer": "Open a ticket from the order and mark the missing item. A refund is issued when the claim is confirmed."},
		{"id": "wrong", "category": "order", "question": "I received the wrong item.", "answer": "Open a ticket with the order number. Photos speed the review up."},
		{"id": "damaged", "category": "order", "question": "An item arrived damaged.", "answer": "Open a ticket with a photo. We refund or replace after review."},
		{"id": "cancel", "category": "order", "question": "Can I cancel my order?", "answer": "You can cancel from order details before picking starts. Cancellation is closed after the courier is on the way."},
		{"id": "pay", "category": "payment", "question": "Payment looks unsuccessful.", "answer": "Check the order summary. Open a ticket if you think you were charged twice."},
		{"id": "coupon", "category": "promo", "question": "My coupon did not apply.", "answer": "The code must be active and the basket must meet the minimum. Invalid codes are rejected."},
	}
}

func (h *Handler) listTickets(w http.ResponseWriter, r *http.Request) {
	pid, ok := requirePrincipal(w, r)
	if !ok {
		return
	}
	items := h.Book.ListTickets(tenant(r), pid)
	writeJSON(w, 200, map[string]any{"items": items, "tickets": items})
}

func (h *Handler) getTicket(w http.ResponseWriter, r *http.Request) {
	pid, ok := requirePrincipal(w, r)
	if !ok {
		return
	}
	t, found := h.Book.GetTicket(tenant(r), pid, r.PathValue("id"))
	if !found {
		writeErr(w, domain.ErrNotFound)
		return
	}
	writeJSON(w, 200, t)
}

func (h *Handler) assistantMessage(w http.ResponseWriter, r *http.Request) {
	if _, ok := requirePrincipal(w, r); !ok {
		return
	}
	var body struct {
		Message, SessionID, OrderID string
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	reply := "Open a support ticket from Help if you need an agent. You can also track an active order from Orders."
	lower := strings.ToLower(body.Message)
	switch {
	case strings.Contains(lower, "iptal") || strings.Contains(lower, "cancel"):
		reply = "Cancellation is available from order details while the order is still confirmed and picking has not started."
	case strings.Contains(lower, "kupon") || strings.Contains(lower, "coupon"):
		reply = "Enter WELCOME10 at checkout for 10% off baskets over 150 TL. Expired or unknown codes are rejected."
	case strings.Contains(lower, "geç") || strings.Contains(lower, "late"):
		reply = "Open live tracking. If the courier is more than 10 minutes late, create a late-delivery ticket."
	}
	writeJSON(w, 200, map[string]any{
		"id": time.Now().UTC().Format("20060102150405"),
		"role": "assistant", "content": reply, "message": reply,
		"session_id": body.SessionID, "order_id": body.OrderID,
	})
}

func customerCouponView(raw map[string]any) map[string]any {
	code := strings.ToUpper(asString(raw["code"]))
	status := asString(raw["status"])
	if status == "" {
		if raw["enabled"] == false || raw["active"] == false {
			status = "disabled"
		} else {
			status = "active"
		}
	}
	active := status == "active"
	out := map[string]any{
		"id": firstNonEmpty(asString(raw["id"]), strings.ToLower(code)),
		"code": code, "title": firstNonEmpty(asString(raw["title"]), code),
		"description": asString(raw["description"]),
		"discount_type": firstNonEmpty(asString(raw["discount_type"]), asString(raw["kind"])),
		"discount_value": raw["discount_value"],
		"min_order_minor": raw["min_order_minor"],
		"currency": firstNonEmpty(asString(raw["currency"]), "TRY"),
		"active": active, "status": status, "enabled": active,
		"maxRedemptions": raw["maxRedemptions"], "redeemedCount": raw["redeemedCount"],
		"startsAt": raw["startsAt"], "endsAt": firstNonEmptyAny(raw["endsAt"], raw["expires_at"]),
		"promotionId": raw["promotionId"],
	}
	return out
}

func firstNonEmptyAny(vals ...any) any {
	for _, v := range vals {
		if v == nil {
			continue
		}
		if s, ok := v.(string); ok && s == "" {
			continue
		}
		return v
	}
	return nil
}

func (h *Handler) listCoupons(w http.ResponseWriter, r *http.Request) {
	if _, ok := requirePrincipal(w, r); !ok {
		return
	}
	if h.Deps == nil || h.Deps.Promo == nil {
		writeErr(w, domain.ErrUpstream)
		return
	}
	raw, err := h.Deps.Promo.ListCoupons(r.Context(), tenant(r))
	if err != nil {
		writeErr(w, err)
		return
	}
	items := make([]map[string]any, 0, len(raw))
	for _, c := range raw {
		items = append(items, customerCouponView(c))
	}
	writeJSON(w, 200, map[string]any{"items": items, "coupons": items})
}

func (h *Handler) getCoupon(w http.ResponseWriter, r *http.Request) {
	if _, ok := requirePrincipal(w, r); !ok {
		return
	}
	if h.Deps == nil || h.Deps.Promo == nil {
		writeErr(w, domain.ErrUpstream)
		return
	}
	raw, err := h.Deps.Promo.GetCoupon(r.Context(), tenant(r), r.PathValue("code"))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, 200, customerCouponView(raw))
}

func (h *Handler) validateCoupon(w http.ResponseWriter, r *http.Request) {
	if _, ok := requirePrincipal(w, r); !ok {
		return
	}
	if h.Deps == nil || h.Deps.Promo == nil {
		writeErr(w, domain.ErrUpstream)
		return
	}
	var body struct {
		Code              string `json:"code"`
		CartSubtotalMinor int64  `json:"cart_subtotal_minor"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	raw, err := h.Deps.Promo.GetCoupon(r.Context(), tenant(r), body.Code)
	if err != nil {
		writeErr(w, err)
		return
	}
	view := customerCouponView(raw)
	status := asString(view["status"])
	if status == "expired" || status == "disabled" || status == "exhausted" || view["active"] == false {
		writeJSON(w, 409, map[string]any{"code": "conflict", "message": "coupon has expired"})
		return
	}
	eval, err := h.Deps.Promo.EvaluateCoupon(r.Context(), tenant(r), body.Code, body.CartSubtotalMinor)
	if err != nil {
		if err == domain.ErrInvalidArgument {
			writeJSON(w, 400, map[string]any{"code": "invalid_argument", "message": "coupon minimum basket not met"})
			return
		}
		writeErr(w, err)
		return
	}
	discount := int64(0)
	if eval != nil {
		discount = asInt64Any(eval["totalDiscountMinor"])
		if discount == 0 {
			for _, d := range asAnyMaps(eval["discounts"]) {
				if strings.EqualFold(asString(d["couponCode"]), body.Code) {
					discount += asInt64Any(d["amountMinor"])
				}
			}
		}
	}
	writeJSON(w, 200, map[string]any{
		"coupon": view, "discount_minor": discount, "message": "applied",
		"evaluate": eval,
	})
}

func asInt64Any(v any) int64 {
	switch n := v.(type) {
	case int:
		return int64(n)
	case int64:
		return n
	case float64:
		return int64(n)
	default:
		return 0
	}
}

func asAnyMaps(v any) []map[string]any {
	switch t := v.(type) {
	case []map[string]any:
		return t
	case []any:
		out := make([]map[string]any, 0, len(t))
		for _, item := range t {
			if m, ok := item.(map[string]any); ok {
				out = append(out, m)
			}
		}
		return out
	default:
		return nil
	}
}

func (h *Handler) listPaymentCards(w http.ResponseWriter, r *http.Request) {
	if _, ok := requirePrincipal(w, r); !ok {
		return
	}
	writeJSON(w, 200, map[string]any{"cards": []any{}})
}

func (h *Handler) walletPay(w http.ResponseWriter, r *http.Request) {
	if _, ok := requirePrincipal(w, r); !ok {
		return
	}
	writeJSON(w, 400, map[string]any{"code": "invalid_argument", "message": "wallet payment is not available"})
}

func (h *Handler) retryPayment(w http.ResponseWriter, r *http.Request) {
	if _, ok := requirePrincipal(w, r); !ok {
		return
	}
	var body struct {
		SessionID string `json:"session_id"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	writeJSON(w, 200, map[string]any{
		"id": body.SessionID, "status": "payment_failed",
		"message": "payment did not go through, retry from checkout",
	})
}
