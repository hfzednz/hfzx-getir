// Package inventory provides an InventoryClient HTTP adapter for warehouse-service.
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
	"github.com/nexora/warehouse-service/internal/app/ports"
)

// Client talks to inventory-service over HTTP when BaseURL is set; otherwise local synthetic IDs.
type Client struct {
	BaseURL    string
	log        *slog.Logger
	HTTPClient *http.Client
}

// NewClient returns an inventory-service client.
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
	if baseURL != "" {
		log.Info("inventory.client.http", "baseURL", baseURL)
	}
	return c
}

func (c *Client) SoftReserve(ctx context.Context, req ports.SoftReserveRequest) (ports.SoftReserveResult, error) {
	if c.BaseURL == "" {
		id := uuid.New()
		c.log.Debug("inventory.soft_reserve.local", "externalRef", req.ExternalRef, "reservationId", id)
		return ports.SoftReserveResult{ReservationID: id, Status: "active"}, nil
	}
	lines := make([]map[string]any, 0, len(req.Lines))
	for _, l := range req.Lines {
		wh := l.WarehouseID
		if wh == uuid.Nil {
			wh = req.WarehouseID
		}
		lines = append(lines, map[string]any{
			"warehouseId": wh,
			"variantId":   l.VariantID,
			"skuCode":     l.SKUCode,
			"qty":         l.Qty,
		})
	}
	body := map[string]any{
		"warehouseId":    req.WarehouseID,
		"externalRef":    req.ExternalRef,
		"idempotencyKey": req.IdempotencyKey,
		"lines":          lines,
	}
	var raw struct {
		ID     uuid.UUID `json:"ID"`
		Status string    `json:"Status"`
	}
	if err := c.doJSON(ctx, http.MethodPost, "/v1/inventory/reservations/soft", req.TenantID.String(), req.IdempotencyKey, body, &raw); err != nil {
		return ports.SoftReserveResult{}, err
	}
	if raw.ID == uuid.Nil {
		return ports.SoftReserveResult{}, fmt.Errorf("inventory soft reserve: empty reservation id")
	}
	status := raw.Status
	if status == "" {
		status = "active"
	}
	return ports.SoftReserveResult{ReservationID: raw.ID, Status: status}, nil
}

func (c *Client) ConfirmHard(ctx context.Context, req ports.ConfirmHardRequest) error {
	if c.BaseURL == "" {
		c.log.Debug("inventory.confirm_hard.local", "reservationId", req.ReservationID)
		return nil
	}
	path := "/v1/inventory/reservations/" + url.PathEscape(req.ReservationID.String()) + "/confirm"
	body := map[string]any{"idempotencyKey": req.IdempotencyKey}
	return c.doJSON(ctx, http.MethodPost, path, req.TenantID.String(), req.IdempotencyKey, body, nil)
}

func (c *Client) Release(ctx context.Context, req ports.ReleaseRequest) error {
	if c.BaseURL == "" {
		c.log.Debug("inventory.release.local", "reservationId", req.ReservationID)
		return nil
	}
	path := "/v1/inventory/reservations/" + url.PathEscape(req.ReservationID.String()) + "/release"
	body := map[string]any{"idempotencyKey": req.IdempotencyKey}
	return c.doJSON(ctx, http.MethodPost, path, req.TenantID.String(), req.IdempotencyKey, body, nil)
}

func (c *Client) Consume(ctx context.Context, req ports.ConsumeRequest) error {
	if c.BaseURL == "" {
		c.log.Debug("inventory.consume.local", "reservationId", req.ReservationID)
		return nil
	}
	path := "/v1/inventory/reservations/" + url.PathEscape(req.ReservationID.String()) + "/consume"
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
