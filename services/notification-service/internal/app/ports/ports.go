package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/notification-service/internal/domain"
)

// Clock provides the current time.
type Clock interface {
	Now() time.Time
}

// IDGen generates UUIDs.
type IDGen interface {
	New() uuid.UUID
}

// EventPublisher publishes domain events.
type EventPublisher interface {
	Publish(ctx context.Context, topic, key string, payload any) error
}

// TemplateRepo persists templates.
type TemplateRepo interface {
	Upsert(ctx context.Context, t domain.Template) (domain.Template, error)
	Get(ctx context.Context, tenantID, id uuid.UUID) (domain.Template, error)
	GetByKey(ctx context.Context, tenantID uuid.UUID, key string, channel domain.Channel, locale string) (domain.Template, error)
	Approve(ctx context.Context, tenantID, id uuid.UUID, now time.Time) (domain.Template, error)
	List(ctx context.Context, tenantID uuid.UUID, key string) ([]domain.Template, error)
}

// MessageRepo persists messages.
type MessageRepo interface {
	Create(ctx context.Context, m domain.Message) error
	Update(ctx context.Context, m domain.Message) error
	Get(ctx context.Context, tenantID, id uuid.UUID) (domain.Message, error)
	GetByIdempotency(ctx context.Context, tenantID uuid.UUID, key string) (domain.Message, error)
	ListFailed(ctx context.Context, tenantID uuid.UUID, limit int) ([]domain.Message, error)
	CountByStatus(ctx context.Context, tenantID uuid.UUID) (map[domain.MessageStatus]int, error)
}

// PreferenceRepo persists preferences and consents.
type PreferenceRepo interface {
	Get(ctx context.Context, tenantID, principalID uuid.UUID) (domain.Preference, error)
	Upsert(ctx context.Context, p domain.Preference) error
	RecordConsent(ctx context.Context, c domain.Consent) error
	ListConsents(ctx context.Context, tenantID, principalID uuid.UUID) ([]domain.Consent, error)
}

// DeviceRepo persists push devices.
type DeviceRepo interface {
	Upsert(ctx context.Context, d domain.Device) (domain.Device, error)
	ListActive(ctx context.Context, tenantID, principalID uuid.UUID) ([]domain.Device, error)
	Deactivate(ctx context.Context, tenantID, id uuid.UUID) error
}

// InboxRepo persists inbox items.
type InboxRepo interface {
	Create(ctx context.Context, item domain.InboxItem) error
	List(ctx context.Context, tenantID, principalID uuid.UUID, includeArchived bool) ([]domain.InboxItem, error)
	MarkRead(ctx context.Context, tenantID, id uuid.UUID, now time.Time) (domain.InboxItem, error)
	Archive(ctx context.Context, tenantID, id uuid.UUID) error
}

// ScheduleRepo persists schedules.
type ScheduleRepo interface {
	Create(ctx context.Context, s domain.Schedule) error
	Get(ctx context.Context, tenantID, id uuid.UUID) (domain.Schedule, error)
	Update(ctx context.Context, s domain.Schedule) error
	ListDue(ctx context.Context, now time.Time, limit int) ([]domain.Schedule, error)
	Cancel(ctx context.Context, tenantID, id uuid.UUID, now time.Time) (domain.Schedule, error)
}

// DeliveryRepo persists deliveries, events, DLQ, routes.
type DeliveryRepo interface {
	CreateAttempt(ctx context.Context, a domain.DeliveryAttempt) error
	ListAttempts(ctx context.Context, tenantID, messageID uuid.UUID) ([]domain.DeliveryAttempt, error)
	CreateEvent(ctx context.Context, e domain.DeliveryEvent) error
	MoveToDLQ(ctx context.Context, item domain.DLQItem) error
	ListDLQ(ctx context.Context, tenantID uuid.UUID, limit int) ([]domain.DLQItem, error)
	UpsertRoute(ctx context.Context, r domain.ProviderRoute) error
	ListRoutes(ctx context.Context, tenantID uuid.UUID) ([]domain.ProviderRoute, error)
}

// OutboxRepository is the transactional outbox.
type OutboxRepository interface {
	Enqueue(ctx context.Context, m domain.OutboxMessage) error
	ListPending(ctx context.Context, limit int) ([]domain.OutboxMessage, error)
	Update(ctx context.Context, m domain.OutboxMessage) error
}

// PushSendRequest is a push provider request.
type PushSendRequest struct {
	TenantID    uuid.UUID
	PrincipalID uuid.UUID
	Token       string
	Platform    domain.DevicePlatform
	Title       string
	Body        string
	Data        map[string]string
}

// PushSendResult is a push provider receipt.
type PushSendResult struct {
	ProviderRef string
}

// PushProvider sends push notifications (FCM/APNs).
type PushProvider interface {
	Send(ctx context.Context, req PushSendRequest) (PushSendResult, error)
}

// EmailSendRequest is an SMTP request.
type EmailSendRequest struct {
	TenantID uuid.UUID
	To       string
	Subject  string
	Body     string
}

// EmailSendResult is an email receipt.
type EmailSendResult struct {
	ProviderRef string
}

// EmailProvider sends email.
type EmailProvider interface {
	Send(ctx context.Context, req EmailSendRequest) (EmailSendResult, error)
}

// SMSSendRequest is an SMS request.
type SMSSendRequest struct {
	TenantID uuid.UUID
	To       string
	Body     string
}

// SMSSendResult is an SMS receipt.
type SMSSendResult struct {
	ProviderRef string
}

// SMSProvider sends SMS.
type SMSProvider interface {
	Send(ctx context.Context, req SMSSendRequest) (SMSSendResult, error)
}

// WhatsAppSendRequest is a WhatsApp request.
type WhatsAppSendRequest struct {
	TenantID uuid.UUID
	To       string
	Body     string
}

// WhatsAppSendResult is a WhatsApp receipt.
type WhatsAppSendResult struct {
	ProviderRef string
}

// WhatsAppProvider sends WhatsApp messages.
type WhatsAppProvider interface {
	Send(ctx context.Context, req WhatsAppSendRequest) (WhatsAppSendResult, error)
}
