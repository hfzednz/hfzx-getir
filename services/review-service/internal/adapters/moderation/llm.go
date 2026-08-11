package moderation

import (
	"context"

	"github.com/google/uuid"
	"github.com/nexora/review-service/internal/app/ports"
	"github.com/nexora/review-service/internal/app/memory"
)

// LLMClient wraps the mock AI for production wiring.
type LLMClient struct {
	Endpoint string
	APIKey   string
	inner    ports.ModerationClient
}

// NewLLMClient returns an LLM moderation client (falls back to heuristic mock).
func NewLLMClient(endpoint, apiKey string) *LLMClient {
	return &LLMClient{Endpoint: endpoint, APIKey: apiKey, inner: memory.MockAI{}}
}

// Analyze delegates to inner client until real LLM is configured.
func (c *LLMClient) Analyze(ctx context.Context, tenantID uuid.UUID, title, body, locale string) (ports.ModerationResult, error) {
	return c.inner.Analyze(ctx, tenantID, title, body, locale)
}

// Summarize delegates to inner client.
func (c *LLMClient) Summarize(ctx context.Context, tenantID uuid.UUID, bodies []string) (string, error) {
	return c.inner.Summarize(ctx, tenantID, bodies)
}

// ExtractTopics delegates to inner client.
func (c *LLMClient) ExtractTopics(ctx context.Context, tenantID uuid.UUID, body string) ([]string, float64, error) {
	return c.inner.ExtractTopics(ctx, tenantID, body)
}
