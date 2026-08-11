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

	"github.com/google/uuid"
	"github.com/nexora/tracking-service/internal/app"
	"github.com/nexora/tracking-service/internal/app/ports"
	"github.com/nexora/tracking-service/internal/domain"
)

// Client talks to geofence-service over HTTP when BaseURL is set; otherwise noop.
type Client struct {
	BaseURL    string
	log        *slog.Logger
	inner      app.NoopGeofenceClient
	HTTPClient *http.Client
}

// NewClient returns a GeofenceClient. Empty baseURL keeps noop (no hits).
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
		log.Info("geofence.client.http", "baseURL", baseURL)
	}
	return c
}

// Check posts POST /v1/geofence/serviceability and maps matching zones to enter hits.
func (c *Client) Check(ctx context.Context, req ports.GeofenceCheckRequest) (ports.GeofenceCheckResult, error) {
	if c.BaseURL == "" {
		return c.inner.Check(ctx, req)
	}
	body := map[string]any{
		"point": map[string]any{"lat": req.Lat, "lng": req.Lon},
	}
	var raw struct {
		MatchingZones   []uuid.UUID `json:"matchingZones"`
		RestrictedZones []uuid.UUID `json:"restrictedZones"`
		DeliveryZones   []uuid.UUID `json:"deliveryZones"`
	}
	if err := c.doJSON(ctx, http.MethodPost, "/v1/geofence/serviceability", req.TenantID.String(), body, &raw); err != nil {
		return ports.GeofenceCheckResult{}, err
	}
	seen := make(map[uuid.UUID]struct{})
	hits := make([]ports.GeofenceHit, 0, len(raw.MatchingZones))
	for _, id := range raw.MatchingZones {
		if id == uuid.Nil {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		hits = append(hits, ports.GeofenceHit{
			ZoneID: id.String(),
			Kind:   domain.GeofenceEnter,
		})
	}
	return ports.GeofenceCheckResult{Hits: hits}, nil
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
		return fmt.Errorf("geofence http %s %s: status %d: %s", method, path, resp.StatusCode, string(raw))
	}
	if out == nil || len(raw) == 0 {
		return nil
	}
	return json.Unmarshal(raw, out)
}

var _ ports.GeofenceClient = (*Client)(nil)
