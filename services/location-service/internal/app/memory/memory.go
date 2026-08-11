package memory

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/location-service/internal/domain"
)

type geocodeEntry struct {
	Result    domain.GeocodeResult
	ExpiresAt time.Time
}

// Store is the in-memory aggregate for location-service.
type Store struct {
	mu sync.RWMutex

	addresses map[uuid.UUID]domain.NormalizedAddress
	pois      map[uuid.UUID]domain.POI
	history   []domain.LocationHistory
	geocode   map[string]geocodeEntry
	manifests map[string]domain.OfflineManifest // tenant|region
	heat      map[string]domain.HeatCell        // tenant|grid
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
		addresses: make(map[uuid.UUID]domain.NormalizedAddress),
		pois:      make(map[uuid.UUID]domain.POI),
		geocode:   make(map[string]geocodeEntry),
		manifests: make(map[string]domain.OfflineManifest),
		heat:      make(map[string]domain.HeatCell),
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

// Published returns a copy of recorded events.
func (s *Store) Published() []PublishedEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]PublishedEvent, len(s.published))
	copy(out, s.published)
	return out
}
