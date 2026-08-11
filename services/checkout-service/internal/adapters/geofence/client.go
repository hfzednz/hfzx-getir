// Package geofence provides an HTTP GeofenceClient for checkout-service.
package geofence

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

	"github.com/nexora/checkout-service/internal/app/ports"
)

// Client talks to geofence-service when baseURL is set.
// Empty baseURL uses inner when set; otherwise InZone:true (dev).
// Never returns context.Canceled.
type Client struct {
	baseURL    string
	log        *slog.Logger
	inner      ports.GeofenceClient
	HTTPClient *http.Client
}

// NewClient returns a geofence client. Empty baseURL keeps inner / in-zone stub.
func NewClient(baseURL string, log *slog.Logger, inner ports.GeofenceClient) *Client {
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
	if baseURL != "" {
		log.Info("geofence.client.http", "baseURL", baseURL)
	}
	return c
}

// CheckZone posts POST /v1/geofence/serviceability when baseURL is set.
func (c *Client) CheckZone(ctx context.Context, req ports.GeofenceRequest) (ports.GeofenceResult, error) {
	if c.baseURL == "" {
		if c.inner != nil {
			return c.inner.CheckZone(ctx, req)
		}
		c.log.Debug("geofence.fallback.dev", "note", "GEOFENCE_URL empty; using inner/InZone")
		return ports.GeofenceResult{InZone: true}, nil
	}

	body := map[string]any{
		"city":  req.CityID,
		"point": map[string]any{"lat": req.Lat, "lng": req.Lng},
	}
	var raw struct {
		Serviceable   bool     `json:"serviceable"`
		Blocked       bool     `json:"blocked"`
		Reason        string   `json:"reason"`
		DeliveryZones []string `json:"deliveryZones"`
		MatchingZones []string `json:"matchingZones"`
	}
	if err := c.doJSON(ctx, http.MethodPost, "/v1/geofence/serviceability", req.TenantID.String(), body, &raw); err != nil {
		return ports.GeofenceResult{}, err
	}

	inZone := raw.Serviceable && !raw.Blocked
	zoneID := ""
	if len(raw.DeliveryZones) > 0 {
		zoneID = raw.DeliveryZones[0]
	} else if len(raw.MatchingZones) > 0 {
		zoneID = raw.MatchingZones[0]
	}
	return ports.GeofenceResult{
		InZone: inZone,
		ZoneID: zoneID,
		Reason: raw.Reason,
	}, nil
}

func (c *Client) doJSON(ctx context.Context, method, path, tenantID string, in any, out any) error {
	b, err := json.Marshal(in)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
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
		return fmt.Errorf("geofence http %s %s: status %d: %s", method, path, resp.StatusCode, string(raw))
	}
	if out == nil || len(raw) == 0 {
		return nil
	}
	return json.Unmarshal(raw, out)
}

var _ ports.GeofenceClient = (*Client)(nil)
