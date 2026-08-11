package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/warehouse-service/internal/app/ports"
	"github.com/nexora/warehouse-service/internal/domain"
)

// ReceiveLineInput is a line on a receive command.
type ReceiveLineInput struct {
	VariantID    uuid.UUID
	SKUCode      string
	Barcode      string
	LocationCode string
	Qty          int64
}

// ReceiveFulfillmentCmd creates a fulfillment projection and queues a pick task.
type ReceiveFulfillmentCmd struct {
	TenantID        uuid.UUID
	WarehouseID     uuid.UUID
	ExternalOrderID string
	Strategy        domain.PickStrategy
	Priority        int
	Lines           []ReceiveLineInput
	IdempotencyKey  string
	ActorID         *uuid.UUID
}

// ReceiveFulfillment creates order projection + lines, SoftReserves via inventory port, queues pick task.
func (d *Deps) ReceiveFulfillment(ctx context.Context, in ReceiveFulfillmentCmd) (domain.FulfillmentOrder, error) {
	if in.TenantID == uuid.Nil || in.WarehouseID == uuid.Nil || in.ExternalOrderID == "" || len(in.Lines) == 0 {
		return domain.FulfillmentOrder{}, domain.ErrInvalidArgument
	}
	for _, l := range in.Lines {
		if l.VariantID == uuid.Nil || l.Qty <= 0 || l.Barcode == "" {
			return domain.FulfillmentOrder{}, fmt.Errorf("%w: line requires variant, qty>0, barcode", domain.ErrInvalidArgument)
		}
	}

	if existing, err := d.Fulfillments.GetByExternalOrderID(ctx, in.TenantID, in.ExternalOrderID); err == nil {
		return existing, nil
	}

	now := d.now()
	strategy := in.Strategy
	if strategy == "" {
		strategy = domain.PickStrategySingle
	}

	fo := domain.FulfillmentOrder{
		ID:              d.newID(),
		TenantID:        in.TenantID,
		WarehouseID:     in.WarehouseID,
		ExternalOrderID: in.ExternalOrderID,
		Status:          domain.FulfillmentStatusReceived,
		Strategy:        strategy,
		Priority:        in.Priority,
		Metadata:        map[string]any{},
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	invLines := make([]ports.SoftReserveLine, 0, len(in.Lines))
	for i, l := range in.Lines {
		fo.Lines = append(fo.Lines, domain.FulfillmentLine{
			ID:            d.newID(),
			FulfillmentID: fo.ID,
			VariantID:     l.VariantID,
			SKUCode:       l.SKUCode,
			Barcode:       l.Barcode,
			LocationCode:  l.LocationCode,
			QtyOrdered:    l.Qty,
			Sequence:      i + 1,
		})
		invLines = append(invLines, ports.SoftReserveLine{
			WarehouseID: in.WarehouseID,
			VariantID:   l.VariantID,
			SKUCode:     l.SKUCode,
			Qty:         l.Qty,
		})
	}

	idem := in.IdempotencyKey
	if idem == "" {
		idem = "recv:" + in.ExternalOrderID
	}
	res, err := d.Inventory.SoftReserve(ctx, ports.SoftReserveRequest{
		TenantID:       in.TenantID,
		WarehouseID:    in.WarehouseID,
		ExternalRef:    in.ExternalOrderID,
		IdempotencyKey: idem,
		Lines:          invLines,
	})
	if err != nil {
		return domain.FulfillmentOrder{}, err
	}
	fo.ReservationID = &res.ReservationID
	fo.Status = domain.FulfillmentStatusReserved
	fo.UpdatedAt = d.now()

	if err := d.Fulfillments.Create(ctx, fo); err != nil {
		return domain.FulfillmentOrder{}, err
	}

	d.publishEvent(ctx, domain.EventOrderReceived, fo.TenantID, fo.WarehouseID, fo.ID, map[string]any{
		"externalOrderId": fo.ExternalOrderID,
		"reservationId":   res.ReservationID,
	})

	task := domain.Task{
		ID:            d.newID(),
		TenantID:      fo.TenantID,
		WarehouseID:   fo.WarehouseID,
		FulfillmentID: fo.ID,
		Type:          domain.TaskTypePick,
		Status:        domain.TaskStatusQueued,
		Priority:      fo.Priority,
		CreatedAt:     d.now(),
		UpdatedAt:     d.now(),
	}
	appendTaskHistory(&task, d.now(), "created", in.ActorID, "", domain.TaskStatusQueued, "pick queued")
	if err := d.Tasks.Create(ctx, task); err != nil {
		return domain.FulfillmentOrder{}, err
	}

	routeLines := make([]ports.RouteLine, 0, len(fo.Lines))
	for _, l := range fo.Lines {
		routeLines = append(routeLines, ports.RouteLine{
			LineID: l.ID, LocationCode: l.LocationCode, SKUCode: l.SKUCode, Sequence: l.Sequence,
		})
	}
	if d.RouteAI != nil {
		if optimized, err := d.RouteAI.OptimizePickRoute(ctx, fo.WarehouseID, routeLines); err == nil {
			routeLines = optimized
		}
	}

	session := domain.PickSession{
		ID:            d.newID(),
		TenantID:      fo.TenantID,
		WarehouseID:   fo.WarehouseID,
		FulfillmentID: fo.ID,
		TaskID:        task.ID,
		CreatedAt:     d.now(),
		UpdatedAt:     d.now(),
	}
	lineByID := map[uuid.UUID]domain.FulfillmentLine{}
	for _, l := range fo.Lines {
		lineByID[l.ID] = l
	}
	for i, rl := range routeLines {
		fl := lineByID[rl.LineID]
		session.Lines = append(session.Lines, domain.PickLine{
			ID: d.newID(), SessionID: session.ID, LineID: fl.ID,
			VariantID: fl.VariantID, SKUCode: fl.SKUCode, Barcode: fl.Barcode,
			LocationCode: fl.LocationCode, QtyRequired: fl.QtyOrdered, Sequence: i + 1,
			Status: domain.PickLineStatusPending,
		})
	}
	if err := d.Picks.CreateSession(ctx, session); err != nil {
		return domain.FulfillmentOrder{}, err
	}

	fo.Status = domain.FulfillmentStatusPickQueued
	fo.UpdatedAt = d.now()
	if err := d.Fulfillments.Update(ctx, fo); err != nil {
		return domain.FulfillmentOrder{}, err
	}

	d.publishEvent(ctx, domain.EventTaskAssigned, fo.TenantID, fo.WarehouseID, fo.ID, map[string]any{
		"taskId": task.ID, "type": task.Type,
	})

	return fo, nil
}

// GetFulfillment returns a fulfillment by id.
func (d *Deps) GetFulfillment(ctx context.Context, tenantID, id uuid.UUID) (domain.FulfillmentOrder, error) {
	return d.Fulfillments.GetByID(ctx, tenantID, id)
}

// ListFulfillments lists fulfillments with filters.
func (d *Deps) ListFulfillments(ctx context.Context, f ports.FulfillmentFilter) ([]domain.FulfillmentOrder, int, error) {
	return d.Fulfillments.List(ctx, f)
}

// CancelFulfillmentCmd cancels an in-progress fulfillment and releases inventory.
type CancelFulfillmentCmd struct {
	TenantID      uuid.UUID
	FulfillmentID uuid.UUID
	Reason        string
	ActorID       *uuid.UUID
}

// CancelFulfillment cancels fulfillment, cancels open tasks, releases reservation.
func (d *Deps) CancelFulfillment(ctx context.Context, in CancelFulfillmentCmd) (domain.FulfillmentOrder, error) {
	fo, err := d.Fulfillments.GetByID(ctx, in.TenantID, in.FulfillmentID)
	if err != nil {
		return domain.FulfillmentOrder{}, err
	}
	if fo.Status == domain.FulfillmentStatusDispatched || fo.Status == domain.FulfillmentStatusCancelled {
		return domain.FulfillmentOrder{}, domain.ErrInvalidTransition
	}
	if fo.ReservationID != nil && d.Inventory != nil {
		_ = d.Inventory.Release(ctx, ports.ReleaseRequest{
			TenantID: in.TenantID, ReservationID: *fo.ReservationID,
			IdempotencyKey: "release:" + fo.ID.String(),
		})
	}
	tasks, _, _ := d.Tasks.List(ctx, ports.TaskFilter{
		TenantID: in.TenantID, WarehouseID: fo.WarehouseID, Limit: 100,
	})
	for _, t := range tasks {
		if t.FulfillmentID != fo.ID {
			continue
		}
		if t.Status == domain.TaskStatusCompleted || t.Status == domain.TaskStatusCancelled {
			continue
		}
		from := t.Status
		t.Status = domain.TaskStatusCancelled
		t.UpdatedAt = d.now()
		appendTaskHistory(&t, d.now(), "cancelled", in.ActorID, from, domain.TaskStatusCancelled, in.Reason)
		_ = d.Tasks.Update(ctx, t)
	}
	fo.Status = domain.FulfillmentStatusCancelled
	fo.UpdatedAt = d.now()
	if err := d.Fulfillments.Update(ctx, fo); err != nil {
		return domain.FulfillmentOrder{}, err
	}
	return fo, nil
}
