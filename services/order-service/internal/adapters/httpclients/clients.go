package httpclients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/nexora/order-service/internal/app/ports"
)

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

func (c *Client) doJSON(ctx context.Context, method, path string, in any, out any, headers map[string]string) error {
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
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode >= 300 {
		return fmt.Errorf("http %s %s: status %d: %s", method, path, resp.StatusCode, string(raw))
	}
	if out == nil || len(raw) == 0 {
		return nil
	}
	return json.Unmarshal(raw, out)
}

type InventoryHTTP struct{ *Client }

func (c InventoryHTTP) SoftReserve(ctx context.Context, req ports.SoftReserveRequest) (ports.SoftReserveResult, error) {
	var out ports.SoftReserveResult
	err := c.doJSON(ctx, http.MethodPost, "/v1/inventory/reservations/soft", req, &out, map[string]string{
		"Idempotency-Key": req.IdempotencyKey,
		"X-Tenant-Id":     req.TenantID.String(),
	})
	return out, err
}

func (c InventoryHTTP) ConfirmHard(ctx context.Context, req ports.ConfirmHardRequest) error {
	return c.doJSON(ctx, http.MethodPost, "/v1/inventory/reservations/confirm", req, nil, map[string]string{
		"Idempotency-Key": req.IdempotencyKey,
		"X-Tenant-Id":     req.TenantID.String(),
	})
}

func (c InventoryHTTP) Release(ctx context.Context, req ports.ReleaseRequest) error {
	return c.doJSON(ctx, http.MethodPost, "/v1/inventory/reservations/release", req, nil, map[string]string{
		"Idempotency-Key": req.IdempotencyKey,
		"X-Tenant-Id":     req.TenantID.String(),
	})
}

var _ ports.InventoryClient = InventoryHTTP{}

type PaymentHTTP struct{ *Client }

func (c PaymentHTTP) Authorize(ctx context.Context, req ports.AuthorizeRequest) (ports.AuthorizeResult, error) {
	var out ports.AuthorizeResult
	err := c.doJSON(ctx, http.MethodPost, "/v1/payments/intents/authorize", req, &out, map[string]string{
		"Idempotency-Key": req.IdempotencyKey,
		"X-Tenant-Id":     req.TenantID.String(),
	})
	return out, err
}

func (c PaymentHTTP) Void(ctx context.Context, req ports.VoidRequest) error {
	return c.doJSON(ctx, http.MethodPost, "/v1/payments/intents/void", req, nil, map[string]string{
		"Idempotency-Key": req.IdempotencyKey,
		"X-Tenant-Id":     req.TenantID.String(),
	})
}

func (c PaymentHTTP) Refund(ctx context.Context, req ports.RefundPaymentRequest) (ports.RefundPaymentResult, error) {
	var out ports.RefundPaymentResult
	err := c.doJSON(ctx, http.MethodPost, "/v1/payments/refunds", req, &out, map[string]string{
		"Idempotency-Key": req.IdempotencyKey,
		"X-Tenant-Id":     req.TenantID.String(),
	})
	return out, err
}

var _ ports.PaymentClient = PaymentHTTP{}

type WarehouseHTTP struct{ *Client }

func (c WarehouseHTTP) ReceiveFulfillment(ctx context.Context, req ports.ReceiveFulfillmentRequest) (ports.ReceiveFulfillmentResult, error) {
	var out ports.ReceiveFulfillmentResult
	err := c.doJSON(ctx, http.MethodPost, "/v1/warehouse/fulfillments/receive", req, &out, map[string]string{
		"Idempotency-Key": req.IdempotencyKey,
		"X-Tenant-Id":     req.TenantID.String(),
	})
	return out, err
}

var _ ports.WarehouseClient = WarehouseHTTP{}

type DispatchHTTP struct{ *Client }

func (c DispatchHTTP) RequestDispatch(ctx context.Context, req ports.RequestDispatchRequest) (ports.RequestDispatchResult, error) {
	var out ports.RequestDispatchResult
	err := c.doJSON(ctx, http.MethodPost, "/v1/dispatch/jobs", req, &out, map[string]string{
		"Idempotency-Key": req.IdempotencyKey,
		"X-Tenant-Id":     req.TenantID.String(),
	})
	return out, err
}

var _ ports.DispatchClient = DispatchHTTP{}
