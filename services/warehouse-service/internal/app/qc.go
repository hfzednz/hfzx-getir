package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/nexora/warehouse-service/internal/domain"
)

// CreateQCCmd starts a QC inspection.
type CreateQCCmd struct {
	TenantID      uuid.UUID
	WarehouseID   uuid.UUID
	FulfillmentID uuid.UUID
	InspectorID   *uuid.UUID
}

// CreateQCInspection creates a pending QC inspection.
func (d *Deps) CreateQCInspection(ctx context.Context, in CreateQCCmd) (domain.QCInspection, error) {
	if in.TenantID == uuid.Nil || in.FulfillmentID == uuid.Nil {
		return domain.QCInspection{}, domain.ErrInvalidArgument
	}
	if _, err := d.Fulfillments.GetByID(ctx, in.TenantID, in.FulfillmentID); err != nil {
		return domain.QCInspection{}, err
	}
	now := d.now()
	insp := domain.QCInspection{
		ID: d.newID(), TenantID: in.TenantID, WarehouseID: in.WarehouseID,
		FulfillmentID: in.FulfillmentID, InspectorID: in.InspectorID,
		Result: domain.QCResultPending, CreatedAt: now, UpdatedAt: now,
	}
	if err := d.QC.Create(ctx, insp); err != nil {
		return domain.QCInspection{}, err
	}
	return insp, nil
}

// QCPassCmd marks inspection as passed.
type QCPassCmd struct {
	TenantID     uuid.UUID
	InspectionID uuid.UUID
	Notes        string
	InspectorID  *uuid.UUID
}

// QCPass completes inspection with pass.
func (d *Deps) QCPass(ctx context.Context, in QCPassCmd) (domain.QCInspection, error) {
	insp, err := d.QC.GetByID(ctx, in.TenantID, in.InspectionID)
	if err != nil {
		return domain.QCInspection{}, err
	}
	if insp.Result != domain.QCResultPending {
		return domain.QCInspection{}, domain.ErrInvalidTransition
	}
	now := d.now()
	insp.Result = domain.QCResultPass
	insp.Notes = in.Notes
	insp.InspectorID = in.InspectorID
	insp.CompletedAt = &now
	insp.UpdatedAt = now
	if err := d.QC.Update(ctx, insp); err != nil {
		return domain.QCInspection{}, err
	}
	d.publishEvent(ctx, domain.EventQCPassed, insp.TenantID, insp.WarehouseID, insp.FulfillmentID, map[string]any{
		"inspectionId": insp.ID,
	})
	return insp, nil
}

// QCFailCmd marks inspection as failed.
type QCFailCmd struct {
	TenantID     uuid.UUID
	InspectionID uuid.UUID
	Notes        string
	DefectCodes  []string
	InspectorID  *uuid.UUID
}

// QCFail completes inspection with fail.
func (d *Deps) QCFail(ctx context.Context, in QCFailCmd) (domain.QCInspection, error) {
	insp, err := d.QC.GetByID(ctx, in.TenantID, in.InspectionID)
	if err != nil {
		return domain.QCInspection{}, err
	}
	if insp.Result != domain.QCResultPending {
		return domain.QCInspection{}, domain.ErrInvalidTransition
	}
	now := d.now()
	insp.Result = domain.QCResultFail
	insp.Notes = in.Notes
	insp.DefectCodes = in.DefectCodes
	insp.InspectorID = in.InspectorID
	insp.CompletedAt = &now
	insp.UpdatedAt = now
	if err := d.QC.Update(ctx, insp); err != nil {
		return domain.QCInspection{}, err
	}
	d.publishEvent(ctx, domain.EventQCFailed, insp.TenantID, insp.WarehouseID, insp.FulfillmentID, map[string]any{
		"inspectionId": insp.ID, "defectCodes": in.DefectCodes,
	})
	return insp, nil
}
