// Package ports defines application-layer dependency interfaces (hexagonal ports).
package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/warehouse-service/internal/domain"
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

// --- Inventory client (no ledger in this service) ---

// SoftReserveLine is a soft-hold request line for inventory-service.
type SoftReserveLine struct {
	WarehouseID uuid.UUID
	VariantID   uuid.UUID
	SKUCode     string
	Qty         int64
}

// SoftReserveRequest creates a soft reservation via inventory-service.
type SoftReserveRequest struct {
	TenantID       uuid.UUID
	WarehouseID    uuid.UUID
	ExternalRef    string
	IdempotencyKey string
	Lines          []SoftReserveLine
}

// SoftReserveResult is the inventory SoftReserve response.
type SoftReserveResult struct {
	ReservationID uuid.UUID
	Status        string
}

// ConfirmHardRequest confirms a soft reservation to hard.
type ConfirmHardRequest struct {
	TenantID       uuid.UUID
	ReservationID  uuid.UUID
	IdempotencyKey string
}

// ReleaseRequest releases a reservation.
type ReleaseRequest struct {
	TenantID       uuid.UUID
	ReservationID  uuid.UUID
	IdempotencyKey string
}

// ConsumeRequest deducts reserved stock on ship/handoff.
type ConsumeRequest struct {
	TenantID       uuid.UUID
	ReservationID  uuid.UUID
	IdempotencyKey string
}

// InventoryClient is the outbound port to inventory-service.
// SoftReserve / ConfirmHard / Release / Consume — stubs only; no ledger here.
type InventoryClient interface {
	SoftReserve(ctx context.Context, req SoftReserveRequest) (SoftReserveResult, error)
	ConfirmHard(ctx context.Context, req ConfirmHardRequest) error
	Release(ctx context.Context, req ReleaseRequest) error
	Consume(ctx context.Context, req ConsumeRequest) error
}

// RouteLine is an unordered pick line for AI route optimization.
type RouteLine struct {
	LineID       uuid.UUID
	LocationCode string
	SKUCode      string
	Sequence     int
}

// RouteOptimizer is an AI port that sorts pick lines by location.
type RouteOptimizer interface {
	OptimizePickRoute(ctx context.Context, warehouseID uuid.UUID, lines []RouteLine) ([]RouteLine, error)
}

// --- Repositories ---

// FulfillmentFilter lists fulfillment projections.
type FulfillmentFilter struct {
	TenantID    uuid.UUID
	WarehouseID *uuid.UUID
	Status      *domain.FulfillmentStatus
	Limit       int
	Offset      int
}

// FulfillmentRepo persists fulfillment order projections.
type FulfillmentRepo interface {
	Create(ctx context.Context, o domain.FulfillmentOrder) error
	Update(ctx context.Context, o domain.FulfillmentOrder) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (domain.FulfillmentOrder, error)
	GetByExternalOrderID(ctx context.Context, tenantID uuid.UUID, externalOrderID string) (domain.FulfillmentOrder, error)
	List(ctx context.Context, f FulfillmentFilter) ([]domain.FulfillmentOrder, int, error)
}

// TaskFilter lists tasks in a warehouse queue.
type TaskFilter struct {
	TenantID    uuid.UUID
	WarehouseID uuid.UUID
	Type        *domain.TaskType
	Status      *domain.TaskStatus
	AssigneeID  *uuid.UUID
	Limit       int
	Offset      int
}

// TaskRepo persists pick/pack/dispatch tasks.
type TaskRepo interface {
	Create(ctx context.Context, t domain.Task) error
	Update(ctx context.Context, t domain.Task) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (domain.Task, error)
	List(ctx context.Context, f TaskFilter) ([]domain.Task, int, error)
	CountByStatus(ctx context.Context, tenantID, warehouseID uuid.UUID) (map[domain.TaskStatus]int, error)
}

// PickRepo persists pick sessions and lines.
type PickRepo interface {
	CreateSession(ctx context.Context, s domain.PickSession) error
	UpdateSession(ctx context.Context, s domain.PickSession) error
	GetSessionByID(ctx context.Context, tenantID, id uuid.UUID) (domain.PickSession, error)
	GetSessionByTaskID(ctx context.Context, tenantID, taskID uuid.UUID) (domain.PickSession, error)
}

// PackRepo persists pack sessions.
type PackRepo interface {
	Create(ctx context.Context, s domain.PackSession) error
	Update(ctx context.Context, s domain.PackSession) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (domain.PackSession, error)
	GetByTaskID(ctx context.Context, tenantID, taskID uuid.UUID) (domain.PackSession, error)
}

// DispatchRepo persists dispatch units.
type DispatchRepo interface {
	Create(ctx context.Context, u domain.DispatchUnit) error
	Update(ctx context.Context, u domain.DispatchUnit) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (domain.DispatchUnit, error)
	GetByFulfillmentID(ctx context.Context, tenantID, fulfillmentID uuid.UUID) (domain.DispatchUnit, error)
	ListQueued(ctx context.Context, tenantID, warehouseID uuid.UUID, limit, offset int) ([]domain.DispatchUnit, int, error)
}

// StationRepo persists stations.
type StationRepo interface {
	Create(ctx context.Context, s domain.Station) error
	Update(ctx context.Context, s domain.Station) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (domain.Station, error)
	ListByWarehouse(ctx context.Context, tenantID, warehouseID uuid.UUID) ([]domain.Station, error)
}

// WorkforceRepo persists employees and shifts.
type WorkforceRepo interface {
	CreateEmployee(ctx context.Context, e domain.Employee) error
	GetEmployee(ctx context.Context, tenantID, id uuid.UUID) (domain.Employee, error)
	ListEmployees(ctx context.Context, tenantID, warehouseID uuid.UUID) ([]domain.Employee, error)
	CreateShift(ctx context.Context, s domain.Shift) error
	UpdateShift(ctx context.Context, s domain.Shift) error
	GetActiveShift(ctx context.Context, tenantID, employeeID uuid.UUID) (domain.Shift, error)
	GetShift(ctx context.Context, tenantID, id uuid.UUID) (domain.Shift, error)
}

// EquipmentRepo persists equipment registry.
type EquipmentRepo interface {
	Create(ctx context.Context, e domain.Equipment) error
	Update(ctx context.Context, e domain.Equipment) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (domain.Equipment, error)
	ListByWarehouse(ctx context.Context, tenantID, warehouseID uuid.UUID) ([]domain.Equipment, error)
}

// QCRepo persists QC inspections.
type QCRepo interface {
	Create(ctx context.Context, i domain.QCInspection) error
	Update(ctx context.Context, i domain.QCInspection) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (domain.QCInspection, error)
	ListByFulfillment(ctx context.Context, tenantID, fulfillmentID uuid.UUID) ([]domain.QCInspection, error)
}

// LabelRepo persists shipping labels.
type LabelRepo interface {
	Create(ctx context.Context, l domain.Label) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (domain.Label, error)
}
