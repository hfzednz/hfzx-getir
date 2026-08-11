package app

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/warehouse-service/internal/app/ports"
	"github.com/nexora/warehouse-service/internal/domain"
)

// Deps aggregates application ports for warehouse use cases.
type Deps struct {
	Fulfillments ports.FulfillmentRepo
	Tasks        ports.TaskRepo
	Picks        ports.PickRepo
	Packs        ports.PackRepo
	Dispatches   ports.DispatchRepo
	Stations     ports.StationRepo
	Workforce    ports.WorkforceRepo
	Equipment    ports.EquipmentRepo
	QC           ports.QCRepo
	Labels       ports.LabelRepo
	Inventory    ports.InventoryClient
	RouteAI      ports.RouteOptimizer
	Events       ports.EventPublisher
	Clock        ports.Clock
	IDs          ports.IDGen

	// WeightToleranceG is default pack weight tolerance in grams.
	WeightToleranceG int64
}

func (d *Deps) now() time.Time {
	if d.Clock != nil {
		return d.Clock.Now().UTC()
	}
	return time.Now().UTC()
}

func (d *Deps) newID() uuid.UUID {
	if d.IDs != nil {
		return d.IDs.New()
	}
	return uuid.New()
}

func (d *Deps) weightTol() int64 {
	if d.WeightToleranceG > 0 {
		return d.WeightToleranceG
	}
	return 50
}

// SystemClock is a real-time Clock.
type SystemClock struct{}

func (SystemClock) Now() time.Time { return time.Now().UTC() }

// UUIDGen generates random UUIDs.
type UUIDGen struct{}

func (UUIDGen) New() uuid.UUID { return uuid.New() }

func (d *Deps) publishEvent(ctx context.Context, eventType string, tenantID, warehouseID, fulfillmentID uuid.UUID, payload map[string]any) {
	if d.Events == nil {
		return
	}
	ev := domain.NewDomainEvent(eventType, tenantID, warehouseID, fulfillmentID, payload)
	ev.OccurredAt = d.now()
	topic := domain.TopicForEvent(eventType)
	_ = d.Events.Publish(ctx, topic, fulfillmentID.String(), ev)
}

func appendTaskHistory(t *domain.Task, at time.Time, action string, actor *uuid.UUID, from, to domain.TaskStatus, note string) {
	t.History = append(t.History, domain.TaskHistoryEntry{
		At: at, Action: action, ActorID: actor, FromStatus: from, ToStatus: to, Note: note,
	})
}
