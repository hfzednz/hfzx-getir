// Package memory provides in-memory port implementations for tests and dev mode.
package memory

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/cart-service/internal/domain"
)

// Store is a shared in-memory database for all repositories.
type Store struct {
	mu sync.RWMutex

	Carts            map[uuid.UUID]domain.Cart
	CartByGuest      map[string]uuid.UUID // tenant:guest
	CartByPrincipal  map[string]uuid.UUID // tenant:principal
	Events           map[uuid.UUID]domain.CartEvent
	Outbox           map[uuid.UUID]domain.OutboxMessage
	Saved            map[uuid.UUID]domain.SavedCart
	Published        []publishedEvent
}

type publishedEvent struct {
	Topic   string
	Key     string
	Payload any
}

// NewStore returns an empty in-memory store.
func NewStore() *Store {
	return &Store{
		Carts:           make(map[uuid.UUID]domain.Cart),
		CartByGuest:     make(map[string]uuid.UUID),
		CartByPrincipal: make(map[string]uuid.UUID),
		Events:          make(map[uuid.UUID]domain.CartEvent),
		Outbox:          make(map[uuid.UUID]domain.OutboxMessage),
		Saved:           make(map[uuid.UUID]domain.SavedCart),
	}
}

func tenantKey(tenantID uuid.UUID, parts ...string) string {
	b := strings.Builder{}
	b.WriteString(tenantID.String())
	for _, p := range parts {
		b.WriteByte(':')
		b.WriteString(p)
	}
	return b.String()
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
