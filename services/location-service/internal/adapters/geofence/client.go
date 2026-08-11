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
	"github.com/nexora/location-service/internal/app"
	"github.com/nexora/location-service/internal/app/ports"
	"github.com/nexora/location-service/internal/domain"
)

// Client talks to geofence-service over HTTP when BaseURL is set; otherwise memory.
type Client struct {
	BaseURL    string
	log        *slog.Logger
	inner      *app.MemoryGeofenceClient
	HTTPClient *http.Client
}

// NewClient returns a GeofenceClient. Empty baseURL keeps in-process memory.
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
		c.inner = app.NewMemoryGeofenceClient()
	} else {
		log.Info("geofence.client.http", "baseURL", baseURL)
	}
	return c
}

// CheckServiceability posts POST /v1/geofence/serviceability (or memory when BaseURL empty).
func (c *Client) CheckServiceability(ctx context.Context, tenantID uuid.UUID, lat, lng float64) (domain.DeliveryFeasibility, error) {
	if c.BaseURL == "" {
		return c.inner.CheckServiceability(ctx, tenantID, lat, lng)
	}
	body := map[string]any{
		"point": map[string]any{"lat": lat, "lng": lng},
	}
	var raw struct {
		Serviceable     bool        `json:"serviceable"`
		Blocked         bool        `json:"blocked"`
		Reason          string      `json:"reason"`
		DeliveryZones   []uuid.UUID `json:"deliveryZones"`
		MatchingZones   []uuid.UUID `json:"matchingZones"`
		RestrictedZones []uuid.UUID `json:"restrictedZones"`
	}
	if err := c.doJSON(ctx, http.MethodPost, "/v1/geofence/serviceability", tenantID.String(), body, &raw); err != nil {
		return domain.DeliveryFeasibility{}, err
	}
	feas := domain.DeliveryFeasibility{
		Serviceable: raw.Serviceable && !raw.Blocked,
		Reason:      raw.Reason,
		Score:       0,
	}
	if feas.Serviceable {
		feas.Score = 1
		if len(raw.DeliveryZones) > 0 {
			zid := raw.DeliveryZones[0]
			feas.ZoneID = &zid
		} else if len(raw.MatchingZones) > 0 {
			zid := raw.MatchingZones[0]
			feas.ZoneID = &zid
		}
	}
	return feas, nil
}

// Contains posts POST /v1/geofence/contains (or memory when BaseURL empty).
func (c *Client) Contains(ctx context.Context, tenantID, zoneID uuid.UUID, lat, lng float64) (bool, error) {
	if c.BaseURL == "" {
		return c.inner.Contains(ctx, tenantID, zoneID, lat, lng)
	}
	body := map[string]any{
		"zoneId": zoneID.String(),
		"point":  map[string]any{"lat": lat, "lng": lng},
	}
	var raw struct {
		Inside bool `json:"inside"`
	}
	if err := c.doJSON(ctx, http.MethodPost, "/v1/geofence/contains", tenantID.String(), body, &raw); err != nil {
		return false, err
	}
	return raw.Inside, nil
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
