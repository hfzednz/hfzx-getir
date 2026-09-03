// Package payout provides a bank/PSP PayoutClient for settlement-service.
//
// Behavior:
//   - Empty BaseURL → honest provider-unavailable (never fake a paid rail).
//   - Non-empty BaseURL → real HTTP POST to the payout provider.
package payout

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

	"github.com/nexora/settlement-service/internal/app/ports"
)

// Client executes payouts via an external PSP/provider when BaseURL is set.
// When BaseURL is empty, Execute reports provider_unavailable rather than
// fabricating a paid settlement.
type Client struct {
	BaseURL    string
	log        *slog.Logger
	HTTPClient *http.Client
}

// NewClient returns a PayoutClient. Empty baseURL keeps the rail unavailable.
func NewClient(baseURL string, log *slog.Logger) *Client {
	if log == nil {
		log = slog.Default()
	}
	baseURL = strings.TrimRight(baseURL, "/")
	c := &Client{
		BaseURL:    baseURL,
		log:        log,
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
	}
	if baseURL != "" {
		log.Info("payout.client.http", "baseURL", baseURL)
	} else {
		log.Warn("payout.client.unavailable", "reason", "PAYOUT_PROVIDER_URL empty")
	}
	return c
}

// Execute posts a payout to the provider, or reports unavailable when BaseURL is empty.
func (c *Client) Execute(ctx context.Context, req ports.PayoutRequest) (ports.PayoutResult, error) {
	if c.BaseURL == "" {
		c.log.Warn("payout.execute.unavailable",
			"instructionId", req.InstructionID.String(),
			"payeeType", string(req.PayeeType),
			"payeeRef", req.PayeeRef,
			"amountMinor", req.AmountMinor,
			"currency", req.Currency,
		)
		return ports.PayoutResult{
			Succeeded: false,
			Error:     "provider_unavailable",
		}, nil
	}

	body := map[string]any{
		"instructionId": req.InstructionID,
		"tenantId":      req.TenantID,
		"payeeType":     string(req.PayeeType),
		"payeeRef":      req.PayeeRef,
		"amountMinor":   req.AmountMinor,
		"currency":      req.Currency,
	}

	var raw struct {
		ProviderRef string `json:"providerRef"`
		Succeeded   *bool  `json:"succeeded"`
		Success     *bool  `json:"success"`
		Status      string `json:"status"`
		Error       string `json:"error"`
		ID          string `json:"id"`
	}
	// Canonical provider path; some PSPs expose /v1/payments/payout equivalently.
	if err := c.doJSON(ctx, http.MethodPost, "/v1/payouts", req.TenantID.String(), body, &raw); err != nil {
		return ports.PayoutResult{Succeeded: false, Error: err.Error()}, err
	}

	ref := raw.ProviderRef
	if ref == "" {
		ref = raw.ID
	}
	succeeded := false
	switch {
	case raw.Succeeded != nil:
		succeeded = *raw.Succeeded
	case raw.Success != nil:
		succeeded = *raw.Success
	case strings.EqualFold(raw.Status, "succeeded"), strings.EqualFold(raw.Status, "success"), strings.EqualFold(raw.Status, "completed"):
		succeeded = true
	case ref != "" && raw.Error == "":
		succeeded = true
	}

	return ports.PayoutResult{
		ProviderRef: ref,
		Succeeded:   succeeded,
		Error:       raw.Error,
	}, nil
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
		return fmt.Errorf("payout http %s %s: status %d: %s", method, path, resp.StatusCode, string(raw))
	}
	if out == nil || len(raw) == 0 {
		return nil
	}
	return json.Unmarshal(raw, out)
}

var _ ports.PayoutClient = (*Client)(nil)
