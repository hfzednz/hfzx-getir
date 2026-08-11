package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/nexora/warehouse-service/internal/domain"
)

// RegisterEquipmentCmd registers a device.
type RegisterEquipmentCmd struct {
	TenantID    uuid.UUID
	WarehouseID uuid.UUID
	Code        string
	Kind        string
	Metadata    map[string]any
}

// RegisterEquipment adds equipment to the registry.
func (d *Deps) RegisterEquipment(ctx context.Context, in RegisterEquipmentCmd) (domain.Equipment, error) {
	if in.TenantID == uuid.Nil || in.WarehouseID == uuid.Nil || in.Code == "" {
		return domain.Equipment{}, domain.ErrInvalidArgument
	}
	now := d.now()
	e := domain.Equipment{
		ID: d.newID(), TenantID: in.TenantID, WarehouseID: in.WarehouseID,
		Code: in.Code, Kind: in.Kind, Status: domain.EquipmentStatusOffline,
		Metadata: in.Metadata, CreatedAt: now, UpdatedAt: now,
	}
	if e.Metadata == nil {
		e.Metadata = map[string]any{}
	}
	if err := d.Equipment.Create(ctx, e); err != nil {
		return domain.Equipment{}, err
	}
	return e, nil
}

// EquipmentHeartbeatCmd records a device heartbeat.
type EquipmentHeartbeatCmd struct {
	TenantID    uuid.UUID
	EquipmentID uuid.UUID
	Status      domain.EquipmentStatus
}

// EquipmentHeartbeat updates last heartbeat and status.
func (d *Deps) EquipmentHeartbeat(ctx context.Context, in EquipmentHeartbeatCmd) (domain.Equipment, error) {
	e, err := d.Equipment.GetByID(ctx, in.TenantID, in.EquipmentID)
	if err != nil {
		return domain.Equipment{}, err
	}
	now := d.now()
	e.LastHeartbeat = &now
	if in.Status != "" {
		e.Status = in.Status
	} else {
		e.Status = domain.EquipmentStatusOnline
	}
	e.UpdatedAt = now
	if err := d.Equipment.Update(ctx, e); err != nil {
		return domain.Equipment{}, err
	}
	return e, nil
}

// ListEquipment lists devices for a warehouse.
func (d *Deps) ListEquipment(ctx context.Context, tenantID, warehouseID uuid.UUID) ([]domain.Equipment, error) {
	return d.Equipment.ListByWarehouse(ctx, tenantID, warehouseID)
}

// CreateStationCmd registers a station.
type CreateStationCmd struct {
	TenantID    uuid.UUID
	WarehouseID uuid.UUID
	Code        string
	Name        string
	Type        domain.StationType
	Zone        string
}

// CreateStation registers a workstation.
func (d *Deps) CreateStation(ctx context.Context, in CreateStationCmd) (domain.Station, error) {
	if in.TenantID == uuid.Nil || in.WarehouseID == uuid.Nil || in.Code == "" {
		return domain.Station{}, domain.ErrInvalidArgument
	}
	stType := in.Type
	if stType == "" {
		stType = domain.StationTypePack
	}
	now := d.now()
	s := domain.Station{
		ID: d.newID(), TenantID: in.TenantID, WarehouseID: in.WarehouseID,
		Code: in.Code, Name: in.Name, Type: stType, Zone: in.Zone,
		Status: domain.StationStatusAvailable, CreatedAt: now, UpdatedAt: now,
	}
	if err := d.Stations.Create(ctx, s); err != nil {
		return domain.Station{}, err
	}
	return s, nil
}

// ListStations lists stations for a warehouse.
func (d *Deps) ListStations(ctx context.Context, tenantID, warehouseID uuid.UUID) ([]domain.Station, error) {
	return d.Stations.ListByWarehouse(ctx, tenantID, warehouseID)
}
