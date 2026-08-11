package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/nexora/catalog-service/internal/domain"
)

// ListProductVersions returns version history.
func (d *Deps) ListProductVersions(ctx context.Context, tenantID, productID uuid.UUID, limit, offset int) ([]domain.ProductVersion, error) {
	if limit <= 0 {
		limit = 20
	}
	return d.Versions.ListByProduct(ctx, tenantID, productID, limit, offset)
}

// GetProductVersion returns a specific version.
func (d *Deps) GetProductVersion(ctx context.Context, tenantID, versionID uuid.UUID) (domain.ProductVersion, error) {
	return d.Versions.GetByID(ctx, tenantID, versionID)
}

// DiffProductVersions compares two version snapshots.
func (d *Deps) DiffProductVersions(ctx context.Context, tenantID, versionA, versionB uuid.UUID) (map[string]any, error) {
	a, err := d.Versions.GetByID(ctx, tenantID, versionA)
	if err != nil {
		return nil, err
	}
	b, err := d.Versions.GetByID(ctx, tenantID, versionB)
	if err != nil {
		return nil, err
	}
	if a.ProductID != b.ProductID {
		return nil, domain.ErrInvalidArgument
	}
	return diffSnapshots(a.Snapshot, b.Snapshot), nil
}

// RollbackProductInput rolls back to a prior version snapshot (draft state).
type RollbackProductInput struct {
	TenantID  uuid.UUID
	ProductID uuid.UUID
	VersionID uuid.UUID
	ActorID   uuid.UUID
	Comment   string
}

// RollbackProduct restores product fields from a version snapshot into draft.
func (d *Deps) RollbackProduct(ctx context.Context, in RollbackProductInput) (domain.Product, domain.ApprovalAction, error) {
	ver, err := d.Versions.GetByID(ctx, in.TenantID, in.VersionID)
	if err != nil {
		return domain.Product{}, domain.ApprovalAction{}, err
	}
	if ver.ProductID != in.ProductID {
		return domain.Product{}, domain.ApprovalAction{}, domain.ErrInvalidArgument
	}
	p, err := d.getProduct(ctx, in.TenantID, in.ProductID)
	if err != nil {
		return domain.Product{}, domain.ApprovalAction{}, err
	}
	from := p.Status
	// Rollback always lands in draft for re-editing.
	if err := p.TransitionTo(domain.ProductStatusDraft, d.now()); err != nil {
		return domain.Product{}, domain.ApprovalAction{}, err
	}
	if prodRaw, ok := ver.Snapshot["product"].(domain.Product); ok {
		p.Slug = prodRaw.Slug
		p.SKUCode = prodRaw.SKUCode
		p.Metadata = prodRaw.Metadata
		p.BrandID = prodRaw.BrandID
	}
	p.UpdatedAt = d.now()
	if err := p.Validate(); err != nil {
		return domain.Product{}, domain.ApprovalAction{}, err
	}
	if err := d.Products.Update(ctx, p); err != nil {
		return domain.Product{}, domain.ApprovalAction{}, err
	}
	to := p.Status
	action := domain.ApprovalAction{
		ID:         d.newID(),
		ProductID:  in.ProductID,
		VersionID:  &ver.ID,
		TenantID:   in.TenantID,
		Action:     domain.ApprovalActionRollback,
		FromStatus: &from,
		ToStatus:   &to,
		ActorID:    in.ActorID,
		ActorRole:  "publisher",
		Comment:    in.Comment,
		Metadata:   map[string]any{"rolledBackToVersion": ver.VersionNumber},
		CreatedAt:  d.now(),
	}
	if err := action.Validate(); err != nil {
		return domain.Product{}, domain.ApprovalAction{}, err
	}
	if err := d.Workflow.CreateAction(ctx, action); err != nil {
		return domain.Product{}, domain.ApprovalAction{}, err
	}
	d.indexProduct(ctx, in.TenantID, in.ProductID)
	return p, action, nil
}
