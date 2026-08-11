// Package inventory provides an InventoryClient HTTP adapter.
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
	"github.com/nexora/cart-service/internal/app/memory"
	"github.com/nexora/cart-service/internal/app/ports"
)

// Client talks to inventory-service over HTTP when BaseURL is set; otherwise memory.
type Client struct {
	BaseURL    string
	log        *slog.Logger
	inner      *memory.InventoryClient
	HTTPClient *http.Client
}

// NewClient returns an inventory client. Empty baseURL keeps in-process memory.
func NewClient(baseURL string, log *slog.Logger) *Client {
	if log == nil {
		log = slog.Default()
	}
	baseURL = strings.TrimRight(baseURL, "/")
	c := &Client{
		BaseURL:    baseURL,
		log:        log,
		HTTPClient: &http.Client{Timeout: 15 * time.Second},
	}
	if baseURL == "" {
		c.inner = memory.NewInventoryClient()
	} else {
		log.Info("inventory.client.http", "baseURL", baseURL)
	}
	return c
}

func (c *Client) ATP(ctx context.Context, req ports.ATPRequest) (ports.ATPResult, error) {
	if c.BaseURL == "" {
		return c.inner.ATP(ctx, req)
	}

	out := ports.ATPResult{Lines: make([]ports.ATPLineResult, 0, len(req.Lines))}
	for _, line := range req.Lines {
		q := url.Values{}
		q.Set("variantId", line.VariantID.String())
		if req.CityID != nil {
			q.Set("regionId", req.CityID.String())
		}
		path := "/v1/inventory/atp?" + q.Encode()
		var raw struct {
			Items []struct {
				VariantID   uuid.UUID  `json:"VariantID"`
				WarehouseID uuid.UUID  `json:"WarehouseID"`
				Available   int64      `json:"Available"`
				ATP         int64      `json:"ATP"`
			} `json:"items"`
		}
		if err := c.doJSON(ctx, http.MethodGet, path, req.TenantID.String(), "", nil, &raw); err != nil {
			return ports.ATPResult{}, err
		}
		avail := 0
		var wh *uuid.UUID
		for _, it := range raw.Items {
			n := int(it.ATP)
			if it.Available > it.ATP {
				n = int(it.Available)
			}
			if n > avail {
				avail = n
				if it.WarehouseID != uuid.Nil {
					id := it.WarehouseID
					wh = &id
				}
			}
		}
		out.Lines = append(out.Lines, ports.ATPLineResult{
			VariantID: line.VariantID, Available: avail, WarehouseID: wh,
		})
	}
	return out, nil
}

func (c *Client) SoftReserve(ctx context.Context, req ports.SoftReserveRequest) (ports.SoftReserveResult, error) {
	if c.BaseURL == "" {
		return c.inner.SoftReserve(ctx, req)
	}

	lines := make([]map[string]any, 0, len(req.Lines))
	for _, l := range req.Lines {
		lines = append(lines, map[string]any{
			"variantId": l.VariantID,
			"qty":       l.Qty,
		})
	}
	body := map[string]any{
		"externalRef":    req.CartID.String(),
		"idempotencyKey": req.IdempotencyKey,
		"ttlSeconds":     900,
		"lines":          lines,
	}
	var raw struct {
		ID        uuid.UUID  `json:"ID"`
		ExpiresAt *time.Time `json:"ExpiresAt"`
	}
	if err := c.doJSON(ctx, http.MethodPost, "/v1/inventory/reservations/soft", req.TenantID.String(), req.IdempotencyKey, body, &raw); err != nil {
		return ports.SoftReserveResult{}, err
	}
	ref := raw.ID.String()
	if ref == uuid.Nil.String() {
		return ports.SoftReserveResult{}, fmt.Errorf("inventory soft reserve: empty reservation id")
	}
	return ports.SoftReserveResult{ReservationRef: ref, ExpiresAt: raw.ExpiresAt}, nil
}

func (c *Client) Release(ctx context.Context, req ports.ReleaseRequest) error {
	if c.BaseURL == "" {
		return c.inner.Release(ctx, req)
	}
	ref := strings.TrimSpace(req.ReservationRef)
	if ref == "" {
		return fmt.Errorf("inventory release: reservation ref required")
	}
	path := "/v1/inventory/reservations/" + url.PathEscape(ref) + "/release"
	body := map[string]any{"idempotencyKey": req.IdempotencyKey}
	return c.doJSON(ctx, http.MethodPost, path, req.TenantID.String(), req.IdempotencyKey, body, nil)
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
		return fmt.Errorf("inventory http %s %s: status %d: %s", method, path, resp.StatusCode, string(raw))
	}
	if out == nil || len(raw) == 0 {
		return nil
	}
	return json.Unmarshal(raw, out)
}

var _ ports.InventoryClient = (*Client)(nil)
