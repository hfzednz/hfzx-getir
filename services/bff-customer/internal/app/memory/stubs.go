package memory

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/bff-customer/internal/domain"
)

type Stubs struct {
	Orders map[string]string
	Carts  map[string][]map[string]any
}

func NewStubs() *Stubs {
	return &Stubs{Orders: map[string]string{}, Carts: map[string][]map[string]any{}}
}

func (s *Stubs) StartOTP(_ context.Context, _, phone string) (string, error) {
	if phone == "" {
		return "", domain.ErrInvalidArgument
	}
	return "ch_" + phone, nil
}

func (s *Stubs) VerifyOTP(_ context.Context, _, challengeID, code string) (domain.SessionView, error) {
	if code != "000000" && code != "123456" {
		return domain.SessionView{}, domain.ErrUnauthorized
	}
	return domain.SessionView{
		AccessToken: "atk_" + challengeID, RefreshToken: "rtk_" + challengeID,
		CustomerID: "cust_" + challengeID, ExpiresIn: 3600,
	}, nil
}

func (s *Stubs) Search(_ context.Context, _, query string) ([]map[string]any, error) {
	return []map[string]any{{"id": "SKU1", "sku": "SKU1", "name": query, "title": query, "priceMinor": int64(1999)}}, nil
}

func (s *Stubs) Categories(_ context.Context, _ string) ([]map[string]any, error) {
	return []map[string]any{{"id": "cat-dairy", "title": "Süt & Kahvaltı", "slug": "sut-kahvalti"}}, nil
}

func (s *Stubs) Product(_ context.Context, _, id string) (map[string]any, error) {
	return map[string]any{"id": id, "title": "Fresh Milk", "name": "Fresh Milk", "priceMinor": int64(1999), "currency": "TRY"}, nil
}

func (s *Stubs) ListStores(_ context.Context, _ string) ([]map[string]any, error) {
	return []map[string]any{
		{"id": "store-kadikoy", "name": "Nexora Market Kadıköy", "status": "open", "open": true, "etaMinutes": 12, "deliveryFeeMinor": 0, "minOrderMinor": 5000},
		{"id": "store-besiktas", "name": "Nexora Market Beşiktaş", "status": "open", "open": true, "etaMinutes": 18, "deliveryFeeMinor": 1499, "minOrderMinor": 7500},
		{"id": "store-bakirkoy", "name": "Nexora Market Bakırköy", "status": "closed", "open": false, "etaMinutes": 25, "deliveryFeeMinor": 1999, "minOrderMinor": 10000},
	}, nil
}

func (s *Stubs) StoreStock(_ context.Context, _, storeID string) ([]map[string]any, error) {
	milk := map[string]any{
		"sku": "SKU1", "skuCode": "SKU1", "name": "Fresh Milk", "title": "Fresh Milk",
		"available": int64(80), "outOfStock": false, "priceMinor": int64(1999),
	}
	bread := map[string]any{
		"sku": "SKU2", "skuCode": "SKU2", "name": "Village Bread", "title": "Village Bread",
		"available": int64(40), "outOfStock": false, "priceMinor": int64(1299),
	}
	yogurt := map[string]any{
		"sku": "SKU3", "skuCode": "SKU3", "name": "Strained Yogurt", "title": "Strained Yogurt",
		"available": int64(25), "outOfStock": false, "priceMinor": int64(3499),
	}
	switch storeID {
	case "store-besiktas":
		// Milk is not carried at this warehouse.
		bread["priceMinor"] = int64(1599)
		return []map[string]any{bread, yogurt}, nil
	case "store-bakirkoy":
		milk["available"] = int64(0)
		milk["outOfStock"] = true
		return []map[string]any{milk, bread}, nil
	default:
		return []map[string]any{milk, bread, yogurt}, nil
	}
}

func (s *Stubs) ForYou(_ context.Context, _, _ string) ([]map[string]any, error) {
	return []map[string]any{{"rail": "for_you", "skus": []string{"SKU1", "SKU2"}}}, nil
}

func (s *Stubs) Get(_ context.Context, _, cartID string) (map[string]any, error) {
	items := s.Carts[cartID]
	if items == nil {
		items = []map[string]any{}
	}
	return map[string]any{"cartId": cartID, "items": items}, nil
}

func (s *Stubs) AddItem(_ context.Context, _, cartID, sku string, qty, unitMinor int64) (map[string]any, error) {
	if cartID == "" {
		cartID = uuid.NewString()
	}
	if s.Carts == nil {
		s.Carts = map[string][]map[string]any{}
	}
	line := map[string]any{
		"id": sku, "sku": sku, "productId": sku, "product_id": sku,
		"name": sku, "title": sku, "quantity": qty, "qty": qty,
		"unit_price_minor": unitMinor, "unitPriceMinor": unitMinor, "priceMinor": unitMinor,
	}
	s.Carts[cartID] = append(s.Carts[cartID], line)
	return map[string]any{
		"cartId": cartID, "sku": sku, "qty": qty,
		"lineTotalMinor": qty * unitMinor, "items": s.Carts[cartID],
	}, nil
}

func (s *Stubs) Preview(_ context.Context, _, cartID string) (domain.CheckoutPreview, error) {
	return domain.CheckoutPreview{
		SessionID: "sess_" + cartID, CartID: cartID, Currency: "TRY", SubtotalMinor: 1999, TotalMinor: 1999, PaymentReady: true,
	}, nil
}

