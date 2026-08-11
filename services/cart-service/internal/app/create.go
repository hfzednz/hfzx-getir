package app

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/nexora/cart-service/internal/domain"
)

// CreateCartInput creates a guest or principal cart.
type CreateCartInput struct {
	TenantID    uuid.UUID
	GuestToken  string
	PrincipalID *uuid.UUID
	CityID      *uuid.UUID
	Currency    string
	Metadata    map[string]any
}

// CreateCart creates a new active cart for guest or principal.
func (d *Deps) CreateCart(ctx context.Context, in CreateCartInput) (domain.Cart, error) {
	if in.TenantID == uuid.Nil {
		return domain.Cart{}, fmt.Errorf("%w: tenant_id required", domain.ErrInvalidArgument)
	}
	guest := strings.TrimSpace(in.GuestToken)
	hasPrincipal := in.PrincipalID != nil && *in.PrincipalID != uuid.Nil
	if guest == "" && !hasPrincipal {
		return domain.Cart{}, fmt.Errorf("%w", domain.ErrOwnerRequired)
	}
	if guest != "" && hasPrincipal {
		return domain.Cart{}, fmt.Errorf("%w: provide guest or principal, not both", domain.ErrInvalidArgument)
	}
	currency := strings.ToUpper(strings.TrimSpace(in.Currency))
	if currency == "" {
		currency = "TRY"
	}
	if _, err := domain.ZeroMoney(currency); err != nil {
		return domain.Cart{}, err
	}

	// Return existing active cart if present.
	if hasPrincipal {
		if existing, err := d.Carts.GetActiveByPrincipal(ctx, in.TenantID, *in.PrincipalID); err == nil {
			return existing, nil
		}
	} else {
		if existing, err := d.Carts.GetActiveByGuest(ctx, in.TenantID, guest); err == nil {
			return existing, nil
		}
	}

	now := d.now()
	c := domain.Cart{
		ID:          d.newID(),
		TenantID:    in.TenantID,
		GuestToken:  guest,
		PrincipalID: in.PrincipalID,
		CityID:      in.CityID,
		Status:      domain.CartStatusActive,
		Currency:    currency,
		Lines:       nil,
		Coupons:     nil,
		Version:     1,
		Metadata:    in.Metadata,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if c.Metadata == nil {
		c.Metadata = map[string]any{}
	}
	if err := c.Validate(); err != nil {
		return domain.Cart{}, err
	}
	if err := d.Carts.Create(ctx, c); err != nil {
		return domain.Cart{}, err
	}
	_ = d.appendEvent(ctx, c.ID, c.TenantID, domain.EventCartCreated, map[string]any{
		"guestToken":  c.GuestToken != "",
		"principalId": ptrUUIDString(c.PrincipalID),
	})
	return c, nil
}

// GetCartInput loads a cart by id or by owner headers.
type GetCartInput struct {
	TenantID    uuid.UUID
	CartID      *uuid.UUID
	GuestToken  string
	PrincipalID *uuid.UUID
}

// GetCart returns a cart by id or active cart for guest/principal.
func (d *Deps) GetCart(ctx context.Context, in GetCartInput) (domain.Cart, error) {
	if in.TenantID == uuid.Nil {
		return domain.Cart{}, fmt.Errorf("%w: tenant_id required", domain.ErrInvalidArgument)
	}
	if in.CartID != nil && *in.CartID != uuid.Nil {
		return d.loadCart(ctx, in.TenantID, *in.CartID)
	}
	if in.PrincipalID != nil && *in.PrincipalID != uuid.Nil {
		return d.Carts.GetActiveByPrincipal(ctx, in.TenantID, *in.PrincipalID)
	}
	guest := strings.TrimSpace(in.GuestToken)
	if guest != "" {
		return d.Carts.GetActiveByGuest(ctx, in.TenantID, guest)
	}
	return domain.Cart{}, fmt.Errorf("%w", domain.ErrOwnerRequired)
}

// GetOrCreateCart returns the active cart or creates one.
func (d *Deps) GetOrCreateCart(ctx context.Context, in CreateCartInput) (domain.Cart, error) {
	c, err := d.GetCart(ctx, GetCartInput{
		TenantID: in.TenantID, GuestToken: in.GuestToken, PrincipalID: in.PrincipalID,
	})
	if err == nil {
		return c, nil
	}
	if err != nil && !isNotFound(err) {
		return domain.Cart{}, err
	}
	return d.CreateCart(ctx, in)
}

func isNotFound(err error) bool {
	return errors.Is(err, domain.ErrNotFound)
}

func ptrUUIDString(id *uuid.UUID) string {
	if id == nil {
		return ""
	}
	return id.String()
}
