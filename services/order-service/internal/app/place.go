package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/nexora/order-service/internal/app/ports"
	"github.com/nexora/order-service/internal/domain"
)

// destDefaultWarehouseID is used when cart/checkout lines have no warehouse
// (in-memory phone-test stack has no warehouse assignment).
var destDefaultWarehouseID = uuid.MustParse("55555555-5555-5555-5555-555555555555")

// PlaceOrderInput starts (or resumes) the place-order saga.
type PlaceOrderInput struct {
	TenantID       uuid.UUID
	OrderID        uuid.UUID
	IdempotencyKey string
}

// PlaceOrder runs validate → soft reserve → authorize → confirm hard → start fulfillment.
// Same idempotency key returns the same order without double-reserving.
// Later warehouse/dispatch events advance state beyond warehouse_assigned.
func (d *Deps) PlaceOrder(ctx context.Context, in PlaceOrderInput) (domain.Order, error) {
	if in.TenantID == uuid.Nil || in.OrderID == uuid.Nil {
		return domain.Order{}, fmt.Errorf("%w: tenant_id and order_id required", domain.ErrInvalidArgument)
	}
	key := strings.TrimSpace(in.IdempotencyKey)
	if key == "" {
		return domain.Order{}, fmt.Errorf("%w: idempotency_key required", domain.ErrInvalidArgument)
	}

	var result domain.Order
	err := d.withPlaceLock(ctx, "place:"+in.TenantID.String()+":"+key, func() error {
		o, err := d.placeOrderLocked(ctx, in, key)
		result = o
		return err
	})
	return result, err
}

