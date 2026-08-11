// Package memory provides in-memory port implementations for tests and dev mode.
package memory

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/inventory-service/internal/app/ports"
	"github.com/nexora/inventory-service/internal/domain"
)

// Store is a shared in-memory database for all repositories.
type Store struct {
	mu sync.RWMutex

	Warehouses   map[uuid.UUID]domain.Warehouse
	WHByCode     map[string]uuid.UUID // tenant:code
	Locations    map[uuid.UUID]domain.Location
	Balances     map[uuid.UUID]domain.StockBalance
	BalanceKey   map[string]uuid.UUID
	Lots         map[uuid.UUID]domain.Lot
	Reservations map[uuid.UUID]domain.Reservation
	Movements    map[uuid.UUID]domain.Movement
	MovByIdem    map[string]uuid.UUID
	Transfers    map[uuid.UUID]domain.Transfer
	Counts       map[uuid.UUID]domain.CountSession
	Returns      map[uuid.UUID]domain.InventoryReturn
	Forecasts    map[uuid.UUID]domain.StockForecast
	ForecastKey  map[string]uuid.UUID

	SearchDocs map[uuid.UUID]ports.SearchDocument
	Events     []publishedEvent
	Idem       map[string]any

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
		Warehouses:   make(map[uuid.UUID]domain.Warehouse),
		WHByCode:     make(map[string]uuid.UUID),
		Locations:    make(map[uuid.UUID]domain.Location),
		Balances:     make(map[uuid.UUID]domain.StockBalance),
		BalanceKey:   make(map[string]uuid.UUID),
		Lots:         make(map[uuid.UUID]domain.Lot),
		Reservations: make(map[uuid.UUID]domain.Reservation),
		Movements:    make(map[uuid.UUID]domain.Movement),
		MovByIdem:    make(map[string]uuid.UUID),
		Transfers:    make(map[uuid.UUID]domain.Transfer),
		Counts:       make(map[uuid.UUID]domain.CountSession),
		Returns:      make(map[uuid.UUID]domain.InventoryReturn),
		Forecasts:    make(map[uuid.UUID]domain.StockForecast),
		ForecastKey:  make(map[string]uuid.UUID),
		SearchDocs:   make(map[uuid.UUID]ports.SearchDocument),
		Idem:         make(map[string]any),
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

func balanceKeyStr(k ports.BalanceKey) string {
	loc := "nil"
	if k.LocationID != nil {
		loc = k.LocationID.String()
	}
	return tenantKey(k.TenantID, k.WarehouseID.String(), k.VariantID.String(), loc)
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
	p.S.Events = append(p.S.Events, publishedEvent{Topic: topic, Key: key, Payload: payload})
	return nil
}

// IdempotencyStore is an in-memory idempotency cache.
type IdempotencyStore struct{ S *Store }

func (i *IdempotencyStore) Get(_ context.Context, key string) (any, bool, error) {
	i.S.mu.RLock()
	defer i.S.mu.RUnlock()
	v, ok := i.S.Idem[key]
	return v, ok, nil
}

func (i *IdempotencyStore) Put(_ context.Context, key string, value any) error {
	i.S.mu.Lock()
	defer i.S.mu.Unlock()
	i.S.Idem[key] = value
	return nil
}

// StockLocker provides per-key mutexes for concurrent-safe stock mutations.
type StockLocker struct{ S *Store }

func (l *StockLocker) WithLock(_ context.Context, key string, fn func() error) error {
	v, _ := l.S.keyMu.LoadOrStore(key, &sync.Mutex{})
	m := v.(*sync.Mutex)
	m.Lock()
	defer m.Unlock()
	return fn()
}

// SearchIndexer is an in-memory search index.
type SearchIndexer struct{ S *Store }

func (i *SearchIndexer) IndexStock(_ context.Context, doc ports.SearchDocument) error {
	i.S.mu.Lock()
	defer i.S.mu.Unlock()
	i.S.SearchDocs[doc.BalanceID] = doc
	return nil
}

func (i *SearchIndexer) DeleteStock(_ context.Context, _, balanceID uuid.UUID) error {
	i.S.mu.Lock()
	defer i.S.mu.Unlock()
	delete(i.S.SearchDocs, balanceID)
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
		if q.WarehouseID != nil && doc.WarehouseID != *q.WarehouseID {
			continue
		}
		if q.VariantID != nil && doc.VariantID != *q.VariantID {
			continue
		}
		if q.SKUCode != "" && !strings.EqualFold(doc.SKUCode, q.SKUCode) {
			continue
		}
		if query != "" &&
			!strings.Contains(strings.ToLower(doc.SKUCode), query) &&
			!strings.Contains(strings.ToLower(doc.VariantID.String()), query) {
			matchedLot := false
			for _, lc := range doc.LotCodes {
				if strings.Contains(strings.ToLower(lc), query) {
					matchedLot = true
					break
				}
			}
			if !matchedLot {
				continue
			}
		}
		hits = append(hits, doc)
	}
	sort.Slice(hits, func(a, b int) bool { return hits[a].BalanceID.String() < hits[b].BalanceID.String() })
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

func (i *SearchIndexer) ReindexAll(_ context.Context, _ uuid.UUID) error { return nil }

// ForecastAIClient stub returns a simple demand projection.
type ForecastAIClient struct{}

func (ForecastAIClient) Predict(_ context.Context, tenantID, warehouseID, variantID uuid.UUID, horizonDays int) (domain.StockForecast, error) {
	now := time.Now().UTC()
	conf := 0.72
	return domain.StockForecast{
		ID: uuid.New(), TenantID: tenantID, WarehouseID: warehouseID, VariantID: variantID,
		HorizonStart: now, HorizonEnd: now.AddDate(0, 0, horizonDays),
		PredictedDemand: float64(horizonDays) * 10,
		Confidence:      &conf,
		ModelID:         "stub-ai", ModelVersion: "v0",
		Metadata: map[string]any{}, CreatedAt: now, UpdatedAt: now,
	}, nil
}

// NewDeps wires memory repositories into app.Deps fields via returned structs.
func NewRepos(s *Store) (
	*WarehouseRepo, *LocationRepo, *BalanceRepo, *LotRepo, *ReservationRepo,
	*MovementRepo, *TransferRepo, *CountRepo, *ReturnRepo, *ForecastRepo,
) {
	return &WarehouseRepo{S: s}, &LocationRepo{S: s}, &BalanceRepo{S: s}, &LotRepo{S: s},
		&ReservationRepo{S: s}, &MovementRepo{S: s}, &TransferRepo{S: s}, &CountRepo{S: s},
		&ReturnRepo{S: s}, &ForecastRepo{S: s}
}
