package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/nexora/cart-service/internal/domain"
)

// MergeCartsInput merges a guest cart into the principal cart on login.
type MergeCartsInput struct {
	TenantID    uuid.UUID
	GuestToken  string
	PrincipalID uuid.UUID
	Policy      domain.MergePolicy
	CityID      *uuid.UUID
	Currency    string
}

// MergeCarts merges guest ∪ auth (qty sum by default) and marks guest as merged.
func (d *Deps) MergeCarts(ctx context.Context, in MergeCartsInput) (domain.Cart, error) {
	if in.TenantID == uuid.Nil || in.PrincipalID == uuid.Nil {
		return domain.Cart{}, fmt.Errorf("%w: tenant and principal required", domain.ErrInvalidArgument)
	}
	guestToken := strings.TrimSpace(in.GuestToken)
	if guestToken == "" {
		return domain.Cart{}, fmt.Errorf("%w: guest_token required", domain.ErrInvalidArgument)
	}

	guest, err := d.Carts.GetActiveByGuest(ctx, in.TenantID, guestToken)
	if err != nil {
		// No guest cart — ensure principal cart exists.
		return d.GetOrCreateCart(ctx, CreateCartInput{
			TenantID: in.TenantID, PrincipalID: &in.PrincipalID,
			CityID: in.CityID, Currency: in.Currency,
		})
	}

	target, err := d.Carts.GetActiveByPrincipal(ctx, in.TenantID, in.PrincipalID)
	if isNotFound(err) {
		// Adopt guest cart as principal cart (re-key ownership).
		now := d.now()
		guest.GuestToken = ""
		pid := in.PrincipalID
		guest.PrincipalID = &pid
		guest.UpdatedAt = now
		guest.Version++
		if err := d.Carts.Update(ctx, guest); err != nil {
			return domain.Cart{}, err
		}
		_ = d.appendEvent(ctx, guest.ID, guest.TenantID, domain.EventCartMerged, map[string]any{
			"mode": "adopt_guest",
		})
		return guest, nil
	}
	if err != nil {
		return domain.Cart{}, err
	}

	merged, abandoned, err := domain.MergeCarts(target, guest, in.Policy, d.now(), d.newID)
	if err != nil {
		return domain.Cart{}, err
	}
	if err := d.Carts.Update(ctx, abandoned); err != nil {
		return domain.Cart{}, err
	}
	if err := d.Carts.Update(ctx, merged); err != nil {
		return domain.Cart{}, err
	}
	_ = d.appendEvent(ctx, merged.ID, merged.TenantID, domain.EventCartMerged, map[string]any{
		"guestCartId": guest.ID.String(),
		"policy":      string(in.Policy),
		"lineCount":   len(merged.Lines),
	})
	return merged, nil
}
