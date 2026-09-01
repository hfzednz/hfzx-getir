package inventory

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/supplier-service/internal/app/ports"
	"github.com/nexora/supplier-service/internal/domain"
)

// Client posts receiving movements to inventory-service.
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

func New(baseURL string) *Client {
	return &Client{
		BaseURL:    strings.TrimRight(baseURL, "/"),
		HTTPClient: &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *Client) AnnounceASN(context.Context, uuid.UUID, domain.InboundShipment) error {
	return nil
}

func (c *Client) ReceiveStock(ctx context.Context, req ports.ReceiveStockRequest) error {
	if c == nil || c.BaseURL == "" {
		return nil
	}
	body, err := json.Marshal(map[string]any{
		"warehouseId":    req.WarehouseID.String(),
		"skuCode":        req.SKUCode,
		"qty":            req.Qty,
		"idempotencyKey": req.IdempotencyKey,
		"reason":         req.Reason,
	})
	if err != nil {
		return err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/v1/inventory/stock/receive", bytes.NewReader(body))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	if req.TenantID != uuid.Nil {
		httpReq.Header.Set("X-Tenant-Id", req.TenantID.String())
	}
	if req.IdempotencyKey != "" {
		httpReq.Header.Set("Idempotency-Key", req.IdempotencyKey)
	}
	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode >= 300 {
		return fmt.Errorf("inventory receive: status %d: %s", resp.StatusCode, string(raw))
	}
	return nil
}

var _ ports.InventoryClient = (*Client)(nil)
