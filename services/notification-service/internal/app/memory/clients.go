package memory

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/nexora/notification-service/internal/app/ports"
)

// EventPublisher records published events.
type EventPublisher struct {
	S *Store
}

func (p *EventPublisher) Publish(_ context.Context, topic, key string, payload any) error {
	if p.S == nil {
		return nil
	}
	p.S.mu.Lock()
	defer p.S.mu.Unlock()
	p.S.published = append(p.S.published, PublishedEvent{Topic: topic, Key: key, Payload: payload})
	return nil
}

var _ ports.EventPublisher = (*EventPublisher)(nil)

// MockEmail is a mock EmailProvider (succeeds by default).
type MockEmail struct {
	mu      sync.Mutex
	Fail    error
	Calls   []ports.EmailSendRequest
	NextRef string
}

func (m *MockEmail) Send(_ context.Context, req ports.EmailSendRequest) (ports.EmailSendResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = append(m.Calls, req)
	if m.Fail != nil {
		return ports.EmailSendResult{}, m.Fail
	}
	ref := m.NextRef
	if ref == "" {
		ref = "smtp-" + uuid.NewString()
	}
	return ports.EmailSendResult{ProviderRef: ref}, nil
}

var _ ports.EmailProvider = (*MockEmail)(nil)

// MockSMS is a mock SMSProvider.
type MockSMS struct {
	mu    sync.Mutex
	Fail  error
	Calls []ports.SMSSendRequest
}

func (m *MockSMS) Send(_ context.Context, req ports.SMSSendRequest) (ports.SMSSendResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = append(m.Calls, req)
	if m.Fail != nil {
		return ports.SMSSendResult{}, m.Fail
	}
	return ports.SMSSendResult{ProviderRef: "sms-" + uuid.NewString()}, nil
}

var _ ports.SMSProvider = (*MockSMS)(nil)

// MockWhatsApp is a mock WhatsAppProvider.
type MockWhatsApp struct {
	mu    sync.Mutex
	Fail  error
	Calls []ports.WhatsAppSendRequest
}

func (m *MockWhatsApp) Send(_ context.Context, req ports.WhatsAppSendRequest) (ports.WhatsAppSendResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = append(m.Calls, req)
	if m.Fail != nil {
		return ports.WhatsAppSendResult{}, m.Fail
	}
	return ports.WhatsAppSendResult{ProviderRef: "wa-" + uuid.NewString()}, nil
}

var _ ports.WhatsAppProvider = (*MockWhatsApp)(nil)

// MockPush is a mock PushProvider (FCM/APNs).
type MockPush struct {
	mu    sync.Mutex
	Fail  error
	Calls []ports.PushSendRequest
}

func (m *MockPush) Send(_ context.Context, req ports.PushSendRequest) (ports.PushSendResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = append(m.Calls, req)
	if m.Fail != nil {
		return ports.PushSendResult{}, m.Fail
	}
	return ports.PushSendResult{ProviderRef: "push-" + uuid.NewString()}, nil
}

var _ ports.PushProvider = (*MockPush)(nil)
