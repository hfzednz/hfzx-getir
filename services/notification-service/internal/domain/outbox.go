package domain

import (
	"time"

	"github.com/google/uuid"
)

// Outbox statuses.
const (
	OutboxStatusPending   = "pending"
	OutboxStatusPublished = "published"
	OutboxStatusFailed    = "failed"
)

// Event type constants (outbound).
const (
	EventNotificationQueued = "NotificationQueued"
	EventNotificationSent   = "NotificationSent"
	EventNotificationFailed = "NotificationFailed"
	EventNotificationOpened = "NotificationOpened"
	EventNotificationClicked = "NotificationClicked"
)

// TopicNotificationDelivery is the primary outbound topic.
const TopicNotificationDelivery = "notification.delivery"

// TopicForEvent maps event type to Kafka topic.
func TopicForEvent(eventType string) string {
	return TopicNotificationDelivery
}

// OutboxMessage is a transactional outbox row.
type OutboxMessage struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	MessageID   uuid.UUID
	Topic       string
	Key         string
	Payload     map[string]any
	Status      string
	Attempts    int
	LastError   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	PublishedAt *time.Time
}

// Domain event names consumed (stub handlers).
const (
	DomainEventOrderCreated         = "OrderCreated"
	DomainEventOrderDelivered       = "OrderDelivered"
	DomainEventPaymentSuccess       = "PaymentSuccess"
	DomainEventRefundCompleted      = "RefundCompleted"
	DomainEventRewardEarned         = "RewardEarned"
	DomainEventCouponIssued         = "CouponIssued"
	DomainEventMembershipUpgraded   = "MembershipUpgraded"
	DomainEventSupportTicketCreated = "SupportTicketCreated"
)

// TemplateKeyForDomainEvent maps inbound domain events to template keys.
func TemplateKeyForDomainEvent(eventType string) (string, bool) {
	switch eventType {
	case DomainEventOrderCreated:
		return "order.created", true
	case DomainEventOrderDelivered:
		return "order.delivered", true
	case DomainEventPaymentSuccess:
		return "payment.success", true
	case DomainEventRefundCompleted:
		return "payment.refund", true
	case DomainEventRewardEarned:
		return "loyalty.reward_earned", true
	case DomainEventCouponIssued:
		return "promo.coupon_issued", true
	case DomainEventMembershipUpgraded:
		return "loyalty.membership_upgraded", true
	case DomainEventSupportTicketCreated:
		return "support.ticket_created", true
	default:
		return "", false
	}
}
