package memory

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/wallet-service/internal/domain"
)

// Store is the in-memory aggregate for wallet-service.
type Store struct {
	mu         sync.RWMutex
	wallets    map[uuid.UUID]domain.Wallet
	byPrin     map[string]uuid.UUID
	accounts   map[uuid.UUID]domain.Account
	acctByType map[string]uuid.UUID // tenant|wallet|type
	entries    []domain.Entry
	entryIdem  map[string]uuid.UUID
	holds      map[uuid.UUID]domain.Hold
	holdIdem   map[string]uuid.UUID
	transfers  map[uuid.UUID]domain.Transfer
	xferIdem   map[string]uuid.UUID
	limits     map[string]domain.Limit
	audits     []domain.AuditEntry
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
		wallets:    make(map[uuid.UUID]domain.Wallet),
		byPrin:     make(map[string]uuid.UUID),
		accounts:   make(map[uuid.UUID]domain.Account),
		acctByType: make(map[string]uuid.UUID),
		entryIdem:  make(map[string]uuid.UUID),
		holds:      make(map[uuid.UUID]domain.Hold),
		holdIdem:   make(map[string]uuid.UUID),
		transfers:  make(map[uuid.UUID]domain.Transfer),
		xferIdem:   make(map[string]uuid.UUID),
		limits:     make(map[string]domain.Limit),
	}
}

func prinKey(tenant, principal uuid.UUID) string {
	return tenant.String() + "|" + principal.String()
}

func acctKey(tenant, wallet uuid.UUID, t domain.AccountType) string {
	return tenant.String() + "|" + wallet.String() + "|" + string(t)
}

func idemKey(tenant uuid.UUID, key string) string {
	return tenant.String() + "|" + key
}

// Clock is a fixed clock for tests.
type Clock struct{ T time.Time }

func (c *Clock) Now() time.Time { return c.T }

// IDGen returns random UUIDs.
type IDGen struct{}

func (IDGen) New() uuid.UUID { return uuid.New() }
