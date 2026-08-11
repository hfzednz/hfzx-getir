package sms

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/identity-service/internal/app/ports"
)

// TwilioSender delivers OTP SMS via Twilio REST API.
type TwilioSender struct {
	AccountSID string
	AuthToken  string
	FromNumber string
	HTTPClient *http.Client
	BaseURL    string
}

func NewTwilioFromEnv() (*TwilioSender, error) {
	sid := os.Getenv("TWILIO_ACCOUNT_SID")
	token := os.Getenv("TWILIO_AUTH_TOKEN")
	from := os.Getenv("TWILIO_FROM_NUMBER")
	if sid == "" || token == "" || from == "" {
		return nil, fmt.Errorf("sms: TWILIO_ACCOUNT_SID, TWILIO_AUTH_TOKEN, TWILIO_FROM_NUMBER required")
	}
	return &TwilioSender{
		AccountSID: sid,
		AuthToken:  token,
		FromNumber: from,
		HTTPClient: &http.Client{Timeout: 15 * time.Second},
		BaseURL:    "https://api.twilio.com",
	}, nil
}

func (s *TwilioSender) SendOTP(ctx context.Context, tenantID uuid.UUID, phone, code string) error {
	_ = tenantID
	phone = strings.TrimSpace(phone)
	if phone == "" || code == "" {
		return fmt.Errorf("sms: phone and code required")
	}
	endpoint := fmt.Sprintf("%s/2010-04-01/Accounts/%s/Messages.json", strings.TrimRight(s.BaseURL, "/"), s.AccountSID)
	form := url.Values{}
	form.Set("To", phone)
	form.Set("From", s.FromNumber)
	form.Set("Body", fmt.Sprintf("NEXORA verification code: %s", code))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.SetBasicAuth(s.AccountSID, s.AuthToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("sms: twilio request: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("sms: twilio status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// HTTPWebhookSender posts OTP to a configurable webhook (generic SMS gateway).
type HTTPWebhookSender struct {
	URL        string
	APIKey     string
	HTTPClient *http.Client
}

func NewWebhookFromEnv() (*HTTPWebhookSender, error) {
	u := os.Getenv("SMS_WEBHOOK_URL")
	if u == "" {
		return nil, fmt.Errorf("sms: SMS_WEBHOOK_URL required")
	}
	return &HTTPWebhookSender{
		URL:        u,
		APIKey:     os.Getenv("SMS_WEBHOOK_API_KEY"),
		HTTPClient: &http.Client{Timeout: 15 * time.Second},
	}, nil
}

func (s *HTTPWebhookSender) SendOTP(ctx context.Context, tenantID uuid.UUID, phone, code string) error {
	payload, _ := json.Marshal(map[string]any{
		"tenantId": tenantID.String(),
		"phone":    phone,
		"code":     code,
		"channel":  "sms",
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.URL, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if s.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+s.APIKey)
	}
	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("sms: webhook: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return fmt.Errorf("sms: webhook status %d: %s", resp.StatusCode, string(b))
	}
	return nil
}

var (
	_ ports.OTPSender = (*TwilioSender)(nil)
	_ ports.OTPSender = (*HTTPWebhookSender)(nil)
)
