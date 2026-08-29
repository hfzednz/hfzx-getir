// Package cart provides a CartClient HTTP adapter.
package cart

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/checkout-service/internal/app/ports"
	"github.com/nexora/checkout-service/internal/domain"
)

// Client talks to cart-service over HTTP when baseURL is set and inner is nil.
type Client struct {
	baseURL    string
	log        *slog.Logger
	inner      ports.CartClient
	HTTPClient *http.Client
}

// NewClient returns a cart client. When inner is non-nil it is used (memory/dev).
func NewClient(baseURL string, log *slog.Logger, inner ports.CartClient) *Client {
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
		log.Info("cart.client.http", "baseURL", baseURL)
	}
	return c
}

// GetCart fetches GET /v1/cart/{id} or delegates to inner.
func (c *Client) GetCart(ctx context.Context, tenantID, cartID uuid.UUID) (ports.CartView, error) {
	if c.baseURL == "" {
		if c.inner != nil {
			return c.inner.GetCart(ctx, tenantID, cartID)
		}
		c.log.Warn("cart.client.unconfigured", "cartId", cartID)
		return ports.CartView{}, fmt.Errorf("%w: cart client unconfigured", domain.ErrInvariant)
	}

	var raw struct {
		ID          uuid.UUID  `json:"ID"`
		TenantID    uuid.UUID  `json:"TenantID"`
		GuestToken  string     `json:"GuestToken"`
		PrincipalID *uuid.UUID `json:"PrincipalID"`
		CityID      *uuid.UUID `json:"CityID"`
		Status      string     `json:"Status"`
		Currency    string     `json:"Currency"`
		Lines       []struct {
			VariantID uuid.UUID `json:"VariantID"`
			Qty       int       `json:"Qty"`
			Notes     string    `json:"Notes"`
		} `json:"Lines"`
		Coupons []struct {
			Code string `json:"Code"`
		} `json:"Coupons"`
		Quote *struct {
			LineQuotes []struct {
				VariantID      uuid.UUID `json:"variantId"`
				UnitPriceMinor int64     `json:"unitPriceMinor"`
			} `json:"LineQuotes"`
		} `json:"Quote"`
	}
	path := "/v1/cart/" + url.PathEscape(cartID.String())
	if err := c.doJSON(ctx, http.MethodGet, path, tenantID.String(), "", nil, &raw); err != nil {
		return ports.CartView{}, err
	}
	if raw.ID == uuid.Nil || raw.TenantID != tenantID {
		return ports.CartView{}, domain.ErrNotFound
	}

	unitByVariant := map[uuid.UUID]int64{}
	if raw.Quote != nil {
		for _, lq := range raw.Quote.LineQuotes {
			unitByVariant[lq.VariantID] = lq.UnitPriceMinor
		}
	}

	lines := make([]ports.CartLine, 0, len(raw.Lines))
	for _, l := range raw.Lines {
		lines = append(lines, ports.CartLine{
			VariantID: l.VariantID, Qty: l.Qty, Notes: l.Notes,
			UnitPriceMinor: unitByVariant[l.VariantID],
		})
	}
	codes := make([]string, 0, len(raw.Coupons))
	for _, cp := range raw.Coupons {
		if cp.Code != "" {
			codes = append(codes, cp.Code)
		}
	}
	view := ports.CartView{
		ID: raw.ID, TenantID: raw.TenantID, GuestID: raw.GuestToken,
		Currency: raw.Currency, CouponCodes: codes, Lines: lines,
		Active: strings.EqualFold(raw.Status, "active"),
	}
	if raw.PrincipalID != nil {
		view.PrincipalID = *raw.PrincipalID
	}
	if raw.CityID != nil {
		view.CityID = raw.CityID.String()
	}
	return view, nil
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
	if resp.StatusCode == http.StatusNotFound {
		return domain.ErrNotFound
	}
	if resp.StatusCode >= 300 {
		return fmt.Errorf("cart http %s %s: status %d: %s", method, path, resp.StatusCode, string(raw))
	}
	if out == nil || len(raw) == 0 {
		return nil
	}
	return json.Unmarshal(raw, out)
}

var _ ports.CartClient = (*Client)(nil)
