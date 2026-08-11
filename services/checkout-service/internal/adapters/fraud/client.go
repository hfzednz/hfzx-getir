// Package fraud provides an HTTP FraudClient backed by ai-platform-service.
package fraud

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/nexora/checkout-service/internal/app/ports"
)

// Client scores checkouts via ai-platform when baseURL is set.
// Empty baseURL uses inner when non-nil; otherwise Decision allow.
type Client struct {
	baseURL    string
	log        *slog.Logger
	inner      ports.FraudClient
	HTTPClient *http.Client
}

// NewClient returns a fraud client. Empty baseURL keeps inner / allow stub.
func NewClient(baseURL string, log *slog.Logger, inner ports.FraudClient) *Client {
	if log == nil {
		log = slog.Default()
	}
	baseURL = strings.TrimRight(baseURL, "/")
	c := &Client{
		baseURL:    baseURL,
		log:        log,
		inner:      inner,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}
	if baseURL != "" {
		log.Info("fraud.client.http", "baseURL", baseURL)
	}
	return c
}

// Score posts POST /v1/ai/fraud/score when baseURL is set.
func (c *Client) Score(ctx context.Context, req ports.FraudRequest) (ports.FraudResult, error) {
	if c.baseURL == "" {
		if c.inner != nil {
			return c.inner.Score(ctx, req)
		}
		c.log.Debug("fraud.fallback.dev", "note", "FRAUD_URL empty; using inner/allow")
		return ports.FraudResult{Decision: "allow"}, nil
	}

	features := map[string]float64{
		"amountMinor": float64(req.TotalMinor),
	}
	body := map[string]any{
		"entityType": "checkout_session",
		"entityId":   req.CheckoutID.String(),
		"features":   features,
	}
	var raw struct {
		Predictions map[string]float64 `json:"Predictions"`
		Outputs     map[string]any     `json:"Outputs"`
	}
	if err := c.doJSON(ctx, http.MethodPost, "/v1/ai/fraud/score", req.TenantID.String(), body, &raw); err != nil {
		return ports.FraudResult{}, err
	}

	scoreF := 0.0
	if v, ok := raw.Predictions["risk"]; ok {
		scoreF = v
	} else if v, ok := raw.Predictions["score"]; ok {
		scoreF = v
	} else {
		for _, v := range raw.Predictions {
			scoreF = v
			break
		}
	}
	score := int(math.Round(scoreF))
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	decision := "allow"
	if score >= 80 {
		decision = "block"
	} else if score >= 50 {
		decision = "review"
	}
	return ports.FraudResult{Score: float64(score), Decision: decision}, nil
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
		return fmt.Errorf("fraud http %s %s: status %d: %s", method, path, resp.StatusCode, string(raw))
	}
	if out == nil || len(raw) == 0 {
		return nil
	}
	return json.Unmarshal(raw, out)
}

var _ ports.FraudClient = (*Client)(nil)
