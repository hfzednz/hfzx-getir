package memory

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/checkout-service/internal/app/ports"
	"github.com/nexora/checkout-service/internal/domain"
)

// CartClient serves carts from the in-memory store.
type CartClient struct{ S *Store }

func (c *CartClient) GetCart(_ context.Context, tenantID, cartID uuid.UUID) (ports.CartView, error) {
	c.S.mu.RLock()
	defer c.S.mu.RUnlock()
	cart, ok := c.S.Carts[cartID]
	if !ok || cart.TenantID != tenantID {
		return ports.CartView{}, domain.ErrNotFound
	}
	cp := cart
	cp.Lines = append([]ports.CartLine(nil), cart.Lines...)
	cp.CouponCodes = append([]string(nil), cart.CouponCodes...)
	return cp, nil
}

// PricingClient returns a simple sum-based quote.
type PricingClient struct {
	Fail bool
	Calls atomic.Int64
}

func (p *PricingClient) Quote(_ context.Context, req ports.QuoteRequest) (ports.QuoteResult, error) {
	p.Calls.Add(1)
	if p.Fail {
		return ports.QuoteResult{}, fmt.Errorf("%w: pricing unavailable", domain.ErrInvariant)
	}
	var sub int64
	for _, l := range req.Lines {
		// Default unit 1000 minor if unknown — tests seed unit via qty*1000.
		sub += int64(l.Qty) * 1000
	}
	delivery := int64(1500)
	tax := sub / 10
	tip := req.TipMinor
	total := sub + delivery + tax + tip
	return ports.QuoteResult{
		QuoteID: "q-" + req.CheckoutID.String()[:8],
		Currency: func() string {
			if req.Currency != "" {
				return req.Currency
			}
			return "TRY"
		}(),
		SubtotalMinor: sub, TaxMinor: tax, DeliveryMinor: delivery,
		TipMinor: tip, TotalMinor: total, QuotedAt: time.Now().UTC(),
	}, nil
}

// InventoryClient ATP check with injectable failure.
type InventoryClient struct {
	FailAll bool
	Calls   atomic.Int64
}

func NewInventoryClient() *InventoryClient { return &InventoryClient{} }

func (i *InventoryClient) CheckATP(_ context.Context, req ports.ATPRequest) (ports.ATPResult, error) {
	i.Calls.Add(1)
	out := ports.ATPResult{AllAvailable: !i.FailAll, Lines: make([]ports.ATPLineResult, 0, len(req.Lines))}
	for _, l := range req.Lines {
		lr := ports.ATPLineResult{VariantID: l.VariantID, Available: !i.FailAll, AvailableQty: l.Qty}
		if i.FailAll {
			lr.AvailableQty = 0
			lr.Reason = "out of stock"
		}
		out.Lines = append(out.Lines, lr)
	}
	return out, nil
}

// GeofenceClient zone check with injectable failure.
type GeofenceClient struct {
	FailInZone bool
	MinOrder   int64
	Calls      atomic.Int64
}

func (g *GeofenceClient) CheckZone(_ context.Context, _ ports.GeofenceRequest) (ports.GeofenceResult, error) {
	g.Calls.Add(1)
	if g.FailInZone {
		return ports.GeofenceResult{InZone: false, Reason: "outside zone"}, nil
	}
	return ports.GeofenceResult{InZone: true, ZoneID: "zone-1", MinOrderMinor: g.MinOrder}, nil
}

// FraudClient risk scoring.
type FraudClient struct {
	Block bool
	Calls atomic.Int64
}

func (f *FraudClient) Score(_ context.Context, _ ports.FraudRequest) (ports.FraudResult, error) {
	f.Calls.Add(1)
	if f.Block {
		return ports.FraudResult{Score: 0.99, Decision: "block", Reason: "high risk"}, nil
	}
	return ports.FraudResult{Score: 0.1, Decision: "allow"}, nil
}

// PaymentEligibilityClient eligibility without capture.
type PaymentEligibilityClient struct {
	Ineligible bool
	Calls      atomic.Int64
}

func (p *PaymentEligibilityClient) Check(_ context.Context, _ ports.PaymentEligibilityRequest) (ports.PaymentEligibilityResult, error) {
	p.Calls.Add(1)
	if p.Ineligible {
		return ports.PaymentEligibilityResult{Eligible: false, Reason: "no methods"}, nil
	}
	return ports.PaymentEligibilityResult{Eligible: true, Methods: []string{"card", "wallet"}}, nil
}

// OrderClient CreateFromCheckout stub.
type OrderClient struct {
	FailCreate bool
	Calls      atomic.Int64
	LastReq    ports.CreateFromCheckoutRequest
}

func (o *OrderClient) CreateFromCheckout(_ context.Context, req ports.CreateFromCheckoutRequest) (ports.CreateFromCheckoutResult, error) {
	o.Calls.Add(1)
	o.LastReq = req
	if o.FailCreate {
		return ports.CreateFromCheckoutResult{}, fmt.Errorf("%w: order create failed", domain.ErrInvariant)
	}
	return ports.CreateFromCheckoutResult{
		OrderID: "ord-" + req.CheckoutSessionID.String()[:8],
		Status:  "pending_payment",
	}, nil
}

func (o *OrderClient) PlaceOrder(_ context.Context, req ports.PlaceOrderRequest) (ports.PlaceOrderResult, error) {
	return ports.PlaceOrderResult{OrderID: req.OrderID, Status: "warehouse_assigned"}, nil
}

// PromoClient coupon validation.
type PromoClient struct {
	Invalid bool
	Calls   atomic.Int64
}

func (p *PromoClient) Validate(_ context.Context, req ports.PromoRequest) (ports.PromoResult, error) {
	p.Calls.Add(1)
	if p.Invalid {
		return ports.PromoResult{Valid: false, Reason: "expired"}, nil
	}
	return ports.PromoResult{Valid: true, CodesApplied: req.Codes, DiscountMinor: 100}, nil
}

// CustomerClient principal check.
type CustomerClient struct {
	Inactive bool
	Age      int
	Calls    atomic.Int64
}

func (c *CustomerClient) Check(_ context.Context, _ ports.CustomerCheckRequest) (ports.CustomerCheckResult, error) {
	c.Calls.Add(1)
	age := c.Age
	if age == 0 {
		age = 25
	}
	if c.Inactive {
		return ports.CustomerCheckResult{Active: false, Reason: "suspended", Age: age}, nil
	}
	return ports.CustomerCheckResult{Active: true, Age: age}, nil
}

var _ ports.CartClient = (*CartClient)(nil)
var _ ports.PricingClient = (*PricingClient)(nil)
var _ ports.InventoryClient = (*InventoryClient)(nil)
var _ ports.GeofenceClient = (*GeofenceClient)(nil)
var _ ports.FraudClient = (*FraudClient)(nil)
var _ ports.PaymentEligibilityClient = (*PaymentEligibilityClient)(nil)
var _ ports.OrderClient = (*OrderClient)(nil)
var _ ports.PromoClient = (*PromoClient)(nil)
var _ ports.CustomerClient = (*CustomerClient)(nil)
