package domain

import (
	"fmt"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

// StationType is the workstation role.
type StationType string

const (
	StationTypePick     StationType = "pick"
	StationTypePack     StationType = "pack"
	StationTypeDispatch StationType = "dispatch"
	StationTypeQC       StationType = "qc"
)

func (t StationType) Valid() bool {
	switch t {
	case StationTypePick, StationTypePack, StationTypeDispatch, StationTypeQC:
		return true
	default:
		return false
	}
}

// StationStatus is occupancy / availability.
type StationStatus string

const (
	StationStatusAvailable   StationStatus = "available"
	StationStatusOccupied    StationStatus = "occupied"
	StationStatusOffline     StationStatus = "offline"
	StationStatusMaintenance StationStatus = "maintenance"
)

func (s StationStatus) Valid() bool {
	switch s {
	case StationStatusAvailable, StationStatusOccupied, StationStatusOffline, StationStatusMaintenance:
		return true
	default:
		return false
	}
}

// WarehouseType classifies local WH config.
type WarehouseType string

const (
	WarehouseTypeDarkStore        WarehouseType = "dark_store"
	WarehouseTypeRegional         WarehouseType = "regional"
	WarehouseTypeHub              WarehouseType = "hub"
	WarehouseTypeMicroFulfillment WarehouseType = "micro_fulfillment"
	WarehouseTypePartner          WarehouseType = "partner"
)

func (t WarehouseType) Valid() bool {
	switch t {
	case WarehouseTypeDarkStore, WarehouseTypeRegional, WarehouseTypeHub,
		WarehouseTypeMicroFulfillment, WarehouseTypePartner:
		return true
	default:
		return false
	}
}

// WarehouseStatus is local WH ops lifecycle.
type WarehouseStatus string

const (
	WarehouseStatusActive      WarehouseStatus = "active"
	WarehouseStatusInactive    WarehouseStatus = "inactive"
	WarehouseStatusMaintenance WarehouseStatus = "maintenance"
	WarehouseStatusClosed      WarehouseStatus = "closed"
)

func (s WarehouseStatus) Valid() bool {
	switch s {
	case WarehouseStatusActive, WarehouseStatusInactive,
		WarehouseStatusMaintenance, WarehouseStatusClosed:
		return true
	default:
		return false
	}
}

const maxStationCodeLen = 64

// Warehouse is local ops config; id may match inventory warehouse id.
type Warehouse struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	Code      string
	Name      string
	Type      WarehouseType
	Status    WarehouseStatus
	Timezone  string
	Metadata  map[string]any
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

// Station is a physical workstation inside a warehouse.
type Station struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	WarehouseID uuid.UUID
	Code        string
	Type        StationType
	Status      StationStatus
	Name        string
	Zone        string
	ZoneCode    string
	ClaimedBy   *uuid.UUID
	Metadata    map[string]any
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}

// Validate checks warehouse invariants.
func (w Warehouse) Validate() error {
	if w.ID == uuid.Nil {
		return fmt.Errorf("%w: warehouse id required", ErrInvalidArgument)
	}
	if w.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if w.Code == "" {
		return fmt.Errorf("%w: warehouse code required", ErrInvalidArgument)
	}
	if !w.Type.Valid() {
		return fmt.Errorf("%w: invalid warehouse type %q", ErrInvalidArgument, w.Type)
	}
	if !w.Status.Valid() {
		return fmt.Errorf("%w: invalid warehouse status %q", ErrInvalidArgument, w.Status)
	}
	if w.Timezone == "" {
		return fmt.Errorf("%w: timezone required", ErrInvalidArgument)
	}
	return nil
}

// IsOperable reports whether the warehouse can accept fulfillment work.
func (w Warehouse) IsOperable() bool {
	return w.Status == WarehouseStatusActive && w.DeletedAt == nil
}

// Validate checks station invariants.
func (s Station) Validate() error {
	if s.ID == uuid.Nil {
		return fmt.Errorf("%w: station id required", ErrInvalidArgument)
	}
	if s.WarehouseID == uuid.Nil {
		return fmt.Errorf("%w: warehouse_id required", ErrInvalidArgument)
	}
	if s.Code == "" {
		return fmt.Errorf("%w: station code required", ErrInvalidArgument)
	}
	if utf8.RuneCountInString(s.Code) > maxStationCodeLen {
		return fmt.Errorf("%w: station code too long", ErrInvalidArgument)
	}
	if !s.Type.Valid() {
		return fmt.Errorf("%w: invalid station type %q", ErrInvalidArgument, s.Type)
	}
	if !s.Status.Valid() {
		return fmt.Errorf("%w: invalid station status %q", ErrInvalidArgument, s.Status)
	}
	return nil
}

// Claim occupies an available station for an employee.
func (s *Station) Claim(employeeID uuid.UUID) error {
	if employeeID == uuid.Nil {
		return fmt.Errorf("%w: employee_id required", ErrInvalidArgument)
	}
	if s.Status != StationStatusAvailable {
		return fmt.Errorf("%w: status %s", ErrStationBusy, s.Status)
	}
	if s.DeletedAt != nil {
		return fmt.Errorf("%w: station deleted", ErrInvariant)
	}
	s.Status = StationStatusOccupied
	s.ClaimedBy = &employeeID
	s.UpdatedAt = time.Now().UTC()
	return nil
}

// Release frees an occupied station.
func (s *Station) Release() error {
	if s.Status != StationStatusOccupied {
		return fmt.Errorf("%w: station not occupied", ErrInvalidTransition)
	}
	s.Status = StationStatusAvailable
	s.ClaimedBy = nil
	s.UpdatedAt = time.Now().UTC()
	return nil
}

// SetOffline marks the station offline.
func (s *Station) SetOffline() {
	s.Status = StationStatusOffline
	s.ClaimedBy = nil
	s.UpdatedAt = time.Now().UTC()
}

// SetMaintenance marks the station under maintenance.
func (s *Station) SetMaintenance() {
	s.Status = StationStatusMaintenance
	s.ClaimedBy = nil
	s.UpdatedAt = time.Now().UTC()
}
