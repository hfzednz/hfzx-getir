// Package ports defines application-layer dependency interfaces (hexagonal ports).
package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/order-service/internal/domain"
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

// OrderFilter lists orders.
type OrderFilter struct {
	TenantID   uuid.UUID
	CustomerID *uuid.UUID
	Status     *domain.OrderStatus
	Query      string
	Limit      int
	Offset     int
}

// OrderRepository persists the order aggregate.
type OrderRepository interface {
	Create(ctx context.Context, o domain.Order) error
	Update(ctx context.Context, o domain.Order) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (domain.Order, error)
	GetByIdempotencyKey(ctx context.Context, tenantID uuid.UUID, key string) (domain.Order, error)
	List(ctx context.Context, f OrderFilter) ([]domain.Order, int, error)
}

// EventStore persists append-only order timeline events.
type EventStore interface {
	Append(ctx context.Context, e domain.OrderEvent) error
	ListByOrder(ctx context.Context, tenantID, orderID uuid.UUID) ([]domain.OrderEvent, error)
}

// SagaRepository persists saga instances and steps.
type SagaRepository interface {
	Create(ctx context.Context, s domain.SagaInstance) error
	Update(ctx context.Context, s domain.SagaInstance) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (domain.SagaInstance, error)
	GetByIdempotencyKey(ctx context.Context, tenantID uuid.UUID, key string) (domain.SagaInstance, error)
	ListByOrder(ctx context.Context, tenantID, orderID uuid.UUID) ([]domain.SagaInstance, error)
}

// OutboxRepository persists transactional outbox rows.
type OutboxRepository interface {
	Enqueue(ctx context.Context, m domain.OutboxMessage) error
	Update(ctx context.Context, m domain.OutboxMessage) error
	ListPending(ctx context.Context, limit int) ([]domain.OutboxMessage, error)
}

// FulfillmentRepository persists split fulfillment units.
type FulfillmentRepository interface {
	Create(ctx context.Context, f domain.Fulfillment) error
	Update(ctx context.Context, f domain.Fulfillment) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (domain.Fulfillment, error)
	ListByOrder(ctx context.Context, tenantID, orderID uuid.UUID) ([]domain.Fulfillment, error)
}

// ReturnRepository persists return requests.
type ReturnRepository interface {
	Create(ctx context.Context, r domain.Return) error
	Update(ctx context.Context, r domain.Return) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (domain.Return, error)
	ListByOrder(ctx context.Context, tenantID, orderID uuid.UUID) ([]domain.Return, error)
}

// RefundRepository persists refund requests.
type RefundRepository interface {
	Create(ctx context.Context, r domain.Refund) error
	Update(ctx context.Context, r domain.Refund) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (domain.Refund, error)
	ListByOrder(ctx context.Context, tenantID, orderID uuid.UUID) ([]domain.Refund, error)
}

// SearchDocument is the OpenSearch projection for an order.
type SearchDocument struct {
	OrderID             uuid.UUID
	TenantID            uuid.UUID
	CustomerPrincipalID uuid.UUID
	Status              string
	Type                string
	Currency            string
	TotalMinor          int64
	Priority            int
	IdempotencyKey      string
	SKUCodes            []string
	WarehouseIDs        []uuid.UUID
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// SearchQuery filters order search.
type SearchQuery struct {
	TenantID   uuid.UUID
	Query      string
	Status     *domain.OrderStatus
	CustomerID *uuid.UUID
	Limit      int
	Offset     int
}

// SearchResult is a page of search hits.
type SearchResult struct {
	Total int
	Hits  []SearchDocument
}

// SearchIndexer indexes and queries order documents.
type SearchIndexer interface {
	IndexOrder(ctx context.Context, doc SearchDocument) error
	DeleteOrder(ctx context.Context, tenantID, orderID uuid.UUID) error
	Search(ctx context.Context, q SearchQuery) (SearchResult, error)
}

// ReserveLine is a line for inventory soft/hard reservation.
type ReserveLine struct {
	VariantID   uuid.UUID
	SKUCode     string
	Qty         int
	WarehouseID uuid.UUID
}

// SoftReserveRequest asks inventory-service for a soft hold.
type SoftReserveRequest struct {
	TenantID       uuid.UUID
	OrderID        uuid.UUID
	IdempotencyKey string
	Lines          []ReserveLine
}

// SoftReserveResult is the opaque reservation ref from inventory.
type SoftReserveResult struct {
	ReservationRef string
}

// ConfirmHardRequest converts a soft hold to a hard reservation.
type ConfirmHardRequest struct {
	TenantID       uuid.UUID
	ReservationRef string
	IdempotencyKey string
}

// ReleaseRequest releases a soft or hard reservation.
type ReleaseRequest struct {
	TenantID       uuid.UUID
	ReservationRef string
	IdempotencyKey string
}

// InventoryClient talks to inventory-service (opaque refs only).
type InventoryClient interface {
	SoftReserve(ctx context.Context, req SoftReserveRequest) (SoftReserveResult, error)
	ConfirmHard(ctx context.Context, req ConfirmHardRequest) error
	Release(ctx context.Context, req ReleaseRequest) error
}

// AuthorizeRequest asks payment-service to authorize funds.
type AuthorizeRequest struct {
	TenantID       uuid.UUID
	OrderID        uuid.UUID
	AmountMinor    int64
	Currency       string
	IdempotencyKey string
}

// AuthorizeResult is the opaque payment intent ref.
type AuthorizeResult struct {
	PaymentIntentRef string
}

// VoidRequest voids an authorization.
type VoidRequest struct {
	TenantID         uuid.UUID
	PaymentIntentRef string
	IdempotencyKey   string
}

// RefundPaymentRequest asks payment-service to refund.
type RefundPaymentRequest struct {
	TenantID         uuid.UUID
	PaymentIntentRef string
	AmountMinor      int64
	Currency         string
	IdempotencyKey   string
}

// RefundPaymentResult is the opaque payment refund ref.
type RefundPaymentResult struct {
	PaymentRefundRef string
}

// PaymentClient talks to payment-service (opaque refs only).
type PaymentClient interface {
	Authorize(ctx context.Context, req AuthorizeRequest) (AuthorizeResult, error)
	Void(ctx context.Context, req VoidRequest) error
	Refund(ctx context.Context, req RefundPaymentRequest) (RefundPaymentResult, error)
}

// ReceiveFulfillmentRequest hands work to warehouse-service.
type ReceiveFulfillmentRequest struct {
	TenantID       uuid.UUID
	OrderID        uuid.UUID
	WarehouseID    uuid.UUID
	Priority       int
	IdempotencyKey string
	Lines          []ReserveLine
}

// ReceiveFulfillmentResult is the opaque warehouse fulfillment ref.
type ReceiveFulfillmentResult struct {
	FulfillmentRef string
}

// WarehouseClient talks to warehouse-service.
type WarehouseClient interface {
	ReceiveFulfillment(ctx context.Context, req ReceiveFulfillmentRequest) (ReceiveFulfillmentResult, error)
}

// RequestDispatchRequest asks dispatch-service to assign a courier.
type RequestDispatchRequest struct {
	TenantID       uuid.UUID
	OrderID        uuid.UUID
	FulfillmentRef string
	IdempotencyKey string
}

// RequestDispatchResult is the opaque dispatch assignment ref.
type RequestDispatchResult struct {
	DispatchRef string
}

// DispatchClient talks to dispatch-service.
type DispatchClient interface {
	RequestDispatch(ctx context.Context, req RequestDispatchRequest) (RequestDispatchResult, error)
}
