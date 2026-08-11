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

	"github.com/nexora/payment-service/internal/app/memory"
	"github.com/nexora/payment-service/internal/app/ports"
	"github.com/nexora/payment-service/internal/domain"
)

// Client scores payments via ai-platform when BaseURL is set.
type Client struct {
	BaseURL    string
	log        *slog.Logger
	inner      *memory.FraudClient
	HTTPClient *http.Client
}

// NewClient returns a fraud client. Empty baseURL keeps memory stub.
func NewClient(baseURL string, log *slog.Logger) *Client {
	if log == nil {
		log = slog.Default()
	}
	baseURL = strings.TrimRight(baseURL, "/")
	c := &Client{
		BaseURL:    baseURL,
		log:        log,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
		inner:      &memory.FraudClient{RiskScore: 5, Decision: domain.FraudAllow},
	}
	if baseURL != "" {
		log.Info("fraud.client.http", "baseURL", baseURL)
	}
	return c
}

func (c *Client) Score(ctx context.Context, req ports.FraudRequest) (ports.FraudResult, error) {
	if c.BaseURL == "" {
		return c.inner.Score(ctx, req)
	}
	features := map[string]float64{
		"amountMinor": float64(req.AmountMinor),
	}
	body := map[string]any{
		"entityType": "payment_intent",
		"entityId":   req.IntentID,
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
	decision := domain.FraudAllow
	if score >= 80 {
		decision = domain.FraudBlock
	} else if score >= 50 {
		decision = domain.FraudChallenge
	}
	return ports.FraudResult{Score: score, Decision: decision}, nil
}

func (c *Client) doJSON(ctx context.Context, method, path, tenantID string, in any, out any) error {
	b, err := json.Marshal(in)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, bytes.NewReader(b))
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
