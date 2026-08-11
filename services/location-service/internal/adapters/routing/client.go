package routing

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
)

// Client talks to routing-service over HTTP when BaseURL is set; otherwise memory.
type Client struct {
	BaseURL    string
	log        *slog.Logger
	inner      *app.MemoryRoutingClient
	HTTPClient *http.Client
}

// NewClient returns a RoutingClient. Empty baseURL keeps in-process memory.
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
		c.inner = &app.MemoryRoutingClient{}
	} else {
		log.Info("routing.client.http", "baseURL", baseURL)
	}
	return c
}

// CreateRoute posts POST /v1/routing/routes (or memory when BaseURL empty).
func (c *Client) CreateRoute(ctx context.Context, req ports.CreateRouteRequest) (ports.CreateRouteResult, error) {
	if c.BaseURL == "" {
		return c.inner.CreateRoute(ctx, req)
	}
	waypoints := make([]map[string]any, 0, 2+len(req.Waypoints))
	waypoints = append(waypoints, map[string]any{"kind": "warehouse", "lat": req.Origin.Lat, "lon": req.Origin.Lng})
	for _, wp := range req.Waypoints {
		waypoints = append(waypoints, map[string]any{"kind": "stop", "lat": wp.Lat, "lon": wp.Lng})
	}
	waypoints = append(waypoints, map[string]any{"kind": "stop", "lat": req.Dest.Lat, "lon": req.Dest.Lng})

	body := map[string]any{"waypoints": waypoints}
	var raw struct {
		ID              uuid.UUID `json:"id"`
		DistanceMeters  float64   `json:"distanceMeters"`
		DurationSeconds float64   `json:"durationSeconds"`
	}
	if err := c.doJSON(ctx, http.MethodPost, "/v1/routing/routes", req.TenantID.String(), body, &raw); err != nil {
		return ports.CreateRouteResult{}, err
	}
	if raw.ID == uuid.Nil {
		return ports.CreateRouteResult{}, fmt.Errorf("routing create route: empty route id")
	}
	return ports.CreateRouteResult{
		RouteID:         raw.ID.String(),
		DistanceMeters:  raw.DistanceMeters,
		DurationSeconds: raw.DurationSeconds,
		Provider:        "routing-service",
	}, nil
}

// ETA creates a transient two-point route and returns distance/duration.
func (c *Client) ETA(ctx context.Context, req ports.ETARequest) (ports.ETAResult, error) {
	if c.BaseURL == "" {
		return c.inner.ETA(ctx, req)
	}
	body := map[string]any{
		"waypoints": []map[string]any{
			{"kind": "warehouse", "lat": req.Origin.Lat, "lon": req.Origin.Lng},
			{"kind": "stop", "lat": req.Dest.Lat, "lon": req.Dest.Lng},
		},
	}
	var raw struct {
		DistanceMeters  float64 `json:"distanceMeters"`
		DurationSeconds float64 `json:"durationSeconds"`
	}
	if err := c.doJSON(ctx, http.MethodPost, "/v1/routing/routes", req.TenantID.String(), body, &raw); err != nil {
		return ports.ETAResult{}, err
	}
	return ports.ETAResult{
		DistanceMeters:  raw.DistanceMeters,
		DurationSeconds: raw.DurationSeconds,
		Provider:        "routing-service",
	}, nil
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
		return fmt.Errorf("routing http %s %s: status %d: %s", method, path, resp.StatusCode, string(raw))
	}
	if out == nil || len(raw) == 0 {
		return nil
	}
	return json.Unmarshal(raw, out)
}

var _ ports.RoutingClient = (*Client)(nil)
