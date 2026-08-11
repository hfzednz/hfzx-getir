// Package wallet provides an HTTP WalletClient for payment-service.
package wallet

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
	"github.com/nexora/payment-service/internal/app/memory"
	"github.com/nexora/payment-service/internal/app/ports"
)

// Client talks to wallet-service over HTTP when BaseURL is set.
type Client struct {
	BaseURL    string
	log        *slog.Logger
	inner      *memory.WalletClient
	HTTPClient *http.Client
}

// NewClient returns a wallet client. Empty baseURL keeps memory stub.
func NewClient(baseURL string, log *slog.Logger) *Client {
	if log == nil {
		log = slog.Default()
	}
	baseURL = strings.TrimRight(baseURL, "/")
	c := &Client{
		BaseURL:    baseURL,
		log:        log,
		HTTPClient: &http.Client{Timeout: 15 * time.Second},
		inner:      &memory.WalletClient{Success: true, EntryID: "wallet-entry-demo"},
	}
	if baseURL != "" {
		log.Info("wallet.client.http", "baseURL", baseURL)
	}
	return c
}

func (c *Client) Debit(ctx context.Context, req ports.WalletDebitRequest) (ports.WalletDebitResult, error) {
	if c.BaseURL == "" {
		return c.inner.Debit(ctx, req)
	}

	acctType := req.AccountType
	if acctType == "" {
		acctType = "cash"
	}
	currency := req.Currency
	if currency == "" {
		currency = "TRY"
	}

	var created struct {
		ID uuid.UUID `json:"id"`
	}
	if err := c.doJSON(ctx, http.MethodPost, "/v1/wallets", req.TenantID.String(), "", map[string]any{
		"principalId": req.PrincipalID,
		"currency":    currency,
	}, &created); err != nil {
		return ports.WalletDebitResult{Success: false, Reason: err.Error()}, err
	}

	var raw struct {
		Entry struct {
			ID uuid.UUID `json:"ID"`
		} `json:"entry"`
	}
	path := "/v1/wallets/" + created.ID.String() + "/debit"
	if err := c.doJSON(ctx, http.MethodPost, path, req.TenantID.String(), req.IdempotencyKey, map[string]any{
		"accountType":    acctType,
		"amountMinor":    req.AmountMinor,
		"idempotencyKey": req.IdempotencyKey,
		"reference":      req.Reference,
	}, &raw); err != nil {
		return ports.WalletDebitResult{Success: false, Reason: err.Error()}, err
	}
	entryID := raw.Entry.ID.String()
	if entryID == uuid.Nil.String() {
		entryID = req.IdempotencyKey
	}
	return ports.WalletDebitResult{
		WalletID: created.ID.String(),
		EntryID:  entryID,
		Success:  true,
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
		return fmt.Errorf("wallet http %s %s: status %d: %s", method, path, resp.StatusCode, string(raw))
	}
	if out == nil || len(raw) == 0 {
		return nil
	}
	return json.Unmarshal(raw, out)
}

var _ ports.WalletClient = (*Client)(nil)
