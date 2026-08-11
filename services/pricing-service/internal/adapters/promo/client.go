// Package promo provides an HTTP PromoClient for pricing-service.
package promo

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

	"github.com/nexora/pricing-service/internal/app"
	"github.com/nexora/pricing-service/internal/app/ports"
)

// Client talks to promotion-service Evaluate when BaseURL is set.
type Client struct {
	BaseURL    string
	log        *slog.Logger
	inner      ports.PromoClient
	HTTPClient *http.Client
}

// NewClient returns a promo client. Empty baseURL keeps noop.
func NewClient(baseURL string, log *slog.Logger) *Client {
	if log == nil {
		log = slog.Default()
	}
	baseURL = strings.TrimRight(baseURL, "/")
	c := &Client{
		BaseURL:    baseURL,
		log:        log,
		inner:      app.NoopPromoClient{},
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}
	if baseURL != "" {
		log.Info("promo.client.http", "baseURL", baseURL)
	}
	return c
}

func (c *Client) Evaluate(ctx context.Context, req ports.PromoEvaluateRequest) (ports.PromoEvaluateResult, error) {
	if c.BaseURL == "" {
		return c.inner.Evaluate(ctx, req)
	}
	lines := make([]map[string]any, 0, len(req.Lines))
	for i, l := range req.Lines {
		lines = append(lines, map[string]any{
			"lineId":         fmt.Sprintf("line-%d", i),
			"variantId":      l.VariantID.String(),
			"quantity":       l.Qty,
			"unitPriceMinor": l.UnitPriceMinor,
		})
	}
	principal := ""
	if req.CustomerID != nil {
		principal = req.CustomerID.String()
	}
	body := map[string]any{
		"principalId": principal,
		"currency":    req.Currency,
		"couponCodes": req.CouponCodes,
		"lines":       lines,
	}
	var raw struct {
		TotalDiscountMinor int64 `json:"totalDiscountMinor"`
		Discounts          []struct {
			PromotionID   string `json:"promotionId"`
			CouponCode    string `json:"couponCode"`
			AmountMinor   int64  `json:"amountMinor"`
			Description   string `json:"description"`
			AppliedLineIDs []string `json:"appliedLineIds"`
		} `json:"discounts"`
	}
	if err := c.doJSON(ctx, http.MethodPost, "/v1/promo/evaluate", req.TenantID.String(), body, &raw); err != nil {
		c.log.Warn("promo.evaluate.fallback", "err", err)
		return c.inner.Evaluate(ctx, req)
	}
	out := ports.PromoEvaluateResult{DiscountMinor: raw.TotalDiscountMinor}
	for _, d := range raw.Discounts {
		item := ports.PromoDiscountResult{
			PromotionID: d.PromotionID, Code: d.CouponCode, DiscountMinor: d.AmountMinor, Description: d.Description,
		}
		out.Discounts = append(out.Discounts, item)
	}
	if out.DiscountMinor == 0 {
		for _, d := range out.Discounts {
			out.DiscountMinor += d.DiscountMinor
		}
	}
	return out, nil
}

func (c *Client) doJSON(ctx context.Context, method, path, tenantID string, in any, out any) error {
	b, err := json.Marshal(in)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if tenantID != "" {
		req.Header.Set("X-Tenant-Id", tenantID)
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode >= 300 {
		return fmt.Errorf("promo http %s %s: status %d: %s", method, path, resp.StatusCode, string(raw))
	}
	if out == nil || len(raw) == 0 {
		return nil
	}
	return json.Unmarshal(raw, out)
}

var _ ports.PromoClient = (*Client)(nil)
