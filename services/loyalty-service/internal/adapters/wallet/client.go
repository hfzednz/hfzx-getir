// Package wallet provides an HTTP WalletClient for loyalty-service.
//
// Behavior:
//   - Empty BaseURL → synthetic in-process success (unit tests / local stubs).
//   - Non-empty BaseURL → ensure wallet via POST /v1/wallets, then POST credit.
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
	"github.com/nexora/loyalty-service/internal/app/ports"
)

// Client credits cashback/promo accounts on wallet-service when BaseURL is set.
// When BaseURL is empty, returns synthetic success for local/unit tests.
type Client struct {
	BaseURL    string
	Log        *slog.Logger
	HTTPClient *http.Client
}

// NewClient returns a WalletClient. Empty baseURL keeps synthetic success.
func NewClient(baseURL string, log *slog.Logger) *Client {
	if log == nil {
		log = slog.Default()
	}
	baseURL = strings.TrimRight(baseURL, "/")
	c := &Client{
		BaseURL:    baseURL,
		Log:        log,
		HTTPClient: &http.Client{Timeout: 15 * time.Second},
	}
	if baseURL != "" {
		log.Info("wallet.client.http", "baseURL", baseURL)
	} else {
		log.Warn("wallet.client.synthetic", "reason", "WALLET_BASE_URL empty; in-process success only")
	}
	return c
}

// Credit ensures a wallet for the principal then posts a credit to wallet-service.
func (c *Client) Credit(ctx context.Context, req ports.WalletCreditRequest) (ports.WalletCreditResult, error) {
	if c.BaseURL == "" {
		c.Log.Info("wallet.credit.synthetic",
			"principalId", req.PrincipalID.String(),
			"amountMinor", req.AmountMinor,
			"accountType", req.AccountType,
			"idempotencyKey", req.IdempotencyKey,
		)
		return ports.WalletCreditResult{
			WalletID: "stub-wallet",
			EntryID:  "stub-entry:" + req.IdempotencyKey,
			Credited: true,
		}, nil
	}

	acctType := req.AccountType
	if acctType == "" {
		acctType = "cashback"
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
		return ports.WalletCreditResult{}, err
	}

	var raw struct {
		Entry struct {
			ID uuid.UUID `json:"ID"`
		} `json:"entry"`
	}
	path := "/v1/wallets/" + created.ID.String() + "/credit"
	if err := c.doJSON(ctx, http.MethodPost, path, req.TenantID.String(), req.IdempotencyKey, map[string]any{
		"accountType":    acctType,
		"amountMinor":    req.AmountMinor,
		"idempotencyKey": req.IdempotencyKey,
		"reference":      req.Reference,
	}, &raw); err != nil {
		return ports.WalletCreditResult{}, err
	}

	entryID := raw.Entry.ID.String()
	if entryID == uuid.Nil.String() {
		entryID = req.IdempotencyKey
	}
	return ports.WalletCreditResult{
		WalletID: created.ID.String(),
		EntryID:  entryID,
		Credited: true,
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
