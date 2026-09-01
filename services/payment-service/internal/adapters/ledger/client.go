// Package ledger provides an HTTP LedgerClient for payment-service.
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

	"github.com/nexora/payment-service/internal/app/memory"
	"github.com/nexora/payment-service/internal/app/ports"
)

// Client posts journals to finance-ledger-service when BaseURL is set.
type Client struct {
	BaseURL    string
	log        *slog.Logger
	inner      *memory.LedgerClient
	HTTPClient *http.Client
}

// NewClient returns a ledger client. Empty baseURL keeps memory stub.
func NewClient(baseURL string, log *slog.Logger) *Client {
	if log == nil {
		log = slog.Default()
	}
	baseURL = strings.TrimRight(baseURL, "/")
	c := &Client{
		BaseURL:    baseURL,
		log:        log,
		HTTPClient: &http.Client{Timeout: 15 * time.Second},
		inner:      &memory.LedgerClient{},
	}
	if baseURL != "" {
		log.Info("ledger.client.http", "baseURL", baseURL)
	}
	return c
}

func (c *Client) PostJournal(ctx context.Context, req ports.PostJournalRequest) (ports.PostJournalResult, error) {
	if c.BaseURL == "" {
		return c.inner.PostJournal(ctx, req)
	}
	lines := make([]map[string]any, 0, len(req.Lines))
	for _, l := range req.Lines {
		lines = append(lines, map[string]any{
			"accountCode": l.AccountCode,
			"debitMinor":  l.DebitMinor,
			"creditMinor": l.CreditMinor,
		})
	}
	currency := "TRY"
	if len(req.Lines) > 0 && req.Lines[0].Currency != "" {
		currency = req.Lines[0].Currency
	}
	body := map[string]any{
		"currency":       currency,
		"reference":      req.Reference,
		"idempotencyKey": req.IdempotencyKey,
		"lines":          lines,
	}
	var raw struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	if err := c.doJSON(ctx, http.MethodPost, "/v1/ledger/journals", req.TenantID.String(), req.IdempotencyKey, body, &raw); err != nil {
		return ports.PostJournalResult{}, err
	}
	return ports.PostJournalResult{JournalID: raw.ID, Posted: raw.Status == "posted" || raw.ID != ""}, nil
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
	if tok := strings.TrimSpace(os.Getenv("LEDGER_INTERNAL_TOKEN")); tok != "" {
		req.Header.Set("X-Ledger-Internal-Token", tok)
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
		return fmt.Errorf("ledger http %s %s: status %d: %s", method, path, resp.StatusCode, string(raw))
	}
	if out == nil || len(raw) == 0 {
		return nil
	}
	return json.Unmarshal(raw, out)
}

var _ ports.LedgerClient = (*Client)(nil)
