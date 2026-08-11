package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/smtp"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/nexora/notification-service/internal/app/ports"
)

// SMTPEmail sends email via SMTP (STARTTLS optional via server address).
type SMTPEmail struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
}

func NewSMTPFromEnv() (*SMTPEmail, error) {
	host := os.Getenv("SMTP_HOST")
	from := os.Getenv("SMTP_FROM")
	if host == "" || from == "" {
		return nil, fmt.Errorf("email: SMTP_HOST and SMTP_FROM required")
	}
	port := os.Getenv("SMTP_PORT")
	if port == "" {
		port = "587"
	}
	return &SMTPEmail{
		Host: host, Port: port, Username: os.Getenv("SMTP_USERNAME"),
		Password: os.Getenv("SMTP_PASSWORD"), From: from,
	}, nil
}

func (s *SMTPEmail) Send(_ context.Context, req ports.EmailSendRequest) (ports.EmailSendResult, error) {
	addr := s.Host + ":" + s.Port
	msg := []byte(fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		s.From, req.To, req.Subject, req.Body))
	var auth smtp.Auth
	if s.Username != "" {
		auth = smtp.PlainAuth("", s.Username, s.Password, s.Host)
	}
	if err := smtp.SendMail(addr, auth, s.From, []string{req.To}, msg); err != nil {
		return ports.EmailSendResult{}, err
	}
	return ports.EmailSendResult{ProviderRef: "smtp:" + req.To}, nil
}

var _ ports.EmailProvider = (*SMTPEmail)(nil)

// TwilioSMS sends SMS via Twilio.
type TwilioSMS struct {
	AccountSID string
	AuthToken  string
	From       string
	HTTPClient *http.Client
}

func NewTwilioSMSFromEnv() (*TwilioSMS, error) {
	sid := os.Getenv("TWILIO_ACCOUNT_SID")
	token := os.Getenv("TWILIO_AUTH_TOKEN")
	from := os.Getenv("TWILIO_FROM_NUMBER")
	if sid == "" || token == "" || from == "" {
		return nil, fmt.Errorf("sms: TWILIO_* required")
	}
	return &TwilioSMS{AccountSID: sid, AuthToken: token, From: from, HTTPClient: &http.Client{Timeout: 15 * time.Second}}, nil
}

func (s *TwilioSMS) Send(ctx context.Context, req ports.SMSSendRequest) (ports.SMSSendResult, error) {
	endpoint := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", s.AccountSID)
	form := url.Values{}
	form.Set("To", req.To)
	form.Set("From", s.From)
	form.Set("Body", req.Body)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return ports.SMSSendResult{}, err
	}
	httpReq.SetBasicAuth(s.AccountSID, s.AuthToken)
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := s.HTTPClient.Do(httpReq)
	if err != nil {
		return ports.SMSSendResult{}, err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if resp.StatusCode >= 300 {
		return ports.SMSSendResult{}, fmt.Errorf("sms: twilio %d: %s", resp.StatusCode, string(b))
	}
	var out struct {
		SID string `json:"sid"`
	}
	_ = json.Unmarshal(b, &out)
	return ports.SMSSendResult{ProviderRef: out.SID}, nil
}

var _ ports.SMSProvider = (*TwilioSMS)(nil)

// FCMPush sends Android/web push via Firebase Cloud Messaging HTTP v1.
type FCMPush struct {
	ServerKey  string
	HTTPClient *http.Client
}

func NewFCMFromEnv() (*FCMPush, error) {
	key := os.Getenv("FCM_SERVER_KEY")
	if key == "" {
		return nil, fmt.Errorf("push: FCM_SERVER_KEY required")
	}
	return &FCMPush{ServerKey: key, HTTPClient: &http.Client{Timeout: 15 * time.Second}}, nil
}

func (p *FCMPush) Send(ctx context.Context, req ports.PushSendRequest) (ports.PushSendResult, error) {
	payload := map[string]any{
		"to": req.Token,
		"notification": map[string]string{
			"title": req.Title,
			"body":  req.Body,
		},
		"data": req.Data,
	}
	body, _ := json.Marshal(payload)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://fcm.googleapis.com/fcm/send", bytes.NewReader(body))
	if err != nil {
		return ports.PushSendResult{}, err
	}
	httpReq.Header.Set("Authorization", "key="+p.ServerKey)
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := p.HTTPClient.Do(httpReq)
	if err != nil {
		return ports.PushSendResult{}, err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if resp.StatusCode >= 300 {
		return ports.PushSendResult{}, fmt.Errorf("push: fcm %d: %s", resp.StatusCode, string(b))
	}
	var out struct {
		MessageID int64 `json:"message_id"`
	}
	_ = json.Unmarshal(b, &out)
	return ports.PushSendResult{ProviderRef: fmt.Sprintf("%d", out.MessageID)}, nil
}

var _ ports.PushProvider = (*FCMPush)(nil)

// WhatsAppCloudAPI sends via Meta WhatsApp Cloud API.
type WhatsAppCloudAPI struct {
	Token      string
	PhoneID    string
	HTTPClient *http.Client
}

func NewWhatsAppFromEnv() (*WhatsAppCloudAPI, error) {
	token := os.Getenv("WHATSAPP_TOKEN")
	phoneID := os.Getenv("WHATSAPP_PHONE_NUMBER_ID")
	if token == "" || phoneID == "" {
		return nil, fmt.Errorf("whatsapp: WHATSAPP_TOKEN and WHATSAPP_PHONE_NUMBER_ID required")
	}
	return &WhatsAppCloudAPI{Token: token, PhoneID: phoneID, HTTPClient: &http.Client{Timeout: 15 * time.Second}}, nil
}

func (w *WhatsAppCloudAPI) Send(ctx context.Context, req ports.WhatsAppSendRequest) (ports.WhatsAppSendResult, error) {
	endpoint := fmt.Sprintf("https://graph.facebook.com/v19.0/%s/messages", w.PhoneID)
	payload, _ := json.Marshal(map[string]any{
		"messaging_product": "whatsapp",
		"to":                req.To,
		"type":              "text",
		"text":              map[string]string{"body": req.Body},
	})
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return ports.WhatsAppSendResult{}, err
	}
	httpReq.Header.Set("Authorization", "Bearer "+w.Token)
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := w.HTTPClient.Do(httpReq)
	if err != nil {
		return ports.WhatsAppSendResult{}, err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if resp.StatusCode >= 300 {
		return ports.WhatsAppSendResult{}, fmt.Errorf("whatsapp: %d: %s", resp.StatusCode, string(b))
	}
	var out struct {
		Messages []struct {
			ID string `json:"id"`
		} `json:"messages"`
	}
	_ = json.Unmarshal(b, &out)
	ref := ""
	if len(out.Messages) > 0 {
		ref = out.Messages[0].ID
	}
	return ports.WhatsAppSendResult{ProviderRef: ref}, nil
}

var _ ports.WhatsAppProvider = (*WhatsAppCloudAPI)(nil)
