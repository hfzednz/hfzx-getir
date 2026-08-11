package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/nexora/catalog-service/internal/domain"
)

// WorkflowInput carries a workflow action request.
type WorkflowInput struct {
	TenantID  uuid.UUID
	ProductID uuid.UUID
	Action    domain.ApprovalActionType
	ActorID   uuid.UUID
	ActorRole string
	Comment   string
	ScheduledAt *domain.Product // optional schedule target time via product.ScheduledAt
}

// ApplyWorkflow applies submit/approve/reject/publish/hide/etc.
func (d *Deps) ApplyWorkflow(ctx context.Context, in WorkflowInput) (domain.Product, domain.ApprovalAction, error) {
	p, err := d.getProduct(ctx, in.TenantID, in.ProductID)
	if err != nil {
		return domain.Product{}, domain.ApprovalAction{}, err
	}
	from := p.Status
	target, err := domain.ApplyWorkflowAction(from, in.Action)
	if err != nil {
		return domain.Product{}, domain.ApprovalAction{}, err
	}
	if in.Action == domain.ApprovalActionSchedule && in.ScheduledAt != nil && in.ScheduledAt.ScheduledAt != nil {
		p.ScheduledAt = in.ScheduledAt.ScheduledAt
	}
	now := d.now()
	if err := p.TransitionTo(target, now); err != nil {
		return domain.Product{}, domain.ApprovalAction{}, err
	}
	if err := p.Validate(); err != nil {
		return domain.Product{}, domain.ApprovalAction{}, err
	}
	if err := d.Products.Update(ctx, p); err != nil {
		return domain.Product{}, domain.ApprovalAction{}, err
	}

	to := p.Status
	var versionID *uuid.UUID
	if in.Action == domain.ApprovalActionPublish {
		ver, err := d.createVersion(ctx, p, in.ActorID, "published")
		if err != nil {
			return domain.Product{}, domain.ApprovalAction{}, err
		}
		versionID = &ver.ID
	}
	action := domain.ApprovalAction{
		ID:         d.newID(),
		ProductID:  in.ProductID,
		VersionID:  versionID,
		TenantID:   in.TenantID,
		Action:     in.Action,
		FromStatus: &from,
		ToStatus:   &to,
		ActorID:    in.ActorID,
		ActorRole:  in.ActorRole,
		Comment:    in.Comment,
		Metadata:   map[string]any{},
		CreatedAt:  now,
	}
	if err := action.Validate(); err != nil {
		return domain.Product{}, domain.ApprovalAction{}, err
	}
	if err := d.Workflow.CreateAction(ctx, action); err != nil {
		return domain.Product{}, domain.ApprovalAction{}, err
	}

	switch in.Action {
	case domain.ApprovalActionPublish:
		d.publishEvent(ctx, domain.EventProductPublished, in.TenantID, in.ProductID, map[string]any{"versionId": versionID})
	case domain.ApprovalActionSubmit, domain.ApprovalActionApprove:
		d.publishEvent(ctx, domain.EventProductUpdated, in.TenantID, in.ProductID, map[string]any{"workflow": in.Action})
	case domain.ApprovalActionArchive:
		d.publishEvent(ctx, domain.EventProductArchived, in.TenantID, in.ProductID, nil)
	}
	d.indexProduct(ctx, in.TenantID, in.ProductID)
	return p, action, nil
}

// SubmitProduct submits draft for review.
func (d *Deps) SubmitProduct(ctx context.Context, tenantID, productID, actorID uuid.UUID, comment string) (domain.Product, error) {
	p, _, err := d.ApplyWorkflow(ctx, WorkflowInput{
		TenantID: tenantID, ProductID: productID,
		Action: domain.ApprovalActionSubmit, ActorID: actorID, ActorRole: "author", Comment: comment,
	})
	return p, err
}

// ApproveProduct approves pending product.
func (d *Deps) ApproveProduct(ctx context.Context, tenantID, productID, actorID uuid.UUID, comment string) (domain.Product, error) {
	p, _, err := d.ApplyWorkflow(ctx, WorkflowInput{
		TenantID: tenantID, ProductID: productID,
		Action: domain.ApprovalActionApprove, ActorID: actorID, ActorRole: "approver", Comment: comment,
	})
	return p, err
}

// RejectProduct rejects back to draft.
func (d *Deps) RejectProduct(ctx context.Context, tenantID, productID, actorID uuid.UUID, comment string) (domain.Product, error) {
	p, _, err := d.ApplyWorkflow(ctx, WorkflowInput{
		TenantID: tenantID, ProductID: productID,
		Action: domain.ApprovalActionReject, ActorID: actorID, ActorRole: "reviewer", Comment: comment,
	})
	return p, err
}

// PublishProduct publishes an approved product and creates a version snapshot.
func (d *Deps) PublishProduct(ctx context.Context, tenantID, productID, actorID uuid.UUID, comment string) (domain.Product, domain.ProductVersion, error) {
	p, action, err := d.ApplyWorkflow(ctx, WorkflowInput{
		TenantID: tenantID, ProductID: productID,
		Action: domain.ApprovalActionPublish, ActorID: actorID, ActorRole: "publisher", Comment: comment,
	})
	if err != nil {
		return domain.Product{}, domain.ProductVersion{}, err
	}
	var ver domain.ProductVersion
	if action.VersionID != nil {
		ver, _ = d.Versions.GetByID(ctx, tenantID, *action.VersionID)
	} else {
		ver, _ = d.Versions.GetLatest(ctx, tenantID, productID)
	}
	return p, ver, nil
}

// HideProduct unpublishes to hidden.
func (d *Deps) HideProduct(ctx context.Context, tenantID, productID, actorID uuid.UUID, comment string) (domain.Product, error) {
	p, _, err := d.ApplyWorkflow(ctx, WorkflowInput{
		TenantID: tenantID, ProductID: productID,
		Action: domain.ApprovalActionUnpublish, ActorID: actorID, ActorRole: "publisher", Comment: comment,
	})
	return p, err
}

// ListWorkflowActions returns approval audit trail.
func (d *Deps) ListWorkflowActions(ctx context.Context, tenantID, productID uuid.UUID, limit int) ([]domain.ApprovalAction, error) {
	if limit <= 0 {
		limit = 50
	}
	return d.Workflow.ListByProduct(ctx, tenantID, productID, limit)
}

func (d *Deps) createVersion(ctx context.Context, p domain.Product, createdBy uuid.UUID, summary string) (domain.ProductVersion, error) {
	num, err := d.Versions.NextVersionNumber(ctx, p.TenantID, p.ID)
	if err != nil {
		return domain.ProductVersion{}, err
	}
	snap := d.snapshotProduct(ctx, p)
	ver := domain.ProductVersion{
		ID:            d.newID(),
		ProductID:     p.ID,
		TenantID:      p.TenantID,
		VersionNumber: num,
		Snapshot:      snap,
		Status:        p.Status,
		ChangeSummary: summary,
		CreatedBy:     &createdBy,
		CreatedAt:     d.now(),
	}
	if err := ver.Validate(); err != nil {
		return domain.ProductVersion{}, err
	}
	if err := d.Versions.Create(ctx, ver); err != nil {
		return domain.ProductVersion{}, err
	}
	return ver, nil
}
