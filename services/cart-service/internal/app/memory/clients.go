package memory

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/cart-service/internal/app/ports"
	"github.com/nexora/cart-service/internal/domain"
)

// ErrInjectedFailure is returned when a client failure flag is set.
var ErrInjectedFailure = errors.New("injected client failure")

// PricingClient is a memory pricing client that returns deterministic quotes.
type PricingClient struct {
	mu sync.Mutex

	FailQuote  bool
	QuoteCalls atomic.Int64
	LastReq    *ports.QuoteRequest

	// UnitPriceMinor is the default unit price per variant (minor units).
	UnitPriceMinor int64
}

// NewPricingClient returns a succeeding pricing client.
func NewPricingClient() *PricingClient {
	return &PricingClient{UnitPriceMinor: 1000}
}

func (c *PricingClient) Quote(_ context.Context, req ports.QuoteRequest) (ports.QuoteResult, error) {
	c.QuoteCalls.Add(1)
	c.mu.Lock()
	defer c.mu.Unlock()
	cp := req
	c.LastReq = &cp
	if c.FailQuote {
		return ports.QuoteResult{}, fmt.Errorf("%w: quote", ErrInjectedFailure)
	}
	unit := c.UnitPriceMinor
	if unit <= 0 {
		unit = 1000
	}
	var subtotal int64
	lineQuotes := make([]domain.LineQuote, 0, len(req.Lines))
	for _, l := range req.Lines {
		lt := unit * int64(l.Qty)
		subtotal += lt
		lineQuotes = append(lineQuotes, domain.LineQuote{
			VariantID: l.VariantID, Qty: l.Qty,
			UnitPriceMinor: unit, LineTotalMinor: lt,
		})
	}
	var discount int64
	if len(req.CouponCodes) > 0 {
		discount = subtotal / 10 // 10% preview
	}
	tax := (subtotal - discount) / 20
	delivery := int64(1500)
	total := subtotal - discount + tax + delivery
	return ports.QuoteResult{
		QuoteID:       uuid.New(),
		Currency:      req.Currency,
		SubtotalMinor: subtotal,
		DiscountMinor: discount,
		TaxMinor:      tax,
		DeliveryMinor: delivery,
		TotalMinor:    total,
		LineQuotes:    lineQuotes,
		QuotedAt:      time.Now().UTC(),
	}, nil
}

var _ ports.PricingClient = (*PricingClient)(nil)

// InventoryClient is a memory inventory client with ATP + soft reserve.
type InventoryClient struct {
	mu sync.Mutex

	FailATP         bool
	FailSoftReserve bool
	FailRelease     bool

	ATPCalls         atomic.Int64
	SoftReserveCalls atomic.Int64
	ReleaseCalls     atomic.Int64

	// DefaultAvailable is returned for ATP when no override.
	DefaultAvailable int
	SeenSoftKeys     map[string]string
}

// NewInventoryClient returns a succeeding inventory client.
func NewInventoryClient() *InventoryClient {
	return &InventoryClient{
		DefaultAvailable: 50,
		SeenSoftKeys:     make(map[string]string),
	}
}

func (c *InventoryClient) ATP(_ context.Context, req ports.ATPRequest) (ports.ATPResult, error) {
	c.ATPCalls.Add(1)
	if c.FailATP {
		return ports.ATPResult{}, fmt.Errorf("%w: atp", ErrInjectedFailure)
	}
	avail := c.DefaultAvailable
	if avail <= 0 {
		avail = 50
	}
	lines := make([]ports.ATPLineResult, 0, len(req.Lines))
	for _, l := range req.Lines {
		lines = append(lines, ports.ATPLineResult{VariantID: l.VariantID, Available: avail})
	}
	return ports.ATPResult{Lines: lines}, nil
}

func (c *InventoryClient) SoftReserve(_ context.Context, req ports.SoftReserveRequest) (ports.SoftReserveResult, error) {
	c.SoftReserveCalls.Add(1)
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.FailSoftReserve {
		return ports.SoftReserveResult{}, fmt.Errorf("%w: soft reserve", ErrInjectedFailure)
	}
	if c.SeenSoftKeys == nil {
		c.SeenSoftKeys = make(map[string]string)
	}
	if ref, ok := c.SeenSoftKeys[req.IdempotencyKey]; ok {
		return ports.SoftReserveResult{ReservationRef: ref}, nil
	}
	ref := "res-" + uuid.NewString()
	c.SeenSoftKeys[req.IdempotencyKey] = ref
	exp := time.Now().UTC().Add(15 * time.Minute)
	return ports.SoftReserveResult{ReservationRef: ref, ExpiresAt: &exp}, nil
}

func (c *InventoryClient) Release(_ context.Context, _ ports.ReleaseRequest) error {
	c.ReleaseCalls.Add(1)
	if c.FailRelease {
		return fmt.Errorf("%w: release", ErrInjectedFailure)
	}
	return nil
}

var _ ports.InventoryClient = (*InventoryClient)(nil)

// RecommendClient returns deterministic recommendation stubs.
type RecommendClient struct {
	RecommendCalls atomic.Int64
	Items          []ports.RecommendItem
}

func (c *RecommendClient) Recommend(_ context.Context, req ports.RecommendRequest) (ports.RecommendResult, error) {
	c.RecommendCalls.Add(1)
	if len(c.Items) > 0 {
		return ports.RecommendResult{Items: c.Items}, nil
	}
	limit := req.Limit
	if limit <= 0 {
		limit = 3
	}
	items := make([]ports.RecommendItem, 0, limit)
	for i := 0; i < limit; i++ {
		items = append(items, ports.RecommendItem{
			VariantID: uuid.New(),
			Score:     1.0 - float64(i)*0.1,
			Reason:    "frequently_bought_together",
		})
	}
	return ports.RecommendResult{Items: items}, nil
}

var _ ports.RecommendClient = (*RecommendClient)(nil)
