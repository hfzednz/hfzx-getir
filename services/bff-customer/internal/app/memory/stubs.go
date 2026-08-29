package memory

import (
	"context"
	"strconv"

	"github.com/google/uuid"
	"github.com/nexora/bff-customer/internal/domain"
)

type Stubs struct {
	Orders map[string]string
}

func NewStubs() *Stubs { return &Stubs{Orders: map[string]string{}} }

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
	return []map[string]any{{"sku": "SKU1", "name": query, "priceMinor": int64(1999)}}, nil
}

func (s *Stubs) ForYou(_ context.Context, _, _ string) ([]map[string]any, error) {
	return []map[string]any{{"rail": "for_you", "skus": []string{"SKU1", "SKU2"}}}, nil
}

func (s *Stubs) Get(_ context.Context, _, cartID string) (map[string]any, error) {
	return map[string]any{"cartId": cartID, "items": []any{}}, nil
}

func (s *Stubs) AddItem(_ context.Context, _, cartID, sku string, qty, unitMinor int64) (map[string]any, error) {
	return map[string]any{
		"cartId": cartID, "sku": sku, "qty": qty,
		"lineTotalMinor": qty * unitMinor,
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
	return map[string]any{"orderId": orderID, "status": "confirmed"}, nil
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