func (d *Deps) placeOrderLocked(ctx context.Context, in PlaceOrderInput, key string) (domain.Order, error) {
	if saga, err := d.Sagas.GetByIdempotencyKey(ctx, in.TenantID, key); err == nil {
		o, gerr := d.Orders.GetByID(ctx, in.TenantID, saga.OrderID)
		if gerr != nil {
			return domain.Order{}, gerr
		}
		if saga.Status == domain.SagaInstanceSucceeded ||
			o.Status == domain.OrderStatusWarehouseAssigned ||
			o.Status == domain.OrderStatusPicking ||
			!isPreFulfillment(o.Status) {
			return o, nil
		}
		return d.runPlaceSaga(ctx, o, saga)
	}

	o, err := d.Orders.GetByID(ctx, in.TenantID, in.OrderID)
	if err != nil {
		return domain.Order{}, err
	}
	if o.Status != domain.OrderStatusDraft && o.Status != domain.OrderStatusPendingPayment {
		if o.IdempotencyKey == key || !isPreFulfillment(o.Status) {
			return o, nil
		}
		return domain.Order{}, fmt.Errorf("%w: cannot place order in status %s", domain.ErrConflict, o.Status)
	}

	now := d.now()
	sagaID := d.newID()
	allSteps := domain.NewPlaceSagaSteps(sagaID, o.ID, o.TenantID, key)
	// Runner executes through StartFulfillment; later events advance state.
	placeSteps := allSteps[:5]
	saga := domain.SagaInstance{
		ID:             sagaID,
		OrderID:        o.ID,
		TenantID:       o.TenantID,
		SagaType:       domain.SagaTypePlace,
		Status:         domain.SagaInstancePending,
		CurrentStep:    domain.SagaStepValidate,
		CorrelationID:  key,
		IdempotencyKey: key,
		Steps:          placeSteps,
		Metadata:       map[string]any{},
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := saga.Validate(); err != nil {
		return domain.Order{}, err
	}
	if err := d.Sagas.Create(ctx, saga); err != nil {
		if existing, gerr := d.Sagas.GetByIdempotencyKey(ctx, in.TenantID, key); gerr == nil {
			o2, getErr := d.Orders.GetByID(ctx, in.TenantID, existing.OrderID)
			if getErr != nil {
				return domain.Order{}, getErr
			}
			return o2, nil
		}
		return domain.Order{}, err
	}
	return d.runPlaceSaga(ctx, o, saga)
}

func isPreFulfillment(s domain.OrderStatus) bool {
	switch s {
	case domain.OrderStatusDraft, domain.OrderStatusPendingPayment,
		domain.OrderStatusPaymentProcessing, domain.OrderStatusInventoryReservation:
		return true
	default:
		return false
	}
}

func (d *Deps) runPlaceSaga(ctx context.Context, o domain.Order, saga domain.SagaInstance) (domain.Order, error) {
	now := d.now()
	saga.Status = domain.SagaInstanceRunning
	if saga.StartedAt == nil {
		saga.StartedAt = &now
	}
	saga.UpdatedAt = now

	for i := range saga.Steps {
		step := &saga.Steps[i]
		if step.Status == domain.SagaStepSucceeded {
			continue
		}
		saga.CurrentStep = step.Name
		step.Attempt++
		started := d.now()
		step.StartedAt = &started
		step.UpdatedAt = started

		err := d.executePlaceStep(ctx, &o, step)
		// Retry stub: one extra attempt before permanent failure.
		if err != nil && step.Attempt < 2 {
			step.Attempt++
			step.LastError = err.Error()
			err = d.executePlaceStep(ctx, &o, step)
		}
		if err != nil {
			step.Status = domain.SagaStepFailed
			step.LastError = err.Error()
			step.UpdatedAt = d.now()
			saga.LastError = err.Error()
			saga.Status = domain.SagaInstanceCompensating
			saga.UpdatedAt = d.now()
			d.markPlaceStepFailed(&o, step.Name, err)
			_ = d.Sagas.Update(ctx, saga)
			_ = d.Orders.Update(ctx, o)

			_ = d.compensatePlace(ctx, &o, saga, i)
			saga.Status = domain.SagaInstanceCompensated
			completed := d.now()
			saga.CompletedAt = &completed
			saga.UpdatedAt = completed
			_ = d.Sagas.Update(ctx, saga)
			_ = d.Orders.Update(ctx, o)
			d.indexOrder(ctx, o)
			return o, err
		}

		completed := d.now()
		step.Status = domain.SagaStepSucceeded
		step.CompletedAt = &completed
		step.UpdatedAt = completed
		saga.UpdatedAt = completed
		_ = d.Sagas.Update(ctx, saga)
		_ = d.Orders.Update(ctx, o)
	}

	saga.Status = domain.SagaInstanceSucceeded
	done := d.now()
	saga.CompletedAt = &done
	saga.UpdatedAt = done
	_ = d.Sagas.Update(ctx, saga)
	d.indexOrder(ctx, o)
	return o, nil
}

func (d *Deps) executePlaceStep(ctx context.Context, o *domain.Order, step *domain.SagaStep) error {
	switch step.Name {
	case domain.SagaStepValidate:
		if err := o.Validate(); err != nil {
			return err
		}
		if len(o.Lines) == 0 {
			return fmt.Errorf("%w: no lines", domain.ErrInvalidArgument)
		}
		switch o.Status {
		case domain.OrderStatusDraft:
			if err := d.transition(o, domain.OrderStatusPendingPayment); err != nil {
				return err
			}
		case domain.OrderStatusPendingPayment:
			// ok
		default:
			return fmt.Errorf("%w: place from %s", domain.ErrConflict, o.Status)
		}
		return d.appendEvent(ctx, o.ID, o.TenantID, domain.EventOrderValidated, map[string]any{
			"status": string(o.Status),
		})

	case domain.SagaStepSoftReserve:
		if o.Status == domain.OrderStatusPendingPayment {
			if err := d.transition(o, domain.OrderStatusInventoryReservation); err != nil {
				return err
			}
		}
		lines := reserveLinesFromOrder(*o)
		if len(lines) == 0 {
			return fmt.Errorf("%w: warehouse_id required on lines for reserve", domain.ErrInvalidArgument)
		}
		res, err := d.Inventory.SoftReserve(ctx, ports.SoftReserveRequest{
			TenantID: o.TenantID, OrderID: o.ID,
			IdempotencyKey: step.IdempotencyKey,
			Lines:          lines,
		})
		if err != nil {
			return err
		}
		o.ReservationRef = res.ReservationRef
		return d.appendEvent(ctx, o.ID, o.TenantID, domain.EventInventoryReserved, map[string]any{
			"reservationRef": res.ReservationRef,
		})

	case domain.SagaStepAuthorizePayment:
		if o.Status == domain.OrderStatusInventoryReservation {
			if err := d.transition(o, domain.OrderStatusPaymentProcessing); err != nil {
				return err
			}
		}
		auth, err := d.Payment.Authorize(ctx, ports.AuthorizeRequest{
			TenantID: o.TenantID, OrderID: o.ID,
			AmountMinor: o.TotalMinor, Currency: o.Currency,
			IdempotencyKey: step.IdempotencyKey,
		})
		if err != nil {
			return err
		}
		o.PaymentIntentRef = auth.PaymentIntentRef
		return d.appendEvent(ctx, o.ID, o.TenantID, domain.EventPaymentAuthorized, map[string]any{
			"paymentIntentRef": auth.PaymentIntentRef,
		})

	case domain.SagaStepConfirmHard:
		if o.ReservationRef == "" {
			return fmt.Errorf("%w: reservation_ref missing", domain.ErrInvariant)
		}
		return d.Inventory.ConfirmHard(ctx, ports.ConfirmHardRequest{
			TenantID: o.TenantID, ReservationRef: o.ReservationRef,
			IdempotencyKey: step.IdempotencyKey,
		})

	case domain.SagaStepStartFulfillment:
		lines := reserveLinesFromOrder(*o)
		whID := primaryWarehouse(*o)
		if whID == uuid.Nil {
			return fmt.Errorf("%w: warehouse_id required for fulfillment", domain.ErrInvalidArgument)
		}
		recv, err := d.Warehouse.ReceiveFulfillment(ctx, ports.ReceiveFulfillmentRequest{
			TenantID: o.TenantID, OrderID: o.ID, WarehouseID: whID,
			Priority: o.Priority, IdempotencyKey: step.IdempotencyKey, Lines: lines,
		})
		if err != nil {
			return err
		}
		now := d.now()
		ful := domain.Fulfillment{
			ID: d.newID(), OrderID: o.ID, TenantID: o.TenantID, WarehouseID: whID,
			Status: domain.FulfillmentStatusAssigned, ReservationID: o.ReservationRef,
			FulfillmentRef: recv.FulfillmentRef, Priority: o.Priority,
			Metadata: map[string]any{}, CreatedAt: now, UpdatedAt: now,
		}
		if d.Fulfillments != nil {
			_ = d.Fulfillments.Create(ctx, ful)
		}
		if err := d.transition(o, domain.OrderStatusWarehouseAssigned); err != nil {
			return err
		}
		return d.appendEvent(ctx, o.ID, o.TenantID, domain.EventWarehouseAssigned, map[string]any{
			"fulfillmentRef": recv.FulfillmentRef,
			"warehouseId":    whID.String(),
		})

	default:
		return fmt.Errorf("%w: unknown place step %s", domain.ErrInvalidArgument, step.Name)
	}
}

func (d *Deps) markPlaceStepFailed(o *domain.Order, stepName string, err error) {
	switch stepName {
	case domain.SagaStepSoftReserve:
		if o.CanTransitionTo(domain.OrderStatusInventoryFailed) {
			_ = d.transition(o, domain.OrderStatusInventoryFailed)
		}
	case domain.SagaStepAuthorizePayment:
		if o.CanTransitionTo(domain.OrderStatusPaymentFailed) {
			_ = d.transition(o, domain.OrderStatusPaymentFailed)
		}
	}
	_ = err
}

func (d *Deps) compensatePlace(ctx context.Context, o *domain.Order, saga domain.SagaInstance, failedIdx int) error {
	needRelease := false
	needVoid := false
	for i := 0; i < failedIdx; i++ {
		st := saga.Steps[i]
		if st.Status != domain.SagaStepSucceeded {
			continue
		}
		switch st.Name {
		case domain.SagaStepSoftReserve, domain.SagaStepConfirmHard:
			needRelease = true
		case domain.SagaStepAuthorizePayment:
			needVoid = true
		}
	}
	if needRelease && o.ReservationRef != "" && d.Inventory != nil {
		_ = d.Inventory.Release(ctx, ports.ReleaseRequest{
			TenantID: o.TenantID, ReservationRef: o.ReservationRef,
			IdempotencyKey: saga.IdempotencyKey + ":ReleaseReserve",
		})
	}
	if needVoid && o.PaymentIntentRef != "" && d.Payment != nil {
		_ = d.Payment.Void(ctx, ports.VoidRequest{
			TenantID: o.TenantID, PaymentIntentRef: o.PaymentIntentRef,
			IdempotencyKey: saga.IdempotencyKey + ":VoidPayment",
		})
	}

	switch o.Status {
	case domain.OrderStatusPaymentFailed:
		_ = d.transition(o, domain.OrderStatusCancelled)
		_ = d.appendEvent(ctx, o.ID, o.TenantID, domain.EventCancelled, map[string]any{
			"reason": "payment_failed",
		})
	case domain.OrderStatusInventoryFailed:
		_ = d.transition(o, domain.OrderStatusCancelled)
		_ = d.appendEvent(ctx, o.ID, o.TenantID, domain.EventCancelled, map[string]any{
			"reason": "inventory_failed",
		})
	case domain.OrderStatusInventoryReservation, domain.OrderStatusPaymentProcessing,
		domain.OrderStatusPendingPayment, domain.OrderStatusWarehouseAssigned:
		if o.CanTransitionTo(domain.OrderStatusCancelled) {
			_ = d.transition(o, domain.OrderStatusCancelled)
			_ = d.appendEvent(ctx, o.ID, o.TenantID, domain.EventCancelled, map[string]any{
				"reason": "place_saga_compensated",
			})
		}
	}
	return nil
}

func reserveLinesFromOrder(o domain.Order) []ports.ReserveLine {
	out := make([]ports.ReserveLine, 0, len(o.Lines))
	for _, l := range o.Lines {
		wh := uuid.Nil
		if l.WarehouseID != nil {
			wh = *l.WarehouseID
		} else if len(o.WarehouseIDs) > 0 {
			wh = o.WarehouseIDs[0]
		}
		if wh == uuid.Nil {
			wh = destDefaultWarehouseID
		}
		out = append(out, ports.ReserveLine{
			VariantID: l.VariantID, SKUCode: l.SKUCode, Qty: l.Qty, WarehouseID: wh,
		})
	}
	return out
}

func primaryWarehouse(o domain.Order) uuid.UUID {
	if len(o.WarehouseIDs) > 0 {
		return o.WarehouseIDs[0]
	}
	for _, l := range o.Lines {
		if l.WarehouseID != nil {
			return *l.WarehouseID
		}
	}
	return destDefaultWarehouseID
}
