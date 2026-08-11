// Package ports defines application-layer dependency interfaces (hexagonal ports).
package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/inventory-service/internal/domain"
)

// Clock abstracts time for deterministic tests.
type Clock interface {
	Now() time.Time
}

// IDGen abstracts UUID generation.
type IDGen interface {
	New() uuid.UUID
}

// EventPublisher publishes domain events (Kafka adapters).
type EventPublisher interface {
	Publish(ctx context.Context, topic string, key string, payload any) error
}

// SearchDocument is the OpenSearch projection for a stock balance.
type SearchDocument struct {
	BalanceID   uuid.UUID
	TenantID    uuid.UUID
	WarehouseID uuid.UUID
	VariantID   uuid.UUID
	SKUCode     string
	LocationID  *uuid.UUID
	OnHand      int64
	Reserved    int64
	Blocked     int64
	Available   int64
	LotCodes    []string
}

// SearchQuery filters stock search.
type SearchQuery struct {
	TenantID    uuid.UUID
	Query       string
	WarehouseID *uuid.UUID
	VariantID   *uuid.UUID
	SKUCode     string
	Limit       int
	Offset      int
}

// SearchResult is a page of search hits.
type SearchResult struct {
	Total int
	Hits  []SearchDocument
}

// SearchIndexer indexes and queries stock documents (OpenSearch adapter).
type SearchIndexer interface {
	IndexStock(ctx context.Context, doc SearchDocument) error
	DeleteStock(ctx context.Context, tenantID, balanceID uuid.UUID) error
	Search(ctx context.Context, q SearchQuery) (SearchResult, error)
	ReindexAll(ctx context.Context, tenantID uuid.UUID) error
}

// ForecastAIClient is a stub AI demand-forecast port.
type ForecastAIClient interface {
	Predict(ctx context.Context, tenantID, warehouseID, variantID uuid.UUID, horizonDays int) (domain.StockForecast, error)
}

// WarehouseFilter lists warehouses.
type WarehouseFilter struct {
	TenantID uuid.UUID
	Status   *domain.WarehouseStatus
	Query    string
	Limit    int
	Offset   int
}

// WarehouseRepository persists warehouses.
type WarehouseRepository interface {
	Create(ctx context.Context, w domain.Warehouse) error
	Update(ctx context.Context, w domain.Warehouse) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (domain.Warehouse, error)
	GetByCode(ctx context.Context, tenantID uuid.UUID, code string) (domain.Warehouse, error)
	List(ctx context.Context, f WarehouseFilter) ([]domain.Warehouse, int, error)
	Delete(ctx context.Context, tenantID, id uuid.UUID, at time.Time) error
}

// LocationRepository persists location tree nodes.
type LocationRepository interface {
	Create(ctx context.Context, l domain.Location) error
	Update(ctx context.Context, l domain.Location) error
	GetByID(ctx context.Context, id uuid.UUID) (domain.Location, error)
	ListByWarehouse(ctx context.Context, warehouseID uuid.UUID) ([]domain.Location, error)
	ListChildren(ctx context.Context, parentID uuid.UUID) ([]domain.Location, error)
	Delete(ctx context.Context, id uuid.UUID, at time.Time) error
}

// BalanceKey uniquely identifies a stock balance row.
type BalanceKey struct {
	TenantID    uuid.UUID
	WarehouseID uuid.UUID
	VariantID   uuid.UUID
	LocationID  *uuid.UUID
}

// BalanceRepository persists stock balances with optimistic versioning.
type BalanceRepository interface {
	Create(ctx context.Context, b domain.StockBalance) error
	Update(ctx context.Context, b domain.StockBalance) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (domain.StockBalance, error)
	GetByKey(ctx context.Context, key BalanceKey) (domain.StockBalance, error)
	ListByWarehouse(ctx context.Context, tenantID, warehouseID uuid.UUID, limit, offset int) ([]domain.StockBalance, int, error)
	ListByVariant(ctx context.Context, tenantID, variantID uuid.UUID) ([]domain.StockBalance, error)
}

// LotRepository persists lots/batches.
type LotRepository interface {
	Create(ctx context.Context, l domain.Lot) error
	Update(ctx context.Context, l domain.Lot) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (domain.Lot, error)
	ListByBalance(ctx context.Context, balanceID uuid.UUID) ([]domain.Lot, error)
	ListByWarehouseVariant(ctx context.Context, tenantID, warehouseID, variantID uuid.UUID) ([]domain.Lot, error)
	ListNearExpiry(ctx context.Context, tenantID uuid.UUID, warehouseID *uuid.UUID, withinDays int, asOf time.Time) ([]domain.Lot, error)
}

// ReservationRepository persists reservation headers and lines.
type ReservationRepository interface {
	Create(ctx context.Context, r domain.Reservation) error
	Update(ctx context.Context, r domain.Reservation) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (domain.Reservation, error)
	ListByExternalRef(ctx context.Context, tenantID uuid.UUID, externalRef string) ([]domain.Reservation, error)
}

// MovementRepository persists append-only ledger rows.
type MovementRepository interface {
	Create(ctx context.Context, m domain.Movement) error
	GetByIdempotencyKey(ctx context.Context, key string) (domain.Movement, error)
	ListByBalance(ctx context.Context, balanceID uuid.UUID, limit, offset int) ([]domain.Movement, int, error)
}

// TransferRepository persists transfers.
type TransferRepository interface {
	Create(ctx context.Context, t domain.Transfer) error
	Update(ctx context.Context, t domain.Transfer) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (domain.Transfer, error)
	List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]domain.Transfer, int, error)
}

// CountRepository persists count sessions.
type CountRepository interface {
	Create(ctx context.Context, s domain.CountSession) error
	Update(ctx context.Context, s domain.CountSession) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (domain.CountSession, error)
	List(ctx context.Context, tenantID, warehouseID uuid.UUID, limit, offset int) ([]domain.CountSession, int, error)
}

// ReturnRepository persists inventory returns.
type ReturnRepository interface {
	Create(ctx context.Context, r domain.InventoryReturn) error
	Update(ctx context.Context, r domain.InventoryReturn) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (domain.InventoryReturn, error)
}

// ForecastRepository persists AI forecast projections.
type ForecastRepository interface {
	Upsert(ctx context.Context, f domain.StockForecast) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (domain.StockForecast, error)
	List(ctx context.Context, tenantID, warehouseID, variantID uuid.UUID) ([]domain.StockForecast, error)
}

// IdempotencyStore remembers command results by key.
type IdempotencyStore interface {
	Get(ctx context.Context, key string) (any, bool, error)
	Put(ctx context.Context, key string, value any) error
}

// StockLocker provides per stock-key mutual exclusion for concurrent-safe mutations.
type StockLocker interface {
	WithLock(ctx context.Context, key string, fn func() error) error
}
