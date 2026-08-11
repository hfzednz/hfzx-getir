package tracking

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
	"github.com/nexora/dispatch-service/internal/app/memory"
	"github.com/nexora/dispatch-service/internal/app/ports"
)

// Client talks to tracking-service over HTTP when BaseURL is set; otherwise memory.
type Client struct {
	BaseURL    string
	log        *slog.Logger
	inner      *memory.TrackingClient
	HTTPClient *http.Client
}

// NewClient returns a tracking client. Empty baseURL keeps in-process memory.
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
		c.inner = &memory.TrackingClient{}
	} else {
		log.Info("tracking.client.http", "baseURL", baseURL)
	}
	return c
}

// SubscribeDispatch registers dispatch interest via POST /v1/tracking/orders/{id}/timeline.
func (c *Client) SubscribeDispatch(ctx context.Context, tenantID, dispatchID, courierPrincipalID uuid.UUID) error {
	if c.BaseURL == "" {
		return c.inner.SubscribeDispatch(ctx, tenantID, dispatchID, courierPrincipalID)
	}
	path := "/v1/tracking/orders/" + url.PathEscape(dispatchID.String()) + "/timeline"
	body := map[string]any{
		"courierId": courierPrincipalID.String(),
		"type":      "Custom",
		"message":   "dispatch.subscribed",
		"meta": map[string]any{
			"dispatchId":         dispatchID.String(),
			"courierPrincipalId": courierPrincipalID.String(),
		},
	}
	return c.doJSON(ctx, http.MethodPost, path, tenantID.String(), body, nil)
}

func (c *Client) doJSON(ctx context.Context, method, path, tenantID string, in any, out any) error {
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
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode >= 300 {
		return fmt.Errorf("tracking http %s %s: status %d: %s", method, path, resp.StatusCode, string(raw))
	}
	if out == nil || len(raw) == 0 {
		return nil
	}
	return json.Unmarshal(raw, out)
}

var _ ports.TrackingClient = (*Client)(nil)