func (s *Stubs) Place(_ context.Context, _, cartID, _, sessionID string, _ domain.CheckoutAddress) (string, error) {
	id := "ord_" + uuid.NewString()[:8]
	s.Orders[cartID] = id
	if sessionID != "" {
		s.Orders[sessionID] = id
	}
	return id, nil
}

func (s *Stubs) GetOrder(_ context.Context, _, orderID string) (map[string]any, error) {
	return map[string]any{"orderId": orderID, "id": orderID, "status": "confirmed", "customerPrincipalId": "cust_1"}, nil
}

func (s *Stubs) ListOrders(_ context.Context, _, principalID string) ([]map[string]any, error) {
	return []map[string]any{{"orderId": "ord_1", "id": "ord_1", "status": "confirmed", "customerPrincipalId": principalID}}, nil
}

func (s *Stubs) CancelOrder(_ context.Context, _, orderID, _ string) (map[string]any, error) {
	return map[string]any{"orderId": orderID, "id": orderID, "status": "cancelled"}, nil
}

func (s *Stubs) Track(_ context.Context, _, orderID string) (domain.OrderTrack, error) {
	return domain.OrderTrack{OrderID: orderID, Status: "out_for_delivery", CourierID: "c1", ETASeconds: 720, Lat: 41.01, Lng: 28.97}, nil
}

func (s *Stubs) Serviceable(_ context.Context, _ string, lat, lng float64) (bool, error) {
	return lat != 0 && lng != 0, nil
}

func (s *Stubs) RegisterDevice(_ context.Context, _, _, _ string) error { return nil }

func (s *Stubs) OpenTicket(_ context.Context, _, _, subject string) (string, error) {
	return "tkt_" + strconv.Itoa(len(subject)), nil
}

func (s *Stubs) Submit(_ context.Context, _, _ string, _ int, _ string) error { return nil }

// OrderClient adapter name clash — wrap.
type OrderStub struct{ S *Stubs }

func (o OrderStub) Get(ctx context.Context, tenantID, orderID string) (map[string]any, error) {
	return o.S.GetOrder(ctx, tenantID, orderID)
}

func (o OrderStub) List(ctx context.Context, tenantID, principalID string) ([]map[string]any, error) {
	return o.S.ListOrders(ctx, tenantID, principalID)
}

func (o OrderStub) Cancel(ctx context.Context, tenantID, orderID, reason string) (map[string]any, error) {
	return o.S.CancelOrder(ctx, tenantID, orderID, reason)
}

func (o OrderStub) Refund(_ context.Context, _, orderID, reason string, amountMinor int64) (map[string]any, error) {
	return map[string]any{
		"orderId": orderID, "id": orderID, "status": "refunded",
		"reason": reason, "amountMinor": amountMinor,
	}, nil
}

func couponCatalog() []map[string]any {
	expired := time.Now().UTC().Add(-24 * time.Hour).Format(time.RFC3339)
	return []map[string]any{
		{
			"id": "welcome10", "code": "WELCOME10", "title": "Welcome 10%",
			"description": "10% off baskets over 150 TL", "discount_type": "percent",
			"discount_value": 10, "min_order_minor": 15000, "currency": "TRY",
			"active": true, "status": "active", "enabled": true,
		},
		{
			"id": "fresh50", "code": "FRESH50", "title": "Fresh 50 TL",
			"description": "50 TL off baskets over 200 TL", "discount_type": "fixed",
			"discount_value": 5000, "min_order_minor": 20000, "currency": "TRY",
			"active": true, "status": "active", "enabled": true,
		},
		{
			"id": "expired", "code": "EXPIRED", "title": "Expired coupon",
			"description": "Used only to test expiry", "discount_type": "percent",
			"discount_value": 20, "min_order_minor": 0, "currency": "TRY",
			"active": false, "status": "expired", "enabled": false,
			"expires_at": expired, "endsAt": expired,
		},
	}
}

func (s *Stubs) ListCoupons(_ context.Context, _ string) ([]map[string]any, error) {
	return couponCatalog(), nil
}

func (s *Stubs) GetCoupon(_ context.Context, _, code string) (map[string]any, error) {
	want := strings.ToUpper(strings.TrimSpace(code))
	for _, c := range couponCatalog() {
		if strings.ToUpper(asStubString(c["code"])) == want {
			return c, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (s *Stubs) EvaluateCoupon(_ context.Context, _, code string, cartSubtotalMinor int64) (map[string]any, error) {
	c, err := s.GetCoupon(context.Background(), "", code)
	if err != nil {
		return nil, err
	}
	if asStubString(c["status"]) == "expired" || c["active"] == false {
		return nil, domain.ErrConflict
	}
	min := asStubInt64(c["min_order_minor"])
	if cartSubtotalMinor > 0 && min > cartSubtotalMinor {
		return nil, domain.ErrInvalidArgument
	}
	discount := int64(0)
	if asStubString(c["discount_type"]) == "percent" {
		discount = cartSubtotalMinor * asStubInt64(c["discount_value"]) / 100
	} else {
		discount = asStubInt64(c["discount_value"])
	}
	return map[string]any{
		"totalDiscountMinor": discount,
		"discounts": []any{map[string]any{"couponCode": c["code"], "amountMinor": discount, "currency": "TRY"}},
	}, nil
}

func asStubString(v any) string {
	s, _ := v.(string)
	return s
}

func asStubInt64(v any) int64 {
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
