// Package memory provides in-memory port implementations for tests and dev mode.
package memory

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/order-service/internal/app"
	"github.com/nexora/order-service/internal/app/ports"
	"github.com/nexora/order-service/internal/domain"
)

// Store is a shared in-memory database for all repositories.
type Store struct {
	mu sync.RWMutex

	Orders       map[uuid.UUID]domain.Order
	OrdersByIdem map[string]uuid.UUID // tenant:key
	Events       map[uuid.UUID]domain.OrderEvent
	Sagas        map[uuid.UUID]domain.SagaInstance
	SagasByIdem  map[string]uuid.UUID
	Outbox       map[uuid.UUID]domain.OutboxMessage
	Fulfillments map[uuid.UUID]domain.Fulfillment
	Returns      map[uuid.UUID]domain.Return
	Refunds      map[uuid.UUID]domain.Refund
	SearchDocs   map[uuid.UUID]ports.SearchDocument
	Published    []publishedEvent

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
		Orders:       make(map[uuid.UUID]domain.Order),
		OrdersByIdem: make(map[string]uuid.UUID),
		Events:       make(map[uuid.UUID]domain.OrderEvent),
		Sagas:        make(map[uuid.UUID]domain.SagaInstance),
		SagasByIdem:  make(map[string]uuid.UUID),
		Outbox:       make(map[uuid.UUID]domain.OutboxMessage),
		Fulfillments: make(map[uuid.UUID]domain.Fulfillment),
		Returns:      make(map[uuid.UUID]domain.Return),
		Refunds:      make(map[uuid.UUID]domain.Refund),
		SearchDocs:   make(map[uuid.UUID]ports.SearchDocument),
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

// PlaceLocker provides per-key mutexes for concurrent PlaceOrder.
type PlaceLocker struct{ S *Store }

func (l *PlaceLocker) WithLock(_ context.Context, key string, fn func() error) error {
	v, _ := l.S.keyMu.LoadOrStore(key, &sync.Mutex{})
	m := v.(*sync.Mutex)
	m.Lock()
	defer m.Unlock()
	return fn()
}

var _ app.PlaceLocker = (*PlaceLocker)(nil)

// SearchIndexer is an in-memory order search index.
type SearchIndexer struct{ S *Store }

func (i *SearchIndexer) IndexOrder(_ context.Context, doc ports.SearchDocument) error {
	i.S.mu.Lock()
	defer i.S.mu.Unlock()
	i.S.SearchDocs[doc.OrderID] = doc
	return nil
}

func (i *SearchIndexer) DeleteOrder(_ context.Context, _, orderID uuid.UUID) error {
	i.S.mu.Lock()
	defer i.S.mu.Unlock()
	delete(i.S.SearchDocs, orderID)
	return nil
}

func (i *SearchIndexer) Search(_ context.Context, q ports.SearchQuery) (ports.SearchResult, error) {
	i.S.mu.RLock()
	defer i.S.mu.RUnlock()
	query := strings.ToLower(strings.TrimSpace(q.Query))
	hits := make([]ports.SearchDocument, 0)
	for _, doc := range i.S.SearchDocs {
		if doc.TenantID != q.TenantID {
			continue
		}
		if q.Status != nil && doc.Status != string(*q.Status) {
			continue
		}
		if q.CustomerID != nil && doc.CustomerPrincipalID != *q.CustomerID {
			continue
		}
		if query != "" {
			match := strings.Contains(strings.ToLower(doc.OrderID.String()), query) ||
				strings.Contains(strings.ToLower(doc.IdempotencyKey), query) ||
				strings.Contains(strings.ToLower(doc.Status), query)
			if !match {
				for _, sku := range doc.SKUCodes {
					if strings.Contains(strings.ToLower(sku), query) {
						match = true
						break
					}
				}
			}
			if !match {
				continue
			}
		}
		hits = append(hits, doc)
	}
	sort.Slice(hits, func(a, b int) bool { return hits[a].OrderID.String() < hits[b].OrderID.String() })
	total := len(hits)
	if q.Offset >= len(hits) {
		return ports.SearchResult{Total: total}, nil
	}
	end := q.Offset + q.Limit
	if q.Limit <= 0 {
		end = len(hits)
	}
	if end > len(hits) {
		end = len(hits)
	}
	return ports.SearchResult{Total: total, Hits: hits[q.Offset:end]}, nil
}
