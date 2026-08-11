package memory

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/crm-service/internal/app/ports"
)

// EventPublisher records published events in the store.
type EventPublisher struct{ S *Store }

func (p *EventPublisher) Publish(_ context.Context, topic, key string, payload any) error {
	if p.S == nil {
		return nil
	}
	p.S.mu.Lock()
	defer p.S.mu.Unlock()
	m := map[string]any{"topic": topic, "key": key, "payload": payload}
	p.S.PublishedEvents = append(p.S.PublishedEvents, m)
	return nil
}

var _ ports.EventPublisher = (*EventPublisher)(nil)

// MockProfile is a stub ProfileReadClient.
type MockProfile struct {
	Profiles map[uuid.UUID]ports.ProfileSummary
}

func (m *MockProfile) GetProfile(_ context.Context, tenantID, customerID uuid.UUID) (ports.ProfileSummary, error) {
	if m.Profiles != nil {
		if p, ok := m.Profiles[customerID]; ok {
			return p, nil
		}
	}
	return ports.ProfileSummary{
		CustomerID:  customerID,
		DisplayName: "Stub Customer",
		Email:       "customer@example.com",
		Phone:       "+900000000000",
		Tier:        "standard",
	}, nil
}

var _ ports.ProfileReadClient = (*MockProfile)(nil)

// MockOrders is a stub OrderReadClient.
type MockOrders struct {
	Orders map[uuid.UUID][]ports.OrderSummary
}

func (m *MockOrders) ListOrders(_ context.Context, tenantID, customerID uuid.UUID, limit int) ([]ports.OrderSummary, error) {
	if m.Orders != nil {
		if o, ok := m.Orders[customerID]; ok {
			if limit > 0 && len(o) > limit {
				return o[:limit], nil
			}
			return o, nil
		}
	}
	return []ports.OrderSummary{{
		OrderID:   uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
		Status:    "delivered",
		Total:     "149.90",
		Currency:  "TRY",
		CreatedAt: time.Date(2026, 8, 1, 10, 0, 0, 0, time.UTC),
	}}, nil
}

var _ ports.OrderReadClient = (*MockOrders)(nil)

// MockNotify is a stub NotificationClient.
type MockNotify struct {
	mu    sync.Mutex
	Calls []map[string]any
}

func (m *MockNotify) Notify(_ context.Context, tenantID, principalID uuid.UUID, templateKey string, data map[string]any) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = append(m.Calls, map[string]any{
		"tenantId": tenantID, "principalId": principalID, "templateKey": templateKey, "data": data,
	})
	return nil
}

var _ ports.NotificationClient = (*MockNotify)(nil)

// MockRefund is a stub RefundRequestClient.
type MockRefund struct {
	mu    sync.Mutex
	Calls []ports.RefundRequest
}

func (m *MockRefund) RequestRefund(_ context.Context, req ports.RefundRequest) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = append(m.Calls, req)
	return "refund-req-" + req.OrderID.String()[:8], nil
}

var _ ports.RefundRequestClient = (*MockRefund)(nil)

// MockLLM is a configurable LLMClient mock.
type MockLLM struct {
	mu              sync.Mutex
	Intent          ports.IntentResult
	Reply           ports.ReplyResult
	Sentiment       string
	Summary         string
	ForceLowConf    bool
	ForceNegSent    bool
	DetectCalls     int
	DraftCalls      int
	SentimentCalls  int
	SummarizeCalls  int
}

func (m *MockLLM) DetectIntent(_ context.Context, text string) (ports.IntentResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.DetectCalls++
	if m.Intent.Intent != "" {
		return m.Intent, nil
	}
	intent := "general"
	lower := strings.ToLower(text)
	switch {
	case strings.Contains(lower, "refund"):
		intent = "refund"
	case strings.Contains(lower, "order"):
		intent = "order_status"
	case strings.Contains(lower, "delivery"):
		intent = "delivery"
	}
	return ports.IntentResult{Intent: intent, Confidence: 0.9}, nil
}

func (m *MockLLM) DraftReply(_ context.Context, text string, kbSnippets []string) (ports.ReplyResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.DraftCalls++
	if m.ForceLowConf {
		return ports.ReplyResult{Reply: "I'm not sure — connecting you to an agent.", Confidence: 0.3, Sources: kbSnippets}, nil
	}
	if m.Reply.Reply != "" {
		return m.Reply, nil
	}
	reply := "Thanks for reaching out. How can I help?"
	if len(kbSnippets) > 0 {
		reply = "Based on our knowledge base: " + kbSnippets[0]
	}
	return ports.ReplyResult{Reply: reply, Confidence: 0.85, Sources: kbSnippets}, nil
}

func (m *MockLLM) AnalyzeSentiment(_ context.Context, text string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.SentimentCalls++
	if m.ForceNegSent {
		return "negative", nil
	}
	if m.Sentiment != "" {
		return m.Sentiment, nil
	}
	lower := strings.ToLower(text)
	switch {
	case strings.Contains(lower, "angry") || strings.Contains(lower, "terrible") || strings.Contains(lower, "hate"):
		return "negative", nil
	case strings.Contains(lower, "thanks") || strings.Contains(lower, "great") || strings.Contains(lower, "love"):
		return "positive", nil
	default:
		return "neutral", nil
	}
}

func (m *MockLLM) Summarize(_ context.Context, text string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.SummarizeCalls++
	if m.Summary != "" {
		return m.Summary, nil
	}
	if len(text) > 80 {
		return text[:80] + "...", nil
	}
	return text, nil
}

var _ ports.LLMClient = (*MockLLM)(nil)
