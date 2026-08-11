// Package memory provides in-memory port implementations for tests and dev mode.
package memory

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/settlement-service/internal/app/ports"
	"github.com/nexora/settlement-service/internal/domain"
)

// Store is a shared in-memory database.
type Store struct {
	mu sync.RWMutex

	Batches       map[uuid.UUID]domain.SettlementBatch
	BatchesByIdem map[string]uuid.UUID
	Payouts       map[uuid.UUID]domain.PayoutInstruction
	Reconciles    map[uuid.UUID]domain.Reconciliation
	Mismatches    map[uuid.UUID]domain.Mismatch
	Events        map[uuid.UUID]domain.SettlementEvent
	Outbox        map[uuid.UUID]domain.OutboxMessage
	Published     []publishedEvent
}

type publishedEvent struct {
	Topic   string
	Key     string
	Payload any
}

// NewStore returns an empty in-memory store.
func NewStore() *Store {
	return &Store{
		Batches:       make(map[uuid.UUID]domain.SettlementBatch),
		BatchesByIdem: make(map[string]uuid.UUID),
		Payouts:       make(map[uuid.UUID]domain.PayoutInstruction),
		Reconciles:    make(map[uuid.UUID]domain.Reconciliation),
		Mismatches:    make(map[uuid.UUID]domain.Mismatch),
		Events:        make(map[uuid.UUID]domain.SettlementEvent),
		Outbox:        make(map[uuid.UUID]domain.OutboxMessage),
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

// LedgerClient is an in-memory finance-ledger stub.
type LedgerClient struct {
	Fail bool
	Calls int
}

func (c *LedgerClient) PostSettlementJournal(_ context.Context, req ports.LedgerPostRequest) (ports.LedgerPostResult, error) {
	c.Calls++
	if c.Fail {
		return ports.LedgerPostResult{}, domain.ErrInvariant
	}
	return ports.LedgerPostResult{JournalID: "ledger-" + req.IdempotencyKey, Posted: true}, nil
}

// PayoutClient is an in-memory bank/PSP stub.
type PayoutClient struct {
	Fail  bool
	Calls int
}

func (c *PayoutClient) Execute(_ context.Context, req ports.PayoutRequest) (ports.PayoutResult, error) {
	c.Calls++
	if c.Fail {
		return ports.PayoutResult{Succeeded: false, Error: "provider declined"}, nil
	}
	return ports.PayoutResult{ProviderRef: "psp-" + req.InstructionID.String(), Succeeded: true}, nil
}

var (
	_ ports.EventPublisher = (*EventPublisher)(nil)
	_ ports.LedgerClient   = (*LedgerClient)(nil)
	_ ports.PayoutClient   = (*PayoutClient)(nil)
)
