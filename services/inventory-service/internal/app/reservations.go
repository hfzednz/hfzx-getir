package app

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/inventory-service/internal/domain"
)

// SoftReserveLine is a soft hold request line.
type SoftReserveLine struct {
	WarehouseID uuid.UUID
	VariantID   uuid.UUID
	SKUCode     string
	LocationID  *uuid.UUID
	Qty         int64
	UseFEFO     bool
}

// SoftReserveCmd creates a soft reservation (idempotent).
type SoftReserveCmd struct {
	TenantID       uuid.UUID
	WarehouseID    *uuid.UUID
	ExternalRef    string
	Priority       int
	TTL            time.Duration
	Lines          []SoftReserveLine
	IdempotencyKey string
	ActorID        *uuid.UUID
}

// SoftReserve holds available stock for a short TTL. Concurrent-safe via per-key mutex.
func (d *Deps) SoftReserve(ctx context.Context, in SoftReserveCmd) (domain.Reservation, error) {
	if in.TenantID == uuid.Nil || len(in.Lines) == 0 {
		return domain.Reservation{}, domain.ErrInvalidArgument
	}
	if in.IdempotencyKey == "" {
		return domain.Reservation{}, fmt.Errorf("%w: idempotency_key required", domain.ErrInvalidArgument)
	}
	if v, ok := d.idemGet(ctx, "soft:"+in.IdempotencyKey); ok {
		if r, ok := v.(domain.Reservation); ok {
			return r, nil
		}
	}

	ttl := in.TTL
	if ttl <= 0 {
		ttl = d.softTTL()
	}
	expires := d.now().Add(ttl)

	// Serialize multi-line reserves by locking each stock key in sorted order to avoid deadlock.
	keys := make([]string, 0, len(in.Lines))
	for _, line := range in.Lines {
		keys = append(keys, stockKey(line.WarehouseID, line.VariantID, line.LocationID))
	}
	sortStrings(keys)

	var out domain.Reservation
	err := d.withMultiLock(ctx, keys, func() error {
		// re-check idempotency inside locks
		if v, ok := d.idemGet(ctx, "soft:"+in.IdempotencyKey); ok {
			if r, ok := v.(domain.Reservation); ok {
				out = r
				return nil
			}
		}

		now := d.now()
		res := domain.Reservation{
			ID:          d.newID(),
			TenantID:    in.TenantID,
			WarehouseID: in.WarehouseID,
			Type:        domain.ReservationTypeSoft,
			Status:      domain.ReservationStatusActive,
			ExpiresAt:   &expires,
			Priority:    in.Priority,
			ExternalRef: in.ExternalRef,
			ActorID:     in.ActorID,
			Metadata:    map[string]any{},
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		for _, line := range in.Lines {
			if line.Qty <= 0 {
				return domain.ErrNegativeQuantity
			}
			b, err := d.EnsureBalance(ctx, EnsureBalanceInput{
				TenantID: in.TenantID, WarehouseID: line.WarehouseID,
				VariantID: line.VariantID, SKUCode: line.SKUCode, LocationID: line.LocationID,
			})
			if err != nil {
				return err
			}
			var lotID *uuid.UUID
			if line.UseFEFO {
				lots, err := d.Lots.ListByWarehouseVariant(ctx, in.TenantID, line.WarehouseID, line.VariantID)
				if err != nil {
					return err
				}
				allocs, err := domain.AllocateLotsFEFO(lots, line.Qty, now)
				if err != nil {
					return err
				}
				if len(allocs) > 0 {
					id := allocs[0].LotID
					lotID = &id
				}
			}
			if err := b.Reserve(line.Qty); err != nil {
				return err
			}
			if err := d.Balances.Update(ctx, b); err != nil {
				return err
			}
			rl := domain.ReservationLine{
				ID:            d.newID(),
				ReservationID: res.ID,
				WarehouseID:   line.WarehouseID,
				VariantID:     line.VariantID,
				SKUCode:       line.SKUCode,
				Qty:           line.Qty,
				LotID:         lotID,
				BalanceID:     &b.ID,
				LocationID:    line.LocationID,
				Metadata:      map[string]any{},
				CreatedAt:     now,
			}
			res.Lines = append(res.Lines, rl)
			d.indexBalance(ctx, b)
			d.publishEvent(ctx, domain.EventStockReserved, in.TenantID, line.WarehouseID, line.VariantID, map[string]any{
				"reservationId": res.ID, "qty": line.Qty, "type": "soft",
			})
		}
		if err := res.Validate(); err != nil {
			return err
		}
		if err := d.Reservations.Create(ctx, res); err != nil {
			return err
		}
		out = res
		d.idemPut(ctx, "soft:"+in.IdempotencyKey, res)
		return nil
	})
	return out, err
}

// ConfirmHardCmd upgrades a soft reservation to hard (idempotent).
type ConfirmHardCmd struct {
	TenantID       uuid.UUID
	ReservationID  uuid.UUID
	IdempotencyKey string
}

// ConfirmHard transitions Soft → Hard.
func (d *Deps) ConfirmHard(ctx context.Context, in ConfirmHardCmd) (domain.Reservation, error) {
	if in.IdempotencyKey == "" {
		return domain.Reservation{}, fmt.Errorf("%w: idempotency_key required", domain.ErrInvalidArgument)
	}
	if v, ok := d.idemGet(ctx, "hard:"+in.IdempotencyKey); ok {
		if r, ok := v.(domain.Reservation); ok {
			return r, nil
		}
	}
	res, err := d.Reservations.GetByID(ctx, in.TenantID, in.ReservationID)
	if err != nil {
		return domain.Reservation{}, err
	}
	if err := d.expireIfNeeded(ctx, &res); err != nil {
		return domain.Reservation{}, err
	}
	if err := res.ConfirmHard(); err != nil {
		return domain.Reservation{}, err
	}
	if err := d.Reservations.Update(ctx, res); err != nil {
		return domain.Reservation{}, err
	}
	wh := uuid.Nil
	if res.WarehouseID != nil {
		wh = *res.WarehouseID
	} else if len(res.Lines) > 0 {
		wh = res.Lines[0].WarehouseID
	}
	var vid uuid.UUID
	if len(res.Lines) > 0 {
		vid = res.Lines[0].VariantID
	}
	d.publishEvent(ctx, domain.EventReservationConfirmed, res.TenantID, wh, vid, map[string]any{
		"reservationId": res.ID,
	})
	d.idemPut(ctx, "hard:"+in.IdempotencyKey, res)
	return res, nil
}

// ExtendReservationCmd extends a soft hold TTL.
type ExtendReservationCmd struct {
	TenantID       uuid.UUID
	ReservationID  uuid.UUID
	ExtendBy       time.Duration
	IdempotencyKey string
}

// Extend pushes expires_at forward for an active soft reservation.
func (d *Deps) Extend(ctx context.Context, in ExtendReservationCmd) (domain.Reservation, error) {
	if in.IdempotencyKey == "" {
		return domain.Reservation{}, fmt.Errorf("%w: idempotency_key required", domain.ErrInvalidArgument)
	}
	if v, ok := d.idemGet(ctx, "extend:"+in.IdempotencyKey); ok {
		if r, ok := v.(domain.Reservation); ok {
			return r, nil
		}
	}
	res, err := d.Reservations.GetByID(ctx, in.TenantID, in.ReservationID)
	if err != nil {
		return domain.Reservation{}, err
	}
	if err := d.expireIfNeeded(ctx, &res); err != nil {
		return domain.Reservation{}, err
	}
	base := d.now()
	if res.ExpiresAt != nil && res.ExpiresAt.After(base) {
		base = *res.ExpiresAt
	}
	ext := in.ExtendBy
	if ext <= 0 {
		ext = d.softTTL()
	}
	if err := res.ExtendSoft(base.Add(ext)); err != nil {
		return domain.Reservation{}, err
	}
	if err := d.Reservations.Update(ctx, res); err != nil {
		return domain.Reservation{}, err
	}
	d.idemPut(ctx, "extend:"+in.IdempotencyKey, res)
	return res, nil
}

// ReleaseReservationCmd releases a soft/hard hold and restores available.
type ReleaseReservationCmd struct {
	TenantID       uuid.UUID
	ReservationID  uuid.UUID
	IdempotencyKey string
}

// Release frees reserved qty back to available (idempotent).
func (d *Deps) Release(ctx context.Context, in ReleaseReservationCmd) (domain.Reservation, error) {
	if in.IdempotencyKey == "" {
		return domain.Reservation{}, fmt.Errorf("%w: idempotency_key required", domain.ErrInvalidArgument)
	}
	if v, ok := d.idemGet(ctx, "release:"+in.IdempotencyKey); ok {
		if r, ok := v.(domain.Reservation); ok {
			return r, nil
		}
	}
	res, err := d.Reservations.GetByID(ctx, in.TenantID, in.ReservationID)
	if err != nil {
		return domain.Reservation{}, err
	}
	if res.Status == domain.ReservationStatusReleased || res.Status == domain.ReservationStatusExpired {
		d.idemPut(ctx, "release:"+in.IdempotencyKey, res)
		return res, nil
	}
	keys := reservationKeys(res)
	err = d.withMultiLock(ctx, keys, func() error {
		fresh, err := d.Reservations.GetByID(ctx, in.TenantID, in.ReservationID)
		if err != nil {
			return err
		}
		if !fresh.IsActive() {
			res = fresh
			return nil
		}
		if err := fresh.Release(); err != nil {
			return err
		}
		for _, line := range fresh.Lines {
			b, err := d.Balances.GetByID(ctx, in.TenantID, *line.BalanceID)
			if err != nil {
				return err
			}
			if err := b.ReleaseReserved(line.Qty); err != nil {
				return err
			}
			if err := d.Balances.Update(ctx, b); err != nil {
				return err
			}
			d.indexBalance(ctx, b)
			d.publishEvent(ctx, domain.EventReservationReleased, in.TenantID, line.WarehouseID, line.VariantID, map[string]any{
				"reservationId": fresh.ID, "qty": line.Qty,
			})
		}
		if err := d.Reservations.Update(ctx, fresh); err != nil {
			return err
		}
		res = fresh
		return nil
	})
	if err != nil {
		return domain.Reservation{}, err
	}
	d.idemPut(ctx, "release:"+in.IdempotencyKey, res)
	return res, nil
}

// ConsumeReservationCmd ships/deducts a hard reservation (idempotent).
type ConsumeReservationCmd struct {
	TenantID       uuid.UUID
	ReservationID  uuid.UUID
	IdempotencyKey string
	ActorID        *uuid.UUID
}

// Consume deducts on_hand for a hard reservation. Soft without ConfirmHard fails.
func (d *Deps) Consume(ctx context.Context, in ConsumeReservationCmd) (domain.Reservation, error) {
	if in.IdempotencyKey == "" {
		return domain.Reservation{}, fmt.Errorf("%w: idempotency_key required", domain.ErrInvalidArgument)
	}
	if v, ok := d.idemGet(ctx, "consume:"+in.IdempotencyKey); ok {
		if r, ok := v.(domain.Reservation); ok {
			return r, nil
		}
	}
	res, err := d.Reservations.GetByID(ctx, in.TenantID, in.ReservationID)
	if err != nil {
		return domain.Reservation{}, err
	}
	keys := reservationKeys(res)
	err = d.withMultiLock(ctx, keys, func() error {
		fresh, err := d.Reservations.GetByID(ctx, in.TenantID, in.ReservationID)
		if err != nil {
			return err
		}
		if fresh.Status == domain.ReservationStatusConsumed {
			res = fresh
			return nil
		}
		if err := fresh.Consume(); err != nil {
			return err
		}
		for _, line := range fresh.Lines {
			b, err := d.Balances.GetByID(ctx, in.TenantID, *line.BalanceID)
			if err != nil {
				return err
			}
			if err := b.ConsumeReserved(line.Qty); err != nil {
				return err
			}
			if err := d.Balances.Update(ctx, b); err != nil {
				return err
			}
			m, err := domain.NewMovementFromBalance(
				d.newID(), in.TenantID, b, domain.MovementTypeSale, -line.Qty,
				in.IdempotencyKey+":"+line.ID.String(), in.ActorID, "consume",
			)
			if err != nil {
				return err
			}
			rid := fresh.ID
			m.ReservationID = &rid
			if err := d.Movements.Create(ctx, m); err != nil {
				return err
			}
			d.indexBalance(ctx, b)
			d.publishEvent(ctx, domain.EventReservationConsumed, in.TenantID, line.WarehouseID, line.VariantID, map[string]any{
				"reservationId": fresh.ID, "qty": line.Qty,
			})
		}
		if err := d.Reservations.Update(ctx, fresh); err != nil {
			return err
		}
		res = fresh
		return nil
	})
	if err != nil {
		return domain.Reservation{}, err
	}
	d.idemPut(ctx, "consume:"+in.IdempotencyKey, res)
	return res, nil
}

// GetReservation returns a reservation by id (auto-expires soft holds).
func (d *Deps) GetReservation(ctx context.Context, tenantID, id uuid.UUID) (domain.Reservation, error) {
	res, err := d.Reservations.GetByID(ctx, tenantID, id)
	if err != nil {
		return domain.Reservation{}, err
	}
	_ = d.expireIfNeeded(ctx, &res)
	return res, nil
}

func (d *Deps) expireIfNeeded(ctx context.Context, res *domain.Reservation) error {
	if res == nil || !res.IsActive() || res.Type != domain.ReservationTypeSoft {
		return nil
	}
	if !res.IsExpiredAt(d.now()) {
		return nil
	}
	keys := reservationKeys(*res)
	return d.withMultiLock(ctx, keys, func() error {
		fresh, err := d.Reservations.GetByID(ctx, res.TenantID, res.ID)
		if err != nil {
			return err
		}
		if !fresh.IsActive() || !fresh.IsExpiredAt(d.now()) {
			*res = fresh
			return nil
		}
		if err := fresh.Expire(d.now()); err != nil {
			return err
		}
		for _, line := range fresh.Lines {
			b, err := d.Balances.GetByID(ctx, fresh.TenantID, *line.BalanceID)
			if err != nil {
				return err
			}
			if err := b.ReleaseReserved(line.Qty); err != nil {
				return err
			}
			if err := d.Balances.Update(ctx, b); err != nil {
				return err
			}
			d.indexBalance(ctx, b)
			d.publishEvent(ctx, domain.EventReservationExpired, fresh.TenantID, line.WarehouseID, line.VariantID, map[string]any{
				"reservationId": fresh.ID, "qty": line.Qty,
			})
		}
		if err := d.Reservations.Update(ctx, fresh); err != nil {
			return err
		}
		*res = fresh
		return nil
	})
}

func reservationKeys(res domain.Reservation) []string {
	keys := make([]string, 0, len(res.Lines))
	seen := map[string]struct{}{}
	for _, line := range res.Lines {
		k := stockKey(line.WarehouseID, line.VariantID, line.LocationID)
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		keys = append(keys, k)
	}
	sortStrings(keys)
	return keys
}

func (d *Deps) withMultiLock(ctx context.Context, keys []string, fn func() error) error {
	if len(keys) == 0 {
		return fn()
	}
	var run func(i int) error
	run = func(i int) error {
		if i >= len(keys) {
			return fn()
		}
		return d.withStockLock(ctx, keys[i], func() error { return run(i + 1) })
	}
	return run(0)
}

func sortStrings(s []string) {
	for i := 0; i < len(s); i++ {
		for j := i + 1; j < len(s); j++ {
			if s[j] < s[i] {
				s[i], s[j] = s[j], s[i]
			}
		}
	}
}
