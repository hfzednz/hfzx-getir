package memory

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/nexora/review-service/internal/app/ports"
)

// MockOrders verifies any order with non-nil IDs as purchased+delivered.
type MockOrders struct{}

func (MockOrders) VerifyPurchase(_ context.Context, _, _, orderID, _ uuid.UUID, _ string) (bool, bool, error) {
	if orderID == uuid.Nil {
		return false, false, nil
	}
	return true, true, nil
}

// MockMedia validates non-empty refs.
type MockMedia struct{}

func (MockMedia) ValidateRef(_ context.Context, _ uuid.UUID, mediaRef string) (bool, string, error) {
	if strings.TrimSpace(mediaRef) == "" {
		return false, "", nil
	}
	kind := "image"
	if strings.Contains(mediaRef, "video") {
		kind = "video"
	}
	if strings.Contains(mediaRef, "voice") {
		kind = "voice"
	}
	return true, kind, nil
}

// MockAI is a heuristic moderation stub.
type MockAI struct{}

func (MockAI) Analyze(_ context.Context, _ uuid.UUID, title, body, _ string) (ports.ModerationResult, error) {
	text := strings.ToLower(title + " " + body)
	res := ports.ModerationResult{Sentiment: 0.2}
	if strings.Contains(text, "hate") || strings.Contains(text, "kill") {
		res.UnsafeScore = 0.95
		res.Labels = []string{"hate_speech"}
	}
	if strings.Contains(text, "spam") || strings.Contains(text, "buy now") {
		res.UnsafeScore = 0.7
		res.Labels = append(res.Labels, "spam")
	}
	if strings.Contains(text, "@") || strings.Contains(text, "05") {
		res.PIIFound = true
		res.MaskedBody = strings.ReplaceAll(body, "@", "[at]")
	}
	if strings.Contains(text, "great") || strings.Contains(text, "excellent") {
		res.Sentiment = 0.8
	}
	if strings.Contains(text, "terrible") || strings.Contains(text, "awful") {
		res.Sentiment = -0.8
	}
	return res, nil
}

func (MockAI) Summarize(_ context.Context, _ uuid.UUID, bodies []string) (string, error) {
	if len(bodies) == 0 {
		return "No reviews yet.", nil
	}
	return "Customers mention product quality and delivery speed across " + itoa(len(bodies)) + " reviews.", nil
}

func (MockAI) ExtractTopics(_ context.Context, _ uuid.UUID, body string) ([]string, float64, error) {
	topics := []string{}
	l := strings.ToLower(body)
	for _, t := range []string{"freshness", "packaging", "delivery", "taste", "price", "support"} {
		if strings.Contains(l, t) {
			topics = append(topics, t)
		}
	}
	sent := 0.0
	if strings.Contains(l, "good") || strings.Contains(l, "great") {
		sent = 0.6
	}
	if strings.Contains(l, "bad") || strings.Contains(l, "poor") {
		sent = -0.6
	}
	return topics, sent, nil
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	digits := []byte{}
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}
