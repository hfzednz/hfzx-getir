// Package payment provides a PaymentEligibilityClient HTTP adapter (no capture).
package payment

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

// Client talks to payment-service eligibility API when baseURL is set and inner is nil.
type Client struct {
	baseURL    string
	log        *slog.Logger
	inner      ports.PaymentEligibilityClient
	HTTPClient *http.Client
}

// NewClient returns a payment eligibility client (never captures).
func NewClient(baseURL string, log *slog.Logger, inner ports.PaymentEligibilityClient) *Client {
	if log == nil {
		log = slog.Default()
	}
	baseURL = strings.TrimRight(baseURL, "/")
	c := &Client{
		baseURL:    baseURL,
		log:        log,
		inner:      inner,
		HTTPClient: &http.Client{Timeout: 15 * time.Second},
	}
	if inner == nil && baseURL != "" {
		log.Info("payment.eligibility.http", "baseURL", baseURL)
	}
	return c
}

// Check posts POST /v1/payments/eligibility.
func (c *Client) Check(ctx context.Context, req ports.PaymentEligibilityRequest) (ports.PaymentEligibilityResult, error) {
	if c.inner != nil {
		return c.inner.Check(ctx, req)
	}
	if c.baseURL == "" {
		c.log.Warn("payment.eligibility.unconfigured")
		return ports.PaymentEligibilityResult{Eligible: true, Methods: []string{"card"}}, nil
	}

	body := map[string]any{
		"principalId": req.PrincipalID,
		"amountMinor": req.TotalMinor,
		"currency":    req.Currency,
	}
	var raw struct {
		Eligible bool     `json:"Eligible"`
		Reason   string   `json:"Reason"`
		Methods  []string `json:"Methods"`
	}
	if err := c.doJSON(ctx, http.MethodPost, "/v1/payments/eligibility", req.TenantID.String(), "", body, &raw); err != nil {
		return ports.PaymentEligibilityResult{}, err
	}
	return ports.PaymentEligibilityResult{
		Eligible: raw.Eligible, Reason: raw.Reason, Methods: raw.Methods,
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
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return err
	}
	if in != nil {
		req.Header.Set("Content-Type", "application/json")
	}
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
		return fmt.Errorf("payment http %s %s: status %d: %s", method, path, resp.StatusCode, string(raw))
	}
	if out == nil || len(raw) == 0 {
		return nil
	}
	return json.Unmarshal(raw, out)
}

var _ ports.PaymentEligibilityClient = (*Client)(nil)
