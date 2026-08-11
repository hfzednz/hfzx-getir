package memory

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/pricing-service/internal/domain"
)

// Store is the in-memory aggregate for pricing-service.
type Store struct {
	mu sync.RWMutex

	books    map[uuid.UUID]domain.PriceBook
	entries  map[uuid.UUID]domain.PriceEntry
	taxes    map[uuid.UUID]domain.TaxRule
	dynamics map[uuid.UUID]domain.DynamicRule
	audits   []domain.QuoteAudit
	outbox   []domain.OutboxMessage
	published []PublishedEvent
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
		books:    make(map[uuid.UUID]domain.PriceBook),
		entries:  make(map[uuid.UUID]domain.PriceEntry),
		taxes:    make(map[uuid.UUID]domain.TaxRule),
		dynamics: make(map[uuid.UUID]domain.DynamicRule),
	}
}

// Clock is a fixed-time clock for tests.
type Clock struct{ T time.Time }

func (c *Clock) Now() time.Time {
	if c.T.IsZero() {
		return time.Now().UTC()
	}
	return c.T.UTC()
}

// IDGen generates random UUIDs.
type IDGen struct{}

func (IDGen) New() uuid.UUID { return uuid.New() }

// EventPublisher records publishes into the store.
type EventPublisher struct{ S *Store }

func (p *EventPublisher) Publish(_ context.Context, topic, key string, payload any) error {
	if p.S == nil {
		return nil
	}
	p.S.mu.Lock()
	defer p.S.mu.Unlock()
	p.S.published = append(p.S.published, PublishedEvent{Topic: topic, Key: key, Payload: payload})
	return nil
}
