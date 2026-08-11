package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/cart-service/internal/app/ports"
	"github.com/nexora/cart-service/internal/domain"
)

// SoftReserveLinesInput soft-reserves cart lines via InventoryClient.
type SoftReserveLinesInput struct {
	TenantID       uuid.UUID
	CartID         uuid.UUID
	IdempotencyKey string
}

// SoftReserveLines calls InventoryClient.SoftReserve and stores the opaque ref.
func (d *Deps) SoftReserveLines(ctx context.Context, in SoftReserveLinesInput) (domain.Cart, error) {
	c, err := d.loadCart(ctx, in.TenantID, in.CartID)
	if err != nil {
		return domain.Cart{}, err
	}
	if err := c.RequireActive(); err != nil {
		return domain.Cart{}, err
	}
	if len(c.Lines) == 0 {
		return domain.Cart{}, fmt.Errorf("%w", domain.ErrEmptyCart)
	}
	if d.Inventory == nil {
		return domain.Cart{}, fmt.Errorf("%w: inventory client required", domain.ErrInvariant)
	}

	// Optional ATP check — clamp line max_qty to available when lower.
	atpLines := make([]ports.ATPLine, 0, len(c.Lines))
	for _, l := range c.Lines {
		atpLines = append(atpLines, ports.ATPLine{VariantID: l.VariantID, Qty: l.Qty})
	}
	if atp, err := d.Inventory.ATP(ctx, ports.ATPRequest{
		TenantID: c.TenantID, CityID: c.CityID, Lines: atpLines,
	}); err == nil {
		avail := make(map[uuid.UUID]int, len(atp.Lines))
		for _, al := range atp.Lines {
			avail[al.VariantID] = al.Available
		}
		for i := range c.Lines {
			if a, ok := avail[c.Lines[i].VariantID]; ok && a > 0 && a < c.Lines[i].MaxQty {
				c.Lines[i].MaxQty = a
				if c.Lines[i].Qty > c.Lines[i].MaxQty {
					c.Lines[i].Qty = c.Lines[i].MaxQty
				}
			}
		}
	}

	resLines := make([]ports.SoftReserveLine, 0, len(c.Lines))
	for _, l := range c.Lines {
		resLines = append(resLines, ports.SoftReserveLine{VariantID: l.VariantID, Qty: l.Qty})
	}
	key := in.IdempotencyKey
	if key == "" {
		key = "cart-" + c.ID.String() + "-v" + fmt.Sprintf("%d", c.Version)
	}
	res, err := d.Inventory.SoftReserve(ctx, ports.SoftReserveRequest{
		TenantID: c.TenantID, CartID: c.ID, IdempotencyKey: key, Lines: resLines,
	})
	if err != nil {
		return domain.Cart{}, err
	}
	now := d.now()
	c.ReservationRefs = append(c.ReservationRefs, domain.ReservationRef{
		ID:             d.newID(),
		ReservationRef: res.ReservationRef,
		IdempotencyKey: key,
		ExpiresAt:      res.ExpiresAt,
		CreatedAt:      now,
	})
	c.UpdatedAt = now
	c.Version++
	if err := d.Carts.Update(ctx, c); err != nil {
		return domain.Cart{}, err
	}
	_ = d.appendEvent(ctx, c.ID, c.TenantID, domain.EventSoftReserved, map[string]any{
		"reservationRef": res.ReservationRef,
	})
	return c, nil
}

// ReleaseReservationsInput releases all active soft reservations.
type ReleaseReservationsInput struct {
	TenantID uuid.UUID
	CartID   uuid.UUID
}

// ReleaseReservations releases active soft holds via InventoryClient.
func (d *Deps) ReleaseReservations(ctx context.Context, in ReleaseReservationsInput) (domain.Cart, error) {
	c, err := d.loadCart(ctx, in.TenantID, in.CartID)
	if err != nil {
		return domain.Cart{}, err
	}
	if d.Inventory == nil {
		return c, nil
	}
	now := d.now()
	for i := range c.ReservationRefs {
		r := &c.ReservationRefs[i]
		if !r.Active() {
			continue
		}
		_ = d.Inventory.Release(ctx, ports.ReleaseRequest{
			TenantID: c.TenantID, ReservationRef: r.ReservationRef,
			IdempotencyKey: "release-" + r.ReservationRef,
		})
		r.ReleasedAt = &now
	}
	c.UpdatedAt = now
	c.Version++
	if err := d.Carts.Update(ctx, c); err != nil {
		return domain.Cart{}, err
	}
	return c, nil
}
