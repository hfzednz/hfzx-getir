// Package order provides an OrderClient HTTP adapter (CreateFromCheckout / PlaceOrder).
package order

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

	"github.com/nexora/checkout-service/internal/app/ports"
	"github.com/nexora/checkout-service/internal/domain"
)

// Client talks to order-service over HTTP when baseURL is set and inner is nil.
type Client struct {
	baseURL    string
	log        *slog.Logger
	inner      ports.OrderClient
	HTTPClient *http.Client
}

// NewClient returns an order client.
func NewClient(baseURL string, log *slog.Logger, inner ports.OrderClient) *Client {
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
		log.Info("order.client.http", "baseURL", baseURL)
	}
	return c
}

// CreateFromCheckout posts POST /v1/orders with fromCheckout=true.
func (c *Client) CreateFromCheckout(ctx context.Context, req ports.CreateFromCheckoutRequest) (ports.CreateFromCheckoutResult, error) {
	if c.baseURL == "" {
		if c.inner != nil {
			return c.inner.CreateFromCheckout(ctx, req)
		}
		c.log.Warn("order.client.unconfigured")
		return ports.CreateFromCheckoutResult{}, fmt.Errorf("%w: order client unconfigured", domain.ErrInvariant)
	}

	lines := make([]map[string]any, 0, len(req.Lines))
	for _, l := range req.Lines {
		ln := map[string]any{
			"variantId": l.VariantID, "skuCode": l.SKUCode, "titleSnapshot": l.TitleSnapshot,
			"qty": l.Qty, "unitPriceMinor": l.UnitPriceMinor,
			"discountsMinor": l.DiscountsMinor, "taxMinor": l.TaxMinor,
		}
		if l.WarehouseID != nil {
			ln["warehouseId"] = *l.WarehouseID
		}
		lines = append(lines, ln)
	}
	meta := req.Metadata
	if meta == nil {
		meta = map[string]any{}
	}
	meta["checkoutSessionId"] = req.CheckoutSessionID.String()
	meta["cartId"] = req.CartID.String()
	meta["deliveryOption"] = string(req.DeliveryOption)
	if req.ScheduledAt != nil {
		meta["scheduledAt"] = req.ScheduledAt.UTC()
	}

	body := map[string]any{
		"customerPrincipalId": req.CustomerPrincipalID,
		"currency":            req.Currency,
		"idempotencyKey":      req.IdempotencyKey,
		"fromCheckout":        true,
		"addressSnapshot":     req.AddressSnapshot,
		"notes":               req.Notes,
		"gift":                req.Gift,
		"discountMinor":       req.DiscountMinor,
		"shippingMinor":       req.ShippingMinor,
		"tipMinor":            req.TipMinor,
		"lines":               lines,
		"metadata":            meta,
	}

	var raw struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	if err := c.doJSON(ctx, http.MethodPost, "/v1/orders", req.TenantID.String(), req.IdempotencyKey, body, &raw); err != nil {
		return ports.CreateFromCheckoutResult{}, err
	}
	if raw.ID == "" {
		return ports.CreateFromCheckoutResult{}, fmt.Errorf("%w: empty order id", domain.ErrInvariant)
	}
	return ports.CreateFromCheckoutResult{OrderID: raw.ID, Status: raw.Status}, nil
}

// PlaceOrder posts POST /v1/orders/{id}/place.
func (c *Client) PlaceOrder(ctx context.Context, req ports.PlaceOrderRequest) (ports.PlaceOrderResult, error) {
	if c.baseURL == "" {
		if c.inner != nil {
			return c.inner.PlaceOrder(ctx, req)
		}
		c.log.Warn("order.place.unconfigured")
		return ports.PlaceOrderResult{}, fmt.Errorf("%w: order client unconfigured", domain.ErrInvariant)
	}
	path := "/v1/orders/" + url.PathEscape(req.OrderID) + "/place"
	body := map[string]any{"idempotencyKey": req.IdempotencyKey}
	var raw struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	if err := c.doJSON(ctx, http.MethodPost, path, req.TenantID.String(), req.IdempotencyKey, body, &raw); err != nil {
		return ports.PlaceOrderResult{}, err
	}
	id := raw.ID
	if id == "" {
		id = req.OrderID
	}
	return ports.PlaceOrderResult{OrderID: id, Status: raw.Status}, nil
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
		return fmt.Errorf("order http %s %s: status %d: %s", method, path, resp.StatusCode, string(raw))
	}
	if out == nil || len(raw) == 0 {
		return nil
	}
	return json.Unmarshal(raw, out)
}

var _ ports.OrderClient = (*Client)(nil)
