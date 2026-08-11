package providers

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/nexora/notification-service/internal/app/ports"
)

// MockEmail is the default SMTP stub (succeeds).
type MockEmail struct{ Log *slog.Logger }

func (m *MockEmail) Send(_ context.Context, req ports.EmailSendRequest) (ports.EmailSendResult, error) {
	if m.Log != nil {
		m.Log.Debug("smtp.mock.send", "to", req.To, "subject", req.Subject)
	}
	return ports.EmailSendResult{ProviderRef: "smtp-mock-" + uuid.NewString()}, nil
}

var _ ports.EmailProvider = (*MockEmail)(nil)

// MockSMS is the default SMS stub.
type MockSMS struct{ Log *slog.Logger }

func (m *MockSMS) Send(_ context.Context, req ports.SMSSendRequest) (ports.SMSSendResult, error) {
	if m.Log != nil {
		m.Log.Debug("sms.mock.send", "to", req.To)
	}
	return ports.SMSSendResult{ProviderRef: "sms-mock-" + uuid.NewString()}, nil
}

var _ ports.SMSProvider = (*MockSMS)(nil)

// MockWhatsApp is the default WhatsApp stub.
type MockWhatsApp struct{ Log *slog.Logger }

func (m *MockWhatsApp) Send(_ context.Context, req ports.WhatsAppSendRequest) (ports.WhatsAppSendResult, error) {
	if m.Log != nil {
		m.Log.Debug("whatsapp.mock.send", "to", req.To)
	}
	return ports.WhatsAppSendResult{ProviderRef: "wa-mock-" + uuid.NewString()}, nil
}

var _ ports.WhatsAppProvider = (*MockWhatsApp)(nil)

// MockPush is the default FCM/APNs stub.
type MockPush struct{ Log *slog.Logger }

func (m *MockPush) Send(_ context.Context, req ports.PushSendRequest) (ports.PushSendResult, error) {
	if m.Log != nil {
		m.Log.Debug("push.mock.send", "token", req.Token, "platform", req.Platform)
	}
	return ports.PushSendResult{ProviderRef: "push-mock-" + uuid.NewString()}, nil
}

var _ ports.PushProvider = (*MockPush)(nil)
