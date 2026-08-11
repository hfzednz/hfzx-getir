package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/inventory-service/internal/domain"
)

// CreateTransferLine is a transfer line request.
type CreateTransferLine struct {
	VariantID uuid.UUID
	SKUCode   string
	LotID     *uuid.UUID
	Qty       int64
}

// CreateTransferCmd creates a draft transfer.
type CreateTransferCmd struct {
	TenantID        uuid.UUID
	Code            string
	FromWarehouseID uuid.UUID
	ToWarehouseID   uuid.UUID
	FromLocationID  *uuid.UUID
	ToLocationID    *uuid.UUID
	Reason          string
	RequestedBy     *uuid.UUID
	Lines           []CreateTransferLine
}

// CreateTransfer creates a draft transfer.
func (d *Deps) CreateTransfer(ctx context.Context, in CreateTransferCmd) (domain.Transfer, error) {
	if in.TenantID == uuid.Nil || len(in.Lines) == 0 {
		return domain.Transfer{}, domain.ErrInvalidArgument
	}
	now := d.now()
	t := domain.Transfer{
		ID: d.newID(), TenantID: in.TenantID, Code: in.Code,
		FromWarehouseID: in.FromWarehouseID, ToWarehouseID: in.ToWarehouseID,
		FromLocationID: in.FromLocationID, ToLocationID: in.ToLocationID,
		Status: domain.TransferStatusDraft, RequestedBy: in.RequestedBy,
		Reason: in.Reason, Metadata: map[string]any{}, CreatedAt: now, UpdatedAt: now,
	}
	for _, line := range in.Lines {
		t.Lines = append(t.Lines, domain.TransferLine{
			ID: d.newID(), TransferID: t.ID, VariantID: line.VariantID,
			SKUCode: line.SKUCode, LotID: line.LotID, QtyRequested: line.Qty,
			Metadata: map[string]any{}, CreatedAt: now,
		})
	}
	if err := t.Validate(); err != nil {
		return domain.Transfer{}, err
	}
	if err := d.Transfers.Create(ctx, t); err != nil {
		return domain.Transfer{}, err
	}
	return t, nil
}

// ApproveTransferCmd approves a transfer.
type ApproveTransferCmd struct {
	TenantID   uuid.UUID
	TransferID uuid.UUID
	ActorID    *uuid.UUID
}

// ApproveTransfer moves draft/pending → approved.
func (d *Deps) ApproveTransfer(ctx context.Context, in ApproveTransferCmd) (domain.Transfer, error) {
	t, err := d.Transfers.GetByID(ctx, in.TenantID, in.TransferID)
	if err != nil {
		return domain.Transfer{}, err
	}
	if t.Status == domain.TransferStatusDraft {
		if err := t.TransitionTo(domain.TransferStatusPendingApproval, in.ActorID); err != nil {
			return domain.Transfer{}, err
		}
	}
	if err := t.TransitionTo(domain.TransferStatusApproved, in.ActorID); err != nil {
		return domain.Transfer{}, err
	}
	if err := d.Transfers.Update(ctx, t); err != nil {
		return domain.Transfer{}, err
	}
	return t, nil
}

// CompleteTransferCmd ships and receives transfer lines (moves stock).
type CompleteTransferCmd struct {
	TenantID       uuid.UUID
	TransferID     uuid.UUID
	IdempotencyKey string
	ActorID        *uuid.UUID
}

