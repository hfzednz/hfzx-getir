// Package dynamic provides a DynamicHintClient HTTP adapter for inventory ATP.
package dynamic

import (
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
	"github.com/nexora/pricing-service/internal/app"
	"github.com/nexora/pricing-service/internal/app/ports"
)

// Client talks to inventory-service ATP when BaseURL is set; otherwise noop.
type Client struct {
	BaseURL    string
	log        *slog.Logger
	inner      ports.DynamicHintClient
	HTTPClient *http.Client
}

// NewClient returns a DynamicHintClient. Empty baseURL keeps noop hints.
func NewClient(baseURL string, log *slog.Logger) *Client {
	if log == nil {
		log = slog.Default()
	}
	baseURL = strings.TrimRight(baseURL, "/")
	c := &Client{
		BaseURL:    baseURL,
		log:        log,
		inner:      app.NoopHintClient{},
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}
	if baseURL != "" {
		log.Info("dynamic.hint.http", "baseURL", baseURL, "path", "/v1/inventory/atp")
	} else {
		log.Info("dynamic.hint.noop", "note", "INVENTORY_BASE_URL empty")
	}
	return c
}

// Hint returns available qty from inventory ATP when configured.
func (c *Client) Hint(ctx context.Context, req ports.DynamicHintRequest) (ports.DynamicHintResult, error) {
	if c.BaseURL == "" {
		return c.inner.Hint(ctx, req)
	}
	q := url.Values{}
	q.Set("variantId", req.VariantID.String())
	if req.WarehouseID != nil && *req.WarehouseID != uuid.Nil {
		q.Set("warehouseId", req.WarehouseID.String())
	}
	u := c.BaseURL + "/v1/inventory/atp?" + q.Encode()
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return ports.DynamicHintResult{}, err
	}
	httpReq.Header.Set("X-Tenant-Id", req.TenantID.String())
	httpReq.Header.Set("Accept", "application/json")

	res, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		c.log.Warn("dynamic.hint.atp.fallback", "err", err)
		return c.inner.Hint(ctx, req)
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if res.StatusCode >= 300 {
		c.log.Warn("dynamic.hint.atp.status", "status", res.StatusCode, "body", string(body))
		return c.inner.Hint(ctx, req)
	}

	var raw struct {
		Items []struct {
			Available int64 `json:"Available"`
			ATP       int64 `json:"ATP"`
		} `json:"items"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return ports.DynamicHintResult{}, fmt.Errorf("inventory atp decode: %w", err)
	}
	var sum int64
	for _, it := range raw.Items {
		if it.ATP != 0 {
			sum += it.ATP
		} else {
			sum += it.Available
		}
	}
	qty := int(sum)
	return ports.DynamicHintResult{AvailableQty: &qty}, nil
}

var _ ports.DynamicHintClient = (*Client)(nil)
