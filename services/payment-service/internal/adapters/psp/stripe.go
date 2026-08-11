package psp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/nexora/payment-service/internal/app/ports"
)

// StripeClient implements PSPClient against Stripe PaymentIntents API.
type StripeClient struct {
	SecretKey  string
	HTTPClient *http.Client
	BaseURL    string
	name       string
}

func NewStripeFromEnv() (*StripeClient, error) {
	key := os.Getenv("STRIPE_SECRET_KEY")
	if key == "" {
		return nil, fmt.Errorf("psp: STRIPE_SECRET_KEY required")
	}
	return &StripeClient{
		SecretKey:  key,
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
		BaseURL:    "https://api.stripe.com",
		name:       "stripe",
	}, nil
}

func (s *StripeClient) Name() string { return s.name }

func (s *StripeClient) Authorize(ctx context.Context, req ports.AuthorizeRequest) (ports.AuthorizeResult, error) {
	form := url.Values{}
	form.Set("amount", fmt.Sprintf("%d", req.AmountMinor))
	form.Set("currency", strings.ToLower(req.Currency))
	form.Set("confirm", "true")
	form.Set("capture_method", "manual")
	if req.Token != "" {
		form.Set("payment_method", req.Token)
	}
	if req.IdempotencyKey != "" {
		form.Set("metadata[idempotency_key]", req.IdempotencyKey)
	}
	if req.Metadata != nil {
		if orderID, ok := req.Metadata["orderId"].(string); ok && orderID != "" {
			form.Set("metadata[order_id]", orderID)
		}
	}
	body, status, err := s.post(ctx, "/v1/payment_intents", form, req.IdempotencyKey)
	if err != nil {
		return ports.AuthorizeResult{}, err
	}
	if status >= 400 {
		return ports.AuthorizeResult{Success: false, ErrorCode: "stripe_error", ErrorMessage: string(body)}, nil
	}
	var pi struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	if err := json.Unmarshal(body, &pi); err != nil {
		return ports.AuthorizeResult{}, err
	}
	ok := pi.Status == "requires_capture" || pi.Status == "succeeded"
	return ports.AuthorizeResult{ProviderRef: pi.ID, Success: ok}, nil
}

func (s *StripeClient) Capture(ctx context.Context, req ports.CaptureRequest) (ports.CaptureResult, error) {
	form := url.Values{}
	if req.AmountMinor > 0 {
		form.Set("amount_to_capture", fmt.Sprintf("%d", req.AmountMinor))
	}
	path := fmt.Sprintf("/v1/payment_intents/%s/capture", url.PathEscape(req.ProviderRef))
	body, status, err := s.post(ctx, path, form, req.IdempotencyKey)
	if err != nil {
		return ports.CaptureResult{}, err
	}
	if status >= 400 {
		return ports.CaptureResult{Success: false, ErrorCode: "stripe_error", ErrorMessage: string(body)}, nil
	}
	var pi struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	_ = json.Unmarshal(body, &pi)
	return ports.CaptureResult{ProviderRef: pi.ID, Success: pi.Status == "succeeded"}, nil
}

func (s *StripeClient) Void(ctx context.Context, req ports.VoidRequest) (ports.VoidResult, error) {
	path := fmt.Sprintf("/v1/payment_intents/%s/cancel", url.PathEscape(req.ProviderRef))
	body, status, err := s.post(ctx, path, url.Values{}, req.IdempotencyKey)
	if err != nil {
		return ports.VoidResult{}, err
	}
	if status >= 400 {
		return ports.VoidResult{Success: false, ErrorCode: "stripe_error", ErrorMessage: string(body)}, nil
	}
	var pi struct {
		ID string `json:"id"`
	}
	_ = json.Unmarshal(body, &pi)
	return ports.VoidResult{ProviderRef: pi.ID, Success: true}, nil
}

func (s *StripeClient) Refund(ctx context.Context, req ports.RefundRequest) (ports.RefundResult, error) {
	form := url.Values{}
	form.Set("payment_intent", req.ProviderRef)
	if req.AmountMinor > 0 {
		form.Set("amount", fmt.Sprintf("%d", req.AmountMinor))
	}
	body, status, err := s.post(ctx, "/v1/refunds", form, req.IdempotencyKey)
	if err != nil {
		return ports.RefundResult{}, err
	}
	if status >= 400 {
		return ports.RefundResult{Success: false, ErrorCode: "stripe_error", ErrorMessage: string(body)}, nil
	}
	var rf struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	_ = json.Unmarshal(body, &rf)
	return ports.RefundResult{ProviderRef: rf.ID, Success: rf.Status == "succeeded" || rf.Status == "pending"}, nil
}

func (s *StripeClient) post(ctx context.Context, path string, form url.Values, idem string) ([]byte, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(s.BaseURL, "/")+path, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, 0, err
	}
	req.SetBasicAuth(s.SecretKey, "")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if idem != "" {
		req.Header.Set("Idempotency-Key", idem)
	}
	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	return b, resp.StatusCode, nil
}

var _ ports.PSPClient = (*StripeClient)(nil)
