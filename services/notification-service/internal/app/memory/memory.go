package memory

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/notification-service/internal/domain"
)

// Store is the in-memory aggregate for notification-service.
type Store struct {
	mu sync.RWMutex

	templates    map[uuid.UUID]domain.Template
	templateIdx  map[string]uuid.UUID // tenant|key|channel|locale → latest id
	messages     map[uuid.UUID]domain.Message
	msgIdem      map[string]uuid.UUID
	preferences  map[string]domain.Preference
	consents     []domain.Consent
	devices      map[uuid.UUID]domain.Device
	deviceToken  map[string]uuid.UUID
	inbox        map[uuid.UUID]domain.InboxItem
	schedules    map[uuid.UUID]domain.Schedule
	attempts     []domain.DeliveryAttempt
	events       []domain.DeliveryEvent
	dlq          []domain.DLQItem
	routes       map[uuid.UUID]domain.ProviderRoute
	outbox       []domain.OutboxMessage
	published    []PublishedEvent
}

// PublishedEvent is a recorded EventPublisher call.
type PublishedEvent struct {
	Topic   string
	Key     string
	Payload any
}

// NewStore creates an empty memory store.
func NewStore() *Store {
	return &Store{
		templates:   make(map[uuid.UUID]domain.Template),
		templateIdx: make(map[string]uuid.UUID),
		messages:    make(map[uuid.UUID]domain.Message),
		msgIdem:     make(map[string]uuid.UUID),
		preferences: make(map[string]domain.Preference),
		devices:     make(map[uuid.UUID]domain.Device),
		deviceToken: make(map[string]uuid.UUID),
		inbox:       make(map[uuid.UUID]domain.InboxItem),
		schedules:   make(map[uuid.UUID]domain.Schedule),
		routes:      make(map[uuid.UUID]domain.ProviderRoute),
	}
}

func tenantKey(tenant uuid.UUID, parts ...string) string {
	s := tenant.String()
	for _, p := range parts {
		s += "|" + p
	}
	return s
}

func prinKey(tenant, principal uuid.UUID) string {
	return tenant.String() + "|" + principal.String()
}

func tplKey(tenant uuid.UUID, key string, channel domain.Channel, locale string) string {
	return tenant.String() + "|" + key + "|" + string(channel) + "|" + locale
}

func idemKey(tenant uuid.UUID, key string) string {
	return tenant.String() + "|" + key
}

// Clock is a fixed/testable clock.
type Clock struct{ T time.Time }

func (c *Clock) Now() time.Time {
	if c.T.IsZero() {
		return time.Now().UTC()
	}
	return c.T.UTC()
}

// Advance moves the clock forward.
func (c *Clock) Advance(d time.Duration) { c.T = c.T.Add(d) }

// IDGen generates UUIDs.
type IDGen struct{}

func (IDGen) New() uuid.UUID { return uuid.New() }
