// Package memory provides in-memory port implementations for tests and dev mode.
package memory

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/finance-ledger-service/internal/app/ports"
	"github.com/nexora/finance-ledger-service/internal/domain"
)

// Store is a shared in-memory database for all repositories.
type Store struct {
	mu sync.RWMutex

	Accounts       map[uuid.UUID]domain.Account
	AccountsByCode map[string]uuid.UUID // tenant:code
	Journals       map[uuid.UUID]domain.Journal
	JournalsByIdem map[string]uuid.UUID
	Invoices       map[uuid.UUID]domain.Invoice
	InvoicesByIdem map[string]uuid.UUID
	CreditNotes    map[uuid.UUID]domain.CreditNote
	CreditByIdem   map[string]uuid.UUID
	TaxRules       map[string]domain.TaxRule // tenant:code
	Events         map[uuid.UUID]domain.LedgerEvent
	Outbox         map[uuid.UUID]domain.OutboxMessage
	Published      []publishedEvent
}

type publishedEvent struct {
	Topic   string
	Key     string
	Payload any
}

// NewStore returns an empty in-memory store.
func NewStore() *Store {
	return &Store{
		Accounts:       make(map[uuid.UUID]domain.Account),
		AccountsByCode: make(map[string]uuid.UUID),
		Journals:       make(map[uuid.UUID]domain.Journal),
		JournalsByIdem: make(map[string]uuid.UUID),
		Invoices:       make(map[uuid.UUID]domain.Invoice),
		InvoicesByIdem: make(map[string]uuid.UUID),
		CreditNotes:    make(map[uuid.UUID]domain.CreditNote),
		CreditByIdem:   make(map[string]uuid.UUID),
		TaxRules:       make(map[string]domain.TaxRule),
		Events:         make(map[uuid.UUID]domain.LedgerEvent),
		Outbox:         make(map[uuid.UUID]domain.OutboxMessage),
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

var _ ports.EventPublisher = (*EventPublisher)(nil)
