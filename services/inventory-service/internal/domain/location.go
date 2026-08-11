package domain

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

// LocationKind is a node in the warehouse location hierarchy.
type LocationKind string

const (
	LocationKindBuilding  LocationKind = "building"
	LocationKindFloor     LocationKind = "floor"
	LocationKindZone      LocationKind = "zone"
	LocationKindAisle     LocationKind = "aisle"
	LocationKindRack      LocationKind = "rack"
	LocationKindShelf     LocationKind = "shelf"
	LocationKindBin       LocationKind = "bin"
	LocationKindContainer LocationKind = "container"
)

func (k LocationKind) Valid() bool {
	switch k {
	case LocationKindBuilding, LocationKindFloor, LocationKindZone,
		LocationKindAisle, LocationKindRack, LocationKindShelf,
		LocationKindBin, LocationKindContainer:
		return true
	default:
		return false
	}
}

// ZoneType classifies climate/security for zone nodes.
type ZoneType string

const (
	ZoneTypeAmbient ZoneType = "ambient"
	ZoneTypeCold    ZoneType = "cold"
	ZoneTypeFrozen  ZoneType = "frozen"
	ZoneTypeSecure  ZoneType = "secure"
)

func (z ZoneType) Valid() bool {
	switch z {
	case ZoneTypeAmbient, ZoneTypeCold, ZoneTypeFrozen, ZoneTypeSecure:
		return true
	default:
		return false
	}
}

const maxLocationCodeLen = 64

// Location is a node in Warehouse → Building → … → Bin/Container.
type Location struct {
	ID          uuid.UUID
	WarehouseID uuid.UUID
	ParentID    *uuid.UUID
	Kind        LocationKind
	ZoneType    *ZoneType
	Code        string
	Path        string
	Depth       int
	Name        string
	IsPickable  bool
	IsActive    bool
	Metadata    map[string]any
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}

// Validate checks structural invariants.
func (l Location) Validate() error {
	if l.ID == uuid.Nil {
		return fmt.Errorf("%w: location id required", ErrInvalidArgument)
	}
	if l.WarehouseID == uuid.Nil {
		return fmt.Errorf("%w: warehouse_id required", ErrInvalidArgument)
	}
	if !l.Kind.Valid() {
		return fmt.Errorf("%w: invalid location kind %q", ErrInvalidArgument, l.Kind)
	}
	if l.Code == "" {
		return fmt.Errorf("%w: location code required", ErrInvalidArgument)
	}
	if utf8.RuneCountInString(l.Code) > maxLocationCodeLen {
		return fmt.Errorf("%w: location code too long", ErrInvalidArgument)
	}
	if l.ParentID != nil && *l.ParentID == l.ID {
		return fmt.Errorf("%w: location cannot be its own parent", ErrInvariant)
	}
	if l.Depth < 0 {
		return fmt.Errorf("%w: depth must be >= 0", ErrInvalidArgument)
	}
	if l.Kind == LocationKindZone {
		if l.ZoneType == nil || !l.ZoneType.Valid() {
			return fmt.Errorf("%w: zone locations require zone_type", ErrInvalidArgument)
		}
	} else if l.ZoneType != nil {
		return fmt.Errorf("%w: zone_type only allowed when kind=zone", ErrInvalidArgument)
	}
	return nil
}

// BuildPath constructs a materialized path from parent path and self id.
func BuildPath(parentPath string, id uuid.UUID) string {
	if parentPath == "" || parentPath == "/" {
		return "/" + id.String()
	}
	return strings.TrimRight(parentPath, "/") + "/" + id.String()
}
