package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/cart-service/internal/domain"
)

// AddLineInput adds or increments a cart line.
type AddLineInput struct {
	TenantID        uuid.UUID
	CartID          uuid.UUID
	VariantID       uuid.UUID
	Qty             int
	MaxQty          int
	Notes           string
	Addons          []domain.LineAddon
	ReplacementPref string
}

// AddLine adds a line or sums qty for an existing variant.
func (d *Deps) AddLine(ctx context.Context, in AddLineInput) (domain.Cart, error) {
	if in.VariantID == uuid.Nil {
		return domain.Cart{}, fmt.Errorf("%w: variant_id required", domain.ErrInvalidArgument)
	}
	if in.Qty <= 0 {
		return domain.Cart{}, fmt.Errorf("%w: qty must be > 0", domain.ErrInvalidArgument)
	}
	maxQty := in.MaxQty
	if maxQty <= 0 {
		maxQty = domain.DefaultMaxQty
	}

	c, err := d.loadCart(ctx, in.TenantID, in.CartID)
	if err != nil {
		return domain.Cart{}, err
	}
	if err := c.RequireActive(); err != nil {
		return domain.Cart{}, err
	}

	now := d.now()
	if idx := lineIndex(c.Lines, in.VariantID); idx >= 0 {
		newQty := c.Lines[idx].Qty + in.Qty
		if newQty > c.Lines[idx].MaxQty {
			return domain.Cart{}, fmt.Errorf("%w: qty %d > max %d", domain.ErrMaxQtyExceeded, newQty, c.Lines[idx].MaxQty)
		}
		c.Lines[idx].Qty = newQty
		if in.Notes != "" {
			c.Lines[idx].Notes = in.Notes
		}
		c.Lines[idx].UpdatedAt = now
	} else {
		line := domain.CartLine{
			ID:              d.newID(),
			CartID:          c.ID,
			TenantID:        c.TenantID,
			VariantID:       in.VariantID,
			Qty:             in.Qty,
			MaxQty:          maxQty,
			Notes:           in.Notes,
			Addons:          in.Addons,
			ReplacementPref: in.ReplacementPref,
			CreatedAt:       now,
			UpdatedAt:       now,
		}
		if err := line.Validate(); err != nil {
			return domain.Cart{}, err
		}
		c.Lines = append(c.Lines, line)
	}
	c.Quote = nil
	c.UpdatedAt = now
	c.Version++
	if err := d.Carts.Update(ctx, c); err != nil {
		return domain.Cart{}, err
	}
	_ = d.appendEvent(ctx, c.ID, c.TenantID, domain.EventItemAdded, map[string]any{
		"variantId": in.VariantID.String(),
		"qty":       in.Qty,
	})
	return c, nil
}

// UpdateQtyInput updates a line quantity.
type UpdateQtyInput struct {
	TenantID uuid.UUID
	CartID   uuid.UUID
	LineID   uuid.UUID
	Qty      int
}

// UpdateQty sets line qty (enforces max qty).
func (d *Deps) UpdateQty(ctx context.Context, in UpdateQtyInput) (domain.Cart, error) {
	c, err := d.loadCart(ctx, in.TenantID, in.CartID)
	if err != nil {
		return domain.Cart{}, err
	}
	if err := c.RequireActive(); err != nil {
		return domain.Cart{}, err
	}
	idx := -1
	for i, l := range c.Lines {
		if l.ID == in.LineID {
			idx = i
			break
		}
	}
	if idx < 0 {
		return domain.Cart{}, fmt.Errorf("%w: line", domain.ErrNotFound)
	}
	if err := c.Lines[idx].SetQty(in.Qty); err != nil {
		return domain.Cart{}, err
	}
	now := d.now()
	c.Lines[idx].UpdatedAt = now
	c.Quote = nil
	c.UpdatedAt = now
	c.Version++
	if err := d.Carts.Update(ctx, c); err != nil {
		return domain.Cart{}, err
	}
	_ = d.appendEvent(ctx, c.ID, c.TenantID, domain.EventCartUpdated, map[string]any{
		"lineId": in.LineID.String(),
		"qty":    in.Qty,
	})
	return c, nil
}

// RemoveLineInput removes a cart line.
type RemoveLineInput struct {
	TenantID uuid.UUID
	CartID   uuid.UUID
	LineID   uuid.UUID
}

// RemoveLine removes a line by id.
func (d *Deps) RemoveLine(ctx context.Context, in RemoveLineInput) (domain.Cart, error) {
	c, err := d.loadCart(ctx, in.TenantID, in.CartID)
	if err != nil {
		return domain.Cart{}, err
	}
	if err := c.RequireActive(); err != nil {
		return domain.Cart{}, err
	}
	idx := -1
	var variantID uuid.UUID
	for i, l := range c.Lines {
		if l.ID == in.LineID {
			idx = i
			variantID = l.VariantID
			break
		}
	}
	if idx < 0 {
		return domain.Cart{}, fmt.Errorf("%w: line", domain.ErrNotFound)
	}
	c.Lines = append(c.Lines[:idx], c.Lines[idx+1:]...)
	now := d.now()
	c.Quote = nil
	c.UpdatedAt = now
	c.Version++
	if err := d.Carts.Update(ctx, c); err != nil {
		return domain.Cart{}, err
	}
	_ = d.appendEvent(ctx, c.ID, c.TenantID, domain.EventItemRemoved, map[string]any{
		"lineId":    in.LineID.String(),
		"variantId": variantID.String(),
	})
	return c, nil
}

func lineIndex(lines []domain.CartLine, variantID uuid.UUID) int {
	for i, l := range lines {
		if l.VariantID == variantID {
			return i
		}
	}
	return -1
}
