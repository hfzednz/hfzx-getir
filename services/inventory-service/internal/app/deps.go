package app

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/inventory-service/internal/app/ports"
	"github.com/nexora/inventory-service/internal/domain"
)

// Deps aggregates application ports for inventory use cases.
type Deps struct {
	Warehouses   ports.WarehouseRepository
	Locations    ports.LocationRepository
	Balances     ports.BalanceRepository
	Lots         ports.LotRepository
	Reservations ports.ReservationRepository
	Movements    ports.MovementRepository
	Transfers    ports.TransferRepository
	Counts       ports.CountRepository
	Returns      ports.ReturnRepository
	Forecasts    ports.ForecastRepository
	Search       ports.SearchIndexer
	Events       ports.EventPublisher
	AI           ports.ForecastAIClient
	Idempotency  ports.IdempotencyStore
	Locker       ports.StockLocker
	Clock        ports.Clock
	IDs          ports.IDGen

	// SoftReserveTTL is the default soft hold duration.
	SoftReserveTTL time.Duration
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

func (d *Deps) softTTL() time.Duration {
	if d.SoftReserveTTL > 0 {
		return d.SoftReserveTTL
	}
	return 15 * time.Minute
}

// SystemClock is a real-time Clock.
type SystemClock struct{}

func (SystemClock) Now() time.Time { return time.Now().UTC() }

// UUIDGen generates random UUIDs.
type UUIDGen struct{}

func (UUIDGen) New() uuid.UUID { return uuid.New() }

func (d *Deps) publishEvent(ctx context.Context, eventType string, tenantID, warehouseID, variantID uuid.UUID, payload map[string]any) {
	if d.Events == nil {
		return
	}
	ev := domain.NewDomainEvent(eventType, tenantID, warehouseID, variantID, payload)
	topic := domain.TopicForEvent(eventType)
	_ = d.Events.Publish(ctx, topic, warehouseID.String()+":"+variantID.String(), ev)
}

func (d *Deps) indexBalance(ctx context.Context, b domain.StockBalance) {
	if d.Search == nil {
		return
	}
	doc := ports.SearchDocument{
		BalanceID:   b.ID,
		TenantID:    b.TenantID,
		WarehouseID: b.WarehouseID,
		VariantID:   b.VariantID,
		SKUCode:     b.SKUCode,
		LocationID:  b.LocationID,
		OnHand:      b.OnHand,
		Reserved:    b.Reserved,
		Blocked:     b.Blocked,
		Available:   b.Available(),
	}
	if d.Lots != nil {
		if lots, err := d.Lots.ListByBalance(ctx, b.ID); err == nil {
			for _, l := range lots {
				doc.LotCodes = append(doc.LotCodes, l.LotCode)
			}
		}
	}
	_ = d.Search.IndexStock(ctx, doc)
	d.publishEvent(ctx, domain.EventReindexStock, b.TenantID, b.WarehouseID, b.VariantID, map[string]any{
		"balanceId": b.ID,
	})
}

func stockKey(warehouseID, variantID uuid.UUID, locationID *uuid.UUID) string {
	loc := "nil"
	if locationID != nil {
		loc = locationID.String()
	}
	return warehouseID.String() + ":" + variantID.String() + ":" + loc
}

func (d *Deps) withStockLock(ctx context.Context, key string, fn func() error) error {
	if d.Locker != nil {
		return d.Locker.WithLock(ctx, key, fn)
	}
	return fn()
}

func (d *Deps) idemGet(ctx context.Context, key string) (any, bool) {
	if d.Idempotency == nil || key == "" {
		return nil, false
	}
	v, ok, err := d.Idempotency.Get(ctx, key)
	if err != nil || !ok {
		return nil, false
	}
	return v, true
}

func (d *Deps) idemPut(ctx context.Context, key string, value any) {
	if d.Idempotency == nil || key == "" {
		return
	}
	_ = d.Idempotency.Put(ctx, key, value)
}
