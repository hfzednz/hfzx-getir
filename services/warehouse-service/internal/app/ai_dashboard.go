package app

import (
	"context"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/nexora/warehouse-service/internal/app/ports"
	"github.com/nexora/warehouse-service/internal/domain"
)

// OptimizeRouteCmd requests AI pick-route ordering.
type OptimizeRouteCmd struct {
	TenantID    uuid.UUID
	WarehouseID uuid.UUID
	Lines       []ports.RouteLine
}

// OptimizeRoute sorts pick lines via the RouteOptimizer AI port (stub).
func (d *Deps) OptimizeRoute(ctx context.Context, in OptimizeRouteCmd) ([]ports.RouteLine, error) {
	if in.WarehouseID == uuid.Nil || len(in.Lines) == 0 {
		return nil, domain.ErrInvalidArgument
	}
	if d.RouteAI == nil {
		out := append([]ports.RouteLine(nil), in.Lines...)
		sort.SliceStable(out, func(i, j int) bool {
			return strings.Compare(out[i].LocationCode, out[j].LocationCode) < 0
		})
		for i := range out {
			out[i].Sequence = i + 1
		}
		return out, nil
	}
	return d.RouteAI.OptimizePickRoute(ctx, in.WarehouseID, in.Lines)
}

// DashboardAggregates summarizes warehouse operational counts.
type DashboardAggregates struct {
	WarehouseID         uuid.UUID                    `json:"warehouseId"`
	FulfillmentsByStatus map[string]int              `json:"fulfillmentsByStatus"`
	TasksByStatus       map[domain.TaskStatus]int    `json:"tasksByStatus"`
	DispatchQueued      int                          `json:"dispatchQueued"`
	EquipmentOnline     int                          `json:"equipmentOnline"`
	ActiveShifts        int                          `json:"activeShifts"`
}

// Dashboard returns aggregate counts by status for admin views.
func (d *Deps) Dashboard(ctx context.Context, tenantID, warehouseID uuid.UUID) (DashboardAggregates, error) {
	agg := DashboardAggregates{
		WarehouseID:          warehouseID,
		FulfillmentsByStatus: map[string]int{},
		TasksByStatus:        map[domain.TaskStatus]int{},
	}

	list, _, err := d.Fulfillments.List(ctx, ports.FulfillmentFilter{
		TenantID: tenantID, WarehouseID: &warehouseID, Limit: 10000,
	})
	if err != nil {
		return DashboardAggregates{}, err
	}
	for _, fo := range list {
		agg.FulfillmentsByStatus[string(fo.Status)]++
	}

	if counts, err := d.Tasks.CountByStatus(ctx, tenantID, warehouseID); err == nil {
		agg.TasksByStatus = counts
	}

	units, n, _ := d.Dispatches.ListQueued(ctx, tenantID, warehouseID, 1, 0)
	_ = units
	agg.DispatchQueued = n

	equip, _ := d.Equipment.ListByWarehouse(ctx, tenantID, warehouseID)
	for _, e := range equip {
		if e.Status == domain.EquipmentStatusOnline {
			agg.EquipmentOnline++
		}
	}

	employees, _ := d.Workforce.ListEmployees(ctx, tenantID, warehouseID)
	for _, emp := range employees {
		if _, err := d.Workforce.GetActiveShift(ctx, tenantID, emp.ID); err == nil {
			agg.ActiveShifts++
		}
	}
	return agg, nil
}
