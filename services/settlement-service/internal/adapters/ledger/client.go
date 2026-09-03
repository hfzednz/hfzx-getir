// Package ledger provides an HTTP LedgerClient for settlement-service.
//
// Behavior:
//   - Empty BaseURL → synthetic in-process success (unit tests / local stubs).
//   - Non-empty BaseURL → real HTTP POST to finance-ledger-service.
package ledger

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/nexora/settlement-service/internal/app/ports"
)

// Client posts settlement journals to finance-ledger-service when BaseURL is set.
// When BaseURL is empty, returns a synthetic success for local/unit tests
// (ledger is typically mocked in app tests via memory.LedgerClient).
type Client struct {
	BaseURL        string
	log            *slog.Logger
	HTTPClient     *http.Client
	internalToken  string
}

// NewClient returns a LedgerClient. Empty baseURL keeps synthetic success.
func NewClient(baseURL string, log *slog.Logger) *Client {
	if log == nil {
		log = slog.Default()
	}
	baseURL = strings.TrimRight(baseURL, "/")
	c := &Client{
		BaseURL:        baseURL,
		log:            log,
		HTTPClient:     &http.Client{Timeout: 15 * time.Second},
		internalToken:  strings.TrimSpace(os.Getenv("LEDGER_INTERNAL_TOKEN")),
	}
	if baseURL != "" {
		log.Info("ledger.client.http", "baseURL", baseURL)
	} else {
		log.Warn("ledger.client.synthetic", "reason", "LEDGER_BASE_URL empty; in-process success only")
	}
	return c
}

// PostSettlementJournal posts POST /v1/ledger/journals (or synthetic when BaseURL empty).
func (c *Client) PostSettlementJournal(ctx context.Context, req ports.LedgerPostRequest) (ports.LedgerPostResult, error) {
	if c.BaseURL == "" {
		c.log.Info("ledger.post.synthetic",
			"reference", req.Reference,
			"amountMinor", req.AmountMinor,
			"currency", req.Currency,
		)
		return ports.LedgerPostResult{JournalID: "stub-" + req.IdempotencyKey, Posted: true}, nil
	}

	body := map[string]any{
		"currency":       req.Currency,
		"reference":      req.Reference,
		"description":    "settlement",
		"idempotencyKey": req.IdempotencyKey,
		"lines": []map[string]any{
			{"accountCode": req.DebitAccount, "debitMinor": req.AmountMinor, "creditMinor": 0},
			{"accountCode": req.CreditAccount, "debitMinor": 0, "creditMinor": req.AmountMinor},
		},
	}
	var raw struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	if err := c.doJSON(ctx, http.MethodPost, "/v1/ledger/journals", req.TenantID.String(), req.IdempotencyKey, body, &raw); err != nil {
		return ports.LedgerPostResult{}, err
	}
	posted := strings.EqualFold(raw.Status, "posted") || raw.ID != ""
	return ports.LedgerPostResult{JournalID: raw.ID, Posted: posted}, nil
}

func (c *Client) doJSON(ctx context.Context, method, path, tenantID, idemKey string, in any, out any) error {
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
	if idemKey != "" {
		req.Header.Set("Idempotency-Key", idemKey)
	}
	if c.internalToken != "" {
		req.Header.Set("X-Ledger-Internal-Token", c.internalToken)
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode >= 300 {
		return fmt.Errorf("ledger http %s %s: status %d: %s", method, path, resp.StatusCode, string(raw))
	}
	if out == nil || len(raw) == 0 {
		return nil
	}
	return json.Unmarshal(raw, out)
}

var _ ports.LedgerClient = (*Client)(nil)
