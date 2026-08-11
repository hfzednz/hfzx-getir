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
	"github.com/nexora/dispatch-service/internal/app/memory"
	"github.com/nexora/dispatch-service/internal/app/ports"
	"github.com/nexora/dispatch-service/internal/domain"
)

// Client talks to geofence-service over HTTP when BaseURL is set; otherwise memory.
type Client struct {
	BaseURL    string
	log        *slog.Logger
	inner      *memory.GeofenceClient
	HTTPClient *http.Client
}

// NewClient returns a geofence client. Empty baseURL keeps in-process memory.
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
		c.inner = &memory.GeofenceClient{OK: true}
	} else {
		log.Info("geofence.client.http", "baseURL", baseURL)
	}
	return c
}

// CheckServiceability posts POST /v1/geofence/serviceability (or memory when BaseURL empty).
func (c *Client) CheckServiceability(ctx context.Context, tenantID uuid.UUID, city string, p domain.Point) (bool, error) {
	if c.BaseURL == "" {
		return c.inner.CheckServiceability(ctx, tenantID, city, p)
	}
	body := map[string]any{
		"city":  city,
		"point": map[string]any{"lat": p.Lat, "lng": p.Lng},
	}
	var raw struct {
		Serviceable bool `json:"serviceable"`
		Blocked     bool `json:"blocked"`
	}
	if err := c.doJSON(ctx, http.MethodPost, "/v1/geofence/serviceability", tenantID.String(), body, &raw); err != nil {
		return false, err
	}
	return raw.Serviceable && !raw.Blocked, nil
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
