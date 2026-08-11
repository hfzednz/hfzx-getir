// Package pricing provides a PricingClient HTTP adapter.
package pricing

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/cart-service/internal/app/memory"
	"github.com/nexora/cart-service/internal/app/ports"
	"github.com/nexora/cart-service/internal/domain"
)

// Client talks to pricing-service over HTTP when BaseURL is set; otherwise memory.
type Client struct {
	BaseURL    string
	log        *slog.Logger
	inner      *memory.PricingClient
	HTTPClient *http.Client
}

// NewClient returns a pricing client. Empty baseURL keeps in-process memory.
func NewClient(baseURL string, log *slog.Logger) *Client {
	if log == nil {
		log = slog.Default()
	}
	baseURL = strings.TrimRight(baseURL, "/")
	c := &Client{
		BaseURL:    baseURL,
		log:        log,
		HTTPClient: &http.Client{Timeout: 15 * time.Second},
	}
	if baseURL == "" {
		c.inner = memory.NewPricingClient()
	} else {
		log.Info("pricing.client.http", "baseURL", baseURL)
	}
	return c
}

// Quote calls pricing-service POST /v1/pricing/quote (or memory when BaseURL empty).
func (c *Client) Quote(ctx context.Context, req ports.QuoteRequest) (ports.QuoteResult, error) {
	if c.BaseURL == "" {
		return c.inner.Quote(ctx, req)
	}

	lines := make([]map[string]any, 0, len(req.Lines))
	for _, l := range req.Lines {
		lines = append(lines, map[string]any{
			"variantId": l.VariantID,
			"qty":       l.Qty,
		})
	}
	body := map[string]any{
		"cartId":      req.CartID,
		"currency":    req.Currency,
		"couponCodes": req.CouponCodes,
		"lines":       lines,
	}
	if req.CityID != nil {
		body["regionId"] = *req.CityID
	}

	var raw struct {
		ID             uuid.UUID `json:"id"`
		Currency       string    `json:"currency"`
		SubtotalMinor  int64     `json:"subtotalMinor"`
		DiscountMinor  int64     `json:"discountMinor"`
		TaxMinor       int64     `json:"taxMinor"`
		DeliveryMinor  int64     `json:"deliveryMinor"`
		ServiceMinor   int64     `json:"serviceMinor"`
		PackagingMinor int64     `json:"packagingMinor"`
		TipMinor       int64     `json:"tipMinor"`
		TotalMinor     int64     `json:"totalMinor"`
		QuotedAt       time.Time `json:"quotedAt"`
		Lines          []struct {
			VariantID      uuid.UUID `json:"variantId"`
			Qty            int       `json:"qty"`
			UnitPriceMinor int64     `json:"unitPriceMinor"`
			LineTotalMinor int64     `json:"lineTotalMinor"`
			DiscountMinor  int64     `json:"discountMinor"`
		} `json:"lines"`
	}
	if err := c.doJSON(ctx, http.MethodPost, "/v1/pricing/quote", req.TenantID.String(), "", body, &raw); err != nil {
		return ports.QuoteResult{}, err
	}

	lineQuotes := make([]domain.LineQuote, 0, len(raw.Lines))
	for _, ln := range raw.Lines {
		lineQuotes = append(lineQuotes, domain.LineQuote{
			VariantID: ln.VariantID, Qty: ln.Qty,
			UnitPriceMinor: ln.UnitPriceMinor, LineTotalMinor: ln.LineTotalMinor,
			DiscountMinor: ln.DiscountMinor,
		})
	}
	quotedAt := raw.QuotedAt
	if quotedAt.IsZero() {
		quotedAt = time.Now().UTC()
	}
	return ports.QuoteResult{
		QuoteID: raw.ID, Currency: raw.Currency,
		SubtotalMinor: raw.SubtotalMinor, DiscountMinor: raw.DiscountMinor,
		TaxMinor: raw.TaxMinor, DeliveryMinor: raw.DeliveryMinor,
		ServiceMinor: raw.ServiceMinor, PackagingMinor: raw.PackagingMinor,
		TipMinor: raw.TipMinor, TotalMinor: raw.TotalMinor,
		LineQuotes: lineQuotes, QuotedAt: quotedAt,
	}, nil
}

func (c *Client) doJSON(ctx context.Context, method, path, tenantID, idemKey string, in any, out any) error {
	var body io.Reader
	if in != nil {
		b, err := json.Marshal(in)
		if err != nil {
			return err
		}
		body = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if tenantID != "" {
		req.Header.Set("X-Tenant-Id", tenantID)
	}
	if idemKey != "" {
		req.Header.Set("Idempotency-Key", idemKey)
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode >= 300 {
		return fmt.Errorf("pricing http %s %s: status %d: %s", method, path, resp.StatusCode, string(raw))
	}
	if out == nil || len(raw) == 0 {
		return nil
	}
	return json.Unmarshal(raw, out)
}

var _ ports.PricingClient = (*Client)(nil)
