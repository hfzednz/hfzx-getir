// Package memory provides in-memory port implementations for tests and dev mode.
package memory

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/checkout-service/internal/app"
	"github.com/nexora/checkout-service/internal/app/ports"
	"github.com/nexora/checkout-service/internal/domain"
)

// Store is a shared in-memory database for all repositories.
type Store struct {
	mu sync.RWMutex

	Sessions       map[uuid.UUID]domain.Session
	SessionsByIdem map[string]uuid.UUID // tenant:key
	ByRecovery     map[string]uuid.UUID
	Events         map[uuid.UUID]domain.SessionEvent
	Outbox         map[uuid.UUID]domain.OutboxMessage
	Published      []publishedEvent
	Carts          map[uuid.UUID]ports.CartView

	keyMu sync.Map // string -> *sync.Mutex
}

type publishedEvent struct {
	Topic   string
	Key     string
	Payload any
}

// NewStore returns an empty in-memory store.
func NewStore() *Store {
	return &Store{
		Sessions:       make(map[uuid.UUID]domain.Session),
		SessionsByIdem: make(map[string]uuid.UUID),
		ByRecovery:     make(map[string]uuid.UUID),
		Events:         make(map[uuid.UUID]domain.SessionEvent),
		Outbox:         make(map[uuid.UUID]domain.OutboxMessage),
		Carts:          make(map[uuid.UUID]ports.CartView),
	}
}

func tenantKey(tenantID uuid.UUID, parts ...string) string {
	b := tenantID.String()
	for _, p := range parts {
		b += ":" + p
	}
	return b
}

// Clock is a fixed/mutable clock for tests.
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

// EventPublisher records published events.
type EventPublisher struct{ S *Store }

func (p *EventPublisher) Publish(_ context.Context, topic, key string, payload any) error {
	p.S.mu.Lock()
	defer p.S.mu.Unlock()
	p.S.Published = append(p.S.Published, publishedEvent{Topic: topic, Key: key, Payload: payload})
	return nil
}

// CompleteLocker provides per-key mutexes for concurrent Complete.
type CompleteLocker struct{ S *Store }

func (l *CompleteLocker) WithLock(_ context.Context, key string, fn func() error) error {
	v, _ := l.S.keyMu.LoadOrStore(key, &sync.Mutex{})
	m := v.(*sync.Mutex)
	m.Lock()
	defer m.Unlock()
	return fn()
}

var _ app.CompleteLocker = (*CompleteLocker)(nil)
