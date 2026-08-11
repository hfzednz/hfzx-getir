// Package promo provides an HTTP PromoClient for checkout-service.
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

	"github.com/nexora/checkout-service/internal/app/ports"
)

// Client talks to promotion-service Evaluate when baseURL is set.
// Empty baseURL uses inner when non-nil; otherwise allows with CodesApplied.
type Client struct {
	baseURL    string
	log        *slog.Logger
	inner      ports.PromoClient
	HTTPClient *http.Client
}

// NewClient returns a promo client. Empty baseURL keeps inner / allow stub.
func NewClient(baseURL string, log *slog.Logger, inner ports.PromoClient) *Client {
	if log == nil {
		log = slog.Default()
	}
	baseURL = strings.TrimRight(baseURL, "/")
	c := &Client{
		baseURL:    baseURL,
		log:        log,
		inner:      inner,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}
	if baseURL != "" {
		log.Info("promo.client.http", "baseURL", baseURL)
	}
	return c
}

// Validate posts POST /v1/promo/evaluate when baseURL is set.
func (c *Client) Validate(ctx context.Context, req ports.PromoRequest) (ports.PromoResult, error) {
	if c.baseURL == "" {
		if c.inner != nil {
			return c.inner.Validate(ctx, req)
		}
		c.log.Debug("promo.fallback.dev", "note", "PROMO_URL empty; using inner/allow")
		return ports.PromoResult{Valid: true, CodesApplied: req.Codes}, nil
	}

	lines := []map[string]any{}
	if req.SubtotalMinor > 0 {
		lines = append(lines, map[string]any{
			"lineId":         "checkout",
			"variantId":      "",
			"quantity":       1,
			"unitPriceMinor": req.SubtotalMinor,
		})
	}
	body := map[string]any{
		"principalId": req.PrincipalID.String(),
		"currency":    req.Currency,
		"couponCodes": req.Codes,
		"lines":       lines,
	}
	var raw struct {
		TotalDiscountMinor int64 `json:"totalDiscountMinor"`
		Discounts          []struct {
			PromotionID string `json:"promotionId"`
			CouponCode  string `json:"couponCode"`
			AmountMinor int64  `json:"amountMinor"`
			Description string `json:"description"`
		} `json:"discounts"`
	}
	if err := c.doJSON(ctx, http.MethodPost, "/v1/promo/evaluate", req.TenantID.String(), body, &raw); err != nil {
		c.log.Warn("promo.evaluate.fallback", "err", err)
		if c.inner != nil {
			return c.inner.Validate(ctx, req)
		}
		return ports.PromoResult{}, err
	}

	discount := raw.TotalDiscountMinor
	codes := make([]string, 0, len(raw.Discounts))
	var sumLines int64
	for _, d := range raw.Discounts {
		sumLines += d.AmountMinor
		if d.CouponCode != "" {
			codes = append(codes, d.CouponCode)
		}
	}
	if discount == 0 {
		discount = sumLines
	}
	if len(codes) == 0 {
		codes = req.Codes
	}
	return ports.PromoResult{
		Valid:         true,
		DiscountMinor: discount,
		CodesApplied:  codes,
	}, nil
}

func (c *Client) doJSON(ctx context.Context, method, path, tenantID string, in any, out any) error {
	b, err := json.Marshal(in)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bytes.NewReader(b))
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
