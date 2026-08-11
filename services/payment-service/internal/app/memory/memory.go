package memory

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/payment-service/internal/domain"
)

// Store is the in-memory aggregate for payment-service tests and dev mode.
type Store struct {
	mu          sync.RWMutex
	intents     map[uuid.UUID]domain.PaymentIntent
	byIdem      map[string]uuid.UUID // tenant|key → intent
	attempts    map[uuid.UUID][]domain.PaymentAttempt
	methods     map[uuid.UUID]domain.PaymentMethod
	refunds     map[uuid.UUID]domain.Refund
	refundIdem  map[string]uuid.UUID
	chargebacks []domain.Chargeback
	routes      []domain.ProviderRoute
	fraud       []domain.FraudScore
	audits      []domain.AuditEntry
	outbox      []domain.OutboxMessage
	published   []PublishedEvent
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
		intents:    make(map[uuid.UUID]domain.PaymentIntent),
		byIdem:     make(map[string]uuid.UUID),
		attempts:   make(map[uuid.UUID][]domain.PaymentAttempt),
		methods:    make(map[uuid.UUID]domain.PaymentMethod),
		refunds:    make(map[uuid.UUID]domain.Refund),
		refundIdem: make(map[string]uuid.UUID),
	}
}

func idemKey(tenant uuid.UUID, key string) string {
	return tenant.String() + "|" + key
}

// Clock is a fixed clock for tests.
type Clock struct{ T time.Time }

func (c *Clock) Now() time.Time { return c.T }

// IDGen returns sequential UUIDs for tests when Seq is set; otherwise random.
type IDGen struct{}

func (IDGen) New() uuid.UUID { return uuid.New() }
