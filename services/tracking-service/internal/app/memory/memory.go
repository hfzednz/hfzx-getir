package memory

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/tracking-service/internal/domain"
)

type courierKey struct {
	Tenant  uuid.UUID
	Courier uuid.UUID
}

// Store is the in-memory aggregate for tracking-service.
type Store struct {
	mu sync.RWMutex

	latest    map[courierKey]domain.CourierLocation
	history   map[courierKey][]domain.LocationHistoryEntry
	timelines []domain.TimelineEvent
	geofence  []domain.GeofenceEvent
	outbox    []domain.OutboxMessage
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
		latest:  make(map[courierKey]domain.CourierLocation),
		history: make(map[courierKey][]domain.LocationHistoryEntry),
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

// HistoryLen returns history length for a courier (test helper).
func (s *Store) HistoryLen(tenantID, courierID uuid.UUID) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.history[courierKey{Tenant: tenantID, Courier: courierID}])
}
