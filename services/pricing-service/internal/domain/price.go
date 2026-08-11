package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// PriceScope is the waterfall level for a price entry.
type PriceScope string

const (
	ScopeBase      PriceScope = "base"
	ScopeRegional  PriceScope = "regional"
	ScopeWarehouse PriceScope = "warehouse"
	ScopeCustomer  PriceScope = "customer"
	ScopeVIP       PriceScope = "vip"
	ScopeCorporate PriceScope = "corporate"
)

// Valid reports whether the scope is recognized.
func (s PriceScope) Valid() bool {
	switch s {
	case ScopeBase, ScopeRegional, ScopeWarehouse, ScopeCustomer, ScopeVIP, ScopeCorporate:
		return true
	default:
		return false
	}
}

// Priority returns waterfall specificity (higher wins).
// Order: base → regional → warehouse → customer → vip → corporate.
func (s PriceScope) Priority() int {
	switch s {
	case ScopeBase:
		return 1
	case ScopeRegional:
		return 2
	case ScopeWarehouse:
		return 3
	case ScopeCustomer:
		return 4
	case ScopeVIP:
		return 5
	case ScopeCorporate:
		return 6
	default:
		return 0
	}
}

// PriceBook groups currency-scoped price entries for a tenant.
type PriceBook struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	Name      string
	Currency  string
	Active    bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Validate checks price book invariants.
func (b PriceBook) Validate() error {
	if b.ID == uuid.Nil {
		return fmt.Errorf("%w: price_book id required", ErrInvalidArgument)
	}
	if b.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if strings.TrimSpace(b.Name) == "" {
		return fmt.Errorf("%w: name required", ErrInvalidArgument)
	}
	if _, err := NewMoney(0, b.Currency); err != nil {
		return err
	}
	return nil
}

// PriceEntry is a scoped unit price for an opaque variant_id.
type PriceEntry struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	PriceBookID uuid.UUID
	VariantID   uuid.UUID
	Scope       PriceScope
	ScopeID     *uuid.UUID // nil for base; region/warehouse/customer/etc. otherwise
	AmountMinor int64
	Currency    string
	ValidFrom   time.Time
	ValidTo     *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Validate checks price entry invariants.
func (e PriceEntry) Validate() error {
	if e.ID == uuid.Nil {
		return fmt.Errorf("%w: price_entry id required", ErrInvalidArgument)
	}
	if e.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if e.PriceBookID == uuid.Nil {
		return fmt.Errorf("%w: price_book_id required", ErrInvalidArgument)
	}
	if e.VariantID == uuid.Nil {
		return fmt.Errorf("%w: variant_id required", ErrInvalidArgument)
	}
	if !e.Scope.Valid() {
		return fmt.Errorf("%w: invalid scope %q", ErrInvalidArgument, e.Scope)
	}
	if e.Scope == ScopeBase && e.ScopeID != nil {
		return fmt.Errorf("%w: base scope must not have scope_id", ErrInvalidArgument)
	}
	if e.Scope != ScopeBase && (e.ScopeID == nil || *e.ScopeID == uuid.Nil) {
		return fmt.Errorf("%w: scope_id required for %s", ErrInvalidArgument, e.Scope)
	}
	if _, err := NewMoney(e.AmountMinor, e.Currency); err != nil {
		return err
	}
	if e.ValidFrom.IsZero() {
		return fmt.Errorf("%w: valid_from required", ErrInvalidArgument)
	}
	if e.ValidTo != nil && !e.ValidTo.After(e.ValidFrom) {
		return fmt.Errorf("%w: valid_to must be after valid_from", ErrInvalidArgument)
	}
	return nil
}

// ActiveAt reports whether the entry is valid at t.
func (e PriceEntry) ActiveAt(t time.Time) bool {
	if t.Before(e.ValidFrom) {
		return false
	}
	if e.ValidTo != nil && !t.Before(*e.ValidTo) {
		return false
	}
	return true
}

// ResolveContext carries optional scope keys for the waterfall.
type ResolveContext struct {
	RegionID     *uuid.UUID
	WarehouseID  *uuid.UUID
	CustomerID   *uuid.UUID
	VIPID        *uuid.UUID // optional VIP segment id
	CorporateID  *uuid.UUID
	Currency     string
	At           time.Time
}

// Matches reports whether entry applies in ctx for its scope.
func (e PriceEntry) Matches(ctx ResolveContext) bool {
	if !e.ActiveAt(ctx.At) {
		return false
	}
	if ctx.Currency != "" && !strings.EqualFold(e.Currency, ctx.Currency) {
		return false
	}
	switch e.Scope {
	case ScopeBase:
		return true
	case ScopeRegional:
		return ctx.RegionID != nil && e.ScopeID != nil && *e.ScopeID == *ctx.RegionID
	case ScopeWarehouse:
		return ctx.WarehouseID != nil && e.ScopeID != nil && *e.ScopeID == *ctx.WarehouseID
	case ScopeCustomer:
		return ctx.CustomerID != nil && e.ScopeID != nil && *e.ScopeID == *ctx.CustomerID
	case ScopeVIP:
		return ctx.VIPID != nil && e.ScopeID != nil && *e.ScopeID == *ctx.VIPID
	case ScopeCorporate:
		return ctx.CorporateID != nil && e.ScopeID != nil && *e.ScopeID == *ctx.CorporateID
	default:
		return false
	}
}

// ResolvePrice picks the highest-priority matching entry (waterfall).
// base → regional → warehouse → customer → vip → corporate.
func ResolvePrice(entries []PriceEntry, ctx ResolveContext) (PriceEntry, error) {
	var best PriceEntry
	found := false
	for _, e := range entries {
		if !e.Matches(ctx) {
			continue
		}
		if !found || e.Scope.Priority() > best.Scope.Priority() {
			best = e
			found = true
			continue
		}
		if e.Scope.Priority() == best.Scope.Priority() && e.UpdatedAt.After(best.UpdatedAt) {
			best = e
		}
	}
	if !found {
		return PriceEntry{}, fmt.Errorf("%w: no matching price", ErrPriceNotFound)
	}
	return best, nil
}
