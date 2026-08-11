package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// EquipmentKind classifies warehouse devices.
type EquipmentKind string

const (
	EquipmentKindScanner  EquipmentKind = "scanner"
	EquipmentKindPrinter  EquipmentKind = "printer"
	EquipmentKindForklift EquipmentKind = "forklift"
	EquipmentKindRobot    EquipmentKind = "robot"
	EquipmentKindIoT      EquipmentKind = "iot"
	EquipmentKindScale    EquipmentKind = "scale"
	EquipmentKindConveyor EquipmentKind = "conveyor"
	EquipmentKindOther    EquipmentKind = "other"

	EquipmentScanner  = EquipmentKindScanner
	EquipmentPrinter  = EquipmentKindPrinter
	EquipmentForklift = EquipmentKindForklift
	EquipmentRobot    = EquipmentKindRobot
	EquipmentIoT      = EquipmentKindIoT
	EquipmentScale    = EquipmentKindScale
	EquipmentConveyor = EquipmentKindConveyor
	EquipmentOther    = EquipmentKindOther
)

func (k EquipmentKind) Valid() bool {
	switch k {
	case EquipmentKindScanner, EquipmentKindPrinter, EquipmentKindForklift, EquipmentKindRobot,
		EquipmentKindIoT, EquipmentKindScale, EquipmentKindConveyor, EquipmentKindOther:
		return true
	default:
		return false
	}
}

// EquipmentStatus is device connectivity / health.
type EquipmentStatus string

const (
	EquipmentStatusOnline      EquipmentStatus = "online"
	EquipmentStatusOffline     EquipmentStatus = "offline"
	EquipmentStatusDegraded    EquipmentStatus = "degraded"
	EquipmentStatusMaintenance EquipmentStatus = "maintenance"
	EquipmentStatusRetired     EquipmentStatus = "retired"
)

func (s EquipmentStatus) Valid() bool {
	switch s {
	case EquipmentStatusOnline, EquipmentStatusOffline, EquipmentStatusDegraded,
		EquipmentStatusMaintenance, EquipmentStatusRetired:
		return true
	default:
		return false
	}
}

// SensorMetric is a reading type.
type SensorMetric string

const (
	SensorTemperatureC SensorMetric = "temperature_c"
	SensorHumidityPct  SensorMetric = "humidity_pct"
	SensorCO2PPM       SensorMetric = "co2_ppm"
	SensorDoorOpen     SensorMetric = "door_open"
	SensorVibration    SensorMetric = "vibration"
	SensorOther        SensorMetric = "other"
)

func (m SensorMetric) Valid() bool {
	switch m {
	case SensorTemperatureC, SensorHumidityPct, SensorCO2PPM,
		SensorDoorOpen, SensorVibration, SensorOther:
		return true
	default:
		return false
	}
}

// Equipment is a registered warehouse device.
// Kind is stored as string for flexible edge device types.
type Equipment struct {
	ID            uuid.UUID
	TenantID      uuid.UUID
	WarehouseID   uuid.UUID
	StationID     *uuid.UUID
	Code          string
	Kind          string
	Status        EquipmentStatus
	Name          string
	SerialNumber  string
	Firmware      string
	LastHeartbeat *time.Time
	Metadata      map[string]any
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     *time.Time
}

// EquipmentHeartbeat is a device heartbeat sample.
type EquipmentHeartbeat struct {
	ID          uuid.UUID
	EquipmentID uuid.UUID
	Status      EquipmentStatus
	BatteryPct  *int
	SignalRSSI  *int
	Payload     map[string]any
	RecordedAt  time.Time
}

// SensorReading is an environmental / IoT sample.
type SensorReading struct {
	ID          uuid.UUID
	WarehouseID uuid.UUID
	EquipmentID *uuid.UUID
	ZoneCode    string
	Metric      SensorMetric
	ValueNum    *float64
	ValueText   string
	Unit        string
	RecordedAt  time.Time
	Metadata    map[string]any
	CreatedAt   time.Time
}

// Validate checks equipment invariants.
func (e Equipment) Validate() error {
	if e.ID == uuid.Nil {
		return fmt.Errorf("%w: equipment id required", ErrInvalidArgument)
	}
	if e.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if e.WarehouseID == uuid.Nil {
		return fmt.Errorf("%w: warehouse_id required", ErrInvalidArgument)
	}
	if e.Code == "" {
		return fmt.Errorf("%w: equipment code required", ErrInvalidArgument)
	}
	if e.Status != "" && !e.Status.Valid() {
		return fmt.Errorf("%w: invalid equipment status %q", ErrInvalidArgument, e.Status)
	}
	return nil
}

// RecordHeartbeat updates last heartbeat and status from an edge ping.
func (e *Equipment) RecordHeartbeat(status EquipmentStatus, batteryPct, signalRSSI *int) (EquipmentHeartbeat, error) {
	if e.DeletedAt != nil {
		return EquipmentHeartbeat{}, fmt.Errorf("%w: equipment deleted", ErrInvariant)
	}
	if e.Status == EquipmentStatusRetired {
		return EquipmentHeartbeat{}, fmt.Errorf("%w: equipment retired", ErrAlreadyTerminal)
	}
	if status != "" && !status.Valid() {
		return EquipmentHeartbeat{}, fmt.Errorf("%w: invalid heartbeat status %q", ErrInvalidArgument, status)
	}
	if batteryPct != nil && (*batteryPct < 0 || *batteryPct > 100) {
		return EquipmentHeartbeat{}, fmt.Errorf("%w: battery_pct out of range", ErrInvalidArgument)
	}
	now := time.Now().UTC()
	if status == "" {
		status = EquipmentStatusOnline
	}
	e.Status = status
	e.LastHeartbeat = &now
	e.UpdatedAt = now
	return EquipmentHeartbeat{
		ID:          uuid.New(),
		EquipmentID: e.ID,
		Status:      status,
		BatteryPct:  batteryPct,
		SignalRSSI:  signalRSSI,
		Payload:     map[string]any{},
		RecordedAt:  now,
	}, nil
}

// IsOnline reports whether the device is currently online.
func (e Equipment) IsOnline() bool {
	return e.Status == EquipmentStatusOnline && e.DeletedAt == nil
}

// Validate checks sensor reading invariants.
func (r SensorReading) Validate() error {
	if r.ID == uuid.Nil {
		return fmt.Errorf("%w: sensor reading id required", ErrInvalidArgument)
	}
	if r.WarehouseID == uuid.Nil {
		return fmt.Errorf("%w: warehouse_id required", ErrInvalidArgument)
	}
	if !r.Metric.Valid() {
		return fmt.Errorf("%w: invalid sensor metric %q", ErrInvalidArgument, r.Metric)
	}
	if r.ValueNum == nil && r.ValueText == "" {
		return fmt.Errorf("%w: value_num or value_text required", ErrInvalidArgument)
	}
	return nil
}
