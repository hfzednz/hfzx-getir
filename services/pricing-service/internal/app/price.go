package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/pricing-service/internal/domain"
)

// UpsertPriceBookInput creates or updates a price book.
type UpsertPriceBookInput struct {
	TenantID uuid.UUID
	BookID   *uuid.UUID
	Name     string
	Currency string
	Active   *bool
}

// UpsertPriceBook upserts a price book.
func (d *Deps) UpsertPriceBook(ctx context.Context, in UpsertPriceBookInput) (domain.PriceBook, error) {
	if in.TenantID == uuid.Nil {
		return domain.PriceBook{}, fmt.Errorf("%w: tenant_id required", domain.ErrInvalidArgument)
	}
	now := d.now()
	var book domain.PriceBook
	if in.BookID != nil && *in.BookID != uuid.Nil {
		existing, err := d.Prices.GetBook(ctx, in.TenantID, *in.BookID)
		if err != nil {
			return domain.PriceBook{}, err
		}
		book = existing
		if in.Name != "" {
			book.Name = in.Name
		}
		if in.Currency != "" {
			book.Currency = strings.ToUpper(strings.TrimSpace(in.Currency))
		}
		if in.Active != nil {
			book.Active = *in.Active
		}
		book.UpdatedAt = now
	} else {
		active := true
		if in.Active != nil {
			active = *in.Active
		}
		book = domain.PriceBook{
			ID: d.newID(), TenantID: in.TenantID, Name: in.Name,
			Currency: strings.ToUpper(strings.TrimSpace(in.Currency)),
			Active: active, CreatedAt: now, UpdatedAt: now,
		}
	}
	if err := book.Validate(); err != nil {
		return domain.PriceBook{}, err
	}
	if err := d.Prices.UpsertBook(ctx, book); err != nil {
		return domain.PriceBook{}, err
	}
	return book, nil
}

// UpsertPriceInput creates or updates a scoped price entry.
type UpsertPriceInput struct {
	TenantID    uuid.UUID
	EntryID     *uuid.UUID
	PriceBookID uuid.UUID
	VariantID   uuid.UUID
	Scope       domain.PriceScope
	ScopeID     *uuid.UUID
	AmountMinor int64
	Currency    string
	ValidFrom   *time.Time
	ValidTo     *time.Time
}

// UpsertPrice upserts a price entry and emits PriceChanged.
func (d *Deps) UpsertPrice(ctx context.Context, in UpsertPriceInput) (domain.PriceEntry, error) {
	if in.TenantID == uuid.Nil {
		return domain.PriceEntry{}, fmt.Errorf("%w: tenant_id required", domain.ErrInvalidArgument)
	}
	if in.PriceBookID == uuid.Nil {
		return domain.PriceEntry{}, fmt.Errorf("%w: price_book_id required", domain.ErrInvalidArgument)
	}
	if _, err := d.Prices.GetBook(ctx, in.TenantID, in.PriceBookID); err != nil {
		return domain.PriceEntry{}, err
	}
	now := d.now()
	validFrom := now
	if in.ValidFrom != nil {
		validFrom = in.ValidFrom.UTC()
	}
	var entry domain.PriceEntry
	if in.EntryID != nil && *in.EntryID != uuid.Nil {
		existing, err := d.Prices.GetEntry(ctx, in.TenantID, *in.EntryID)
		if err != nil {
			return domain.PriceEntry{}, err
		}
		entry = existing
		if in.VariantID != uuid.Nil {
			entry.VariantID = in.VariantID
		}
		if in.Scope != "" {
			entry.Scope = in.Scope
		}
		entry.ScopeID = in.ScopeID
		entry.AmountMinor = in.AmountMinor
		if in.Currency != "" {
			entry.Currency = strings.ToUpper(strings.TrimSpace(in.Currency))
		}
		if in.ValidFrom != nil {
			entry.ValidFrom = validFrom
		}
		entry.ValidTo = in.ValidTo
		entry.UpdatedAt = now
	} else {
		entry = domain.PriceEntry{
			ID: d.newID(), TenantID: in.TenantID, PriceBookID: in.PriceBookID,
			VariantID: in.VariantID, Scope: in.Scope, ScopeID: in.ScopeID,
			AmountMinor: in.AmountMinor,
			Currency:    strings.ToUpper(strings.TrimSpace(in.Currency)),
			ValidFrom:   validFrom, ValidTo: in.ValidTo,
			CreatedAt: now, UpdatedAt: now,
		}
	}
	if err := entry.Validate(); err != nil {
		return domain.PriceEntry{}, err
	}
	if err := d.Prices.UpsertEntry(ctx, entry); err != nil {
		return domain.PriceEntry{}, err
	}
	d.emit(ctx, in.TenantID, entry.ID, domain.EventPriceChanged, map[string]any{
		"variantId": entry.VariantID.String(), "scope": string(entry.Scope),
		"amountMinor": entry.AmountMinor, "currency": entry.Currency,
	})
	return entry, nil
}

// GetPriceInput resolves the waterfall price for a variant.
type GetPriceInput struct {
	TenantID    uuid.UUID
	VariantID   uuid.UUID
	Currency    string
	RegionID    *uuid.UUID
	WarehouseID *uuid.UUID
	CustomerID  *uuid.UUID
	VIPID       *uuid.UUID
	CorporateID *uuid.UUID
	At          *time.Time
}

// GetPriceResult is the resolved unit price.
type GetPriceResult struct {
	Entry       domain.PriceEntry
	AmountMinor int64
	Currency    string
	Scope       domain.PriceScope
}

// GetPrice resolves base→regional→warehouse→customer(+vip/corporate) waterfall.
func (d *Deps) GetPrice(ctx context.Context, in GetPriceInput) (GetPriceResult, error) {
	if in.TenantID == uuid.Nil || in.VariantID == uuid.Nil {
		return GetPriceResult{}, fmt.Errorf("%w: tenant_id and variant_id required", domain.ErrInvalidArgument)
	}
	at := d.now()
	if in.At != nil {
		at = in.At.UTC()
	}
	entries, err := d.Prices.ListEntriesForVariant(ctx, in.TenantID, in.VariantID)
	if err != nil {
		return GetPriceResult{}, err
	}
	rctx := domain.ResolveContext{
		RegionID: in.RegionID, WarehouseID: in.WarehouseID, CustomerID: in.CustomerID,
		VIPID: in.VIPID, CorporateID: in.CorporateID,
		Currency: strings.ToUpper(strings.TrimSpace(in.Currency)), At: at,
	}
	entry, err := domain.ResolvePrice(entries, rctx)
	if err != nil {
		return GetPriceResult{}, err
	}
	return GetPriceResult{
		Entry: entry, AmountMinor: entry.AmountMinor,
		Currency: entry.Currency, Scope: entry.Scope,
	}, nil
}
