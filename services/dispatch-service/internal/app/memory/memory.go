package memory

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/dispatch-service/internal/domain"
)

// Store is the in-memory aggregate for dispatch-service.
type Store struct {
	mu sync.RWMutex

	dispatches map[uuid.UUID]domain.Dispatch
	events     []domain.DispatchEvent
	attempts   []domain.AssignmentAttempt
	batches    map[uuid.UUID]domain.Batch
	couriers   map[string]domain.CourierSnapshot // tenant|courierPrincipal
	vehicles   map[uuid.UUID]domain.Vehicle
	outbox     []domain.OutboxMessage
	published  []PublishedEvent
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
		dispatches: make(map[uuid.UUID]domain.Dispatch),
		batches:    make(map[uuid.UUID]domain.Batch),
		couriers:   make(map[string]domain.CourierSnapshot),
		vehicles:   make(map[uuid.UUID]domain.Vehicle),
	}
}

func courierKey(tenantID, courierPrincipalID uuid.UUID) string {
	return tenantID.String() + "|" + courierPrincipalID.String()
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

// Attempts returns assignment attempts (test helper).
func (s *Store) Attempts() []domain.AssignmentAttempt {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]domain.AssignmentAttempt, len(s.attempts))
	copy(out, s.attempts)
	return out
}
