package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ATPPolicy controls how Available-to-Promise is computed.
type ATPPolicy struct {
	// IncludeIncoming adds incoming within IncomingHorizon.
	IncludeIncoming bool
	// IncomingHorizon limits which incoming supply counts; zero = all incoming.
	IncomingHorizon time.Duration
	// IncludeInTransit adds in_transit supply (cross-dock / transfer-in).
	IncludeInTransit bool
}

// DefaultATPPolicy is available = on_hand - reserved - blocked (no supply add-ons).
func DefaultATPPolicy() ATPPolicy {
	return ATPPolicy{
		IncludeIncoming:  false,
		IncludeInTransit: false,
	}
}

// ATPQuery scopes an availability-to-promise request.
// VariantID / SKUCode are opaque catalog references.
type ATPQuery struct {
	TenantID    uuid.UUID
	VariantID   uuid.UUID
	SKUCode     string
	WarehouseID *uuid.UUID
	RegionID    *uuid.UUID
	AsOf        time.Time
	Policy      ATPPolicy
}

// Validate checks query invariants.
func (q ATPQuery) Validate() error {
	if q.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if q.VariantID == uuid.Nil && q.SKUCode == "" {
		return fmt.Errorf("%w: variant_id or sku_code required", ErrInvalidArgument)
	}
	return nil
}

// ATPResult is the computed availability for a warehouse/variant key.
type ATPResult struct {
	TenantID    uuid.UUID
	WarehouseID uuid.UUID
	VariantID   uuid.UUID
	SKUCode     string
	OnHand      int64
	Reserved    int64
	Blocked     int64
	Incoming    int64
	InTransit   int64
	Available   int64
	ATP         int64
	AsOf        time.Time
}

// ComputeATP calculates ATP from a balance using the given policy.
// Base available = on_hand - reserved - blocked; policy may add incoming / in_transit.
func ComputeATP(b StockBalance, policy ATPPolicy, asOf time.Time) ATPResult {
	if asOf.IsZero() {
		asOf = time.Now().UTC()
	}
	available := b.Available()
	atp := available
	if policy.IncludeIncoming {
		atp += b.Incoming
	}
	if policy.IncludeInTransit {
		atp += b.InTransit
	}
	return ATPResult{
		TenantID:    b.TenantID,
		WarehouseID: b.WarehouseID,
		VariantID:   b.VariantID,
		SKUCode:     b.SKUCode,
		OnHand:      b.OnHand,
		Reserved:    b.Reserved,
		Blocked:     b.Blocked,
		Incoming:    b.Incoming,
		InTransit:   b.InTransit,
		Available:   available,
		ATP:         atp,
		AsOf:        asOf,
	}
}

// AggregateATP sums ATP results (e.g. regional rollup).
func AggregateATP(parts []ATPResult) ATPResult {
	if len(parts) == 0 {
		return ATPResult{}
	}
	out := ATPResult{
		TenantID:  parts[0].TenantID,
		VariantID: parts[0].VariantID,
		SKUCode:   parts[0].SKUCode,
		AsOf:      parts[0].AsOf,
	}
	for _, p := range parts {
		out.OnHand += p.OnHand
		out.Reserved += p.Reserved
		out.Blocked += p.Blocked
		out.Incoming += p.Incoming
		out.InTransit += p.InTransit
		out.Available += p.Available
		out.ATP += p.ATP
		if p.AsOf.After(out.AsOf) {
			out.AsOf = p.AsOf
		}
	}
	return out
}

// CanFulfill reports whether ATP covers the requested quantity.
func (r ATPResult) CanFulfill(qty int64) bool {
	return qty > 0 && r.ATP >= qty
}
