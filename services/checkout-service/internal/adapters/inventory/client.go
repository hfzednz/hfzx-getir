// Package inventory provides an InventoryClient HTTP adapter (ATP only).
package inventory

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

// Client talks to inventory-service ATP when baseURL is set and inner is nil.
type Client struct {
	baseURL    string
	log        *slog.Logger
	inner      ports.InventoryClient
	HTTPClient *http.Client
}

// NewClient returns an inventory ATP client.
func NewClient(baseURL string, log *slog.Logger, inner ports.InventoryClient) *Client {
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
		log.Info("inventory.client.http", "baseURL", baseURL)
	}
	return c
}

// CheckATP queries GET /v1/inventory/atp per line.
func (c *Client) CheckATP(ctx context.Context, req ports.ATPRequest) (ports.ATPResult, error) {
	if c.inner != nil {
		return c.inner.CheckATP(ctx, req)
	}
	if c.baseURL == "" {
		c.log.Warn("inventory.client.unconfigured")
		return ports.ATPResult{}, fmt.Errorf("%w: inventory client unconfigured", domain.ErrInvariant)
	}

	out := ports.ATPResult{AllAvailable: true, Lines: make([]ports.ATPLineResult, 0, len(req.Lines))}
	for _, line := range req.Lines {
		q := url.Values{}
		if line.VariantID != uuid.Nil {
			q.Set("variantId", line.VariantID.String())
		}
		if line.SKUCode != "" {
			q.Set("skuCode", line.SKUCode)
		}
		if line.WarehouseID != nil {
			q.Set("warehouseId", line.WarehouseID.String())
		}
		if req.CityID != "" {
			if id, err := uuid.Parse(req.CityID); err == nil {
				q.Set("regionId", id.String())
			}
		}
		path := "/v1/inventory/atp?" + q.Encode()
		var raw struct {
			Items []struct {
				VariantID uuid.UUID `json:"VariantID"`
				Available int64     `json:"Available"`
				ATP       int64     `json:"ATP"`
			} `json:"items"`
		}
		if err := c.doJSON(ctx, http.MethodGet, path, req.TenantID.String(), "", nil, &raw); err != nil {
			return ports.ATPResult{}, err
		}
		availQty := 0
		for _, it := range raw.Items {
			n := int(it.ATP)
			if int(it.Available) > n {
				n = int(it.Available)
			}
			if n > availQty {
				availQty = n
			}
		}
		ok := availQty >= line.Qty
		lr := ports.ATPLineResult{
			VariantID: line.VariantID, Available: ok, AvailableQty: availQty,
		}
		if !ok {
			lr.Reason = "insufficient_atp"
			out.AllAvailable = false
		}
		out.Lines = append(out.Lines, lr)
	}
	return out, nil
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
		return fmt.Errorf("inventory http %s %s: status %d: %s", method, path, resp.StatusCode, string(raw))
	}
	if out == nil || len(raw) == 0 {
		return nil
	}
	return json.Unmarshal(raw, out)
}

var _ ports.InventoryClient = (*Client)(nil)
