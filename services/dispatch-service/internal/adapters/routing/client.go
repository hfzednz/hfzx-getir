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
	"github.com/nexora/dispatch-service/internal/app/memory"
	"github.com/nexora/dispatch-service/internal/app/ports"
	"github.com/nexora/dispatch-service/internal/domain"
)

// Client talks to routing-service over HTTP when BaseURL is set; otherwise memory.
type Client struct {
	BaseURL    string
	log        *slog.Logger
	inner      *memory.RoutingClient
	HTTPClient *http.Client
}

// NewClient returns a routing client. Empty baseURL keeps in-process memory.
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
		c.inner = &memory.RoutingClient{ETA: 900}
	} else {
		log.Info("routing.client.http", "baseURL", baseURL)
	}
	return c
}

// CreateRoute posts POST /v1/routing/routes (or memory when BaseURL empty).
func (c *Client) CreateRoute(ctx context.Context, req ports.RouteRequest) (ports.RouteResult, error) {
	if c.BaseURL == "" {
		return c.inner.CreateRoute(ctx, req)
	}

	waypoints := make([]map[string]any, 0, len(req.Waypoints))
	for i, wp := range req.Waypoints {
		kind := "stop"
		if i == 0 {
			kind = "warehouse"
		}
		waypoints = append(waypoints, map[string]any{
			"kind": kind, "lat": wp.Lat, "lon": wp.Lng,
		})
	}
	dispatchID := req.DispatchID.String()
	body := map[string]any{
		"dispatchId": dispatchID,
		"waypoints":  waypoints,
	}
	var raw struct {
		ID              uuid.UUID `json:"id"`
		DurationSeconds float64   `json:"durationSeconds"`
		DistanceMeters  float64   `json:"distanceMeters"`
	}
	if err := c.doJSON(ctx, http.MethodPost, "/v1/routing/routes", req.TenantID.String(), body, &raw); err != nil {
		return ports.RouteResult{}, err
	}
	if raw.ID == uuid.Nil {
		return ports.RouteResult{}, fmt.Errorf("routing create route: empty route id")
	}
	return ports.RouteResult{
		RouteID:    raw.ID,
		ETASeconds: int(raw.DurationSeconds),
		DistanceM:  raw.DistanceMeters,
	}, nil
}

// EstimateETA creates a transient two-point route and returns durationSeconds.
func (c *Client) EstimateETA(ctx context.Context, from, to domain.Point) (int, error) {
	if c.BaseURL == "" {
		return c.inner.EstimateETA(ctx, from, to)
	}
	body := map[string]any{
		"waypoints": []map[string]any{
			{"kind": "warehouse", "lat": from.Lat, "lon": from.Lng},
			{"kind": "stop", "lat": to.Lat, "lon": to.Lng},
		},
	}
	var raw struct {
		DurationSeconds float64 `json:"durationSeconds"`
	}
	if err := c.doJSON(ctx, http.MethodPost, "/v1/routing/routes", "", body, &raw); err != nil {
		return 0, err
	}
	eta := int(raw.DurationSeconds)
	if eta <= 0 {
		eta = 600
	}
	return eta, nil
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