// CompleteTransfer deducts from source and adds to destination.
func (d *Deps) CompleteTransfer(ctx context.Context, in CompleteTransferCmd) (domain.Transfer, error) {
	if in.IdempotencyKey == "" {
		return domain.Transfer{}, fmt.Errorf("%w: idempotency_key required", domain.ErrInvalidArgument)
	}
	if v, ok := d.idemGet(ctx, "xfer:"+in.IdempotencyKey); ok {
		if t, ok := v.(domain.Transfer); ok {
			return t, nil
		}
	}
	t, err := d.Transfers.GetByID(ctx, in.TenantID, in.TransferID)
	if err != nil {
		return domain.Transfer{}, err
	}
	if t.Status == domain.TransferStatusCompleted {
		d.idemPut(ctx, "xfer:"+in.IdempotencyKey, t)
		return t, nil
	}
	if t.Status == domain.TransferStatusDraft || t.Status == domain.TransferStatusPendingApproval {
		if _, err := d.ApproveTransfer(ctx, ApproveTransferCmd{
			TenantID: in.TenantID, TransferID: in.TransferID, ActorID: in.ActorID,
		}); err != nil {
			return domain.Transfer{}, err
		}
		t, err = d.Transfers.GetByID(ctx, in.TenantID, in.TransferID)
		if err != nil {
			return domain.Transfer{}, err
		}
	}
	if t.Status == domain.TransferStatusApproved {
		if err := t.TransitionTo(domain.TransferStatusInTransit, in.ActorID); err != nil {
			return domain.Transfer{}, err
		}
		if err := d.Transfers.Update(ctx, t); err != nil {
			return domain.Transfer{}, err
		}
	}
	if t.Status != domain.TransferStatusInTransit {
		return domain.Transfer{}, fmt.Errorf("%w: cannot complete from %s", domain.ErrInvalidTransition, t.Status)
	}

	keys := make([]string, 0, len(t.Lines)*2)
	for _, line := range t.Lines {
		keys = append(keys, stockKey(t.FromWarehouseID, line.VariantID, t.FromLocationID))
		keys = append(keys, stockKey(t.ToWarehouseID, line.VariantID, t.ToLocationID))
	}
	sortStrings(keys)
	uniq := keys[:0]
	seen := map[string]struct{}{}
	for _, k := range keys {
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		uniq = append(uniq, k)
	}

	err = d.withMultiLock(ctx, uniq, func() error {
		fresh, err := d.Transfers.GetByID(ctx, in.TenantID, in.TransferID)
		if err != nil {
			return err
		}
		if fresh.Status == domain.TransferStatusCompleted {
			t = fresh
			return nil
		}
		for i := range fresh.Lines {
			line := &fresh.Lines[i]
			src, err := d.EnsureBalance(ctx, EnsureBalanceInput{
				TenantID: in.TenantID, WarehouseID: fresh.FromWarehouseID,
				VariantID: line.VariantID, SKUCode: line.SKUCode, LocationID: fresh.FromLocationID,
			})
			if err != nil {
				return err
			}
			qty := line.QtyRequested
			if qty > src.Available() {
				return fmt.Errorf("%w: transfer qty %d available %d", domain.ErrInsufficientStock, qty, src.Available())
			}
			if err := src.AdjustOnHand(-qty); err != nil {
				return err
			}
			if err := d.Balances.Update(ctx, src); err != nil {
				return err
			}
			dst, err := d.EnsureBalance(ctx, EnsureBalanceInput{
				TenantID: in.TenantID, WarehouseID: fresh.ToWarehouseID,
				VariantID: line.VariantID, SKUCode: line.SKUCode, LocationID: fresh.ToLocationID,
			})
			if err != nil {
				return err
			}
			if err := dst.AdjustOnHand(qty); err != nil {
				return err
			}
			if err := d.Balances.Update(ctx, dst); err != nil {
				return err
			}
			line.QtyShipped = qty
			line.QtyReceived = qty
			mout, err := domain.NewMovementFromBalance(
				d.newID(), in.TenantID, src, domain.MovementTypeTransferOut, -qty,
				in.IdempotencyKey+":out:"+line.ID.String(), in.ActorID, "transfer",
			)
			if err != nil {
				return err
			}
			tid := fresh.ID
			mout.TransferID = &tid
			if err := d.Movements.Create(ctx, mout); err != nil {
				return err
			}
			min, err := domain.NewMovementFromBalance(
				d.newID(), in.TenantID, dst, domain.MovementTypeTransferIn, qty,
				in.IdempotencyKey+":in:"+line.ID.String(), in.ActorID, "transfer",
			)
			if err != nil {
				return err
			}
			min.TransferID = &tid
			if err := d.Movements.Create(ctx, min); err != nil {
				return err
			}
			d.indexBalance(ctx, src)
			d.indexBalance(ctx, dst)
		}
		if err := fresh.TransitionTo(domain.TransferStatusCompleted, in.ActorID); err != nil {
			return err
		}
		if err := d.Transfers.Update(ctx, fresh); err != nil {
			return err
		}
		t = fresh
		d.publishEvent(ctx, domain.EventStockTransferred, in.TenantID, fresh.FromWarehouseID, fresh.Lines[0].VariantID, map[string]any{
			"transferId": fresh.ID, "toWarehouseId": fresh.ToWarehouseID,
		})
		return nil
	})
	if err != nil {
		return domain.Transfer{}, err
	}
	d.idemPut(ctx, "xfer:"+in.IdempotencyKey, t)
	return t, nil
}

// GetTransfer returns a transfer by id.
func (d *Deps) GetTransfer(ctx context.Context, tenantID, id uuid.UUID) (domain.Transfer, error) {
	return d.Transfers.GetByID(ctx, tenantID, id)
}
